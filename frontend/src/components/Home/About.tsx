import { useNavigate } from 'react-router-dom';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Button } from '../ui/Button';
import { useUserRole } from '../../hooks/useUserRole';

export function About() {
  const navigate = useNavigate();
  const { isAdmin, loading: roleLoading } = useUserRole();

  return (
    <PageShell width="narrow">
      <PageHeader
        title="О проекте"
        onBack={() => navigate('/')}
        backLabel="Назад"
      />

      <article className="space-y-6 text-base leading-relaxed text-text-secondary">
        <p className="font-display text-2xl font-semibold leading-snug text-text-primary">
          Event Horizon — платформа мини-игр и витрина для авторов.
        </p>
        <p>
          Восемь игр на главной. За партии начисляются билетики и лампочки. Билетики
          тратятся в магазине: скины для игр и вещи авторов. Лампочки — на буст в
          партии. Результат без буста идёт в лидерборд.
        </p>
        <p>
          Магазин собирают авторы: карточки, брелоки, картины, фенечки и другой мерч
          по подписке Boosty. Скин появляется в игре сразу после покупки; физический
          заказ собирает автор.
        </p>
        <p>
          Чтобы продавать свои работы — Boosty и кабинет автора. Чтобы играть и
          покупать, достаточно аккаунта.
        </p>

        {!roleLoading && isAdmin && (
          <aside className="rounded-md border border-white/10 bg-nebula p-4 text-sm text-text-muted">
            <p className="mb-2 font-semibold text-indigo-soft">Только для админов</p>
            <p className="mb-2">
              Экономика (ориентир): Базовый ~200 ₽ и Расширенный на Boosty; билетики —
              валюта магазина; лампы — бусты. Выплаты авторам в ₽ (C4) отложены — сейчас
              витрина продаж в билетиках (<code className="text-text-secondary">/authors/me/sales</code>
              ). Всё, что не карточная «карточка» и не мелкий физ. тип, ложится в тип
              мерч.
            </p>
            <p>
              Стек: микросервисы на Go (gRPC, NATS, PostgreSQL, Redis, ClickHouse) +
              React. Публичный HTTP — <code className="text-text-secondary">/api/v1/*</code> на
              Gateway.
            </p>
          </aside>
        )}
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
