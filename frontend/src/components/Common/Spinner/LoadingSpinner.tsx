// src/components/common/LoadingSpinner.tsx
import { Spinner } from '../../ui/Spinner';

type LoadingSpinnerProps = {
  /** Fill viewport and center (shop / detail cold load). */
  fullscreen?: boolean;
};

function LoadingSpinner({ fullscreen = false }: LoadingSpinnerProps) {
  return (
    <div
      className={
        fullscreen
          ? 'flex min-h-screen flex-col items-center justify-center gap-4 bg-void'
          : 'flex flex-col items-center gap-4 py-16'
      }
    >
      <Spinner size={48} />
      <p className="text-text-secondary">Загрузка...</p>
    </div>
  );
}

export default LoadingSpinner;
