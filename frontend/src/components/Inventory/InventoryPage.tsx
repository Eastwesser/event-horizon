import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useInventory } from '../../hooks/useInventory';
import { useUserRole } from '../../hooks/useUserRole';
import { InventoryList } from './InventoryList';
import { InventoryCreateModal } from './InventoryCreateModal';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Button } from '../ui/Button';
import { FilterChip } from '../ui/FilterChip';

const TYPE_FILTERS = [
  { value: '', label: 'Все типы' },
  { value: 'брелок', label: 'Брелок' },
  { value: 'картина', label: 'Картина' },
  { value: 'фенечка', label: 'Фенечка' },
];

export const InventoryPage: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthor } = useUserRole();
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

  const handleSearch = (overrideType?: string) => {
    const type = overrideType !== undefined ? overrideType : filters.type;
    const params: Record<string, string | number> = { limit: 20 };
    if (type) params.type = type;
    if (filters.query) params.query = filters.query;
    if (filters.priceMin) params.price_min = parseFloat(filters.priceMin);
    if (filters.priceMax) params.price_max = parseFloat(filters.priceMax);
    fetchItems(params);
  };

  const handleTypeFilter = (type: string) => {
    setFilters((prev) => ({ ...prev, type }));
    handleSearch(type);
  };

  const handleClearFilters = () => {
    setFilters({ type: '', query: '', priceMin: '', priceMax: '' });
    fetchItems({ limit: 20 });
  };

  const inputClass =
    'rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:border-photon-cyan/50';

  return (
    <PageShell width="wide">
      <PageHeader
        title="Каталог товаров"
        onBack={() => navigate('/')}
        backLabel="На главную"
        actions={
          isAuthor ? (
            <Button size="sm" onClick={() => setShowCreateModal(true)}>
              + Создать товар
            </Button>
          ) : undefined
        }
      />

      <div className="mb-4 flex flex-wrap gap-2">
        {TYPE_FILTERS.map((f) => (
          <FilterChip
            key={f.value || 'all'}
            active={filters.type === f.value}
            onClick={() => handleTypeFilter(f.value)}
          >
            {f.label}
          </FilterChip>
        ))}
      </div>

      <div className="mb-4 flex flex-wrap gap-2">
        <input
          type="text"
          placeholder="Поиск по названию..."
          value={filters.query}
          onChange={(e) => setFilters({ ...filters, query: e.target.value })}
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          className={`${inputClass} min-w-[180px] flex-1`}
        />
        <input
          type="number"
          placeholder="Цена от"
          value={filters.priceMin}
          onChange={(e) => setFilters({ ...filters, priceMin: e.target.value })}
          className={`${inputClass} w-28`}
        />
        <input
          type="number"
          placeholder="Цена до"
          value={filters.priceMax}
          onChange={(e) => setFilters({ ...filters, priceMax: e.target.value })}
          className={`${inputClass} w-28`}
        />
        <Button variant="secondary" size="sm" onClick={() => handleSearch()}>
          Найти
        </Button>
        <Button variant="ghost" size="sm" onClick={handleClearFilters}>
          Сбросить
        </Button>
      </div>

      <p className="mb-6 text-sm text-text-secondary">
        Найдено: <strong className="text-text-primary">{total}</strong> товаров
      </p>

      {loading ? (
        <LoadingSpinner />
      ) : (
        <InventoryList
          items={items ?? []}
          filtered={Boolean(filters.type || filters.query || filters.priceMin || filters.priceMax)}
        />
      )}

      {showCreateModal && isAuthor && (
        <InventoryCreateModal onClose={() => setShowCreateModal(false)} />
      )}
    </PageShell>
  );
};
