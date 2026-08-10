import React, { useState } from 'react';
import { useInventory } from '../../hooks/useInventory';
import { InventoryEditModal } from './InventoryEditModal';
import './styles/InventoryItemCard.css';
import type { InventoryItem } from '../../services/inventoryApi';

interface InventoryItemCardProps {
  item: InventoryItem;
}

export const InventoryItemCard: React.FC<InventoryItemCardProps> = ({ item }) => {
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

  return (
    <>
      <div className="inventory-card">
        <div className="inventory-card-image">
          {item.images && item.images.length > 0 ? (
            <img src={item.images[0]} alt={item.name} />
          ) : (
            <div className="inventory-card-image-placeholder">📦</div>
          )}
        </div>
        <div className="inventory-card-body">
          <h3>{item.name}</h3>
          <p className="inventory-card-description">{item.description}</p>
          <div className="inventory-card-meta">
            <span className="inventory-card-type">{item.type}</span>
            <span className="inventory-card-price">{item.price} ₽</span>
            <span className="inventory-card-stock">В наличии: {item.stock}</span>
          </div>
          {item.attributes && Object.keys(item.attributes).length > 0 && (
            <div className="inventory-card-attributes">
              {Object.entries(item.attributes).map(([key, value]) => (
                <span key={key} className="inventory-card-attribute">
                  {key}: {String(value)}
                </span>
              ))}
            </div>
          )}
          <div className="inventory-card-actions">
            <button
              className="btn btn-sm btn-outline"
              onClick={() => setShowEditModal(true)}
            >
              ✏️ Редактировать
            </button>
            <button
              className="btn btn-sm btn-danger"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? '...' : '🗑️ Удалить'}
            </button>
          </div>
        </div>
      </div>

      {showEditModal && (
        <InventoryEditModal
          item={item}
          onClose={() => setShowEditModal(false)}
        />
      )}
    </>
  );
};