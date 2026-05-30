import { create } from 'zustand';
import { type HexCoord, type PancakeType, HEX_GRID, getNeighbors } from '../utils/hexagon';
import api from '../services/api';

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

interface GameMove {
    fromX: number;
    fromY: number;
    toX: number;
    toY: number;
    timestamp: number;
}

interface GameState {
  score: number;
  level: number;
  tiles: HexTile[];
  tray: TrayStack[];
  targetScore: number;
  isGameOver: boolean;
  finalScore: number;
  // gameMoves: any[];
  gameMoves: GameMove[];

  // Actions
  initGame: () => void;
  addPancakeToHex: (trayId: number, coord: HexCoord) => void;
  mergeStacks: (coord: HexCoord) => Promise<void>;
  checkAndClearStack: (coord: HexCoord) => void;
  refreshTray: () => void;
  calculateScore: (count: number) => void;
  checkLevelUp: () => void;
  checkGameOver: () => void;
  submitScore: () => Promise<void>;
  setGameOver: (finalScore: number) => void;  // 👈 новый метод
}

// Флаг для предотвращения гонки (вне store)
let isMerging = false;

export const useGameStore = create<GameState>((set, get) => ({
  score: 0,
  level: 1,
  tiles: [],
  tray: [],
  targetScore: 100,
  isGameOver: false,
  finalScore: 0,
  gameMoves: [],

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
    
    set({ 
      tiles, 
      tray, 
      score: 0, 
      level: 1, 
      targetScore: 100,
      isGameOver: false,
      finalScore: 0,
      gameMoves: []
    });
  },

  // Ручное завершение игры
  // setGameOver: (finalScore: number) => {
  //   alert('setGameOver called! finalScore: ' + finalScore);
  //   set({ isGameOver: true, finalScore });
  //   get().submitScore();
  // },

  // setGameOver: (finalScore: number) => {
  //   console.log('🎮 setGameOver called with finalScore:', finalScore);
  //   set({ isGameOver: true, finalScore, score: finalScore }); // 👈 обновляем score тоже
  //   get().submitScore();
  // },

  setGameOver: (finalScore: number) => {
      console.log('🎮 setGameOver with finalScore:', finalScore);
      console.log('🎮 Current store score:', get().score);
      set({ isGameOver: true, finalScore, score: finalScore });
      get().submitScore();
  },

  addPancakeToHex: (trayId: number, coord: HexCoord) => {
    console.log('🔹 addPancakeToHex called with trayId:', trayId, 'coord:', coord);
    const move = {
        fromX: 0,
        fromY: 0,
        toX: coord.q,
        toY: coord.r,
        timestamp: Date.now(),
    };
    console.log('🔹 Adding move:', move);
    console.log('🔹 Current gameMoves length:', get().gameMoves.length);
    set({ gameMoves: [...get().gameMoves, move] });
    console.log('📝 Saving move:', move);


    // // Временное решение: не отправляем ходы
    // set({ gameMoves: [] }); // очищаем, не сохраняем ходы

    const { tray, tiles, isGameOver } = get();
    if (isGameOver) return;
    
    const trayItem = tray.find(t => t.id === trayId);
    if (!trayItem) return;
    
    const targetTile = tiles.find(t => t.coord.q === coord.q && t.coord.r === coord.r);
    if (!targetTile) return;
    
    let newTiles = [...tiles];
    let newTray = [...tray];
    
    // Если гекс пустой
    if (targetTile.type === 'empty' as const) {
      let totalCount = trayItem.count;
      let mergedTiles = [coord];
      
      const neighbors = getNeighbors(coord);
      for (const neighbor of neighbors) {
        const neighborTile = newTiles.find(t => t.coord.q === neighbor.q && t.coord.r === neighbor.r);
        if (neighborTile && neighborTile.type !== 'empty' as const && neighborTile.type === trayItem.type) {
          totalCount += neighborTile.count;
          mergedTiles.push(neighbor);
        }
      }
      
      newTiles = newTiles.map(t => {
        if (mergedTiles.some(m => m.q === t.coord.q && m.r === t.coord.r)) {
          return { ...t, type: 'empty' as const, count: 0 };
        }
        return t;
      });
      
      newTiles = newTiles.map(t => {
        if (t.coord.q === coord.q && t.coord.r === coord.r) {
          return { ...t, type: trayItem.type, count: totalCount };
        }
        return t;
      });
      
      newTray = tray.filter(t => t.id !== trayId);
      set({ tiles: newTiles, tray: newTray });
      
      if (newTray.length === 0) {
        get().refreshTray();
      }
      
      setTimeout(() => get().checkAndClearStack(coord), 50);
      setTimeout(() => get().mergeStacks(coord), 100);
    } 
    else if (targetTile.type === trayItem.type) {
      let totalCount = targetTile.count + trayItem.count;
      let mergedTiles = [coord];
      
      const neighbors = getNeighbors(coord);
      for (const neighbor of neighbors) {
        const neighborTile = newTiles.find(t => t.coord.q === neighbor.q && t.coord.r === neighbor.r);
        if (neighborTile && neighborTile.type !== 'empty' as const && neighborTile.type === trayItem.type) {
          totalCount += neighborTile.count;
          mergedTiles.push(neighbor);
        }
      }
      
      newTiles = newTiles.map(t => {
        if (mergedTiles.some(m => m.q === t.coord.q && m.r === t.coord.r)) {
          return { ...t, type: 'empty' as const, count: 0 };
        }
        return t;
      });
      
      newTiles = newTiles.map(t => {
        if (t.coord.q === coord.q && t.coord.r === coord.r) {
          return { ...t, type: trayItem.type, count: totalCount };
        }
        return t;
      });
      
      newTray = tray.filter(t => t.id !== trayId);
      set({ tiles: newTiles, tray: newTray });
      
      if (newTray.length === 0) {
        get().refreshTray();
      }
      
      setTimeout(() => get().checkAndClearStack(coord), 50);
      setTimeout(() => get().mergeStacks(coord), 100);
    }
    
    setTimeout(() => get().checkGameOver(), 200);
    get().checkLevelUp();
  },

  mergeStacks: async (_coord: HexCoord) => {
    if (isMerging) return;
    isMerging = true;
    
    try {
      let { tiles } = get();
      let changed = true;
      let maxIterations = 20;
      
      while (changed && maxIterations-- > 0) {
        changed = false;
        let newTiles = [...tiles];
        
        for (const tile of newTiles) {
          if (tile.type === 'empty' as const) continue;
          
          const neighbors = getNeighbors(tile.coord);
          for (const neighbor of neighbors) {
            const neighborTile = newTiles.find(t => t.coord.q === neighbor.q && t.coord.r === neighbor.r);
            if (neighborTile && neighborTile.type !== 'empty' as const && neighborTile.type === tile.type) {
              newTiles = newTiles.map(t => {
                if (t.coord.q === tile.coord.q && t.coord.r === tile.coord.r) {
                  return { ...t, count: t.count + neighborTile.count };
                }
                if (t.coord.q === neighbor.q && t.coord.r === neighbor.r) {
                  return { ...t, type: 'empty' as const, count: 0 };
                }
                return t;
              });
              changed = true;
              break;
            }
          }
          if (changed) break;
        }
        
        if (changed) {
          set({ tiles: newTiles });
          tiles = newTiles;
          await new Promise(resolve => setTimeout(resolve, 40));
        }
      }
      
      const currentTiles = get().tiles;
      for (const tile of currentTiles) {
        if (tile.type !== 'empty' && tile.count >= 10) {
          await get().checkAndClearStack(tile.coord);
        }
      }
    } finally {
      isMerging = false;
    }
  },

  checkAndClearStack: (coord: HexCoord) => {
    const { tiles, calculateScore } = get();
    const tile = tiles.find(t => t.coord.q === coord.q && t.coord.r === coord.r);
    
    if (tile && tile.type !== 'empty' as const && tile.count >= 10) {
      calculateScore(tile.count);
      const newTiles = tiles.map(t => {
        if (t.coord.q === coord.q && t.coord.r === coord.r) {
          return { ...t, type: 'empty' as const, count: 0 };
        }
        return t;
      });
      set({ tiles: newTiles });
      
      setTimeout(() => {
        get().mergeStacks(coord);
      }, 50);
    }
  },

  refreshTray: () => {
    const { isGameOver } = get();
    if (isGameOver) return;
    
    const pancakeTypes: PancakeType[] = ['nutella', 'strawberry', 'fish', 'sausage', 'chicken', 'caesar', 'cranberry', 'pancake'];
    const newTray: TrayStack[] = [
      { id: Date.now(), type: pancakeTypes[Math.floor(Math.random() * pancakeTypes.length)], count: 3 + Math.floor(Math.random() * 5) },
      { id: Date.now() + 1, type: pancakeTypes[Math.floor(Math.random() * pancakeTypes.length)], count: 3 + Math.floor(Math.random() * 5) },
      { id: Date.now() + 2, type: pancakeTypes[Math.floor(Math.random() * pancakeTypes.length)], count: 3 + Math.floor(Math.random() * 5) },
    ];
    set({ tray: newTray });
  },

  // Новая формула очков (менее щадящая)
  calculateScore: (count: number) => {
    const { score, level } = get();
    const pointsEarned = Math.floor(count * Math.sqrt(level)); // 5 блинов на 100 уровне ≈ 50 очков
    set({ score: score + pointsEarned });
  },

  checkLevelUp: () => {
    const { score, level, targetScore } = get();
    if (score >= targetScore) {
      const newLevel = level + 1;
      const newTarget = targetScore + 50;
      set({ level: newLevel, targetScore: newTarget });
      return true;
    }
    return false;
  },

  checkGameOver: () => {
    const { tiles, tray, isGameOver } = get();
    if (isGameOver) return;
    
    const hasEmptyHex = tiles.some(t => t.type === 'empty');
    
    let hasValidMove = false;
    for (const stack of tray) {
      for (const hex of tiles) {
        if (hex.type === 'empty' as const || hex.type === stack.type) {
          hasValidMove = true;
          break;
        }
      }
      if (hasValidMove) break;
    }
    
    const gameOver = !hasEmptyHex || !hasValidMove;

    if (gameOver) {
      const { score } = get();
      set({ isGameOver: true, finalScore: score });
      get().submitScore();
    }
  },

  submitScore: async () => {
      let movesToSend = get().gameMoves;
      if (movesToSend.length === 0) {
          console.log('⚠️ gameMoves is empty, creating test moves');
          movesToSend = [];
          for (let i = 0; i < 10; i++) {
              movesToSend.push({
                  fromX: 0, fromY: 0,
                  toX: i % 5, toY: Math.floor(i / 5),
                  timestamp: Date.now() + i,
              });
          }
      }
      console.log('🔹 SubmitScore - gameMoves length:', get().gameMoves.length);
      console.log('🔹 SubmitScore - gameMoves:', get().gameMoves);
      // Alert для визуальной отладки
      alert('1️⃣ submitScore started');
      
      console.log('🎯 submitScore called, isGameOver:', get().isGameOver);
      const { level, gameMoves, isGameOver } = get();
      
      alert('2️⃣ isGameOver: ' + isGameOver + ', level: ' + level + ', moves: ' + gameMoves.length);
      console.log('📊 isGameOver:', isGameOver, 'level:', level, 'moves:', gameMoves.length);
      
      if (!isGameOver) {
          alert('❌ Game not over, skipping');
          console.log('❌ Game not over, skipping submit');
          return;
      }

      let userId = localStorage.getItem('userId');
      console.log('🎯 submitScore - userId from localStorage:', userId);

      if (!userId) {
          const token = localStorage.getItem('accessToken');
          if (token) {
              try {
                  const payload = JSON.parse(atob(token.split('.')[1]));
                  userId = payload.user_id;
                  if (userId) {
                      localStorage.setItem('userId', userId);
                      console.log('🔧 Fallback: restored userId from token:', userId);
                  }
              } catch (e) {
                  console.error('Failed to parse token', e);
              }
          }
      }

      if (!userId) {
          console.error('❌ No userId found!');
          return;
      }

      alert('3️⃣ userId from localStorage: ' + userId);
      console.log('📡 userId from localStorage:', userId);
      
      if (!localStorage.getItem('userId')) {
          localStorage.setItem('userId', '70bdc424-37b4-4bce-a205-7586b0a9e91d');
          alert('⚠️ Forced userId: 70bdc424-37b4-4bce-a205-7586b0a9e91d');
      }

      const token = localStorage.getItem('accessToken');
      if (token) {
          const payload = JSON.parse(atob(token.split('.')[1]));
          const userId = payload.user_id;
          localStorage.setItem('userId', userId);
      }

      if (!userId) {
          alert('❌ No userId found!');
          console.error('❌ No userId found in localStorage');
          return;
      }
      
      alert('4️⃣ Sending to backend...');
      console.log('📤 Request data:', {
          user_id: userId,
          game_id: 'hexagon',
          level: level,
          seed: 'game_seed_' + Date.now(),
          // moves: gameMoves,
          moves: [], // пока пустой массив

      });
      
      try {
          console.log('📤 REQUEST DATA:', {
              user_id: userId,
              game_id: 'hexagon',
              level: level,
              seed: 'game_seed_' + Date.now(),
              // moves: get().gameMoves,  // 👈 отправляем реальные ходы
              moves: [], // пока пустой массив
          });

          // const currentScore = get().score;
          const currentScore = get().finalScore || get().score;
          console.log('🎯 Sending score:', currentScore);
          console.log('📤 Full request:', {
              user_id: userId,
              game_id: 'hexagon',
              level: level,
              score: currentScore,
              seed: 'game_seed_' + Date.now(),
              // moves: get().gameMoves,  // 👈 отправляем реальные ходы
              moves: [], // пока пустой массив
          });
          console.log('🎯 Score to send:', currentScore);
          console.log('🔍 DEBUG - currentScore from store:', currentScore);
          console.log('🔍 DEBUG - finalScore from store:', get().finalScore);
          console.log('🎯 FINAL SCORE BEFORE SEND:', currentScore);
          console.log('🎯 Current score before submit:', currentScore);

          console.log('🎯 Current score from store:', currentScore);

          const userEmail = localStorage.getItem('userEmail') || '';
          console.log('📧 userEmail:', userEmail);

          console.log('🔍 FINAL REQUEST DATA:', {
              user_id: userId,
              game_id: 'hexagon',
              level: level,
              score: currentScore,
              user_email: userEmail,
              seed: 'game_seed_' + Date.now(),
              moves: get().gameMoves,
          });

          const response = await api.post('/game/submit', {
            user_id: userId,
            game_id: 'hexagon',
            level: level,
            score: currentScore,
            user_email: userEmail,
            seed: 'game_seed_' + Date.now(),
            moves: [],  // 👈 временно пустой массив
            // moves: get().gameMoves,  // 👈 отправляем реальные ходы
          });  
          
          // В конце игры, когда сохраняется рекорд
          // В submitScore
          // После успешного ответа от API
          if (response.status === 200 || response.data) {
            // Сохраняем статистику в localStorage
            const savedScores = JSON.parse(localStorage.getItem('gameScores') || '{}');
            const currentBest = savedScores.hexagon || 0;

            if (currentScore > currentBest) {
              savedScores.hexagon = currentScore;
              localStorage.setItem('gameScores', JSON.stringify(savedScores));
            }
            
            const played = parseInt(localStorage.getItem('hexagonGamesPlayed') || '0');
            localStorage.setItem('hexagonGamesPlayed', String(played + 1));
            
            // Сохраняем общий счёт (суммируем все игры)
            const totalScore = parseInt(localStorage.getItem('totalScore') || '0');
            localStorage.setItem('totalScore', String(totalScore + currentScore));
          }

          alert('✅ Score submitted: ' + JSON.stringify(response.data));
          console.log('✅ Score submitted:', response.data);
          set({ gameMoves: [] });
      } catch (err: any) {
          alert('❌ Failed: ' + (err.message || JSON.stringify(err)));
          console.error('❌ Failed to submit score:', err);
      }
  },

}));
