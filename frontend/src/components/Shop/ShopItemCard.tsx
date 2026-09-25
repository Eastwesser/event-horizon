// frontend/src/components/Shop/ShopItemCard.tsx
import type { ShopItem } from '../../store/shopStore';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { formatTicketPrice } from '../../lib/formatPrice';

interface ShopItemCardProps {
  item: ShopItem;
  balance: number;
  onBuyClick: (item: ShopItem) => void;
}

const categoryEmojis: Record<string, string> = {
  game_skin: '🎨',
  merch: '🎁',
  profile_theme: '🎨',
  other: '🎁',
};

const gameEmojis: Record<string, string> = {
  flappy: '🐦',
  hexagon: '🔶',
  towers: '🗼',
  memory: '🎴',
};

const categoryLabels: Record<string, string> = {
  game_skin: 'Скин',
  merch: 'Мерч',
  profile_theme: 'Тема',
};

function ShopItemCard({ item, balance, onBuyClick }: ShopItemCardProps) {
  const canAfford = balance >= item.price_tickets;
  const isOwned = item.owned || false;

  let emoji = categoryEmojis[item.category] || '🎁';
  if (item.game_id && gameEmojis[item.game_id]) {
    emoji = gameEmojis[item.game_id];
  }

  const categoryLabel = categoryLabels[item.category] || 'Тема';

  return (
    <Card
      interactive={!isOwned && canAfford}
      className={`flex h-full flex-col${!canAfford && !isOwned ? ' opacity-60' : ''}`}
    >
      <div className="mb-4 flex items-center justify-between">
        <span className="text-4xl">{emoji}</span>
        <Badge tone={isOwned ? 'success' : 'indigo'}>
          {isOwned ? '✅ В инвентаре' : categoryLabel}
        </Badge>
      </div>
      <h3 className="font-display text-lg font-semibold text-text-primary">{item.name}</h3>
      <p className="mt-1 text-sm text-text-secondary">{item.description}</p>
      <div className="mt-auto flex items-center justify-between pt-4">
        <span className="font-hud tabular-nums text-horizon-gold">
          {formatTicketPrice(item.price_tickets)}
        </span>
        {isOwned ? (
          <Button variant="ghost" size="sm" disabled>
            В инвентаре
          </Button>
        ) : (
          <Button
            variant={canAfford ? 'primary' : 'ghost'}
            size="sm"
            disabled={!canAfford}
            onClick={() => onBuyClick(item)}
          >
            {canAfford ? 'Купить' : 'Не хватает'}
          </Button>
        )}
      </div>
    </Card>
  );
}

export default ShopItemCard;
