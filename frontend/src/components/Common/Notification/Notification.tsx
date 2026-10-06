// src/components/common/Notification.tsx
import { useEffect, useState } from 'react';
import { cn } from '../../../lib/cn';
import { Icon, type IconName } from '../../ui/Icon';

interface NotificationProps {
  /** Warning must always pair hue with an icon (never color-only). */
  type: 'success' | 'error' | 'info' | 'warning';
  message: string;
  onClose: () => void;
  autoClose?: number; // ms
}

const toneClasses: Record<NotificationProps['type'], string> = {
  success: 'border-success/40 bg-success/10 text-success',
  error: 'border-error/40 bg-error/10 text-error',
  info: 'border-photon-cyan/40 bg-photon-cyan/10 text-photon-cyan',
  warning: 'border-warning/40 bg-warning/10 text-warning',
};

const icons: Record<NotificationProps['type'], IconName> = {
  success: 'check',
  error: 'x',
  info: 'info',
  warning: 'warning',
};

function Notification({
  type,
  message,
  onClose,
  autoClose = 3000,
}: NotificationProps) {
  const [isVisible, setIsVisible] = useState(true);

  useEffect(() => {
    const timer = setTimeout(() => {
      setIsVisible(false);
      setTimeout(onClose, 300);
    }, autoClose);

    return () => clearTimeout(timer);
  }, [autoClose, onClose]);

  const handleClose = () => {
    setIsVisible(false);
    setTimeout(onClose, 300);
  };

  return (
    <div
      role="status"
      className={cn(
        'fixed right-6 top-6 z-[1100] flex items-center gap-3 rounded-md border px-4 py-3 shadow-elevated backdrop-blur-md transition-all duration-300',
        'bg-nebula-elevated/90',
        toneClasses[type],
        isVisible ? 'translate-x-0 opacity-100' : 'translate-x-4 opacity-0',
      )}
    >
      <Icon name={icons[type]} className="h-5 w-5 shrink-0" aria-hidden />
      <span className="text-sm text-text-primary">{message}</span>
      <button
        onClick={handleClose}
        aria-label="Закрыть уведомление"
        className="ml-2 text-text-secondary transition-colors hover:text-text-primary"
      >
        <Icon name="x" className="h-4 w-4" />
      </button>
    </div>
  );
}

export default Notification;
