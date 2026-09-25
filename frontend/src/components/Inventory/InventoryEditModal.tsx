import React, { useState } from 'react';
import { useInventory } from '../../hooks/useInventory';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import type { InventoryItem } from '../../services/inventoryApi';
import { InventoryImageUrlField } from './InventoryImageUrlField';

interface InventoryEditModalProps {
  item: InventoryItem;
  onClose: () => void;
}

const inputClass =
  'w-full rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary focus:border-photon-cyan/50';
const labelClass = 'mb-1 block text-sm text-text-secondary';

export const InventoryEditModal: React.FC<InventoryEditModalProps> = ({ item, onClose }) => {
  const { updateItem } = useInventory();
  const [loading, setLoading] = useState(false);
  const [imageUrl, setImageUrl] = useState(item.images?.[0] ?? '');
  const [formData, setFormData] = useState({
    name: item.name,
    description: item.description || '',
    price: item.price ?? 0,
    stock: item.stock ?? 0,
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const trimmed = imageUrl.trim();
      await updateItem(item.id, {
        name: formData.name,
        description: formData.description,
        price: Number(formData.price) || 0,
        stock: Number(formData.stock) || 0,
        images: trimmed ? [trimmed] : [],
      });
      onClose();
    } catch (error) {
      console.error('Failed to update item:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal open onClose={onClose} title="Редактировать товар">
      <form onSubmit={handleSubmit} className="flex flex-col gap-5">
        <div>
          <label className={labelClass}>Название</label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
            className={inputClass}
          />
        </div>
        <div>
          <label className={labelClass}>Описание</label>
          <textarea
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            className={`${inputClass} min-h-20`}
          />
        </div>
        <div>
          <label className={labelClass}>Цена (₽)</label>
          <input
            type="number"
            step="0.01"
            value={formData.price}
            onChange={(e) => setFormData({ ...formData, price: parseFloat(e.target.value) || 0 })}
            required
            className={inputClass}
          />
        </div>
        <div>
          <label className={labelClass}>Количество</label>
          <input
            type="number"
            value={formData.stock}
            onChange={(e) => setFormData({ ...formData, stock: parseInt(e.target.value) || 0 })}
            className={inputClass}
          />
        </div>
        <InventoryImageUrlField value={imageUrl} onChange={setImageUrl} />
        <div className="mt-2 flex justify-end gap-3">
          <Button type="button" variant="ghost" onClick={onClose}>
            Отмена
          </Button>
          <Button type="submit" variant="primary" disabled={loading}>
            {loading ? 'Сохранение...' : 'Сохранить'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
