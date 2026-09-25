import React, { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useInventory } from '../../hooks/useInventory';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';
import { PageHeader } from '../ui/PageHeader';
import { PageShell } from '../ui/PageShell';
import { Badge } from '../ui/Badge';
import { formatRubPrice } from '../../lib/formatPrice';

export const InventoryItemDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { selectedItem, loading, fetchItem, clearSelected } = useInventory();

  useEffect(() => {
    if (id) {
      fetchItem(id);
    }
    return () => clearSelected();
  }, [id, fetchItem, clearSelected]);

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

  return (
    <PageShell width="narrow">
      <PageHeader title={selectedItem.name} onBack={() => navigate('/inventory')} />

      <div className="rounded-md border border-white/10 bg-nebula p-6">
        <div className="mb-4 flex flex-wrap items-center gap-2 text-sm">
          <Badge tone="cyan">{selectedItem.type}</Badge>
          <span className="font-hud tabular-nums text-horizon-gold">
            {formatRubPrice(selectedItem.price)}
          </span>
          {selectedItem.stock === null ? null : selectedItem.stock === undefined ||
            selectedItem.stock === 0 ? (
            <span className="text-text-muted">Нет в наличии</span>
          ) : (
            <span className="text-text-muted">В наличии: {selectedItem.stock}</span>
          )}
        </div>

        <p className="text-text-secondary">{selectedItem.description}</p>

        {selectedItem.attributes && Object.keys(selectedItem.attributes).length > 0 && (
          <div className="mt-6">
            <h3 className="mb-2 font-display text-base font-semibold text-text-primary">Характеристики</h3>
            <div className="flex flex-col gap-1">
              {Object.entries(selectedItem.attributes).map(([key, value]) => (
                <div key={key} className="text-sm text-text-secondary">
                  <strong className="text-text-primary">{key}:</strong> {String(value)}
                </div>
              ))}
            </div>
          </div>
        )}

        {selectedItem.images && selectedItem.images.length > 0 && (
          <div className="mt-6">
            <h3 className="mb-2 font-display text-base font-semibold text-text-primary">Изображения</h3>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
              {selectedItem.images.map((img, idx) => (
                <img
                  key={idx}
                  src={img}
                  alt={`${selectedItem.name} ${idx + 1}`}
                  className="aspect-square w-full rounded-sm object-cover"
                />
              ))}
            </div>
          </div>
        )}
      </div>
    </PageShell>
  );
};
