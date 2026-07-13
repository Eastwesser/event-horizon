// frontend/src/components/Shop/ShopItemCard.tsx
import React from 'react';
import './ShopItemCard.css';
import type { ShopItem } from '../../store/shopStore';

interface ShopItemCardProps {
  item: ShopItem;
  balance: number;
  onBuyClick: (item: ShopItem) => void;
}

// Цвета для разных категорий
const categoryColors: Record<string, string> = {
  game_skin: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
  merch: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
  profile_theme: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
  other: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
};

// Эмодзи для категорий
const categoryEmojis: Record<string, string> = {
  game_skin: '🎨',
  merch: '🎁',
  profile_theme: '🎨',
  other: '🎁',
};

// Эмодзи для конкретных игр
const gameEmojis: Record<string, string> = {
  flappy: '🐦',
  hexagon: '🔶',
  towers: '🗼',
  memory: '🎴',
};

const ShopItemCard: React.FC<ShopItemCardProps> = ({
  item,
  balance,
  onBuyClick,
}) => {
  const canAfford = balance >= item.price_tickets;
  const isOwned = item.owned || false;
  
  // Получаем эмодзи для товара
  let emoji = categoryEmojis[item.category] || '🎁';
  if (item.game_id && gameEmojis[item.game_id]) {
    emoji = gameEmojis[item.game_id];
  }
  
  const bgColor = categoryColors[item.category] || categoryColors.other;

  if (isOwned) {
    return (
      <div className="shop-item-card owned">
        <div className="item-icon" style={{ background: bgColor }}>
          <span className="item-emoji">{emoji}</span>
          <span className="item-type-badge">{item.category === 'game_skin' ? 'Скин' : item.category === 'merch' ? 'Мерч' : 'Тема'}</span>
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
      <div className="item-icon" style={{ background: bgColor }}>
        <span className="item-emoji">{emoji}</span>
        <span className="item-type-badge">{item.category === 'game_skin' ? 'Скин' : item.category === 'merch' ? 'Мерч' : 'Тема'}</span>
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