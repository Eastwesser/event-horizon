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

const CARD_ELEMENTS = [
  { value: 'woods', label: 'Лес' },
  { value: 'mountains', label: 'Горы' },
  { value: 'swamps', label: 'Болота' },
  { value: 'steppes', label: 'Степи' },
  { value: 'darkness', label: 'Тьма' },
  { value: 'neutral', label: 'Нейтрал' },
] as const;

const CARD_RARITIES = [
  { value: 'common', label: 'Обычная' },
  { value: 'uncommon', label: 'Необычная' },
  { value: 'rare', label: 'Редкая' },
  { value: 'ultra', label: 'Ультра' },
] as const;

const AVAILABLE_ICONS = [
  { type: 'regen', label: 'Регенерация' },
  { type: 'armor', label: 'Доспех' },
  { type: 'counter', label: 'Фишка' },
  { type: 'strike', label: 'Атака' },
  { type: 'tap', label: 'Поворот' },
  { type: 'instant', label: 'Мгновенно' },
  { type: 'direct', label: 'Направленный' },
  { type: 'stamina', label: 'Стойкость' },
] as const;

function suggestTickets(rubles: number, foil: boolean): number {
  const base = Math.max(0, Math.round(rubles * 1000));
  return foil ? base * 2 : base;
}

