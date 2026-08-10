import React, { useState } from 'react';
import { useInventory } from '../../hooks/useInventory';
import './styles/InventoryModal.css';
import type { InventoryItem } from '../../services/inventoryApi';

interface InventoryEditModalProps {
  item: InventoryItem;
  onClose: () => void;
}

export const InventoryEditModal: React.FC<InventoryEditModalProps> = ({ item, onClose }) => {
  const { updateItem } = useInventory();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    name: item.name,
    description: item.description || '',
    price: item.price,
    stock: item.stock,
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await updateItem(item.id, formData);
      onClose();
    } catch (error) {
      console.error('Failed to update item:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="inventory-modal-overlay" onClick={onClose}>
      <div className="inventory-modal" onClick={(e) => e.stopPropagation()}>
        <div className="inventory-modal-header">
          <h2>Редактировать товар</h2>
          <button className="btn-close" onClick={onClose}>×</button>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="inventory-modal-body">
            <div className="form-group">
              <label>Название</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>
            <div className="form-group">
              <label>Описание</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />
            </div>
            <div className="form-group">
              <label>Цена (₽)</label>
              <input
                type="number"
                step="0.01"
                value={formData.price}
                onChange={(e) => setFormData({ ...formData, price: parseFloat(e.target.value) || 0 })}
                required
              />
            </div>
            <div className="form-group">
              <label>Количество</label>
              <input
                type="number"
                value={formData.stock}
                onChange={(e) => setFormData({ ...formData, stock: parseInt(e.target.value) || 0 })}
              />
            </div>
          </div>
          <div className="inventory-modal-footer">
            <button type="button" className="btn btn-outline" onClick={onClose}>
              Отмена
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? 'Сохранение...' : 'Сохранить'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};