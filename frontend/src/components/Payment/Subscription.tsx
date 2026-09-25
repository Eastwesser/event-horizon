// frontend/src/components/Payment/Subscription.tsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { paymentApi, type SubscriptionStatus } from '../../services/paymentApi';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Spinner } from '../ui/Spinner';

const PLANS = [
  {
    id: 'present' as const,
    name: 'Текущий',
    title: '🎁 Текущий план',
    description: 'Подписка на Boosty — доступ к мерчу и бонусам сообщества.',
    titleClass: 'text-horizon-gold',
    buttonVariant: 'primary' as const,
  },
  {
    id: 'future' as const,
    name: 'Будущий',
    title: '🚀 Будущий план',
    description: 'Расширенная подписка с дополнительными привилегиями (когда будет доступна).',
    titleClass: 'text-indigo-soft',
    buttonVariant: 'secondary' as const,
  },
];

const PLAN_LABELS: Record<string, string> = {
  present: 'Текущий',
  future: 'Будущий',
  none: 'Нет плана',
};

const STATUS_LABELS: Record<string, string> = {
  active: 'Активна',
  inactive: 'Не активна',
  cancelled: 'Отменена',
  expired: 'Истекла',
  pending: 'Ожидает оплаты',
  none: 'Нет подписки',
};

function statusLabel(raw?: string | null): string {
  if (!raw || raw === 'none') return 'Нет подписки';
  return STATUS_LABELS[raw.toLowerCase()] || raw;
}

function planLabel(raw?: string | null): string {
  if (!raw || raw === 'none') return 'Нет плана';
  return PLAN_LABELS[raw.toLowerCase()] || raw;
}