export const InventoryEditModal: React.FC<InventoryEditModalProps> = ({ item, onClose }) => {
  const { updateItem } = useInventory();
  const [loading, setLoading] = useState(false);
  const [imageUrl, setImageUrl] = useState(item.images?.[0] ?? '');
  
  const attrs = item.attributes || {};
  
  const [formData, setFormData] = useState({
    name: item.name,
    description: item.description || '',
    price: item.price ?? 0,
    stock: item.stock ?? 0,
  });

  const initialIcons = Array.isArray(attrs.icons) ? attrs.icons : [];
  const initialIconValues = initialIcons.reduce((acc: any, i: any) => {
    acc[i.type] = i.value;
    return acc;
  }, {});

  const [card, setCard] = useState({
    element: (attrs.element as string) || 'woods',
    rarity: (attrs.rarity as string) || 'common',
    set: String(attrs.set || '7'),
    artist: (attrs.artist as string) || '',
    year: String(attrs.year || '2026'),
    // market_rub: String(attrs.market_rub || '10'), <- ALWAYS 10, IRRITATING
    market_rub: (attrs.market_rub !== undefined && attrs.market_rub !== null && attrs.market_rub !== '') ? String(attrs.market_rub) : '10',
    // market_rub: attrs.market_rub !== undefined && attrs.market_rub !== null ? String(attrs.market_rub) : '10',
    foil: Boolean(attrs.foil),
    card_no: (attrs.card_no as string) || '',
    health: String(attrs.health || ''),
    strike: (attrs.strike as string) || '',
    card_text: (attrs.card_text as string) || '',
    flavor_text: (attrs.flavor_text as string) || '',
    type_main: (attrs.type_main as string) || 'creature',
    type_sub: (attrs.type_sub as string) || '',
    cost_tier: (attrs.cost_tier as string) || 'rank_and_file',
    cost: String(attrs.cost || ''),
    move: String(attrs.move || ''),
  });

  const [icons, setIcons] = useState<Record<string, number | string>>(initialIconValues);

  const handleIconChange = (type: string, value: string) => {
    setIcons((prev) => ({ ...prev, [type]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      const trimmed = imageUrl.trim();
      
      const iconsArray = Object.entries(icons)
        .filter(([_, val]) => val !== '' && val !== undefined)
        .map(([type, val]) => ({
          type,
          value: Math.max(0, Number(val) || 0),
        }));

      const rub = Math.max(1, Number(card.market_rub) || 1);
      
      const isCard =
        item.type === 'карточка' ||
        attrs.market_rub !== undefined ||
        attrs.element !== undefined ||
        attrs.card_text !== undefined;

      const attributes: Record<string, unknown> = isCard
        ? {
            ...attrs,
            element: card.element,
            rarity: card.rarity,
            set: Math.max(0, Number(card.set) || 0),
            artist: card.artist.trim(),
            year: Math.max(0, Number(card.year) || 0),
            foil: card.foil,
            card_no: card.card_no.trim(),
            market_rub: rub,
            health: Math.max(0, Number(card.health) || 0),
            strike: card.strike.trim(),
            card_text: card.card_text.trim(),
            flavor_text: card.flavor_text.trim(),
            type_main: card.type_main,
            type_sub: card.type_sub.trim(),
            cost_tier: card.cost_tier,
            icons: iconsArray,
            cost: Math.max(0, Number(card.cost) || 0),
            move: Math.max(0, Number(card.move) || 0),
            class: card.type_sub ? [card.type_sub.trim()] : [],
          }
        : attrs;

      // ЕДИНСТВЕННОЕ ОБЪЯВЛЕНИЕ PRICE
      const computedTickets = suggestTickets(rub, card.foil);
      const price = isCard ? computedTickets : Math.max(0, formData.price);

      await updateItem(item.id, {
        name: formData.name,
        description: formData.description,
        price,
        stock: Math.max(0, Number(formData.stock) || 0),
        attributes,
        images: trimmed ? [trimmed] : [],
        version: item.version,
      });

      onClose();
    } catch (error) {
      console.error('Failed to update item:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal open onClose={onClose} title="Редактировать товар" className="max-w-3xl">
      {/* <form onSubmit={handleSubmit} className="flex max-h-[80vh] flex-col gap-5 overflow-y-auto"> */}
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
          <label className={labelClass}>Описание (краткое)</label>
          <textarea
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            className={`${inputClass} min-h-20`}
          />
        </div>

        {/* {item.type === 'карточка' && ( */}
        {(item.type === 'карточка' || attrs.market_rub !== undefined || attrs.element !== undefined) && (
          <div className="space-y-3 rounded-sm border border-white/10 bg-void/40 p-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className={labelClass}>Стихия</label>
                <select
                  className={inputClass}
                  value={card.element}
                  onChange={(e) => setCard({ ...card, element: e.target.value })}
                >
                  {CARD_ELEMENTS.map((el) => (
                    <option key={el.value} value={el.value}>{el.label}</option>
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
                    <option key={r.value} value={r.value}>{r.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className={labelClass}>Сет</label>
                <input
                  type="number"
                  min="0"
                  className={inputClass}
                  value={card.set}
                  onChange={(e) => setCard({ ...card, set: e.target.value })}
                />
              </div>
              <div>
                <label className={labelClass}>Год</label>
                <input
                  type="number"
                  min="0"
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
                
                {/* <input
                  type="number"
                  min="1"
                  className={inputClass}
                  value={card.market_rub}
                  onChange={(e) => setCard({ ...card, market_rub: e.target.value })}
                /> */}

                <input
                  type="number"
                  min="1"
                  step="1"
                  className={inputClass}
                  value={card.market_rub}
                  onChange={(e) => {
                    const val = e.target.value;
                    if (val === '') {
                      setCard({ ...card, market_rub: '1' }); // Не даем оставить пустым
                    } else {
                      setCard({ ...card, market_rub: val });
                    }
                  }}
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

            <div className="grid grid-cols-2 gap-3 border-t border-white/10 pt-3 mt-3">
              <div>
                <label className={labelClass}>Тип (Карта)</label>
                <select
                  className={inputClass}
                  value={card.type_main}
                  onChange={(e) => setCard({ ...card, type_main: e.target.value })}
                >
                  <option value="creature">Существо</option>
                  <option value="artifact">Артефакт</option>
                  <option value="land">Местность</option>
                </select>
              </div>
              <div>
                <label className={labelClass}>Класс (Карта)</label>
                <input
                  className={inputClass}
                  value={card.type_sub}
                  onChange={(e) => setCard({ ...card, type_sub: e.target.value })}
                  placeholder="Тролль, Постройка..."
                />
              </div>
              <div>
                <label className={labelClass}>Стоимость (Ряд)</label>
                <select
                  className={inputClass}
                  value={card.cost_tier}
                  onChange={(e) => setCard({ ...card, cost_tier: e.target.value })}
                >
                  <option value="rank_and_file">Рядовая</option>
                  <option value="elite">Элитная</option>
                </select>
              </div>
              <div>
                <label className={labelClass}>Стоимость (в углу)</label>
                <input
                  type="number"
                  min="0"
                  className={inputClass}
                  value={card.cost}
                  onChange={(e) => setCard({ ...card, cost: e.target.value })}
                  placeholder="3"
                />
              </div>
              <div>
                <label className={labelClass}>Ход</label>
                <input
                  type="number"
                  min="0"
                  className={inputClass}
                  value={card.move}
                  onChange={(e) => setCard({ ...card, move: e.target.value })}
                  placeholder="1"
                />
              </div>
              <div>
                <label className={labelClass}>Здоровье (Бой)</label>
                <input
                  type="number"
                  min="0"
                  className={inputClass}
                  value={card.health}
                  onChange={(e) => setCard({ ...card, health: e.target.value })}
                />
              </div>
              <div>
                <label className={labelClass}>Удар (Бой)</label>
                <input
                  className={inputClass}
                  value={card.strike}
                  onChange={(e) => setCard({ ...card, strike: e.target.value })}
                />
              </div>
            </div>

            {/* ИКОНКИ В СТОЛБИК */}
            <div className="border-t border-white/10 pt-3 mt-3">
              <label className={labelClass}>Иконки (свойства)</label>
              <div className="flex flex-col gap-2">
                {AVAILABLE_ICONS.map((icon) => (
                  <div key={icon.type} className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={icons[icon.type] !== undefined && icons[icon.type] !== ''}
                      onChange={(e) => handleIconChange(icon.type, e.target.checked ? '1' : '')}
                    />
                    <span className="text-sm text-text-secondary">{icon.label}</span>
                    {icons[icon.type] !== undefined && icons[icon.type] !== '' && (
                      <input
                        type="number"
                        min="0"
                        className="w-14 rounded-sm border border-white/10 bg-nebula px-1 py-0.5 text-xs"
                        value={icons[icon.type] || ''}
                        onChange={(e) => handleIconChange(icon.type, e.target.value)}
                        placeholder="0"
                      />
                    )}
                  </div>
                ))}
              </div>
            </div>

            <div>
              <label className={labelClass}>Текст карты (Правила)</label>
              <textarea
                className={`${inputClass} min-h-20`}
                value={card.card_text}
                onChange={(e) => setCard({ ...card, card_text: e.target.value })}
              />
            </div>
            <div>
              <label className={labelClass}>Художественный текст (Flavor)</label>
              <textarea
                className={`${inputClass} min-h-10`}
                value={card.flavor_text}
                onChange={(e) => setCard({ ...card, flavor_text: e.target.value })}
              />
            </div>
            <p className="text-sm text-horizon-gold">
              Цена в магазине: {suggestTickets(Number(card.market_rub) || 0, card.foil)} билетиков
            </p>
          </div>
        )}

        {/* {item.type !== 'карточка' && ( */}
        {(item.type !== 'карточка' && attrs.market_rub === undefined && attrs.element === undefined) && (
          <div>
            <label className={labelClass}>Цена (билетики)</label>
            <input
              type="number"
              step="1"
              min="0"
              value={formData.price}
              onChange={(e) => setFormData({ ...formData, price: parseFloat(e.target.value) || 0 })}
              required
              className={inputClass}
            />
          </div>
        )}

        <div>
          <label className={labelClass}>Количество</label>
          <input
            type="number"
            min="0"
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
