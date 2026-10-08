import type { ShopItem } from './shopStore';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { CardImage } from '../ui/CardImage';
import { formatTicketAmount } from '../../lib/formatPrice';
import { Icon } from '../ui/Icon';

interface CancelPurchaseModalProps {
  isOpen: boolean;
  item: ShopItem | null;
  refundAmount: number;
  loading: boolean;
  onConfirm: () => void;
  onClose: () => void;
}

export function CancelPurchaseModal({
  isOpen,
  item,
  refundAmount,
  loading,
  onConfirm,
  onClose,
}: CancelPurchaseModalProps) {
  if (!item) return null;

  const amount = refundAmount > 0 ? refundAmount : item.price_tickets;

  return (
    <Modal open={isOpen} onClose={onClose} title="Отменить покупку">
      <div className="flex flex-col items-center text-center">
        <CardImage
          src={item.image_url || item.images?.[0] || item.icon_url}
          alt={item.name}
          className="mb-4 w-[200px]"
          fit="cover"
          fallback={<Icon name="cards" className="h-12 w-12 text-text-muted" />}
        />
        <p className="font-display text-lg font-semibold text-text-primary">{item.name}</p>
        <p className="mt-3 text-sm text-text-secondary">
          Вернуть {formatTicketAmount(amount)}? Предмет будет удалён из инвентаря, товар
          снова станет доступен.
        </p>

        <div className="mt-6 flex w-full gap-3">
          <Button variant="ghost" className="flex-1" onClick={onClose} disabled={loading}>
            Нет
          </Button>
          <Button
            variant="primary"
            className="flex-1"
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? 'Отмена...' : 'Да, вернуть'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}

export default CancelPurchaseModal;
