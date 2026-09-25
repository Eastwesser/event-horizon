// src/components/common/LoadingSpinner.tsx
import { Spinner } from '../../ui/Spinner';

function LoadingSpinner() {
  return (
    <div className="flex flex-col items-center gap-4 py-16">
      <Spinner size={48} />
      <p className="text-text-secondary">Загрузка...</p>
    </div>
  );
}

export default LoadingSpinner;
