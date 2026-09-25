// frontend/src/components/Shop/PurchaseModal.tsx
import type { ShopItem } from '../../store/shopStore';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { formatTicketPrice } from '../../lib/formatPrice';

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

  return (
    <Modal open={isOpen} onClose={onClose} title="Подтверждение покупки">
      <div className="flex flex-col items-center text-center">
        <div className="mb-3 flex h-16 w-16 items-center justify-center rounded-full bg-white/5 text-3xl">
          {item.icon_url ? (
            <img src={item.icon_url} alt={item.name} className="h-full w-full rounded-full object-cover" />
          ) : (
            <span>🎁</span>
          )}
        </div>
        <p className="font-display text-lg font-semibold text-text-primary">{item.name}</p>
        <p className="mt-1 text-sm text-text-secondary">{item.description}</p>

        <div className="mt-4 flex w-full justify-between rounded-sm border border-white/10 bg-nebula px-4 py-3 font-hud text-sm tabular-nums">
          <span className="text-text-secondary">Цена</span>
          <span className="text-horizon-gold">{formatTicketPrice(item.price_tickets)}</span>
        </div>
        <div className="mt-2 flex w-full justify-between rounded-sm border border-white/10 bg-nebula px-4 py-3 font-hud text-sm tabular-nums">
          <span className="text-text-secondary">Ваш баланс</span>
          <span className="text-photon-cyan">🎟️ {balance}</span>
        </div>

        {merchBlocked && (
          <div className="mt-4 w-full rounded-sm border border-error/40 bg-error/10 p-3 text-sm text-error">
            ❌ {merchBlockReason || 'Покупка мерча недоступна без подписки'}
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
            ❌ Недостаточно билетиков!
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
