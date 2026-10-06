import type { ReactNode } from 'react';
import { Button } from '../ui/Button';

type Props = {
  onNewGame: () => void;
  onHome: () => void;
  onSave?: () => void;
  newLabel?: string;
  homeLabel?: string;
  saveLabel?: string;
  busy?: boolean;
  extra?: ReactNode;
};

/** Equal-width GO actions: Новая / optional Сохранить / На главную. */
export function GameOverActions({
  onNewGame,
  onHome,
  onSave,
  newLabel = 'Новая игра',
  homeLabel = 'На главную',
  saveLabel = 'Сохранить',
  busy = false,
  extra,
}: Props) {
  const cols = onSave ? 'sm:grid-cols-3' : 'sm:grid-cols-2';
  return (
    <div className={`mt-6 grid grid-cols-1 gap-3 ${cols}`}>
      <Button
        variant="primary"
        size="sm"
        className="w-full justify-center"
        onClick={onNewGame}
        disabled={busy}
      >
        {newLabel}
      </Button>
      {onSave ? (
        <Button
          variant="secondary"
          size="sm"
          className="w-full justify-center"
          onClick={onSave}
          disabled={busy}
        >
          {saveLabel}
        </Button>
      ) : null}
      <Button
        variant="ghost"
        size="sm"
        className="w-full justify-center"
        onClick={onHome}
        disabled={busy}
      >
        {homeLabel}
      </Button>
      {extra}
    </div>
  );
}
