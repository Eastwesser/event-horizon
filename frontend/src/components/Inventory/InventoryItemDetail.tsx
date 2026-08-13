import React, { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useInventory } from '../../hooks/useInventory';
import './styles/Inventory.css';
import LoadingSpinner from '../Common/Spinner/LoadingSpinner';

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

  if (loading) return <LoadingSpinner />;
  if (!selectedItem) return <div className="inventory-empty">Товар не найден</div>;

  return (
    <div className="inventory-detail">
      <button className="btn btn-outline" onClick={() => navigate('/inventory')}>
        ← Назад к списку
      </button>
      <div className="inventory-detail-content">
        <h1>{selectedItem.name}</h1>
        <div className="inventory-detail-meta">
          <span className="inventory-detail-type">{selectedItem.type}</span>
          <span className="inventory-detail-price">{selectedItem.price} ₽</span>
          <span className="inventory-detail-stock">В наличии: {selectedItem.stock}</span>
        </div>
        <p className="inventory-detail-description">{selectedItem.description}</p>
        {selectedItem.attributes && Object.keys(selectedItem.attributes).length > 0 && (
          <div className="inventory-detail-attributes">
            <h3>Характеристики</h3>
            {Object.entries(selectedItem.attributes).map(([key, value]) => (
              <div key={key} className="inventory-detail-attribute">
                <strong>{key}:</strong> {String(value)}
              </div>
            ))}
          </div>
        )}
        {selectedItem.images && selectedItem.images.length > 0 && (
          <div className="inventory-detail-images">
            <h3>Изображения</h3>
            {selectedItem.images.map((img, idx) => (
              <img key={idx} src={img} alt={`${selectedItem.name} ${idx + 1}`} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
