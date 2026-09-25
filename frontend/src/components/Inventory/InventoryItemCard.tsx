import React, { useState } from 'react';
import { useInventory } from '../../hooks/useInventory';
import { useUserRole } from '../../hooks/useUserRole';
import { InventoryEditModal } from './InventoryEditModal';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import type { InventoryItem } from '../../services/inventoryApi';
import { formatRubPrice } from '../../lib/formatPrice';

interface InventoryItemCardProps {
  item: InventoryItem;
}

export const InventoryItemCard: React.FC<InventoryItemCardProps> = ({ item }) => {
  const { isAuthor } = useUserRole();
  const { deleteItem } = useInventory();
  const [showEditModal, setShowEditModal] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [imgFailed, setImgFailed] = useState(false);

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

  const hasImage = item.images && item.images.length > 0 && !imgFailed;

  return (
    <>
      <Card className="flex h-full flex-col gap-3">
        <div className="flex h-32 items-center justify-center overflow-hidden rounded-sm bg-white/5">
          {hasImage ? (
            <img
              src={item.images[0]}
              alt={item.name}
              className="h-full w-full object-cover"
              onError={() => setImgFailed(true)}
            />
          ) : (
            <span className="text-4xl">📦</span>
          )}
        </div>
        <div>
          <h3 className="font-display text-base font-semibold text-text-primary">{item.name}</h3>
          <p className="mt-1 text-sm text-text-secondary">{item.description}</p>
        </div>
        <div className="flex flex-wrap items-center gap-2 text-sm">
          <Badge tone="indigo">{item.type}</Badge>
          <span className="font-hud tabular-nums text-horizon-gold">{formatRubPrice(item.price)}</span>
          {item.stock === null ? null : item.stock === undefined || item.stock === 0 ? (
            // proto3 JSON omits stock:0 → undefined; treat as out of stock
            <span className="text-text-muted">Нет в наличии</span>
          ) : (
            <span className="text-text-muted">В наличии: {item.stock}</span>
          )}
        </div>
        {item.attributes && Object.keys(item.attributes).length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {Object.entries(item.attributes).map(([key, value]) => (
              <Badge key={key} tone="neutral">
                {key}: {String(value)}
              </Badge>
            ))}
          </div>
        )}
        {isAuthor && (
          <div className="mt-auto grid grid-cols-2 gap-2 pt-1">
            <Button
              variant="ghost"
              size="sm"
              className="w-full min-w-0 px-2"
              onClick={() => setShowEditModal(true)}
            >
              <span className="block truncate">✏️ Ред.</span>
            </Button>
            <Button
              variant="danger"
              size="sm"
              className="w-full min-w-0 px-2"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? '...' : '🗑️ Удалить'}
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
