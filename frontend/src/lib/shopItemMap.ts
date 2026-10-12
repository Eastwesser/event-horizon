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

/** Shop-service row (скины, темы, мерч) into the same card shape. */
export function shopApiToItem(raw: Record<string, unknown>, owned = false): ShopItem {
  const id = String(raw.id ?? raw.Id ?? '');
  const cat = String(raw.category ?? raw.Category ?? 'other');
  const available = raw.available !== false && raw.Available !== false;
  const stockRaw = raw.stock ?? raw.Stock;
  const stock = typeof stockRaw === 'number' ? stockRaw : available ? 999 : 0;
  const image = String(raw.image_url ?? raw.ImageUrl ?? '');
  return {
    id,
    name: String(raw.name ?? raw.Name ?? 'Без названия'),
    description: String(raw.description ?? raw.Description ?? ''),
    price_tickets: Number(raw.price ?? raw.Price ?? 0),
    icon_url: image,
    type: cat,
    category: cat,
    game_id: raw.game_id || raw.GameId
      ? String(raw.game_id ?? raw.GameId)
      : undefined,
    image_url: image,
    available,
    owned,
    stock,
  };
}

export function isShopCosmetic(item: Pick<ShopItem, 'category' | 'type'>): boolean {
  const c = `${item.category || ''} ${item.type || ''}`.toLowerCase();
  return (
    c.includes('game_skin') ||
    c.includes('profile_theme') ||
    c.includes('merch') ||
    c.includes('скин') ||
    c.includes('skin') ||
    c.includes('theme') ||
    c.includes('тема')
  );
}

export function stockLabel(stock: number | null | undefined): string | null {
  if (stock === null || stock === undefined) return null;
  if (stock === 0) return 'Нет в наличии';
  return `В наличии: ${stock}`;
}
