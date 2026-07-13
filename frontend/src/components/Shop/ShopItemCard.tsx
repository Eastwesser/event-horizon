// frontend/src/components/Shop/ShopItemCard.tsx
import React from 'react';
import './ShopItemCard.css';
import type { ShopItem } from '../../store/shopStore';

interface ShopItemCardProps {
  item: ShopItem;
  balance: number;
  onBuyClick: (item: ShopItem) => void;
}

const ShopItemCard: React.FC<ShopItemCardProps> = ({
  item,
  balance,
  onBuyClick,
}) => {
  const canAfford = balance >= item.price_tickets;
  const isOwned = item.owned || false;
  
  const typeEmojis: Record<string, string> = {
    game_skin: '🎨',
    profile_theme: '🎨',
    merch: '🎁',
  };

  const typeLabels: Record<string, string> = {
    game_skin: 'Скин',
    profile_theme: 'Тема',
    merch: 'Мерч',
  };

  if (isOwned) {
    return (
      <div className="shop-item-card owned">
        <div className="item-icon">
          {item.icon_url ? (
            <img src={item.icon_url} alt={item.name} />
          ) : (
            <span className="item-emoji">{typeEmojis[item.type] || '🎁'}</span>
          )}
          <span className="item-type-badge">{typeLabels[item.type] || item.type}</span>
          <div className="owned-badge">✅ Уже куплено</div>
        </div>
        <div className="item-info">
          <h3 className="item-name">{item.name}</h3>
          <p className="item-description">{item.description}</p>
          <div className="item-footer">
            <span className="item-price">🎟️ {item.price_tickets}</span>
            <button className="buy-button owned" disabled>
              В инвентаре
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={`shop-item-card ${!canAfford ? 'locked' : ''}`}>
      <div className="item-icon">
        {item.icon_url ? (
          <img src={item.icon_url} alt={item.name} />
        ) : (
          <span className="item-emoji">{typeEmojis[item.type] || '🎁'}</span>
        )}
        <span className="item-type-badge">{typeLabels[item.type] || item.type}</span>
      </div>
      <div className="item-info">
        <h3 className="item-name">{item.name}</h3>
        <p className="item-description">{item.description}</p>
        <div className="item-footer">
          <span className="item-price">🎟️ {item.price_tickets}</span>
          <button
            className={`buy-button ${canAfford ? 'active' : 'disabled'}`}
            onClick={() => onBuyClick(item)}
            disabled={!canAfford}
          >
            {canAfford ? 'Купить' : 'Не хватает'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default ShopItemCard;