import { useEffect, useState } from 'react';
import { getLeaderboard } from '../services/api';

/** Best published score for a game. `null` if the board is empty or the request failed — never invent 0. */
export function useTopScore(gameId: string | undefined) {
  const [top, setTop] = useState<number | null>(null);

  useEffect(() => {
    if (!gameId) {
      setTop(null);
      return;
    }
    let cancelled = false;
    getLeaderboard(gameId, 10, 1)
      .then(({ data }) => {
        if (cancelled) return;
        const nums = (Array.isArray(data?.entries) ? data.entries : [])
          .map((e: { score?: unknown }) => e?.score)
          .filter((n: unknown): n is number => typeof n === 'number' && Number.isFinite(n));
        setTop(nums.length > 0 ? Math.max(...nums) : null);
      })
      .catch(() => {
        if (!cancelled) setTop(null);
      });
    return () => {
      cancelled = true;
    };
  }, [gameId]);

  return top;
}
