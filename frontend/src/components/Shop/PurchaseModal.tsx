// frontend/src/components/Shop/PurchaseModal.tsx
import React from 'react';
import './PurchaseModal.css';
import type { ShopItem } from '../../store/shopStore';

interface PurchaseModalProps {
  isOpen: boolean;
  item: ShopItem | null;
  balance: number;
  onConfirm: () => void;
  onClose: () => void;
  loading: boolean;
}

const PurchaseModal: React.FC<PurchaseModalProps> = ({
  isOpen,
  item,
  balance,
  onConfirm,
  onClose,
  loading,
}) => {
  if (!isOpen || !item) return null;

  const canAfford = balance >= item.price_tickets;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close" onClick={onClose}>×</button>
        <div className="modal-body">
          <div className="modal-icon">
            {item.icon_url ? (
              <img src={item.icon_url} alt={item.name} />
            ) : (
              <span>🎁</span>
            )}
          </div>
          <h2>Подтверждение покупки</h2>
          <p className="modal-item-name">{item.name}</p>
          <p className="modal-item-description">{item.description}</p>
          <div className="modal-price">
            <span>Цена: 🎟️ {item.price_tickets}</span>
          </div>
          <div className="modal-balance">
            <span>Ваш баланс: 🎟️ {balance}</span>
          </div>
          {!canAfford && (
            <div className="modal-error">
              ❌ Недостаточно билетиков!
            </div>
          )}
          <div className="modal-actions">
            <button
              className="modal-cancel"
              onClick={onClose}
              disabled={loading}
            >
              Отмена
            </button>
            <button
              className="modal-confirm"
              onClick={onConfirm}
              disabled={!canAfford || loading}
            >
              {loading ? 'Покупка...' : 'Да, купить'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PurchaseModal;