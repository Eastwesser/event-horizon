import { useCallback, useState } from 'react';
import api from '../services/api';
import { invalidateBalanceCache } from '../components/Billing/Balance';
import { syncAchievements } from '../lib/achievements';

export const BOOST_COST = 10;

export type ArmBoostResult = { boostId: string | null; boosted: boolean };

/**
 * Pre-game lamp boost for any allowlisted game_id.
 * Boosted runs must send boost_id on submit; BE marks them unranked.
 */
export function useGameBoost(gameId: string) {
  const [useBoost, setUseBoost] = useState(false);
  const [boostBusy, setBoostBusy] = useState(false);
  const [boostId, setBoostId] = useState<string | null>(null);
  const [boosted, setBoosted] = useState(false);
  const [boostError, setBoostError] = useState<string | null>(null);

  const clearBoost = useCallback(() => {
    setBoostId(null);
    setBoosted(false);
  }, []);

  const armBoost = useCallback(async (): Promise<ArmBoostResult> => {
    if (!useBoost) {
      setBoostId(null);
      setBoosted(false);
      return { boostId: null, boosted: false };
    }
    setBoostBusy(true);
    setBoostError(null);
    try {
      const response = await api.post('/game/boost/start', { game_id: gameId });
      const id = response.data?.boost_id as string | undefined;
      if (!id) {
        throw new Error(response.data?.message || 'boost_id missing');
      }
      invalidateBalanceCache();
      setBoostId(id);
      setBoosted(true);
      void syncAchievements();
      return { boostId: id, boosted: true };
    } catch (e: unknown) {
      const err = e as { response?: { data?: { error?: string; message?: string } }; message?: string };
      const msg =
        err?.response?.data?.error ||
        err?.response?.data?.message ||
        err?.message ||
        'Недостаточно лампочек для boost';
      setBoostError(String(msg));
      setBoostId(null);
      setBoosted(false);
      throw e;
    } finally {
      setBoostBusy(false);
    }
  }, [gameId, useBoost]);

  return {
    useBoost,
    setUseBoost,
    boostBusy,
    boostId,
    boosted,
    boostError,
    setBoostError,
    armBoost,
    clearBoost,
  };
}
