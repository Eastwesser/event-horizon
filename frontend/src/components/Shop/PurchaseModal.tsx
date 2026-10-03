// frontend/src/components/Shop/PurchaseModal.tsx
import type { ShopItem } from '../../store/shopStore';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { CardImage } from '../ui/CardImage';
import { formatTicketPrice } from '../../lib/formatPrice';
import { Icon } from '../ui/Icon';
import { CardPurchaseSummary } from './cardAttributes';

interface PurchaseModalProps {
  isOpen: boolean;
  item: ShopItem | null;
  balance: number;
  onConfirm: () => void;
  onClose: () => void;
  loading: boolean;
  merchAllowed?: boolean | null;
  merchBlockReason?: string;
  onGoSubscription?: () => void;
}

function PurchaseModal({
  isOpen,
  item,
  balance,
  onConfirm,
  onClose,
  loading,
  merchAllowed = true,
  merchBlockReason = '',
  onGoSubscription,
}: PurchaseModalProps) {
  if (!item) return null;

  const canAfford = balance >= item.price_tickets;
  const merchBlocked = merchAllowed === false;
  const isCard = (item.type || item.category) === 'карточка';

  return (
    <Modal open={isOpen} onClose={onClose} title="Подтверждение покупки">
      <div className="flex flex-col items-center text-center">
        <CardImage
          src={item.image_url || item.images?.[0] || item.icon_url}
          alt={item.name}
          className="mb-4 w-[240px]"
          fit="cover"
          fallback={<Icon name="gift" className="h-12 w-12 text-text-muted" />}
        />
        <p className="font-display text-lg font-semibold text-text-primary">{item.name}</p>

        {isCard ? (
          <CardPurchaseSummary attrs={item.attributes} />
        ) : item.description ? (
          <p className="mt-1 line-clamp-3 text-sm text-text-secondary">{item.description}</p>
        ) : null}

        <div className="mt-4 flex w-full justify-between rounded-sm border border-white/10 bg-nebula px-4 py-3 font-hud text-sm tabular-nums">
          <span className="text-text-secondary">Цена</span>
          <span className="inline-flex items-center gap-1.5 text-horizon-gold">
            <Icon name="ticket" className="h-4 w-4" />
            {formatTicketPrice(item.price_tickets)}
          </span>
        </div>
        <div className="mt-2 flex w-full justify-between rounded-sm border border-white/10 bg-nebula px-4 py-3 font-hud text-sm tabular-nums">
          <span className="text-text-secondary">Ваш баланс</span>
          <span className="inline-flex items-center gap-1.5 text-photon-cyan">
            <Icon name="ticket" className="h-4 w-4" />
            {balance}
          </span>
        </div>

        {merchBlocked && (
          <div className="mt-4 w-full rounded-sm border border-error/40 bg-error/10 p-3 text-sm text-error">
            {merchBlockReason || 'Покупка мерча недоступна без подписки'}
            {onGoSubscription && (
              <button
                type="button"
                onClick={onGoSubscription}
                className="mt-2 block font-semibold text-indigo-soft underline"
              >
                Оформить подписку →
              </button>
            )}
          </div>
        )}
        {!canAfford && !merchBlocked && (
          <div className="mt-4 w-full rounded-sm border border-error/40 bg-error/10 p-3 text-sm text-error">
            Недостаточно билетиков!
          </div>
        )}

        <div className="mt-6 flex w-full gap-3">
          <Button variant="ghost" className="flex-1" onClick={onClose} disabled={loading}>
            Отмена
          </Button>
          <Button
            variant="primary"
            className="flex-1"
            onClick={onConfirm}
            disabled={!canAfford || loading || merchBlocked}
          >
            {loading ? 'Покупка...' : 'Да, купить'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}

export default PurchaseModal;
