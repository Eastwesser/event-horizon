import React, { useEffect, useState } from 'react';
import { useInventory } from '../../hooks/useInventory';
import { InventoryList } from './InventoryList';
import { InventoryCreateModal } from './InventoryCreateModal';
import './styles/Inventory.css'; 
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';

export const InventoryPage: React.FC = () => {
  const { items, total, loading, fetchItems } = useInventory();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [filters, setFilters] = useState({
    type: '',
    query: '',
    priceMin: '',
    priceMax: '',
  });

  useEffect(() => {
    fetchItems({ limit: 20 });
  }, []);

  const handleSearch = () => {
    const params: any = { limit: 20 };
    if (filters.type) params.type = filters.type;
    if (filters.query) params.query = filters.query;
    if (filters.priceMin) params.price_min = parseFloat(filters.priceMin);
    if (filters.priceMax) params.price_max = parseFloat(filters.priceMax);
    fetchItems(params);
  };

  const handleClearFilters = () => {
    setFilters({ type: '', query: '', priceMin: '', priceMax: '' });
    fetchItems({ limit: 20 });
  };

  return (
    <div className="inventory-page">
      <div className="inventory-header">
        <h1>Каталог товаров</h1>
        <button
          className="btn btn-primary"
          onClick={() => setShowCreateModal(true)}
        >
          + Создать товар
        </button>
      </div>

      <div className="inventory-filters">
        <input
          type="text"
          placeholder="Поиск по названию..."
          value={filters.query}
          onChange={(e) => setFilters({ ...filters, query: e.target.value })}
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
        />
        <select
          value={filters.type}
          onChange={(e) => setFilters({ ...filters, type: e.target.value })}
        >
          <option value="">Все типы</option>
          <option value="брелок">Брелок</option>
          <option value="картина">Картина</option>
          <option value="фенечка">Фенечка</option>
        </select>
        <input
          type="number"
          placeholder="Цена от"
          value={filters.priceMin}
          onChange={(e) => setFilters({ ...filters, priceMin: e.target.value })}
        />
        <input
          type="number"
          placeholder="Цена до"
          value={filters.priceMax}
          onChange={(e) => setFilters({ ...filters, priceMax: e.target.value })}
        />
        <button className="btn btn-secondary" onClick={handleSearch}>
          Найти
        </button>
        <button className="btn btn-outline" onClick={handleClearFilters}>
          Сбросить
        </button>
      </div>

      <div className="inventory-stats">
        Найдено: <strong>{total}</strong> товаров
      </div>

      {loading ? (
        <LoadingSpinner />
      ) : (
        <InventoryList items={items} />
      )}

      {showCreateModal && (
        <InventoryCreateModal onClose={() => setShowCreateModal(false)} />
      )}
    </div>
  );
};
