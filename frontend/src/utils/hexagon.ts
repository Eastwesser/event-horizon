// Axial координаты для гексагонов (q, r)
export interface HexCoord {
  q: number;
  r: number;
}

// 19 гексагонов
export const HEX_GRID: HexCoord[] = [
  // Центр
  { q: 0, r: 0 },
  // Кольцо 1
  { q: -1, r: 0 }, { q: 1, r: 0 },
  { q: -1, r: 1 }, { q: 0, r: 1 }, { q: 1, r: 1 },
  // Кольцо 2
  { q: -2, r: 0 }, { q: -2, r: 1 }, { q: -1, r: 2 }, 
  { q: 0, r: 2 }, { q: 1, r: 2 }, { q: 2, r: 1 }, { q: 2, r: 0 },
  // Кольцо 3
  { q: -2, r: 2 }, { q: -1, r: 3 }, { q: 0, r: 3 }, 
  { q: 1, r: 3 }, { q: 2, r: 2 },
  // Кольцо 4
  { q: -1, r: 4 }, { q: 0, r: 4 }, { q: 1, r: 4 },
];

// // Конвертация axial в пиксельные координаты
// const HEX_WIDTH = 80;   // было 60
// const HEX_HEIGHT = 92;  // было 70

// export function hexToPixel(q: number, r: number): { x: number; y: number } {
//   const x = HEX_WIDTH * (Math.sqrt(3) * q + Math.sqrt(3) / 2 * r);
//   const y = HEX_HEIGHT * (3/2 * r);
//   return { x, y };
// }

export function hexToPixel(q: number, r: number): { x: number; y: number } {
  const size = 40;
  const x = size * (Math.sqrt(3) * q + Math.sqrt(3) / 2 * r);
  const y = size * (3 / 2 * r);
  return { x, y };
}

// Получить соседей гекса
export function getNeighbors(coord: HexCoord): HexCoord[] {
  const directions = [
    { q: 1, r: 0 }, { q: 1, r: -1 }, { q: 0, r: -1 },
    { q: -1, r: 0 }, { q: -1, r: 1 }, { q: 0, r: 1 },
  ];
  return directions.map(d => ({ q: coord.q + d.q, r: coord.r + d.r }));
}

// Типы блинов (начинки)
export type PancakeType = 
  | 'nutella'
  | 'strawberry'
  | 'fish'
  | 'sausage'
  | 'chicken'
  | 'caesar'
  | 'cranberry'
  | 'pancake';

// Эмодзи для типов блинов
export const pancakeEmoji: Record<PancakeType, string> = {
  nutella: '🍫',
  strawberry: '🍓',
  fish: '🐟',
  sausage: '🌭',
  chicken: '🍗',
  caesar: '🥗',
  cranberry: '🍒',
  pancake: '🥞',
};

// Цвета для типов блинов
export const pancakeColor: Record<PancakeType, string> = {
  nutella: '#8B4513',
  strawberry: '#FF6B6B',
  fish: '#FFA500',
  sausage: '#CD5C5C',
  chicken: '#FFD700',
  caesar: '#90EE90',
  cranberry: '#FF69B4',
  pancake: '#DEB887',
};