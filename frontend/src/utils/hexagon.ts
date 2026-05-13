// Axial координаты для гексагонов (q, r)
export interface HexCoord {
  q: number;
  r: number;
}

// 19 гексагонов для стартового поля
export const HEX_GRID: HexCoord[] = [
  // Центральный ряд
  { q: 0, r: 0 },
  // Второй ряд
  { q: -1, r: 0 }, { q: 1, r: 0 },
  { q: -1, r: 1 }, { q: 0, r: 1 }, { q: 1, r: 1 },
  // Третий ряд
  { q: -2, r: 0 }, { q: -2, r: 1 }, { q: -1, r: 2 }, 
  { q: 0, r: 2 }, { q: 1, r: 2 }, { q: 2, r: 1 }, { q: 2, r: 0 },
  // Четвёртый ряд
  { q: -2, r: 2 }, { q: -1, r: 3 }, { q: 0, r: 3 }, 
  { q: 1, r: 3 }, { q: 2, r: 2 },
  // Пятый ряд
  { q: -1, r: 4 }, { q: 0, r: 4 }, { q: 1, r: 4 },
];

// Конвертация axial в пиксельные координаты
export function hexToPixel(q: number, r: number, size: number): { x: number; y: number } {
  const x = size * (Math.sqrt(3) * q + Math.sqrt(3)/2 * r);
  const y = size * (3/2 * r);
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
