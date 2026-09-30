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

const CARD_ELEMENTS = ['лес', 'горы', 'степи', 'тьма', 'болото', 'нейтрал'] as const;
const CARD_RARITIES = [
  { value: 'common', label: 'Обычная' },
  { value: 'uncommon', label: 'Необычная' },
  { value: 'rare', label: 'Редкая' },
  { value: 'ultra', label: 'Ультра' },
] as const;

/** Rough EH tickets from ~₽ market × 2000 (10₽ → 20_000 tickets). Foil ×2. */
function suggestTickets(rubles: number, foil: boolean): number {
  const base = Math.max(0, Math.round(rubles * 2000));
  return foil ? base * 2 : base;
}

export const InventoryCreateModal: React.FC<InventoryCreateModalProps> = ({ onClose }) => {
  const { createItem } = useInventory();
  const [loading, setLoading] = useState(false);
  const [imageUrl, setImageUrl] = useState('');
  const [formData, setFormData] = useState({
    type: 'брелок',
    name: '',
    description: '',
    price: 0,
    stock: 1,
  });
  const [card, setCard] = useState({
    element: 'лес',
    rarity: 'common',
    set: '7',
    artist: '',
    year: '2006',
    market_rub: '10',
    foil: false,
    card_no: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const trimmed = imageUrl.trim();
      const attributes: Record<string, unknown> =
        formData.type === 'карточка'
          ? {
              element: card.element,
              rarity: card.rarity,
              set: Number(card.set) || 0,
              artist: card.artist.trim(),
              year: Number(card.year) || 0,
              foil: card.foil,
              card_no: card.card_no.trim(),
              market_rub: Number(card.market_rub) || 0,
            }
          : {};
      const price =
        formData.type === 'карточка'
          ? suggestTickets(Number(card.market_rub) || 0, card.foil)
          : formData.price;
      await createItem({
        ...formData,
        price,
        attributes,
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
      <form onSubmit={handleSubmit} className="flex max-h-[80vh] flex-col gap-5 overflow-y-auto">
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
            <option value="карточка">Карточка (ККИ)</option>
          </select>
        </div>
        <div>
          <label className={labelClass}>Название</label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
            placeholder={formData.type === 'карточка' ? 'Слепыш' : undefined}
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

        {formData.type === 'карточка' ? (
          <div className="space-y-3 rounded-sm border border-white/10 bg-void/40 p-3">
            <p className="text-xs text-text-muted">
              Берсерк / ККИ: цена в билетиках ≈ ₽×2000 (фойл ×2). Пример: 10₽ → 20 000.
            </p>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className={labelClass}>Стихия</label>
                <select
                  className={inputClass}
                  value={card.element}
                  onChange={(e) => setCard({ ...card, element: e.target.value })}
                >
                  {CARD_ELEMENTS.map((el) => (
                    <option key={el} value={el}>
                      {el}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className={labelClass}>Редкость</label>
                <select
                  className={inputClass}
                  value={card.rarity}
                  onChange={(e) => setCard({ ...card, rarity: e.target.value })}
                >
                  {CARD_RARITIES.map((r) => (
                    <option key={r.value} value={r.value}>
                      {r.label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className={labelClass}>Сет</label>
                <input
                  className={inputClass}
                  value={card.set}
                  onChange={(e) => setCard({ ...card, set: e.target.value })}
                  placeholder="7"
                />
              </div>
              <div>
                <label className={labelClass}>Год</label>
                <input
                  className={inputClass}
                  value={card.year}
                  onChange={(e) => setCard({ ...card, year: e.target.value })}
                />
              </div>
              <div>
                <label className={labelClass}>Автор арта</label>
                <input
                  className={inputClass}
                  value={card.artist}
                  onChange={(e) => setCard({ ...card, artist: e.target.value })}
                  placeholder="ark_cj"
                />
              </div>
              <div>
                <label className={labelClass}>№ в серии</label>
                <input
                  className={inputClass}
                  value={card.card_no}
                  onChange={(e) => setCard({ ...card, card_no: e.target.value })}
                />
              </div>
              <div>
                <label className={labelClass}>Рынок ≈ ₽</label>
                <input
                  type="number"
                  className={inputClass}
                  value={card.market_rub}
                  onChange={(e) => setCard({ ...card, market_rub: e.target.value })}
                />
              </div>
              <div className="flex items-end pb-2">
                <label className="flex items-center gap-2 text-sm text-text-secondary">
                  <input
                    type="checkbox"
                    checked={card.foil}
                    onChange={(e) => setCard({ ...card, foil: e.target.checked })}
                  />
                  Фойл (×2 цена)
                </label>
              </div>
            </div>
            <p className="text-sm text-horizon-gold">
              Цена в магазине: {suggestTickets(Number(card.market_rub) || 0, card.foil)} билетиков
            </p>
          </div>
        ) : (
          <div>
            <label className={labelClass}>Цена (₽ / билетики — как принято)</label>
            <input
              type="number"
              step="0.01"
              value={formData.price}
              onChange={(e) => setFormData({ ...formData, price: parseFloat(e.target.value) || 0 })}
              required
              className={inputClass}
            />
          </div>
        )}

        <div>
          <label className={labelClass}>Количество (поштучно)</label>
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
