import { useEffect, useState } from 'react';
import { authorsApi, type AuthorSaleRow } from '../../services/authorsApi';
import { StatCard } from '../ui/StatCard';
import { Card } from '../ui/Card';
import { IconLabel } from '../ui/Icon';
import { Spinner } from '../ui/Spinner';
import { Badge } from '../ui/Badge';
import { formatTicketPrice } from '../../lib/formatPrice';

export function AuthorSales() {
  const [sales, setSales] = useState<AuthorSaleRow[]>([]);
  const [count, setCount] = useState(0);
  const [totalTickets, setTotalTickets] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    authorsApi.getSales({ limit: 50 })
      .then((res) => {
        setSales(res.purchases);
        setCount(res.sales_count);
        setTotalTickets(res.tickets_earned);
      })
      .catch(() => setError('Не удалось загрузить продажи'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="flex justify-center py-10"><Spinner /></div>;
  if (error) return <Card className="border-error/30 bg-error/10 text-sm text-error">{error}</Card>;

  // Считаем заработок автора (85%) и платформы (15%)
  const authorEarned = Math.floor(totalTickets * 0.85);
//   const platformEarned = Math.floor(totalTickets * 0.15);

  return (
    <div className="flex flex-col gap-6">
      {/* Блок метрик */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <StatCard 
          size="md" 
          value={count} 
          label={<IconLabel name="cart">Всего продаж</IconLabel>} 
        />
        <StatCard 
          size="md" 
          value={authorEarned} 
          label={<IconLabel name="ticket">Ваш заработок (85%)</IconLabel>} 
        />
        {/* <StatCard 
          size="md" 
          value={platformEarned} 
          label={<IconLabel name="chart">Платформа (15%)</IconLabel>} 
        /> */}
      </div>

      {/* Таблица покупок */}
      <div>
        <h3 className="mb-3 font-display text-lg font-semibold text-text-primary">Последние покупки</h3>
        {sales.length === 0 ? (
          <p className="py-8 text-center text-text-secondary">Продаж пока нет</p>
        ) : (
          <div className="overflow-x-auto rounded-md border border-white/10">
            <table className="w-full min-w-[40rem] border-collapse text-left text-sm">
              <thead className="bg-white/5 text-xs uppercase tracking-wide text-text-muted">
                <tr>
                  <th className="px-3 py-2.5 font-medium">Товар</th>
                  <th className="px-3 py-2.5 font-medium">Покупатель</th>
                  <th className="px-3 py-2.5 font-medium">Цена</th>
                  <th className="px-3 py-2.5 font-medium">Когда</th>
                  <th className="px-3 py-2.5 font-medium">Статус</th>
                </tr>
              </thead>
              <tbody>
                {sales.map((row) => (
                  <tr key={row.id} className="border-t border-white/10">
                    <td className="px-3 py-2.5 text-text-primary">{row.item_name || row.item_id}</td>
                    <td className="px-3 py-2.5 text-text-secondary">{row.buyer_email}</td>
                    <td className="px-3 py-2.5 font-hud text-horizon-gold">{formatTicketPrice(row.price)}</td>
                    <td className="px-3 py-2.5 text-text-muted">
                      {row.purchased_at ? new Date(row.purchased_at).toLocaleString('ru-RU') : '—'}
                    </td>
                    <td className="px-3 py-2.5">
                      {row.refunded_at ? (
                        <Badge tone="error">Возврат</Badge>
                      ) : (
                        <Badge tone="indigo">{row.status || 'COMPLETED'}</Badge>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}