// frontend/src/hooks/useSkins.ts
import { useEffect, useState } from 'react';
import { getInventory } from '../services/api';

export interface GameSkins {
  flappy: {
    hasRainbowPipes: boolean;
    hasGoldenBird: boolean;
  };
  hexagon: {
    hasSpacePancakes: boolean;
  };
  towers: {
    hasRainbowBlocks: boolean;
  };
  memory: {
    hasAnimalCards: boolean;
  };
}

export function useSkins() {
  const [skins, setSkins] = useState<GameSkins>({
    flappy: { hasRainbowPipes: false, hasGoldenBird: false },
    hexagon: { hasSpacePancakes: false },
    towers: { hasRainbowBlocks: false },
    memory: { hasAnimalCards: false },
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadSkins = async () => {
      try {
        const response = await getInventory();
        const items = response.data?.items || [];
        
        setSkins({
          flappy: {
            hasRainbowPipes: items.some((item: any) => 
              item.game_id === 'flappy' && item.name?.includes('Радужные трубы')
            ),
            hasGoldenBird: items.some((item: any) => 
              item.game_id === 'flappy' && item.name?.includes('Золотая птичка')
            ),
          },
          hexagon: {
            hasSpacePancakes: items.some((item: any) => 
              item.game_id === 'hexagon' && item.name?.includes('Космические блины')
            ),
          },
          towers: {
            hasRainbowBlocks: items.some((item: any) => 
              item.game_id === 'towers' && item.name?.includes('Радужные блоки')
            ),
          },
          memory: {
            hasAnimalCards: items.some((item: any) => 
              item.game_id === 'memory' && item.name?.includes('Карточки со зверями')
            ),
          },
        });
      } catch (error) {
        console.error('Failed to load skins:', error);
      } finally {
        setLoading(false);
      }
    };
    
    loadSkins();
  }, []);

  return { skins, loading };
}