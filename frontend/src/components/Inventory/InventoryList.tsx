import React from 'react';
import { InventoryItemCard } from './InventoryItemCard';
import type { InventoryItem } from '../../services/inventoryApi';

interface InventoryListProps {
  items: InventoryItem[] | null | undefined;
  /** When a type/search filter is active, use a category-empty message. */
  filtered?: boolean;
}

export const InventoryList: React.FC<InventoryListProps> = ({ items: rawItems, filtered = false }) => {
  const items = rawItems ?? [];

  if (items.length === 0) {
    return (
      <div className="py-16 text-center">
        <p className="text-text-secondary">
          {filtered ? 'В категории пусто' : 'Товаров пока нет'}
        </p>
        <p className="mt-1 text-sm text-text-muted">
          {filtered
            ? 'Сбросьте фильтр или выберите другой тип'
            : 'Создайте первый товар, нажав кнопку выше'}
        </p>
      </div>
    );
 }

  return (
    <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {items.map((item) => (
        <InventoryItemCard key={item.id} item={item} />
      ))}
    </div>
  );
};
