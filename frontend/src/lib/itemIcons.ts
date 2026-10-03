import type { IconName } from '../components/ui/Icon';
import { gameIcon } from './gameIcons';

/** Category / type → fallback Icon when there is no image. */
const TYPE_ICON: Record<string, IconName> = {
  карточка: 'cards',
  game_skin: 'palette',
  skin: 'palette',
  profile_theme: 'sparkle',
  theme: 'sparkle',
  merch: 'gift',
  мерч: 'gift',
  брелок: 'key',
  картина: 'frame',
  фенечка: 'sparkle',
  other: 'package',
};

export type ItemIconSource = {
  category?: string;
  type?: string;
  game_id?: string;
};

/**
 * Per-type fallback for catalog/inventory cards without images.
 * Skins with game_id use that game's chrome icon so «Космические блины» ≠ generic gift.
 */
export function itemFallbackIcon(item: ItemIconSource): IconName {
  const cat = (item.category || item.type || '').toLowerCase().trim();
  const isSkin =
    cat === 'game_skin' || cat === 'skin' || cat.includes('skin') || cat.includes('скин');

  if (item.game_id && (isSkin || !cat || cat === 'other')) {
    return gameIcon(item.game_id);
  }

  if (TYPE_ICON[cat]) return TYPE_ICON[cat];

  // Partial / localized matches
  if (cat.includes('карточ')) return 'cards';
  if (cat.includes('скин') || cat.includes('skin')) return 'palette';
  if (cat.includes('тем')) return 'sparkle';
  if (cat.includes('мерч') || cat.includes('merch')) return 'gift';
  if (cat.includes('брелок')) return 'key';
  if (cat.includes('картин')) return 'frame';
  if (cat.includes('фенеч')) return 'sparkle';

  return 'package';
}
