// frontend/src/components/Shop/ShopItemCard.tsx
import { Link } from 'react-router-dom';
import type { ShopItem } from '../../store/shopStore';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { CardImage } from '../ui/CardImage';
import { formatTicketPrice } from '../../lib/formatPrice';
import { stockLabel } from '../../lib/shopItemMap';
import { cardFlagBadges } from './cardAttributes';

interface ShopItemCardProps {
  item: ShopItem;
  balance: number;
  onBuyClick: (item: ShopItem) => void;
  detailTo?: string;
}

const categoryEmojis: Record<string, string> = {
  game_skin: '🎨',
  merch: '🎁',
  profile_theme: '🎨',
  карточка: '🃏',
  брелок: '🔑',
  картина: '🖼',
  фенечка: '💫',
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
  карточка: 'Карточка',
  брелок: 'Брелок',
  картина: 'Картина',
  фенечка: 'Фенечка',
};

function fallbackEmoji(item: ShopItem): string {
  if (item.game_id && gameEmojis[item.game_id]) return gameEmojis[item.game_id];
  return categoryEmojis[item.category] || categoryEmojis[item.type] || '🎁';
}

function ShopItemCard({ item, balance, onBuyClick, detailTo }: ShopItemCardProps) {
  const imageSrc = item.image_url || item.images?.[0] || item.icon_url || '';
  const canAfford = balance >= item.price_tickets;
  const isOwned = item.owned || false;
  const stockText = stockLabel(item.stock);
  const outOfStock = item.stock === 0;
  const categoryLabel = categoryLabels[item.category] || categoryLabels[item.type] || 'Товар';
  const emoji = fallbackEmoji(item);
  const flagBadges = item.attributes ? cardFlagBadges(item.attributes) : [];
  const to = detailTo || `/shop/item/${item.id}`;

  return (
    <Card
      interactive={!isOwned && canAfford && !outOfStock}
      className={`flex h-full flex-col${!canAfford && !isOwned ? ' opacity-60' : ''}${outOfStock ? ' opacity-70' : ''}`}
    >
      <Link to={to} className="flex min-h-0 flex-1 flex-col text-inherit no-underline">
        <CardImage
          src={imageSrc}
          alt={item.name}
          className="w-full shrink-0"
          fit="cover"
          fallback={
            <span className="text-4xl" aria-hidden>
              {emoji}
            </span>
          }
        />
        <div className="mt-4 flex min-h-0 flex-1 flex-col">
          <div className="mb-2 flex items-start justify-between gap-2">
            <h3 className="font-display text-lg font-semibold leading-snug text-text-primary">
              {item.name}
            </h3>
            <Badge tone={isOwned ? 'success' : 'indigo'}>
              {isOwned ? '✅ В инвентаре' : categoryLabel}
            </Badge>
          </div>
          {flagBadges.length > 0 ? (
            <div className="mb-2 flex flex-wrap gap-1">
              {flagBadges.map((f) => (
                <Badge key={f.key} tone={f.tone}>
                  {f.label}
                </Badge>
              ))}
            </div>
          ) : null}
          <div className="mt-auto pt-3">
            {stockText !== null && (
              <p
                className={`text-xs ${
                  outOfStock ? 'text-text-muted' : 'text-text-secondary'
                }`}
              >
                {stockText}
              </p>
            )}
          </div>
        </div>
      </Link>

      <div className="mt-3 flex items-center justify-between gap-2 border-t border-white/5 pt-3">
        <span className="font-hud tabular-nums text-horizon-gold">
          {formatTicketPrice(item.price_tickets)}
        </span>
        {isOwned ? (
          <Button variant="ghost" size="sm" disabled>
            В инвентаре
          </Button>
        ) : (
          <Button
            variant={canAfford && !outOfStock ? 'primary' : 'ghost'}
            size="sm"
            disabled={!canAfford || outOfStock}
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();
              onBuyClick(item);
            }}
          >
            {outOfStock ? 'Нет в наличии' : canAfford ? 'Купить' : 'Не хватает'}
          </Button>
        )}
      </div>
    </Card>
  );
}

export default ShopItemCard;
