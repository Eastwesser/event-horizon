import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { inventoryApi, type InventoryItem } from '../../services/inventoryApi';
import { useUserRole } from '../../hooks/useUserRole';
import { InventoryList } from './InventoryList';
import { InventoryCreateModal } from './InventoryCreateModal';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Button } from '../ui/Button';
import { FilterChip } from '../ui/FilterChip';
import { CatalogPager } from '../ui/CatalogPager';
import {
  CATALOG_PAGE_SIZE,
  CATALOG_SORT_OPTIONS,
  type CatalogSort,
  filterCatalogItems,
  paginateItems,
  parseCatalogPage,
  parseCatalogSort,
  parseFlag,
  sortCatalogItems,
} from '../../lib/catalogQuery';
import { saveCatalogNav } from '../../lib/catalogNav';

const TYPE_FILTERS = [
  { value: '', label: 'Все типы' },
  { value: 'брелок', label: 'Брелок' },
  { value: 'картина', label: 'Картина' },
  { value: 'фенечка', label: 'Фенечка' },
  { value: 'карточка', label: 'Карточка' },
];

export const InventoryPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { isAuthor } = useUserRole();
  const [allItems, setAllItems] = useState<InventoryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);

  const type = searchParams.get('type') || '';
  const q = searchParams.get('q') || '';
  const priceMin = searchParams.get('price_min') || '';
  const priceMax = searchParams.get('price_max') || '';
  const sort = parseCatalogSort(searchParams.get('sort'));
  const page = parseCatalogPage(searchParams.get('page'));
  const inStock = parseFlag(searchParams.get('in_stock'));
  const foil = parseFlag(searchParams.get('foil'));
  const noir = parseFlag(searchParams.get('noir'));
  const flying = parseFlag(searchParams.get('flying'));

  const [queryDraft, setQueryDraft] = useState(q);
  const [priceMinDraft, setPriceMinDraft] = useState(priceMin);
  const [priceMaxDraft, setPriceMaxDraft] = useState(priceMax);

  useEffect(() => {
    setQueryDraft(q);
  }, [q]);

  const setQuery = (patch: Record<string, string | null>, replace = false) => {
    const next = new URLSearchParams(searchParams);
    for (const [k, v] of Object.entries(patch)) {
      if (v === null || v === '') next.delete(k);
      else if (k === 'page' && v === '1') next.delete('page');
      else if (k === 'sort' && v === 'newest') next.delete('sort');
      else next.set(k, v);
    }
    setSearchParams(next, { replace });
  };

  useEffect(() => {
    const t = window.setTimeout(() => {
      const next = queryDraft.trim();
      if (next === q) return;
      setQuery({ q: next || null, page: '1' }, true);
    }, 300);
    return () => window.clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [queryDraft]);

  const reload = async () => {
    setLoading(true);
    try {
      const res = await inventoryApi.searchAllItems({
        type: type || undefined,
        price_min: priceMin ? Number(priceMin) : undefined,
        price_max: priceMax ? Number(priceMax) : undefined,
      });
      setAllItems(res.items ?? []);
    } catch {
      setAllItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void reload();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [type, priceMin, priceMax]);

  const sorted = useMemo(() => {
    const mapped = allItems.map((it) => ({
      ...it,
      price_tickets: it.price,
    }));
    const filtered = filterCatalogItems(mapped, {
      query: q,
      inStock,
      foil,
      noir,
      flying,
    });
    return sortCatalogItems(filtered, sort);
  }, [allItems, sort, q, inStock, foil, noir, flying]);

  const { page: safePage, pageCount, slice } = useMemo(
    () => paginateItems(sorted, page, CATALOG_PAGE_SIZE),
    [sorted, page]
  );

  useEffect(() => {
    if (safePage !== page) setQuery({ page: String(safePage) }, true);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [safePage, page]);

  useEffect(() => {
    const qs = searchParams.toString();
    saveCatalogNav({
      ids: sorted.map((it) => it.id),
      listPath: qs ? `/inventory?${qs}` : '/inventory',
    });
  }, [sorted, searchParams]);

  const toggleFlag = (key: string, on: boolean) => {
    setQuery({ [key]: on ? null : '1', page: '1' });
  };

  const handlePriceApply = () => {
    setQuery({
      price_min: priceMinDraft.trim() || null,
      price_max: priceMaxDraft.trim() || null,
      page: '1',
    });
  };

  const handleClear = () => {
    setQueryDraft('');
    setPriceMinDraft('');
    setPriceMaxDraft('');
    setSearchParams(new URLSearchParams(), { replace: false });
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

      <div className="mb-3 flex flex-wrap items-center gap-2">
        {TYPE_FILTERS.map((f) => (
          <FilterChip
            key={f.value || 'all'}
            active={type === f.value}
            onClick={() => setQuery({ type: f.value || null, page: '1' })}
          >
            {f.label}
          </FilterChip>
        ))}
        <label className="ml-auto flex items-center gap-2 text-sm text-text-secondary">
          <span className="text-text-muted">Сортировка</span>
          <select
            className="rounded-sm border border-white/10 bg-nebula px-2 py-1.5 text-sm text-text-primary"
            value={sort}
            onChange={(e) =>
              setQuery({ sort: e.target.value as CatalogSort, page: '1' })
            }
          >
            {CATALOG_SORT_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      <div className="mb-3 flex flex-wrap gap-2">
        <FilterChip active={inStock} onClick={() => toggleFlag('in_stock', inStock)}>
          В наличии
        </FilterChip>
        <FilterChip active={foil} onClick={() => toggleFlag('foil', foil)}>
          Фойл
        </FilterChip>
        <FilterChip active={noir} onClick={() => toggleFlag('noir', noir)}>
          Нуар
        </FilterChip>
        <FilterChip active={flying} onClick={() => toggleFlag('flying', flying)}>
          Летающие
        </FilterChip>
      </div>

      <div className="mb-4 flex flex-wrap gap-2">
        <input
          type="search"
          placeholder="Поиск по названию..."
          value={queryDraft}
          onChange={(e) => setQueryDraft(e.target.value)}
          className={`${inputClass} min-w-[180px] flex-1`}
        />
        <input
          type="number"
          placeholder="Цена от"
          value={priceMinDraft}
          onChange={(e) => setPriceMinDraft(e.target.value)}
          className={`${inputClass} w-28`}
        />
        <input
          type="number"
          placeholder="Цена до"
          value={priceMaxDraft}
          onChange={(e) => setPriceMaxDraft(e.target.value)}
          className={`${inputClass} w-28`}
        />
        <Button variant="secondary" size="sm" onClick={handlePriceApply}>
          Цены
        </Button>
        <Button variant="ghost" size="sm" onClick={handleClear}>
          Сбросить
        </Button>
      </div>

      <p className="mb-6 text-sm text-text-secondary">
        Найдено: <strong className="text-text-primary">{sorted.length}</strong> товаров
      </p>

      {loading ? (
        <LoadingSpinner />
      ) : (
        <InventoryList
          items={slice}
          filtered={Boolean(type || q || priceMin || priceMax || inStock || foil || noir || flying)}
        />
      )}

      <CatalogPager
        page={safePage}
        pageCount={pageCount}
        total={sorted.length}
        disabled={loading}
        onPageChange={(p) => setQuery({ page: String(p) })}
      />

      {showCreateModal && isAuthor && (
        <InventoryCreateModal
          onClose={() => {
            setShowCreateModal(false);
            void reload();
          }}
        />
      )}
    </PageShell>
  );
};
