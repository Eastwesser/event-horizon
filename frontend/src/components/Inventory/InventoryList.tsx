import React from 'react';
import { InventoryItemCard } from './InventoryItemCard';
import './styles/Inventory.css';
import type { InventoryItem } from '../../services/inventoryApi';

interface InventoryListProps {
  items: InventoryItem[];
}

export const InventoryList: React.FC<InventoryListProps> = ({ items }) => {
  if (items.length === 0) {
    return (
      <div className="inventory-empty">
        <p>Товаров пока нет</p>
        <p className="text-muted">Создайте первый товар, нажав кнопку выше</p>
      </div>
    );
  }

  return (
    <div className="inventory-grid">
      {items.map((item) => (
        <InventoryItemCard key={item.id} item={item} />
      ))}
    </div>
  );
};
