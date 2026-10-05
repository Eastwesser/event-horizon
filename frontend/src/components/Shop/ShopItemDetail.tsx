import { useEffect, useMemo, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { inventoryApi, type InventoryItem } from '../../services/inventoryApi';
import { paymentApi } from '../../services/paymentApi';
import { useShopStore } from '../../store/shopStore';
import { inventoryToShopItem } from '../../lib/shopItemMap';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import Notification from '../Common/Notification/Notification';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { CardImage } from '../ui/CardImage';
import { Icon } from '../ui/Icon';
import { formatTicketAmount, formatTicketPrice } from '../../lib/formatPrice';
import { stockLabel } from '../../lib/shopItemMap';
import { loadCatalogNav, navNeighbors } from '../../lib/catalogNav';
import { CardAttributesView, CardFlagBadges } from './cardAttributes';
import { NoizReviewBlock } from './NoizReviewBlock';
import PurchaseModal from './PurchaseModal';
import CancelPurchaseModal from './CancelPurchaseModal';

const ELEMENT_RU: Record<string, string> = {
  mountains: 'Горы',
  woods: 'Лес',
  steppes: 'Степи',
  swamps: 'Болота',
  darkness: 'Тьма',
  neutral: 'Нейтралы',
};

export function ShopItemDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const {
    balance,
    inventory,
    buying,
    cancelling,
    fetchBalance,
    fetchInventory,
    buyItem,
    cancelPurchase,
  } = useShopStore();

  const [item, setItem] = useState<InventoryItem | null>(null);
  const [owned, setOwned] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [textOpen, setTextOpen] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [showCancelModal, setShowCancelModal] = useState(false);
  const [merchAllowed, setMerchAllowed] = useState<boolean | null>(null);
  const [merchBlockReason, setMerchBlockReason] = useState('');
  const [notice, setNotice] = useState<{ type: 'success' | 'error'; message: string } | null>(
    null
  );
  const touchX = useRef<number | null>(null);

  const ownedPurchase = useMemo(
    () => inventory.find((p) => p.item_id === id || p.item?.id === id),
    [inventory, id]
  );
  const refundAmount =
    ownedPurchase?.purchase_price || ownedPurchase?.item?.price_tickets || item?.price || 0;

  const nav = useMemo(() => navNeighbors(id || '', loadCatalogNav()), [id, item]);

  useEffect(() => {
    fetchBalance();
    fetchInventory();
  }, [fetchBalance, fetchInventory]);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;
    setLoading(true);
    setError('');
    setTextOpen(false);
    inventoryApi
      .getItem(id)
      .then((data) => {
        if (!cancelled) setItem(data);
      })
      .catch(() => {
        if (!cancelled) setError('Товар не найден');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [id]);

  useEffect(() => {
    if (!id) return;
    const isOwned = inventory.some((p) => p.item_id === id || p.item?.id === id);
    setOwned(isOwned);
  }, [id, inventory]);

  const goNeighbor = (neighborId: string | null) => {
    if (!neighborId) return;
    navigate(`/shop/item/${neighborId}`);
  };

  const handleBuyClick = async () => {
    if (!item || owned || item.stock === 0) return;
    try {
      const { allowed, reason } = await paymentApi.canPurchaseMerch();
      setMerchAllowed(allowed);
      if (!allowed) {
        setMerchBlockReason(reason || 'Покупка мерча недоступна без активной подписки');
        setNotice({
          type: 'error',
          message: `${reason || 'Покупка недоступна'}. Оформите подписку.`,
        });
        return;
      }
    } catch {
      setMerchAllowed(null);
      setMerchBlockReason('');
    }
    setShowModal(true);
  };

  const handleConfirmPurchase = async () => {
    if (!item) return;
    try {
      const result: any = await buyItem(item.id);
      const remaining =
        typeof result?.remaining_stock === 'number'
          ? result.remaining_stock
          : typeof result?.remainingStock === 'number'
            ? result.remainingStock
            : typeof item.stock === 'number'
              ? Math.max(0, item.stock - 1)
              : item.stock;

      setItem((prev) =>
        prev
          ? {
              ...prev,
              stock: remaining ?? prev.stock,
            }
          : prev
      );
      setOwned(true);
      setShowModal(false);
      setNotice({ type: 'success', message: `${item.name} куплен!` });
    } catch (e: any) {
      setNotice({
        type: 'error',
        message: e.message || 'Ошибка при покупке',
      });
    }
  };

  const handleConfirmCancel = async () => {
    if (!item) return;
    try {
      const result: any = await cancelPurchase(item.id);
      const remaining =
        typeof result?.remaining_stock === 'number'
          ? result.remaining_stock
          : typeof result?.remainingStock === 'number'
            ? result.remainingStock
            : typeof item.stock === 'number'
              ? item.stock + 1
              : item.stock;

      setItem((prev) =>
        prev
          ? {
              ...prev,
              stock: remaining ?? prev.stock,
            }
          : prev
      );
      setOwned(false);
      setShowCancelModal(false);
      const refunded =
        result?.refunded_amount ?? result?.refundedAmount ?? refundAmount;
      setNotice({
        type: 'success',
        message: `Покупка отменена. Возвращено ${formatTicketAmount(refunded)}.`,
      });
    } catch (e: any) {
      setNotice({
        type: 'error',
        message: e.message || 'Ошибка при отмене покупки',
      });
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-void">
        <LoadingSpinner />
      </div>
    );
  }

  if (error || !item) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-3 bg-void text-text-secondary">
        <p>{error || 'Товар не найден'}</p>
        <button
          type="button"
          className="text-photon-cyan underline"
          onClick={() => navigate('/shop')}
        >
          Назад в магазин
        </button>
      </div>
    );
  }

  const image = item.images?.[0];
  const attrs = (item.attributes || {}) as Record<string, unknown>;
  const stockText = stockLabel(item.stock);
  const outOfStock = item.stock === 0;
  const cardText =
    (typeof attrs.card_text === 'string' && attrs.card_text.trim()) ||
    (item.description || '').trim();
  const flavor =
    typeof attrs.flavor_text === 'string' ? attrs.flavor_text.trim() : '';

  const setName = typeof attrs.set_name === 'string' ? attrs.set_name : '';
  const year = typeof attrs.year === 'number' ? attrs.year : null;
  const artistId = typeof attrs.artist_id === 'string' ? attrs.artist_id : '';
  const artist =
    (typeof attrs.artist_display === 'string' && attrs.artist_display) ||
    (typeof attrs.artist === 'string' && attrs.artist) ||
    '';
  const element =
    typeof attrs.element === 'string'
      ? ELEMENT_RU[attrs.element] || attrs.element
      : '';

  const metaParts = [
    setName,
    year != null ? String(year) : '',
    artist || artistId,
  ].filter(Boolean);

  const listBack = loadCatalogNav()?.listPath || '/shop';
  const shopItem = inventoryToShopItem(item, owned);

  return (
    <PageShell width="narrow">
      <PageHeader
        title={item.name}
        onBack={() => navigate(listBack)}
        backLabel="Назад к списку"
        actions={
          nav.total > 0 && nav.index >= 0 ? (
            <span className="font-hud text-xs tabular-nums text-text-muted">
              {nav.index + 1} / {nav.total}
            </span>
          ) : undefined
        }
      />

      {notice && (
        <Notification
          type={notice.type}
          message={notice.message}
          onClose={() => setNotice(null)}
        />
      )}

      {(nav.prevId || nav.nextId) && (
        <div className="mb-4 flex items-center justify-between gap-2">
          <Button
            variant="ghost"
            size="sm"
            disabled={!nav.prevId}
            onClick={() => goNeighbor(nav.prevId)}
          >
            ← Пред
          </Button>
          <Button
            variant="ghost"
            size="sm"
            disabled={!nav.nextId}
            onClick={() => goNeighbor(nav.nextId)}
          >
            След →
          </Button>
        </div>
      )}

      <div
        className="overflow-hidden rounded-md border border-white/10 bg-nebula"
        onTouchStart={(e) => {
          touchX.current = e.changedTouches[0]?.clientX ?? null;
        }}
        onTouchEnd={(e) => {
          const start = touchX.current;
          touchX.current = null;
          if (start == null) return;
          const end = e.changedTouches[0]?.clientX ?? start;
          const dx = end - start;
          if (Math.abs(dx) < 60) return;
          if (dx > 0) goNeighbor(nav.prevId);
          else goNeighbor(nav.nextId);
        }}
      >
        <div className="overflow-hidden rounded-md bg-nebula px-3 py-4 sm:px-6">
          <CardImage
            src={image}
            alt={item.name}
            className="mx-auto max-h-[32rem] w-full max-w-sm bg-nebula"
            fit="contain"
            fixedAspect={false}
            fallback={<Icon name="cards" className="h-16 w-16 text-text-muted" />}
          />
        </div>

        <div className="space-y-3 p-5 sm:p-6">
          {metaParts.length > 0 ? (
            <p className="text-sm text-text-secondary">
              {setName ? <span>{setName}</span> : null}
              {setName && year != null ? <span> · </span> : null}
              {year != null ? <span>{year}</span> : null}
              {(setName || year != null) && (artist || artistId) ? (
                <span> · </span>
              ) : null}
              {artistId ? (
                <Link
                  to={`/authors/${encodeURIComponent(artistId)}`}
                  className="text-horizon-cyan underline-offset-2 hover:underline"
                >
                  {artist || artistId}
                </Link>
              ) : artist ? (
                <span className="text-text-primary">{artist}</span>
              ) : null}
            </p>
          ) : null}

          <div className="flex flex-wrap items-center gap-2 text-sm">
            <Badge tone="cyan">{item.type}</Badge>
            {element ? <Badge tone="indigo">{element}</Badge> : null}
            <span className="inline-flex items-center gap-1 font-hud tabular-nums text-horizon-gold">
              <Icon name="ticket" className="h-3.5 w-3.5" />
              {formatTicketPrice(item.price)}
            </span>
            {stockText !== null && (
              <span
                className={
                  outOfStock ? 'text-text-muted' : 'text-text-secondary'
                }
              >
                {stockText}
              </span>
            )}
          </div>

          <CardFlagBadges attrs={attrs} />

          {flavor ? (
            <p className="border-l-2 border-horizon-gold/40 pl-3 text-sm italic text-text-muted">
              {flavor}
            </p>
          ) : null}

          {cardText ? (
            <div>
              <button
                type="button"
                className="text-sm text-horizon-cyan underline-offset-2 hover:underline"
                onClick={() => setTextOpen((v) => !v)}
                aria-expanded={textOpen}
              >
                {textOpen ? 'Скрыть текст карты' : 'Показать текст карты'}
              </button>
              {textOpen ? (
                <p className="mt-2 whitespace-pre-wrap text-sm text-text-secondary">
                  {cardText}
                </p>
              ) : null}
            </div>
          ) : null}

          <CardAttributesView
            attrs={attrs}
            trailingAction={
              <div className="flex flex-col items-end gap-1">
                {owned ? (
                  <>
                    {ownedPurchase?.can_cancel ? (
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={cancelling}
                        onClick={() => setShowCancelModal(true)}
                        className="gap-1.5"
                      >
                        <Icon name="undo" className="h-3.5 w-3.5" />
                        Отменить
                      </Button>
                    ) : (
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled
                        className="gap-1.5 opacity-50"
                        title={
                          ownedPurchase?.fulfilled_at
                            ? 'Товар уже отправлен — возврат недоступен.'
                            : 'Срок возврата истёк (7 дней с покупки).'
                        }
                      >
                        <Icon name="undo" className="h-3.5 w-3.5" />
                        Отменить
                      </Button>
                    )}
                    <Link
                      to="/shop?tab=inventory"
                      className="text-xs text-horizon-cyan underline-offset-2 hover:underline"
                    >
                      Мой инвентарь
                    </Link>
                  </>
                ) : (
                  <Button
                    variant={
                      !outOfStock && balance >= item.price ? 'primary' : 'ghost'
                    }
                    size="sm"
                    disabled={outOfStock || balance < item.price || buying}
                    onClick={() => void handleBuyClick()}
                  >
                    {outOfStock
                      ? 'Нет в наличии'
                      : balance < item.price
                        ? 'Не хватает'
                        : 'Купить'}
                  </Button>
                )}
              </div>
            }
          />

          <NoizReviewBlock attrs={attrs} />
        </div>
      </div>

      <PurchaseModal
        isOpen={showModal}
        item={shopItem}
        balance={balance}
        loading={buying}
        merchAllowed={merchAllowed}
        merchBlockReason={merchBlockReason}
        onConfirm={() => void handleConfirmPurchase()}
        onClose={() => setShowModal(false)}
        onGoSubscription={() => navigate('/subscription')}
      />

      <CancelPurchaseModal
        isOpen={showCancelModal}
        item={shopItem}
        refundAmount={refundAmount}
        loading={cancelling}
        onConfirm={() => void handleConfirmCancel()}
        onClose={() => setShowCancelModal(false)}
      />
    </PageShell>
  );
}
