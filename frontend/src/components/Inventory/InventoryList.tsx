import React from 'react';
import { InventoryItemCard } from './InventoryItemCard';
import type { InventoryItem } from '../../services/inventoryApi';

interface InventoryListProps {
  items: InventoryItem[] | null | undefined;
}

export const InventoryList: React.FC<InventoryListProps> = ({ items: rawItems }) => {
  const items = rawItems ?? [];

  if (items.length === 0) {
    return (
      <div className="py-16 text-center">
        <p className="text-text-secondary">Товаров пока нет</p>
        <p className="mt-1 text-sm text-text-muted">Создайте первый товар, нажав кнопку выше</p>
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
