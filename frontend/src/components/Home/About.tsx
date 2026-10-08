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
          Авторы публикуют свои работы (карточки, брелоки, картины, фенечки и другой
          мерч) по подписке; игроки забирают товары за билетики, а авторы получают долю
          от продаж. Платформа связывает сообщество, экономику и real-time рекорды в
          одном месте.
        </p>
        <p>
          Технически это микросервисный backend на Go (gRPC, NATS, PostgreSQL, Redis,
          ClickHouse) и React-фронтенд — от идеи до production-grade деплоя.
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
