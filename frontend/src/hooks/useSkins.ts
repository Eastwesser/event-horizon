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

const EMPTY_SKINS: GameSkins = {
  flappy: { hasRainbowPipes: false, hasGoldenBird: false },
  hexagon: { hasSpacePancakes: false },
  towers: { hasRainbowBlocks: false },
  memory: { hasAnimalCards: false },
};

function normalizeInventoryPayload(data: unknown): unknown[] {
  if (data == null) return [];
  if (Array.isArray(data)) return data;
  if (typeof data === 'object' && data !== null && Array.isArray((data as { items?: unknown }).items)) {
    return (data as { items: unknown[] }).items;
  }
  return [];
}

export function useSkins() {
  const [skins, setSkins] = useState<GameSkins>(EMPTY_SKINS);
  const [loading, setLoading] = useState(true);
  /** True when the user owns zero skins (empty inventory). */
  const [empty, setEmpty] = useState(false);

  useEffect(() => {
    const loadSkins = async () => {
      try {
        const response = await getInventory();
        const items = normalizeInventoryPayload(response.data) as Array<{
          game_id?: string;
          name?: string;
        }>;

        setEmpty(items.length === 0);

        setSkins({
          flappy: {
            hasRainbowPipes: items.some(
              (item) => item.game_id === 'flappy' && item.name?.includes('Радужные трубы'),
            ),
            hasGoldenBird: items.some(
              (item) => item.game_id === 'flappy' && item.name?.includes('Золотая птичка'),
            ),
          },
          hexagon: {
            hasSpacePancakes: items.some(
              (item) => item.game_id === 'hexagon' && item.name?.includes('Космические блины'),
            ),
          },
          towers: {
            hasRainbowBlocks: items.some(
              (item) => item.game_id === 'towers' && item.name?.includes('Радужные блоки'),
            ),
          },
          memory: {
            hasAnimalCards: items.some(
              (item) => item.game_id === 'memory' && item.name?.includes('Карточки со зверями'),
            ),
          },
        });
      } catch (error) {
        console.error('Failed to load skins:', error);
        setSkins(EMPTY_SKINS);
        setEmpty(true);
      } finally {
        setLoading(false);
      }
    };

    void loadSkins();
  }, []);

  return { skins, loading, empty, emptyMessage: empty ? 'У вас пока нет скинов' : null };
}
