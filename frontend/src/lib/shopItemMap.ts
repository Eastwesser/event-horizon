import type { InventoryItem } from '../services/inventoryApi';
import type { ShopItem } from '../components/Shop/shopStore';

/** Map an inventory catalog item into the shop card shape. */
export function inventoryToShopItem(item: InventoryItem, owned = false): ShopItem {
  const image = item.images?.[0] || '';
  return {
    id: item.id,
    name: item.name,
    description: item.description || '',
    price_tickets: item.price || 0,
    icon_url: image,
    type: item.type || 'other',
    category: item.type || 'other',
    game_id: undefined,
    image_url: image,
    images: item.images ?? [],
    available: (item.stock ?? 0) > 0,
    owned,
    stock: item.stock,
    attributes: item.attributes ?? {},
    created_at: item.created_at,
  };
}

export function stockLabel(stock: number | null | undefined): string | null {
  if (stock === null || stock === undefined) return null;
  if (stock === 0) return 'Нет в наличии';
  return `В наличии: ${stock}`;
}
