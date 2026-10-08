import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { inventoryApi } from '../../services/inventoryApi';
import { inventoryToShopItem } from '../../lib/shopItemMap';
import { useShopStore, type ShopItem } from '../Shop/shopStore';
import ShopItemCard from '../Shop/ShopItemCard';
import PurchaseModal from '../Shop/PurchaseModal';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Spinner } from '../ui/Spinner';
import { CatalogPager } from '../ui/CatalogPager';
import {
  CATALOG_PAGE_SIZE,
  paginateItems,
  parseCatalogPage,
  sortCatalogItems,
} from '../../lib/catalogQuery';
import { saveCatalogNav } from '../../lib/catalogNav';
import { pluralCards } from '../../lib/pluralize';
import { paymentApi } from '../../services/paymentApi';

export function CardArtistPage() {
  const { artistId = '' } = useParams<{ artistId: string }>();
  const decodedId = decodeURIComponent(artistId);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { balance, inventory, fetchBalance, fetchInventory } = useShopStore();

  const [items, setItems] = useState<ShopItem[]>([]);
  const [displayName, setDisplayName] = useState(decodedId);
  const [loading, setLoading] = useState(true);
  const [selectedItem, setSelectedItem] = useState<ShopItem | null>(null);
  const [showModal, setShowModal] = useState(false);

  const page = parseCatalogPage(searchParams.get('page'));

  useEffect(() => {
    fetchBalance();
    fetchInventory();
  }, [fetchBalance, fetchInventory]);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setLoading(true);
      try {
        const ownedIds = new Set(inventory.map((p) => p.item_id));
        const res = await inventoryApi.searchAllItems({ type: 'карточка' });
        if (cancelled) return;
        const matched = (res.items ?? [])
          .filter((it) => {
            const a = it.attributes || {};
            return a.artist_id === decodedId;
          })
          .map((it) => inventoryToShopItem(it, ownedIds.has(it.id)));
        setItems(sortCatalogItems(matched, 'newest'));
        const first = matched[0]?.attributes;
        const name =
          (typeof first?.artist_display === 'string' && first.artist_display) ||
          (typeof first?.artist === 'string' && first.artist) ||
          decodedId;
        setDisplayName(name);
      } catch {
        if (!cancelled) setItems([]);
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, [decodedId, inventory]);

  const { page: safePage, pageCount, slice } = useMemo(
    () => paginateItems(items, page, CATALOG_PAGE_SIZE),
    [items, page]
  );

  useEffect(() => {
    if (safePage !== page) {
      const next = new URLSearchParams(searchParams);
      if (safePage === 1) next.delete('page');
      else next.set('page', String(safePage));
      setSearchParams(next, { replace: true });
    }
  }, [safePage, page, searchParams, setSearchParams]);

  useEffect(() => {
    const qs = searchParams.toString();
    const base = `/authors/${encodeURIComponent(decodedId)}`;
    saveCatalogNav({
      ids: items.map((it) => it.id),
      listPath: qs ? `${base}?${qs}` : base,
    });
  }, [items, searchParams, decodedId]);

  const handleBuy = async (item: ShopItem) => {
    try {
      const { allowed, reason } = await paymentApi.canPurchaseMerch();
      if (!allowed) {
        alert(reason || 'Покупка недоступна без подписки');
        return;
      }
    } catch {
      /* allow modal attempt */
    }
    setSelectedItem(item);
    setShowModal(true);
  };

  return (
    <PageShell width="wide">
      <PageHeader
        title={displayName}
        subtitle={pluralCards(items.length)}
        onBack={() => navigate('/authors')}
        backLabel="К списку художников"
      />

      {loading ? (
        <div className="flex justify-center py-16">
          <Spinner size={48} />
        </div>
      ) : items.length === 0 ? (
        <p className="py-16 text-center text-text-secondary">Карт этого художника нет</p>
      ) : (
        <>
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {slice.map((item) => (
              <ShopItemCard
                key={item.id}
                item={item}
                balance={balance}
                onBuyClick={handleBuy}
              />
            ))}
          </div>
          <CatalogPager
            page={safePage}
            pageCount={pageCount}
            total={items.length}
            onPageChange={(p) => {
              const next = new URLSearchParams(searchParams);
              if (p <= 1) next.delete('page');
              else next.set('page', String(p));
              setSearchParams(next);
            }}
          />
        </>
      )}

      <PurchaseModal
        isOpen={showModal}
        item={selectedItem}
        balance={balance}
        loading={useShopStore.getState().buying}
        onClose={() => {
          setShowModal(false);
          setSelectedItem(null);
        }}
        onConfirm={async () => {
          if (!selectedItem) return;
          try {
            await useShopStore.getState().buyItem(selectedItem.id);
            setShowModal(false);
            setSelectedItem(null);
          } catch (e: any) {
            alert(e.message || 'Ошибка покупки');
          }
        }}
        merchAllowed
        onGoSubscription={() => navigate('/subscription')}
      />
    </PageShell>
  );
}
