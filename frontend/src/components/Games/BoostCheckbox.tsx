import { BOOST_COST } from '../../hooks/useGameBoost';
import { Icon } from '../ui/Icon';

type Props = {
  useBoost: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
  /** Compact: hide long warning until opened (default true — less eye-sore). */
  collapsed?: boolean;
};

/** Shared pre-game boost control — tucked away; full copy in help. */
export function BoostCheckbox({
  useBoost,
  onChange,
  disabled,
  collapsed = true,
}: Props) {
  const body = (
    <label className="flex max-w-sm cursor-pointer flex-col gap-1 text-xs text-text-secondary">
      <span className="inline-flex items-center gap-2 text-sm text-text-primary">
        <input
          type="checkbox"
          checked={useBoost}
          disabled={disabled}
          onChange={(e) => onChange(e.target.checked)}
          className="accent-horizon-gold"
        />
        <Icon name="lamp" className="h-3.5 w-3.5 text-horizon-gold" aria-hidden />
        Boost (−{BOOST_COST})
      </span>
      {useBoost && (
        <span className="pl-6 text-[11px] leading-snug text-horizon-gold/90">
          Не в лидерборд · без награды билетиками
        </span>
      )}
    </label>
  );

  if (!collapsed) return body;

  return (
    <details className="max-w-sm rounded-sm border border-white/10 bg-nebula/40 px-2 py-1.5 text-text-secondary open:pb-2">
      <summary className="cursor-pointer list-none text-xs text-text-muted marker:content-none [&::-webkit-details-marker]:hidden">
        <span className="underline decoration-white/20 underline-offset-2 hover:text-text-secondary">
          Boost (опционально)
        </span>
      </summary>
      <div className="mt-2">{body}</div>
    </details>
  );
}

export function boostUnrankedToast(): string {
  return 'Забег с boost — не попал в лидерборд';
}
