import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useInventory } from '../../hooks/useInventory';
import { useUserRole } from '../../hooks/useUserRole';
import { InventoryEditModal } from './InventoryEditModal';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { CardImage } from '../ui/CardImage';
import { Icon, IconLabel } from '../ui/Icon';
import type { InventoryItem } from '../../services/inventoryApi';
import { formatTicketPrice } from '../../lib/formatPrice';
import { itemFallbackIcon } from '../../lib/itemIcons';
import { stockLabel } from '../../lib/shopItemMap';

interface InventoryItemCardProps {
  item: InventoryItem;
}

export const InventoryItemCard: React.FC<InventoryItemCardProps> = ({ item }) => {
  const { isAuthor } = useUserRole();
  const { deleteItem } = useInventory();
  const [showEditModal, setShowEditModal] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    if (window.confirm(`Удалить товар "${item.name}"?`)) {
      setIsDeleting(true);
      try {
        await deleteItem(item.id);
      } finally {
        setIsDeleting(false);
      }
    }
  };

  const stockText = stockLabel(item.stock);

  return (
    <>
      <Card className="flex h-full flex-col">
        <Link
          to={`/inventory/${item.id}`}
          className="flex min-h-0 flex-1 flex-col text-inherit no-underline"
        >
          <CardImage
            src={item.images?.[0]}
            alt={item.name}
            className="w-full shrink-0"
            fit="cover"
            fallback={
              <Icon
                name={itemFallbackIcon({ type: item.type, category: item.type })}
                className="h-12 w-12 text-horizon-gold"
              />
            }
          />
          <div className="mt-4 flex min-h-0 flex-1 flex-col">
            <h3 className="font-display text-base font-semibold leading-snug text-text-primary">
              {item.name}
            </h3>
            <p className="mt-1 line-clamp-2 text-sm text-text-secondary">{item.description}</p>
            <div className="mt-auto flex flex-wrap items-center gap-2 pt-3 text-sm">
              <Badge tone="indigo">{item.type}</Badge>
              <span className="inline-flex items-center gap-1 font-hud tabular-nums text-horizon-gold">
                <Icon name="ticket" className="h-3.5 w-3.5" />
                {formatTicketPrice(item.price)}
              </span>
              {stockText !== null && (
                <span className="text-text-muted">{stockText}</span>
              )}
            </div>
          </div>
        </Link>
        {isAuthor && (
          <div className="mt-3 grid grid-cols-2 gap-2 border-t border-white/5 pt-3">
            <Button
              variant="ghost"
              size="sm"
              className="w-full min-w-0 px-2"
              onClick={() => setShowEditModal(true)}
            >
              <span className="block truncate">
                <IconLabel name="pen" iconClassName="h-3 w-3">
                  Ред.
                </IconLabel>
              </span>
            </Button>
            <Button
              variant="danger"
              size="sm"
              className="w-full min-w-0 px-2"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? '...' : 'Удалить'}
            </Button>
          </div>
        )}
      </Card>

      {showEditModal && isAuthor && (
        <InventoryEditModal item={item} onClose={() => setShowEditModal(false)} />
      )}
    </>
  );
};
