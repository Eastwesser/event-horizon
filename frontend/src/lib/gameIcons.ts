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

export type GameTabId =
  | 'hexagon'
  | 'flappy'
  | 'towers'
  | 'hanoi'
  | 'memory'
  | 'twenty48'
  | 'gears'
  | 'companion';

/** Canonical game order — Home + Leaderboard (same sequence). */
export const GAME_LB_TABS: { id: GameTabId; label: string; icon: IconName }[] = [
  { id: 'hexagon', label: 'Pancaker', icon: gameIcon('hexagon') },
  { id: 'flappy', label: 'Flappy Bird', icon: gameIcon('flappy') },
  { id: 'towers', label: 'Builder', icon: gameIcon('towers') },
  { id: 'hanoi', label: 'Hanoi', icon: gameIcon('hanoi') },
  { id: 'memory', label: 'Memonia', icon: gameIcon('memory') },
  { id: 'twenty48', label: '2048', icon: gameIcon('twenty48') },
  { id: 'gears', label: 'Gears', icon: gameIcon('gears') },
  { id: 'companion', label: 'Компаньон', icon: gameIcon('companion') },
];

