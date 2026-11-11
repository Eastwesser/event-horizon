Меняем на разные лимиты по сервисам. Один newServiceBreaker(name) не даёт различий. Нужно два параметра: name + maxConcurrent. Патч:

go
// /home/denismatveev/event_horizon/services/gateway/internal/app/gateway.go

// newServiceBreaker builds a circuit breaker + bulkhead.
// maxConcurrent caps in-flight calls to this downstream. Read-heavy services
// (shop/inventory/leaderboard/profile) get a higher cap; write-heavy
// (billing/payment/authors) stay lower so a slow write path cannot starve
// the read path.
func newServiceBreaker(name string, maxConcurrent int) *circuit.Breaker {
	if maxConcurrent <= 0 {
		maxConcurrent = 32
	}
	return circuit.New(circuit.Settings{
		Name:          name,
		MaxRequests:   3,
		Timeout:       10 * time.Second,
		MaxConcurrent: maxConcurrent,
		ReadyToTrip: func(counts circuit.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})
}
Затем заменить все вызовы newServiceBreaker("...") на newServiceBreaker("...", N). По сервисам:

Сервис	MaxConcurrent	Почему
shop	256	read path (shop/items) под 140 browse-VU
inventory	256	read path (inventory/items) под 140 browse-VU
leaderboard	128	Redis Sorted Set, быстрый, но gRPC-хоп
profile	128	агрегатор, read path
game	128	submit, write path, но быстрый
billing	64	write path (SpendCurrency/AddCurrency), деньги
payment	64	merch gate, write path
auth	128	ValidateToken (пока без кэша — Шаг 4), login/register
authors	64	applications, write path
history	64	read path, но редко
notification	64	read path, редко
analytics	32	admin-only, редко
Почему 256, а не 1000: bulkhead защищает downstream. Если Shop реально не тянет >256 одновременных, 257-й получит 503 — быстрый fail, не зависание. 256 — эмпирически: 140 browse-VU × 4 ручки ≈ 560, но Shop получает только shop/items ≈ 140. 256 с запасом. Если P=200 даст 280 одновременных — 24 получат 503, это лучше, чем 30s зависание.

Почему billing 64: деньги, не хочется 256 параллельных SpendCurrency (deдлоки, contention на balance). 64 — норм для 20 purchase-VU.

---

// newServiceBreaker builds a circuit breaker + bulkhead.
// maxConcurrent caps in-flight calls to this downstream. Read-heavy services
// (shop/inventory/leaderboard/profile) get a higher cap; write-heavy
// (billing/payment/authors) stay lower so a slow write path cannot starve
// the read path.
func newServiceBreaker(name string, maxConcurrent int) *circuit.Breaker {
	if maxConcurrent <= 0 {
		maxConcurrent = 32
	}
	return circuit.New(circuit.Settings{
		Name:          name,
		MaxRequests:   3,
		Timeout:       10 * time.Second,
		MaxConcurrent: uint32(maxConcurrent),
		ReadyToTrip: func(counts circuit.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})


Почему uint32 in MaxConcurrent: uint32(maxConcurrent),
circuit.Settings.MaxConcurrent — uint32 (это видно из ошибки). 

Не спрашивай, почему не int — это внутри circuit-пакета, менять не надо. 
Просто кастуем.

Плюс, раз уж правим — проверь, что if maxConcurrent <= 0 { maxConcurrent = 32 } до circuit.New — стоит (ты его уже добавил, вижу в файле). 
Тогда негативные значения отсекаются, uint32(-1) не будет.