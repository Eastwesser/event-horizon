import { useNavigate } from 'react-router-dom';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Button } from '../ui/Button';

export function About() {
  const navigate = useNavigate();

  return (
    <PageShell width="narrow">
      <PageHeader
        title="О проекте"
        onBack={() => navigate('/')}
        backLabel="Назад"
      />

      <article className="space-y-6 text-base leading-relaxed text-text-secondary">
        <p className="font-display text-xl font-semibold text-text-primary">
          Event Horizon — игровая платформа и витрина для авторов.
        </p>
        <p>
          Игроки проходят мини-игры, зарабатывают билетики и лампочки, соревнуются в
          лидерборде и обменивают накопленное на товары в магазине — от внутриигровых
          скинов до физического мерча.
        </p>
        <p>
          Авторы публикуют товары (карточка / брелок / картина / фенечка / мерч и
          косметика) по подписке Boosty; игроки покупают за билетики. Всё, что не
          карточная «карточка» и не мелкий физ. тип — ложится в тип <strong>мерч</strong>.
        </p>
        <p>
          Экономика (ориентир): Базовый ~200 ₽ и Расширенный на Boosty; билетики —
          внутриигровая валюта магазина; лампы — бусты в играх. Выплаты авторам в ₽
          (C4) сознательно отложены — сейчас витрина продаж в билетиках
          (`/authors/me/sales`).
        </p>
        <p>
          Технически: микросервисы на Go (gRPC, NATS, PostgreSQL, Redis, ClickHouse) +
          React; публичный HTTP — <code className="text-text-primary">/api/v1/*</code> на
          Gateway.
        </p>
      </article>

      <div className="mt-10 flex flex-wrap gap-3">
        <Button onClick={() => navigate('/#games')}>К играм</Button>
        <Button variant="ghost" onClick={() => navigate('/shop')}>
          В магазин
        </Button>
        <Button
          variant="secondary"
          onClick={() =>
            window.open('https://boosty.to/eastwesser', '_blank', 'noopener,noreferrer')
          }
        >
          Поддержать на Boosty
        </Button>
      </div>
    </PageShell>
  );
}
