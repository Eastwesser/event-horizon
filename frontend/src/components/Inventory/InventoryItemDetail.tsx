import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { useInventory } from '../../hooks/useInventory';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { CardImage } from '../ui/CardImage';
import { Icon } from '../ui/Icon';
import { formatTicketPrice } from '../../lib/formatPrice';
import { stockLabel } from '../../lib/shopItemMap';
import { loadCatalogNav, navNeighbors } from '../../lib/catalogNav';
import { CardAttributesView, CardFlagBadges } from '../Shop/cardAttributes';
import { NoizReviewBlock } from '../Shop/NoizReviewBlock';

export const InventoryItemDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { selectedItem, loading, fetchItem, clearSelected } = useInventory();
  const [textOpen, setTextOpen] = useState(false);
  const touchX = useRef<number | null>(null);

  const nav = useMemo(() => navNeighbors(id || '', loadCatalogNav()), [id, selectedItem]);

  useEffect(() => {
    setTextOpen(false);
    if (id) fetchItem(id);
    return () => clearSelected();
  }, [id, fetchItem, clearSelected]);

  const goNeighbor = (neighborId: string | null) => {
    if (!neighborId) return;
    navigate(`/inventory/${neighborId}`);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-void">
        <LoadingSpinner />
      </div>
    );
  }

  if (!selectedItem) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-void text-text-secondary">
        Товар не найден
      </div>
    );
  }

  const stockText = stockLabel(selectedItem.stock);
  const attrs = (selectedItem.attributes || {}) as Record<string, unknown>;
  const isCard = selectedItem.type === 'карточка';
  const artistId = typeof attrs.artist_id === 'string' ? attrs.artist_id : '';
  const artistDisplay =
    (typeof attrs.artist_display === 'string' && attrs.artist_display) ||
    (typeof attrs.artist === 'string' && attrs.artist) ||
    '';
  const setName = typeof attrs.set_name === 'string' ? attrs.set_name : '';
  const year = typeof attrs.year === 'number' ? attrs.year : null;
  const cardText =
    (typeof attrs.card_text === 'string' && attrs.card_text.trim()) ||
    (selectedItem.description || '').trim();
  const flavor =
    typeof attrs.flavor_text === 'string' ? attrs.flavor_text.trim() : '';
  const listBack = loadCatalogNav()?.listPath || '/inventory';

  return (
    <PageShell width="narrow">
      <PageHeader
        title={selectedItem.name}
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
        {selectedItem.images?.[0] ? (
          <div className="overflow-hidden rounded-md bg-nebula px-3 py-4 sm:px-6">
            <CardImage
              src={selectedItem.images[0]}
              alt={selectedItem.name}
              className="mx-auto max-h-[32rem] w-full max-w-sm bg-nebula"
              fit="contain"
              fixedAspect={false}
              fallback={<Icon name="package" className="h-16 w-16 text-text-muted" />}
            />
          </div>
        ) : null}

        <div className="space-y-3 p-5 sm:p-6">
          {isCard && (setName || year != null || artistDisplay || artistId) ? (
            <p className="text-sm text-text-secondary">
              {setName ? <span>{setName}</span> : null}
              {setName && year != null ? <span> · </span> : null}
              {year != null ? <span>{year}</span> : null}
              {(setName || year != null) && (artistDisplay || artistId) ? (
                <span> · </span>
              ) : null}
              {artistId ? (
                <Link
                  to={`/authors/${encodeURIComponent(artistId)}`}
                  className="text-horizon-cyan underline-offset-2 hover:underline"
                >
                  {artistDisplay || artistId}
                </Link>
              ) : artistDisplay ? (
                <span className="text-text-primary">{artistDisplay}</span>
              ) : null}
            </p>
          ) : null}

          <div className="flex flex-wrap items-center gap-2 text-sm">
            <Badge tone="cyan">{selectedItem.type}</Badge>
            <span className="font-hud tabular-nums text-horizon-gold">
              {formatTicketPrice(selectedItem.price)}
            </span>
            {stockText !== null && (
              <span className="text-text-muted">{stockText}</span>
            )}
          </div>

          {isCard ? <CardFlagBadges attrs={attrs} /> : null}

          {flavor ? (
            <p className="border-l-2 border-horizon-gold/40 pl-3 text-sm italic text-text-muted">
              {flavor}
            </p>
          ) : null}

          {isCard && cardText ? (
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
          ) : !isCard ? (
            <p className="text-text-secondary">{selectedItem.description}</p>
          ) : null}

          {isCard ? (
            <>
              <CardAttributesView attrs={attrs} />
              <NoizReviewBlock attrs={attrs} />
            </>
          ) : selectedItem.attributes && Object.keys(selectedItem.attributes).length > 0 ? (
            <div>
              <h3 className="mb-2 font-display text-base font-semibold text-text-primary">
                Характеристики
              </h3>
              <div className="flex flex-col gap-1">
                {Object.entries(selectedItem.attributes)
                  .filter(
                    ([key]) =>
                      ![
                        'artist_id',
                        'market_rub',
                        'idempotency_key',
                        'noiz_review',
                      ].includes(key)
                  )
                  .map(([key, value]) => (
                    <div key={key} className="text-sm text-text-secondary">
                      <strong className="text-text-primary">{key}:</strong> {String(value)}
                    </div>
                  ))}
              </div>
            </div>
          ) : null}
        </div>
      </div>
    </PageShell>
  );
};
