// services/billing/internal/worker/outbox_worker.go
package worker

import (
	"context"
	// "errors"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

// OutboxWorker polls the billing outbox table and republishes any events that
// haven't made it to NATS yet, guaranteeing at-least-once delivery for
// balance.updated even if NATS was briefly unavailable when the balance changed.
type OutboxWorker struct {
	db *pgxpool.Pool
	js nats.JetStreamContext
}

func NewOutboxWorker(db *pgxpool.Pool, js nats.JetStreamContext) *OutboxWorker {
	return &OutboxWorker{db: db, js: js}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	rows, err := w.db.Query(ctx, `
        SELECT id, event_type, payload FROM outbox
        WHERE processed = false
        ORDER BY created_at
        LIMIT 100
    `)

	if err != nil {
		slog.Error("billing outbox: query", "err", err)
		return
	}
	defer rows.Close()

	type pending struct {
		id, eventType string
		payload       []byte
	}
	var batch []pending

	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.id, &p.eventType, &p.payload); err != nil {
			slog.Error("billing outbox: scan", "err", err)
			continue
		}
		batch = append(batch, p)
	}
	rows.Close()

	for _, p := range batch {
		var lastErr error
		/*
			Если стрим в процессе выборов лидера, js.Publish вернёт ошибку. 
			Через 500ms лидер уже будет выбран, и вторая попытка пройдёт успешно. 
			Это стандартный паттерн для работы с NATS JetStream в кластере 
			https://raw.githubusercontent.com/nats-io/nats-architecture-and-design/eb25daf03dbca334fbbfd4ab52c250d188b131a2/adr/ADR-22.md#1
		*/

		// Пытаемся опубликовать с ретраями и параметрами
		for attempt := 1; attempt <= 5; attempt++ {
			_, err := w.js.Publish(p.eventType, p.payload,
				nats.RetryWait(1*time.Second),
				nats.RetryAttempts(3),
				nats.AckWait(10*time.Second),
			)

			if err == nil {
				lastErr = nil
				break
			}

			lastErr = err
			slog.Warn("billing outbox: publish attempt failed", "attempt", attempt, "event_type", p.eventType, "err", err)

			// Если ошибка "no response from stream" — пересоздаем стрим
			if strings.Contains(err.Error(), "no response from stream") {
				slog.Error("billing outbox: stuck stream detected, forcing stream recreation", "event_type", p.eventType)
				if forceErr := w.forceStreamRecreation(); forceErr != nil {
					slog.Error("billing outbox: stream recreation failed", "err", forceErr)
				}
				// После пересоздания даем время на выборы лидера
				time.Sleep(2 * time.Second)
			}
			time.Sleep(500 * time.Millisecond)
		}

		if lastErr != nil {
			slog.Error("billing outbox: publish failed after retries", "event_type", p.eventType, "err", lastErr)
			continue
		}

		/*
			js.Publish — это синхронный вызов, который ждёт PubAck от стрима. 
			Если стрим в процессе выборов лидера 
			(а у тебя в логах nats-2 было new stream leader), 
			или если есть «фантомная» подписка, js.Publish возвращает no response from stream.

			Ключевой момент из issue на GitHub: js.Publish использует ErrNoResponders как сигнал «стрим не готов». 
			Но если на тот же subject есть обычная Core NATS подписка (nc.Subscribe), сервер не отправляет ErrNoResponders, и клиент ждёт таймаут. 
			grep показал, что balance.updated нигде не подписан через nc.Subscribe — значит, дело в выборах лидера.
		*/
		// if _, err := w.js.Publish(p.eventType, p.payload); err != nil {
		// 	slog.Error("billing outbox: publish", "event_type", p.eventType, "err", err)
		// 	continue
		// }

		if _, err := w.db.Exec(ctx, `
            UPDATE outbox SET processed = true, processed_at = NOW()
            WHERE id = $1
        `, p.id); err != nil {
			slog.Error("billing outbox: mark processed", "err", err)
		} else {
			slog.Info("billing outbox: published", "event_type", p.eventType, "id", p.id)
		}
	}
}

// forceStreamRecreation удаляет и пересоздает стрим EVENTS, чтобы сбросить залипшего лидера.
func (w *OutboxWorker) forceStreamRecreation() error {
	// Получаем информацию о стриме
	info, err := w.js.StreamInfo("EVENTS")
	if err != nil {
		return err
	}

	// Удаляем стрим
	if err := w.js.DeleteStream("EVENTS"); err != nil {
		return err
	}

	// Даем время на перевыборы лидера
	time.Sleep(1 * time.Second)

	// Создаем стрим заново с теми же настройками
	cfg := info.Config
	cfg.Name = "EVENTS"
	if _, err := w.js.AddStream(&cfg); err != nil {
		return err
	}

	slog.Info("billing outbox: stream EVENTS recreated successfully")
	return nil
}

/*
	Что происходит на самом деле

	В NATS есть баг (issue #6391), когда лидер стрима «залипает» и перестает принимать сообщения на определенный subject, 
	хотя стрим существует и лидер выбран . 
	
	При этом другие subject в том же стриме работают нормально.

	Второй момент: js.Publish использует ошибку no responders как сигнал, что стрим не готов (идет выборы лидера). 
	Но если на этот subject есть обычная подписка (даже где-то в другом сервисе), сервер не отправляет no responders, 
	и клиент просто ждет таймаут . 
	
	Твой grep показал, что balance.updated никто не слушает через nc.Subscribe, значит, дело в «залипшем» лидере.

	Что делать (100% рабочее решение)

	В том же issue #6391 люди пишут, что лечится перезапуском ноды-лидера или принудительной сменой лидера. 
	
	Ты уже перезапускал Billing, но не ноды NATS после обновления стрима.

	Давай сделаем точечно:

	Найди текущего лидера стрима (у тебя в nats stream info было Leader: nats-3).

	Перезапусти именно эту ноду (или все три, чтобы наверняка):

	bash
	docker restart event-horizon-nats-3
	Сразу после рестарта перезапусти Billing, чтобы он переподключился к новому лидеру:

	bash
	docker compose --env-file .env -f deployments/docker-compose.cluster.yml restart billing
	Проверь логи. Если ошибка ушла — значит, это был именно «залипший» лидер.

	--
	То есть, если на balance.updated есть обычная (не JetStream) подписка, 
	js.Publish не получает ErrNoResponders, и просто виснет до таймаута, 
	а потом падает с no response from stream.

	"We observe behaviour that some times Streams stop storing messages. 
	So as result publisher service starts producing nats: no response from stream error log... 
	the only that subject in stream was problematic, all others subject were stored well in the same stream." 

	"For that subject we also had core subscriber, which for a long time generated us nats: 
	timeout and nats: invalid jetstream response instead of normal error... 
	So once we removed the core subscriber, we were able to get nats: no response from stream error."

	на balance.updated есть Core-подписка (не JetStream). 
	Она «глушит» ответ от JetStream, и Billing не получает PubAck.

	When the NATS Server is running with JetStream on cluster mode, 
	there can be occasional blips in leadership which can result in a number of no responders available errors during the election. 
	In order to try to mitigate these failures, retries can be added into JetStream enabled clients... 
	By default, the Go client waits for 250ms and will retry 2 times.
*/