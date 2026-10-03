import type { IconName } from '../components/ui/Icon';

/** Shared game → chrome Icon mapping (Home, Profile, Leaderboard). */
export const GAME_ICON: Record<string, IconName> = {
  hexagon: 'hex',
  flappy: 'bird',
  towers: 'tower',
  hanoi: 'hanoi',
  memory: 'cards',
  twenty48: 'twenty48',
  gears: 'gears',
  companion: 'star',
};

export function gameIcon(id: string): IconName {
  return GAME_ICON[id] || 'diamond';
}
