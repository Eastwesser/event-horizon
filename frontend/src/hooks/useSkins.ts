// frontend/src/hooks/useSkins.ts
import { useEffect, useState } from 'react';
import { getInventory } from '../services/api';

export interface GameSkins {
  flappy: {
    hasCosmicPipes: boolean;
    hasGoldenBird: boolean;
  };
  hexagon: {
    hasSpacePancakes: boolean;
  };
  towers: {
    hasCosmicBlocks: boolean;
  };
  memory: {
    hasAnimalCards: boolean;
  };
}

const EMPTY_SKINS: GameSkins = {
  flappy: { hasCosmicPipes: false, hasGoldenBird: false },
  hexagon: { hasSpacePancakes: false },
  towers: { hasCosmicBlocks: false },
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

function nameMatches(name: string | undefined, needles: string[]): boolean {
  if (!name) return false;
  const n = name.toLowerCase();
  return needles.some((needle) => n.includes(needle.toLowerCase()));
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
            hasCosmicPipes: items.some(
              (item) =>
                item.game_id === 'flappy' &&
                nameMatches(item.name, ['космические трубы', 'радужные трубы']),
            ),
            hasGoldenBird: items.some(
              (item) => item.game_id === 'flappy' && nameMatches(item.name, ['золотая птичка']),
            ),
          },
          hexagon: {
            hasSpacePancakes: items.some(
              (item) =>
                item.game_id === 'hexagon' && nameMatches(item.name, ['космические блины']),
            ),
          },
          towers: {
            hasCosmicBlocks: items.some(
              (item) =>
                item.game_id === 'towers' &&
                nameMatches(item.name, ['космические блоки', 'радужные блоки']),
            ),
          },
          memory: {
            hasAnimalCards: items.some(
              (item) =>
                item.game_id === 'memory' && nameMatches(item.name, ['карточки со зверями']),
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