export function Subscription() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');

  const [status, setStatus] = useState<SubscriptionStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [checkoutLoading, setCheckoutLoading] = useState<string | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!token) {
      navigate('/login');
      return;
    }

    const load = async () => {
      try {
        const data = await paymentApi.getSubscription();
        setStatus(data);
      } catch (err: unknown) {
        const msg =
          (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
          'Не удалось загрузить статус подписки';
        setError(msg);
      } finally {
        setLoading(false);
      }
    };

    load();
  }, [token, navigate]);

  const handleCheckout = async (plan: 'present' | 'future') => {
    setCheckoutLoading(plan);
    setError('');
    try {
      const checkout = await paymentApi.createCheckout(plan);
      if (checkout.checkout_url) {
        window.open(checkout.checkout_url, '_blank', 'noopener,noreferrer');
      }
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        'Ошибка создания оплаты';
      setError(msg);
    } finally {
      setCheckoutLoading(null);
    }
  };

  const formatExpiry = (unix: number) => {
    if (!unix) return '—';
    return new Date(unix * 1000).toLocaleDateString('ru-RU', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    });
  };

  if (!token) return null;

  const isActive = Boolean(status?.active);
  const hasPlan = Boolean(status?.plan && status.plan !== 'none');
  const hasExpiry = Boolean(status?.expires_at_unix);
  const activePlanId = (status?.plan || '').toLowerCase();

  return (
    <PageShell width="narrow">
      <PageHeader
        title="💳 Подписка"
        subtitle="Оформите подписку Boosty для доступа к мерчу и эксклюзивным возможностям"
        onBack={() => navigate('/')}
        backLabel="На главную"
      />

      {error && (
        <div role="alert" className="mb-4 rounded-sm border border-error/30 bg-error/10 px-4 py-3 text-sm text-error">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex justify-center py-16">
          <Spinner size={48} />
        </div>
      ) : (
        <>
          <Card className={isActive ? 'border-success/40' : 'border-white/10'}>
            <Badge tone={isActive ? 'success' : 'neutral'} className="mb-4">
              {isActive ? '✓ Активна' : '○ Не активна'}
            </Badge>
            {hasPlan ? (
              <div className="flex justify-between border-b border-white/10 py-2 text-sm">
                <span className="text-text-secondary">План</span>
                <span className="font-medium text-text-primary">{planLabel(status?.plan)}</span>
              </div>
            ) : null}
            {status?.status && status.status !== 'none' ? (
              <div className="flex justify-between border-b border-white/10 py-2 text-sm">
                <span className="text-text-secondary">Статус</span>
                <span className="font-medium text-text-primary">{statusLabel(status.status)}</span>
              </div>
            ) : null}
            {hasExpiry ? (
              <div className="flex justify-between border-b border-white/10 py-2 text-sm">
                <span className="text-text-secondary">Действует до</span>
                <span className="font-medium text-text-primary">
                  {formatExpiry(status?.expires_at_unix ?? 0)}
                </span>
              </div>
            ) : null}
            {status?.amount_rub ? (
              <div className="flex justify-between py-2 text-sm">
                <span className="text-text-secondary">Сумма</span>
                <span className="font-medium text-text-primary">{status.amount_rub} ₽</span>
              </div>
            ) : null}
            {!isActive && !hasPlan && (
              <p className="text-sm text-text-secondary">
                Подписка ещё не оформлена. Выберите план ниже.
              </p>
            )}
          </Card>

          {isActive ? (
            <div className="mt-8">
              <h2 className="mb-4 font-display text-lg font-semibold text-text-primary">Управление</h2>
              <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
                {PLANS.map((plan) => {
                  const isCurrent = plan.id === activePlanId;
                  return (
                    <Card key={plan.id} interactive className="flex h-full flex-col">
                      <h3 className={`mb-2 font-display text-base font-semibold ${plan.titleClass}`}>
                        {plan.title}
                        {isCurrent ? (
                          <span className="ml-2 text-xs font-normal text-success">· активен</span>
                        ) : null}
                      </h3>
                      <p className="mb-4 flex-1 text-sm leading-relaxed text-text-secondary">
                        {plan.description}
                      </p>
                      {isCurrent ? (
                        <Button variant="ghost" className="mt-auto self-start" disabled>
                          Текущий план
                        </Button>
                      ) : (
                        <Button
                          variant={plan.buttonVariant}
                          className="mt-auto self-start"
                          disabled={checkoutLoading !== null}
                          onClick={() => handleCheckout(plan.id)}
                        >
                          {checkoutLoading === plan.id ? 'Переход к оплате…' : 'Активировать'}
                        </Button>
                      )}
                    </Card>
                  );
                })}
              </div>
              <p className="mt-4 text-sm text-text-secondary">
                Мерч в магазине уже доступен. Продление и отмена появятся позже.
              </p>
            </div>
          ) : (
            <div className="mt-8">
              <h2 className="mb-4 font-display text-lg font-semibold text-text-primary">Оформить подписку</h2>
              <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
                {PLANS.map((plan) => (
                  <Card key={plan.id} interactive className="flex h-full flex-col">
                    <h3 className={`mb-2 font-display text-base font-semibold ${plan.titleClass}`}>
                      {plan.title}
                    </h3>
                    <p className="mb-4 flex-1 text-sm leading-relaxed text-text-secondary">
                      {plan.description}
                    </p>
                    <Button
                      variant={plan.buttonVariant}
                      className="mt-auto self-start"
                      disabled={checkoutLoading !== null}
                      onClick={() => handleCheckout(plan.id)}
                    >
                      {checkoutLoading === plan.id ? 'Переход к оплате…' : 'Активировать'}
                    </Button>
                  </Card>
                ))}
              </div>
            </div>
          )}

          {!isActive && (
            <div className="mt-8 rounded-sm border border-photon-cyan/30 bg-photon-cyan/10 px-4 py-3 text-sm text-photon-cyan">
              ℹ️ Для покупки мерча в магазине нужна активная подписка. После оплаты вернитесь в магазин.
            </div>
          )}
        </>
      )}
    </PageShell>
  );
}
