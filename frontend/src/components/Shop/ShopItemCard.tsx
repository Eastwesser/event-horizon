// frontend/src/components/Shop/ShopItemCard.tsx
import { Link } from 'react-router-dom';
import { type ShopItem } from './shopStore';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { CardImage } from '../ui/CardImage';
import { Icon } from '../ui/Icon';
import { formatTicketPrice } from '../../lib/formatPrice';
import { itemFallbackIcon } from '../../lib/itemIcons';
import { stockLabel, isDigitalItem } from '../../lib/shopItemMap';
import { cardFlagBadges } from './cardAttributes';

interface ShopItemCardProps {
  item: ShopItem;
  balance: number;
  onBuyClick: (item: ShopItem) => void;
  detailTo?: string;
}

// ИСПРАВЛЕНО v.1.1.3.1: Добавлен ключ "карточка"
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
  const isDigital = isDigitalItem(item);
  const imageSrc = isDigital
    ? ''
    : item.image_url || item.images?.[0] || item.icon_url || '';
  const canAfford = balance >= item.price_tickets;
  // const isOwned = item.owned || false;
  // const patch = useShopStore((s) => s.catalogPatches[item.id]);
  // const isOwned = patch?.owned ?? item.owned ?? false;
  const isOwned = item.owned || false;
  const stockText = stockLabel(item.stock);
  const outOfStock = item.stock === 0;
  
  // Теперь категория "карточка" превратится в "Карточка"
  const categoryLabel = categoryLabels[item.category] || categoryLabels[item.type] || 'Товар';
  
  const icon = itemFallbackIcon(item);
  const flagBadges = item.attributes ? cardFlagBadges(item.attributes) : [];
  const to = detailTo || `/shop/item/${item.id}`;

  return (
    <Card
      interactive={!isOwned && canAfford && !outOfStock}
      className={`flex h-full flex-col overflow-hidden${!canAfford && !isOwned ? ' opacity-60' : ''}${outOfStock ? ' opacity-70' : ''}`}
    >
      <Link to={to} className="flex min-h-0 flex-1 flex-col text-inherit no-underline">
        <CardImage
          src={imageSrc}
          alt={item.name}
          className="w-full shrink-0"
          fit="cover"
          fallback={<Icon name={icon} className="h-12 w-12 text-text-muted" />}
        />
        <div className="mt-4 flex min-h-0 flex-1 flex-col">

          {/* 1. ИМЯ КАРТОЧКИ (ТЕПЕРЬ НА ВСЮ ШИРИНУ, БЕЗ БЕЙДЖА) */}
          <div className="mb-2">
            <h3 className="min-w-0 font-display text-lg font-semibold leading-snug text-text-primary line-clamp-2 break-words min-h-[2.8rem]">
              {item.name}
            </h3>
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

          {/* 2. НИЗ: СТОК + БЕЙДЖ "КАРТОЧКА" */}
          <div className="mt-auto pt-3 flex items-center justify-between gap-2">
            <div className="text-xs">
              {stockText !== null && !isDigital ? (
                <p className={outOfStock ? 'text-text-muted' : 'text-text-secondary'}>
                  {stockText}
                </p>
              ) : null}
            </div>
            <Badge tone={isOwned ? 'success' : 'indigo'}>
              {isOwned ? (
                <Icon name="check" className="h-3 w-3" title="В инвентаре" />
              ) : (
                categoryLabel
              )}
            </Badge>
          </div>
        </div>
      </Link>

      {/* 3. НИЖНИЙ БЛОК: ЦЕНА + КНОПКА (ТЕПЕРЬ С FLEX-WRAP) */}
      <div className="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-white/5 pt-3">
        <span className="inline-flex min-w-0 shrink-0 items-center gap-1 font-hud tabular-nums text-horizon-gold">
          <Icon name="ticket" className="h-3.5 w-3.5" />
          {formatTicketPrice(item.price_tickets)}
        </span>
        {isOwned ? (
          <Button variant="ghost" size="sm" disabled className="shrink-0 whitespace-nowrap">
            В инвентаре
          </Button>
        ) : (
          <Button
            variant={canAfford && !outOfStock ? 'primary' : 'ghost'}
            size="sm"
            className="shrink-0 whitespace-nowrap"
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