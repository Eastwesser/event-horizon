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
        console.log('📊 RAW Balance response:', response.data);
        console.log('📊 lamps:', response.data.lamps);
        console.log('📊 tickets:', response.data.tickets);
        console.log('📊 Balance response final:', response.data);
        
        // Сервер возвращает { lamps: X, tickets: Y }
        setBalances({
          lamps: response.data.lamps || 0,
          tickets: response.data.tickets || 0,
        });
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
