import type { InventoryItem } from '../services/inventoryApi';

export interface CardArtist {
  artist_id: string;
  display_name: string;
  count: number;
}

export function collectCardArtists(items: InventoryItem[]): CardArtist[] {
  const map = new Map<string, CardArtist>();
  for (const item of items) {
    if (item.type !== 'карточка') continue;
    const attrs = item.attributes || {};
    const id =
      (typeof attrs.artist_id === 'string' && attrs.artist_id) ||
      slugFallback(
        (typeof attrs.artist_display === 'string' && attrs.artist_display) ||
          (typeof attrs.artist === 'string' && attrs.artist) ||
          ''
      );
    if (!id || id === 'unknown') continue;
    const display =
      (typeof attrs.artist_display === 'string' && attrs.artist_display) ||
      (typeof attrs.artist === 'string' && attrs.artist) ||
      id;
    const cur = map.get(id);
    if (cur) cur.count += 1;
    else map.set(id, { artist_id: id, display_name: display, count: 1 });
  }
  return [...map.values()].sort((a, b) =>
    a.display_name.localeCompare(b.display_name, 'ru', { sensitivity: 'base' })
  );
}

function slugFallback(name: string): string {
  return name
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '_')
    .replace(/[^a-z0-9_\u0400-\u04ff-]/gi, '') || 'unknown';
}
