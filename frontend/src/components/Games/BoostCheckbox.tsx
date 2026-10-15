import { useEffect, useState } from 'react';
import { BOOST_COST } from '../../hooks/useGameBoost';
import { canFetchProtected } from '../../lib/auth';
import { fetchBalancesOnce } from '../Billing/Balance';
import { Icon } from '../ui/Icon';
import { BOOST_EFFECT_BY_GAME } from '../../lib/gameBoostCopy';

// type Props_old = {
//   useBoost: boolean;
//   onChange: (checked: boolean) => void;
//   disabled?: boolean;
//   /** Compact: hide long warning until opened (default true — less eye-sore). */
//   collapsed?: boolean;
// };

type Props = {
  gameId: string;
  useBoost: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
};

// /** Shared pre-game boost control — tucked away; full copy in help. */
// export function BoostCheckbox_old({
//   useBoost,
//   onChange,
//   disabled,
//   collapsed = true,
// }: Props) {
//   const [lamps, setLamps] = useState<number | null>(null);

//   useEffect(() => {
//     if (!canFetchProtected()) {
//       setLamps(0);
//       return;
//     }
//     let cancelled = false;
//     const load = () => {
//       void fetchBalancesOnce()
//         .then((b) => {
//           if (!cancelled) setLamps(b.lamps);
//         })
//         .catch(() => {
//           if (!cancelled) setLamps(0);
//         });
//     };
//     load();
//     const onInvalidate = () => load();
//     window.addEventListener('eh:balance-invalidate', onInvalidate);
//     return () => {
//       cancelled = true;
//       window.removeEventListener('eh:balance-invalidate', onInvalidate);
//     };
//   }, []);

//   // Can't afford boost → force checkbox off and disable.
//   const broke = lamps !== null && lamps < BOOST_COST;
//   useEffect(() => {
//     if (broke && useBoost) onChange(false);
//   }, [broke, useBoost, onChange]);

//   const gated = Boolean(disabled || broke);
//   const body = (
//     <label className="flex max-w-sm cursor-pointer flex-col gap-1 text-xs text-text-secondary">
//       <span className="inline-flex items-center gap-2 text-sm text-text-primary">
//         <input
//           type="checkbox"
//           checked={useBoost && !broke}
//           disabled={gated}
//           onChange={(e) => onChange(e.target.checked)}
//           className="accent-horizon-gold"
//         />
//         <Icon name="lamp" className="h-3.5 w-3.5 text-horizon-gold" aria-hidden />
//         Boost (−{BOOST_COST})
//         {broke ? (
//           <span className="text-[11px] text-text-muted">(мало ламп)</span>
//         ) : null}
//       </span>
//       {useBoost && !broke && (
//         <span className="inline-flex items-center gap-1 pl-6 text-[11px] leading-snug text-horizon-gold/90">
//           Не в лидерборд · без награды
//           <Icon name="ticket" className="h-3 w-3" aria-hidden />
//           билетиками
//         </span>
//       )}
//     </label>
//   );

//   if (!collapsed) return body;

//   return (
//     <details className="max-w-sm rounded-sm border border-white/10 bg-nebula/40 px-2 py-1.5 text-text-secondary open:pb-2">
//       <summary className="cursor-pointer list-none text-xs text-text-muted marker:content-none [&::-webkit-details-marker]:hidden">
//         <span className="underline decoration-white/20 underline-offset-2 hover:text-text-secondary">
//           Boost (опционально)
//         </span>
//       </summary>
//       <div className="mt-2">{body}</div>
//     </details>
//   );
// }

/** Shared pre-game boost control — explains effect right away. */
export function BoostCheckbox({ gameId, useBoost, onChange, disabled }: Props) {
  const [lamps, setLamps] = useState<number | null>(null);

  useEffect(() => {
    if (!canFetchProtected()) {
      setLamps(0);
      return;
    }
    let cancelled = false;
    const load = () => {
      void fetchBalancesOnce()
        .then((b) => {
          if (!cancelled) setLamps(b.lamps);
        })
        .catch(() => {
          if (!cancelled) setLamps(0);
        });
    };
    load();
    const onInvalidate = () => load();
    window.addEventListener('eh:balance-invalidate', onInvalidate);
    return () => {
      cancelled = true;
      window.removeEventListener('eh:balance-invalidate', onInvalidate);
    };
  }, []);

  // Can't afford boost → force checkbox off and disable.
  const broke = lamps !== null && lamps < BOOST_COST;
  useEffect(() => {
    if (broke && useBoost) onChange(false);
  }, [broke, useBoost, onChange]);

  const gated = Boolean(disabled || broke);
  const effectText = BOOST_EFFECT_BY_GAME[gameId] || 'Облегчает игру';

  return (
    <div className="flex max-w-sm flex-col gap-1 text-xs text-text-secondary">
      <label className="flex cursor-pointer items-center gap-2">
        <input
          type="checkbox"
          checked={useBoost && !broke}
          disabled={gated}
          onChange={(e) => onChange(e.target.checked)}
          className="accent-horizon-gold"
        />
        <Icon name="lamp" className="h-3.5 w-3.5 text-horizon-gold" aria-hidden />
        <span className="text-sm text-text-primary">
          Boost за {BOOST_COST} ламп
        </span>
        {broke && <span className="text-[11px] text-text-muted">(мало ламп)</span>}
      </label>
      <p className="pl-6 text-[11px] leading-snug text-horizon-gold/90">
        {effectText} · <span className="text-text-muted">не в лидерборд</span>
      </p>
    </div>
  );
}

export function boostUnrankedToast(): string {
  return 'Забег с Boost не будет учтен в Leaderboard';
}

// Что изменилось v.1.1.2.x:
// Убран collapsed и details. Теперь он всегда развернут.
// Добавлен проп gameId.
// Показывается эффект буста для конкретной игры.
// Убрано дублирование про «билетики».
