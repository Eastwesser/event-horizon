import { BOOST_COST } from '../../hooks/useGameBoost';

type Props = {
  useBoost: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
};

/** Shared pre-game boost checkbox (RU copy matches Flappy). */
export function BoostCheckbox({ useBoost, onChange, disabled }: Props) {
  return (
    <label className="flex max-w-md cursor-pointer flex-col gap-1 text-sm text-text-secondary">
      <span className="inline-flex items-center gap-2 text-text-primary">
        <input
          type="checkbox"
          checked={useBoost}
          disabled={disabled}
          onChange={(e) => onChange(e.target.checked)}
        />
        Использовать boost ({BOOST_COST} лампочек)
      </span>
      {useBoost && (
        <span className="text-xs text-horizon-gold/90">
          Этот забег не попадёт в лидерборд — boost считается нечестным преимуществом
        </span>
      )}
    </label>
  );
}

export function boostUnrankedToast(): string {
  return 'Забег с boost — не попал в лидерборд';
}
