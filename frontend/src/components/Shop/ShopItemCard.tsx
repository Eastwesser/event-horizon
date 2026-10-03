// frontend/src/components/Shop/ShopItemCard.tsx
import { Link } from 'react-router-dom';
import type { ShopItem } from '../../store/shopStore';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { CardImage } from '../ui/CardImage';
import { Icon, IconLabel } from '../ui/Icon';
import { formatTicketPrice } from '../../lib/formatPrice';
import { itemFallbackIcon } from '../../lib/itemIcons';
import { stockLabel } from '../../lib/shopItemMap';
import { cardFlagBadges } from './cardAttributes';

interface ShopItemCardProps {
  item: ShopItem;
  balance: number;
  onBuyClick: (item: ShopItem) => void;
  detailTo?: string;
}

const categoryLabels: Record<string, string> = {
  game_skin: 'Скин',
  merch: 'Мерч',
  profile_theme: 'Тема',
  карточка: 'Карточка',
  брелок: 'Брелок',
  картина: 'Картина',
  фенечка: 'Фенечка',
};

function ShopItemCard({ item, balance, onBuyClick, detailTo }: ShopItemCardProps) {
  const imageSrc = item.image_url || item.images?.[0] || item.icon_url || '';
  const canAfford = balance >= item.price_tickets;
  const isOwned = item.owned || false;
  const stockText = stockLabel(item.stock);
  const outOfStock = item.stock === 0;
  const categoryLabel = categoryLabels[item.category] || categoryLabels[item.type] || 'Товар';
  const icon = itemFallbackIcon(item);
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
            <Icon name={icon} className="h-12 w-12 text-text-muted" />
          }
        />
        <div className="mt-4 flex min-h-0 flex-1 flex-col">
          <div className="mb-2 flex items-start justify-between gap-2">
            <h3 className="font-display text-lg font-semibold leading-snug text-text-primary">
              {item.name}
            </h3>
            <Badge tone={isOwned ? 'success' : 'indigo'}>
              {isOwned ? (
                <IconLabel name="check" iconClassName="h-3 w-3">
                  В инвентаре
                </IconLabel>
              ) : (
                categoryLabel
              )}
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
        <span className="inline-flex items-center gap-1 font-hud tabular-nums text-horizon-gold">
          <Icon name="ticket" className="h-3.5 w-3.5" />
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
