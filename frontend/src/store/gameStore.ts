import { create } from 'zustand';
import { type HexCoord, type PancakeType, HEX_GRID, getNeighbors } from '../utils/hexagon';

interface HexTile {
  coord: HexCoord;
  type: PancakeType | 'empty';
  count: number;
}

interface TrayStack {
  id: number;
  type: PancakeType;
  count: number;
}

interface GameState {
  score: number;
  level: number;
  tiles: HexTile[];
  tray: TrayStack[];
  
  // Actions
  initGame: () => void;
  addPancakeToHex: (trayId: number, coord: HexCoord) => void;
  mergeStacks: (coord: HexCoord) => void;
  checkAndClearStack: (coord: HexCoord) => void;
  refreshTray: () => void;
  calculateScore: (count: number) => void;
}

export const useGameStore = create<GameState>((set, get) => ({
  score: 0,
  level: 1,
  tiles: [],
  tray: [],

  initGame: () => {
    // Создаём пустое поле
    const tiles: HexTile[] = HEX_GRID.map(coord => ({
      coord,
      type: 'empty' as const,
      count: 0,
    }));
    
    // Создаём 3 случайные стопки для подноса
    const pancakeTypes: PancakeType[] = ['nutella', 'strawberry', 'fish', 'sausage', 'chicken', 'caesar', 'cranberry', 'pancake'];
    const tray: TrayStack[] = [
      { id: Date.now(), type: pancakeTypes[Math.floor(Math.random() * pancakeTypes.length)], count: 3 + Math.floor(Math.random() * 5) },
      { id: Date.now() + 1, type: pancakeTypes[Math.floor(Math.random() * pancakeTypes.length)], count: 3 + Math.floor(Math.random() * 5) },
      { id: Date.now() + 2, type: pancakeTypes[Math.floor(Math.random() * pancakeTypes.length)], count: 3 + Math.floor(Math.random() * 5) },
    ];
    
    set({ tiles, tray, score: 0, level: 1 });
  },

  addPancakeToHex: (trayId: number, coord: HexCoord) => {
    const { tray, tiles } = get();
    const trayItem = tray.find(t => t.id === trayId);
    if (!trayItem) return;
    
    // Находим целевой гекс
    const targetTile = tiles.find(t => t.coord.q === coord.q && t.coord.r === coord.r);
    if (!targetTile) return;
    
    // Если гекс пустой — кладём стопку
    if (targetTile.type === 'empty') {
      const newTiles = tiles.map(t => {
        if (t.coord.q === coord.q && t.coord.r === coord.r) {
          return { ...t, type: trayItem.type, count: trayItem.count };
        }
        return t;
      });
      
      // Удаляем использованную стопку из подноса
      const newTray = tray.filter(t => t.id !== trayId);
      
      set({ tiles: newTiles, tray: newTray });
      
      // Проверяем, нужно ли обновить поднос
      if (newTray.length === 0) {
        get().refreshTray();
      }
      
      // Проверяем объединение с соседями
      setTimeout(() => get().mergeStacks(coord), 10);
    } 
    // Тот же тип — складываем
    else if (targetTile.type === trayItem.type) {
      const newCount = targetTile.count + trayItem.count;
      const newTiles = tiles.map(t => {
        if (t.coord.q === coord.q && t.coord.r === coord.r) {
          return { ...t, count: newCount };
        }
        return t;
      });
      
      const newTray = tray.filter(t => t.id !== trayId);
      set({ tiles: newTiles, tray: newTray });
      
      if (newTray.length === 0) {
        get().refreshTray();
      }
      
      setTimeout(() => get().checkAndClearStack(coord), 10);
    }
  },

  mergeStacks: (coord: HexCoord) => {
    const { tiles } = get();
    const currentTile = tiles.find(t => t.coord.q === coord.q && t.coord.r === coord.r);
    if (!currentTile || currentTile.type === 'empty') return;
    
    const neighbors = getNeighbors(coord);
    let newTiles = [...tiles];
    let merged = false;
    
    for (const neighbor of neighbors) {
      const neighborTile = newTiles.find(t => t.coord.q === neighbor.q && t.coord.r === neighbor.r);
      if (neighborTile && neighborTile.type !== 'empty' && neighborTile.type === currentTile.type) {
        // Объединяем соседнюю стопку в текущую
        newTiles = newTiles.map(t => {
          if (t.coord.q === coord.q && t.coord.r === coord.r) {
            return { ...t, count: t.count + neighborTile.count };
          }
          if (t.coord.q === neighbor.q && t.coord.r === neighbor.r) {
            return { ...t, type: 'empty' as const, count: 0 };
          }
          return t;
        });
        merged = true;
        break;
      }
    }
    
    if (merged) {
      set({ tiles: newTiles });
      setTimeout(() => get().checkAndClearStack(coord), 10);
      setTimeout(() => get().mergeStacks(coord), 20);
    }
  },

  checkAndClearStack: (coord: HexCoord) => {
    const { tiles, calculateScore } = get();
    const tile = tiles.find(t => t.coord.q === coord.q && t.coord.r === coord.r);
    
    if (tile && tile.type !== 'empty' && tile.count >= 10) {
      calculateScore(tile.count);
      const newTiles = tiles.map(t => {
        if (t.coord.q === coord.q && t.coord.r === coord.r) {
          return { ...t, type: 'empty' as const, count: 0 };
        }
        return t;
      });
      set({ tiles: newTiles });
    }
  },

  refreshTray: () => {
    const pancakeTypes: PancakeType[] = ['nutella', 'strawberry', 'fish', 'sausage', 'chicken', 'caesar', 'cranberry', 'pancake'];
    const newTray: TrayStack[] = [
      { id: Date.now(), type: pancakeTypes[Math.floor(Math.random() * pancakeTypes.length)], count: 3 + Math.floor(Math.random() * 5) },
      { id: Date.now() + 1, type: pancakeTypes[Math.floor(Math.random() * pancakeTypes.length)], count: 3 + Math.floor(Math.random() * 5) },
      { id: Date.now() + 2, type: pancakeTypes[Math.floor(Math.random() * pancakeTypes.length)], count: 3 + Math.floor(Math.random() * 5) },
    ];
    set({ tray: newTray });
  },

  calculateScore: (count: number) => {
    const { score, level } = get();
    const pointsEarned = count * level;
    set({ score: score + pointsEarned });
  },
}));
