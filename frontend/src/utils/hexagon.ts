// Axial координаты для гексагонов (q, r)
export interface HexCoord {
  q: number;
  r: number;
}

// Правильная 19-гексовая сетка (ромб)
export const HEX_GRID: HexCoord[] = [];
for (let q = -2; q <= 2; q++) {
  for (let r = -2; r <= 2; r++) {
    const s = -q - r;
    if (Math.abs(q) + Math.abs(r) + Math.abs(s) <= 4) {
      HEX_GRID.push({ q, r });
    }
  }
}

// Конвертация axial в пиксельные координаты
export function hexToPixel(q: number, r: number): { x: number; y: number } {
  const size = 40;
  const x = size * (Math.sqrt(3) * q + Math.sqrt(3) / 2 * r);
  const y = size * (3 / 2 * r);
  return { x, y };
}

// Получить вершины гекса для отрисовки
export function getHexagonPoints(radius: number): string {
  const points = [];
  for (let i = 0; i < 6; i++) {
    const angle = Math.PI / 2 + (Math.PI * 2 * i) / 6;
    const x = radius * Math.cos(angle);
    const y = radius * Math.sin(angle);
    points.push(`${x},${y}`);
  }
  return points.join(' ');
}

// Получить соседей гекса
export function getNeighbors(coord: HexCoord): HexCoord[] {
  const directions = [
    { q: 1, r: 0 }, { q: 1, r: -1 }, { q: 0, r: -1 },
    { q: -1, r: 0 }, { q: -1, r: 1 }, { q: 0, r: 1 },
  ];
  return directions.map(d => ({ q: coord.q + d.q, r: coord.r + d.r }));
}

// Типы блинов
export type PancakeType = 
  | 'nutella' | 'strawberry' | 'fish' | 'sausage'
  | 'chicken' | 'caesar' | 'cranberry' | 'pancake';

export const pancakeEmoji: Record<PancakeType, string> = {
  nutella: '🍫', strawberry: '🍓', fish: '🐟', sausage: '🌭',
  chicken: '🍗', caesar: '🥗', cranberry: '🍒', pancake: '🥞',
};

export const pancakeColor: Record<PancakeType, string> = {
  nutella: '#8B4513', strawberry: '#FF6B6B', fish: '#FFA500', sausage: '#CD5C5C',
  chicken: '#FFD700', caesar: '#90EE90', cranberry: '#FF69B4', pancake: '#DEB887',
};

export const EMPTY_COLOR = '#2a2a4a';
export const HEX_STROKE = '#ffd700';