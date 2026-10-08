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
import { inventoryApi, type InventoryItem } from '../../services/inventoryApi';
import { formatTicketPrice } from '../../lib/formatPrice';
import { itemFallbackIcon } from '../../lib/itemIcons';
import { stockLabel } from '../../lib/shopItemMap';

interface InventoryItemCardProps {
  item: InventoryItem;
  /** Author dashboard: soft-delete + restore instead of hard delete. */
  softManage?: boolean;
  onChanged?: () => void;
}

export const InventoryItemCard: React.FC<InventoryItemCardProps> = ({
  item,
  softManage = false,
  onChanged,
}) => {
  const { isAuthor } = useUserRole();
  const { deleteItem } = useInventory();
  const [showEditModal, setShowEditModal] = useState(false);
  const [busy, setBusy] = useState(false);
  const deleted = Boolean(item.deleted || item.attributes?._deleted);

  const handleDelete = async () => {
    if (softManage) {
      if (!window.confirm(`Удалить «${item.name}» из витрины? (soft delete)`)) return;
      setBusy(true);
      try {
        await inventoryApi.softDeleteItem(item.id);
        onChanged?.();
      } finally {
        setBusy(false);
      }
      return;
    }
    if (window.confirm(`Удалить товар "${item.name}"?`)) {
      setBusy(true);
      try {
        await deleteItem(item.id);
        onChanged?.();
      } finally {
        setBusy(false);
      }
    }
  };

  const handleRestore = async () => {
    if (!window.confirm(`Восстановить «${item.name}»?`)) return;
    setBusy(true);
    try {
      await inventoryApi.restoreItem(item.id);
      onChanged?.();
    } finally {
      setBusy(false);
    }
  };

  const stockText = stockLabel(item.stock);

  return (
    <>
      <Card className={`flex h-full flex-col ${deleted ? 'opacity-70' : ''}`}>
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
              {deleted && <Badge tone="error">Удалено</Badge>}
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
            {!deleted && (
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
            )}
            {softManage && deleted ? (
              <Button
                variant="secondary"
                size="sm"
                className="col-span-2 w-full min-w-0 px-2"
                onClick={() => void handleRestore()}
                disabled={busy}
              >
                {busy ? '...' : 'Восстановить'}
              </Button>
            ) : (
              <Button
                variant="danger"
                size="sm"
                className={`w-full min-w-0 px-2 ${deleted ? 'col-span-2' : ''}`}
                onClick={() => void handleDelete()}
                disabled={busy}
              >
                {busy ? '...' : 'Удалить'}
              </Button>
            )}
          </div>
        )}
      </Card>

      {showEditModal && isAuthor && !deleted && (
        <InventoryEditModal
          item={item}
          onClose={() => {
            setShowEditModal(false);
            onChanged?.();
          }}
        />
      )}
    </>
  );
};
