import { useEffect, useState } from 'react';
import api from '../../services/api';

interface Balances {
  lamps: number;
  tickets: number;
}

export function Balance() {
  const [balances, setBalances] = useState<Balances>({ lamps: 0, tickets: 0 });

  useEffect(() => {
    const fetchBalances = async () => {
      try {
        const response = await api.get('/billing/balance/all');
        const lamps = response.data.balances?.find((b: any) => b.currency === 'lamps')?.balance || 0;
        const tickets = response.data.balances?.find((b: any) => b.currency === 'tickets')?.balance || 0;
        setBalances({ lamps, tickets });
      } catch (err) {
        console.error('Failed to fetch balance:', err);
      }
    };
    
    fetchBalances();
    const interval = setInterval(fetchBalances, 30000); // обновляем каждые 30 сек
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="balance">
      <span className="lamps">💡 {balances.lamps}</span>
      <span className="tickets">🎫 {balances.tickets}</span>
    </div>
  );
}
