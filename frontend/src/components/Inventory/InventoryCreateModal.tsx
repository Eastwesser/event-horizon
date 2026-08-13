import React, { useState } from 'react';
import { useInventory } from '../../hooks/useInventory';
import './styles/InventoryModal.css';

interface InventoryCreateModalProps {
  onClose: () => void;
}

export const InventoryCreateModal: React.FC<InventoryCreateModalProps> = ({ onClose }) => {
  const { createItem } = useInventory();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    type: 'брелок',
    name: '',
    description: '',
    price: 0,
    stock: 0,
    attributes: {} as Record<string, any>,
    images: [] as string[],
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await createItem(formData);
      onClose();
    } catch (error) {
      console.error('Failed to create item:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="inventory-modal-overlay" onClick={onClose}>
      <div className="inventory-modal" onClick={(e) => e.stopPropagation()}>
        <div className="inventory-modal-header">
          <h2>Создать товар</h2>
          <button className="btn-close" onClick={onClose}>×</button>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="inventory-modal-body">
            <div className="form-group">
              <label>Тип товара</label>
              <select
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value })}
              >
                <option value="брелок">Брелок</option>
                <option value="картина">Картина</option>
                <option value="фенечка">Фенечка</option>
              </select>
            </div>
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
              {loading ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
