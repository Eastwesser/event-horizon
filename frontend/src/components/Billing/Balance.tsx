import { useEffect, useState } from 'react';
import api from '../../services/api';
import { canFetchProtected } from '../../lib/auth';

interface Balances {
  lamps: number;
  tickets: number;
}

/** In-flight / short-TTL cache so StrictMode + multiple <Balance /> mounts don't spam /billing/balance/all. */
let cached: Balances | null = null;
let cachedAt = 0;
let inflight: Promise<Balances> | null = null;
const CACHE_MS = 5000;

/** Call after shop purchase / reward so the header balance refreshes immediately. */
export function invalidateBalanceCache(): void {
  cached = null;
  cachedAt = 0;
  inflight = null;
  try {
    localStorage.removeItem('shop_balance_cache');
  } catch {
    // ignore
  }
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new Event('eh:balance-invalidate'));
  }
}

async function fetchBalancesOnce(): Promise<Balances> {
  const now = Date.now();
  if (cached && now - cachedAt < CACHE_MS) return cached;
  if (inflight) return inflight;

  inflight = api
    .get('/billing/balance/all')
    .then((response) => {
      const next = {
        lamps: response.data.lamps || 0,
        tickets: response.data.tickets || 0,
      };
      cached = next;
      cachedAt = Date.now();
      return next;
    })
    .finally(() => {
      inflight = null;
    });

  return inflight;
}

export function Balance() {
  const [balances, setBalances] = useState<Balances>(cached ?? { lamps: 0, tickets: 0 });

  useEffect(() => {
    if (!canFetchProtected()) return;

    let cancelled = false;
    const run = async () => {
      try {
        const next = await fetchBalancesOnce();
        if (!cancelled) setBalances(next);
      } catch (err) {
        console.error('Failed to fetch balance:', err);
      }
    };

    void run();
    const interval = setInterval(run, 30000);
    const onInvalidate = () => {
      void run();
    };
    window.addEventListener('eh:balance-invalidate', onInvalidate);
    return () => {
      cancelled = true;
      clearInterval(interval);
      window.removeEventListener('eh:balance-invalidate', onInvalidate);
    };
  }, []);

  return (
    <div className="flex items-center gap-2 font-hud text-sm tabular-nums">
      <span className="flex items-center gap-1.5 rounded-sm border border-horizon-gold/30 bg-horizon-gold/10 px-2.5 py-1 text-horizon-gold">
        💡 {balances.lamps}
      </span>
      <span className="flex items-center gap-1.5 rounded-sm border border-photon-cyan/30 bg-photon-cyan/10 px-2.5 py-1 text-photon-cyan">
        🎫 {balances.tickets}
      </span>
    </div>
  );
}
