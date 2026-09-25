import React, { useState } from 'react';
import { useInventory } from '../../hooks/useInventory';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { InventoryImageUrlField } from './InventoryImageUrlField';

interface InventoryCreateModalProps {
  onClose: () => void;
}

const inputClass =
  'w-full rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary focus:border-photon-cyan/50';
const labelClass = 'mb-1 block text-sm text-text-secondary';

export const InventoryCreateModal: React.FC<InventoryCreateModalProps> = ({ onClose }) => {
  const { createItem } = useInventory();
  const [loading, setLoading] = useState(false);
  const [imageUrl, setImageUrl] = useState('');
  const [formData, setFormData] = useState({
    type: 'брелок',
    name: '',
    description: '',
    price: 0,
    stock: 0,
    attributes: {} as Record<string, any>,
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const trimmed = imageUrl.trim();
      await createItem({
        ...formData,
        images: trimmed ? [trimmed] : [],
      });
      onClose();
    } catch (error) {
      console.error('Failed to create item:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal open onClose={onClose} title="Создать товар">
      <form onSubmit={handleSubmit} className="flex flex-col gap-5">
        <div>
          <label className={labelClass}>Тип товара</label>
          <select
            value={formData.type}
            onChange={(e) => setFormData({ ...formData, type: e.target.value })}
            className={inputClass}
          >
            <option value="брелок">Брелок</option>
            <option value="картина">Картина</option>
            <option value="фенечка">Фенечка</option>
          </select>
        </div>
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
            {loading ? 'Создание...' : 'Создать'}
          </Button>
        </div>
      </form>
    </Modal>
  );
};
