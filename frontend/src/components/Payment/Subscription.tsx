// frontend/src/components/Payment/Subscription.tsx
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { paymentApi, type SubscriptionStatus } from '../../services/paymentApi';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import './Subscription.css';

const PLANS = [
  {
    id: 'present' as const,
    name: 'Present',
    title: '🎁 Текущий план',
    description: 'Подписка на Boosty — доступ к мерчу и бонусам сообщества.',
  },
  {
    id: 'future' as const,
    name: 'Future',
    title: '🚀 Будущий план',
    description: 'Расширенная подписка с дополнительными привилегиями (когда будет доступна).',
  },
];

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

  return (
    <div className="subscription-container">
      <button className="back-btn" onClick={() => navigate('/')}>
        ← Назад
      </button>

      <div className="subscription-header">
        <h1>💳 Подписка</h1>
        <p>Оформите подписку Boosty для доступа к мерчу и эксклюзивным возможностям</p>
      </div>

      {error && <div className="subscription-error">{error}</div>}

      {loading ? (
        <div className="subscription-loading">
          <LoadingSpinner />
        </div>
      ) : (
        <>
          <div className={`subscription-status-card ${status?.active ? 'active' : 'inactive'}`}>
            <span className={`status-badge status-badge--${status?.active ? 'active' : 'inactive'}`}>
              {status?.active ? '✅ Активна' : '❌ Не активна'}
            </span>
            <div className="status-row">
              <span className="status-label">План</span>
              <span className="status-value">{status?.plan || '—'}</span>
            </div>
            <div className="status-row">
              <span className="status-label">Статус</span>
              <span className="status-value">{status?.status || '—'}</span>
            </div>
            <div className="status-row">
              <span className="status-label">Действует до</span>
              <span className="status-value">{formatExpiry(status?.expires_at_unix ?? 0)}</span>
            </div>
            {status?.amount_rub ? (
              <div className="status-row">
                <span className="status-label">Сумма</span>
                <span className="status-value">{status.amount_rub} ₽</span>
              </div>
            ) : null}
          </div>

          <div className="subscription-plans">
            <h2>Оформить подписку</h2>
            <div className="plans-grid">
              {PLANS.map((plan) => (
                <div key={plan.id} className="plan-card">
                  <h3>{plan.title}</h3>
                  <p>{plan.description}</p>
                  <button
                    className="plan-btn"
                    disabled={checkoutLoading !== null}
                    onClick={() => handleCheckout(plan.id)}
                  >
                    {checkoutLoading === plan.id ? '⏳ Переход к оплате...' : '💳 Активировать'}
                  </button>
                </div>
              ))}
            </div>
          </div>

          <div className="merch-note">
            ℹ️ Для покупки мерча в магазине нужна активная подписка. После оплаты вернитесь в магазин.
          </div>
        </>
      )}
    </div>
  );
}
