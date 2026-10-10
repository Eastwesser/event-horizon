// frontend/src/components/Games/Hexagon/HexGrid.tsx
import { useDrop } from 'react-dnd';
import { useEffect, useRef, useState } from 'react';
import {
  type HexCoord,
  hexToPixel,
  getHexagonPoints,
  HEX_GRID,
  pancakeEmoji,
  pancakeColor,
  EMPTY_COLOR,
  HEX_STROKE,
} from '../../../utils/hexagon';

interface HexTile {
  coord: HexCoord;
  type: string;
  count: number;
}

interface DragPancake {
  id: number;
  type: string;
  count: number;
}

interface HexGridProps {
  tiles: HexTile[];
  onDrop: (item: DragPancake, coord: HexCoord) => void;
  skinMode?: 'default' | 'space';
  boostHighlight?: boolean;
}

const spaceEmojis: Record<string, string> = {
  nutella: '🌙',
  strawberry: '⭐',
  fish: '🌌',
  sausage: '☄️',
  chicken: '🪐',
  caesar: '🌠',
  cranberry: '✨',
  pancake: '☀️',
  default: '🌌',
};

const spaceColors: Record<string, string> = {
  nutella: '#1a1228',
  strawberry: '#1c1830',
  fish: '#12182a',
  sausage: '#0f1a2e',
  chicken: '#1c1410',
  caesar: '#1a1220',
  cranberry: '#0e1a18',
  pancake: '#1a1810',
  default: '#12141c',
};

function coordKey(c: HexCoord): string {
  return `${c.q},${c.r}`;
}

function sameCoord(a: HexCoord | null | undefined, b: HexCoord): boolean {
  return !!a && a.q === b.q && a.r === b.r;
}

export function HexGrid({ tiles, onDrop, skinMode = 'default', boostHighlight = false }: HexGridProps) {
  const RADIUS = 35;
  const points = getHexagonPoints(RADIUS);
  const svgRef = useRef<SVGSVGElement>(null);
  const prevTilesRef = useRef<HexTile[] | null>(null);
  const [hoverCoord, setHoverCoord] = useState<HexCoord | null>(null);
  const [pulseKeys, setPulseKeys] = useState<Set<string>>(() => new Set());
  const [boardShake, setBoardShake] = useState(false);

  const allPixels = HEX_GRID.map((coord) => hexToPixel(coord.q, coord.r));
  const minX = Math.min(...allPixels.map((p) => p.x)) - RADIUS;
  const maxX = Math.max(...allPixels.map((p) => p.x)) + RADIUS;
  const minY = Math.min(...allPixels.map((p) => p.y)) - RADIUS;
  const maxY = Math.max(...allPixels.map((p) => p.y)) + RADIUS;

  const width = maxX - minX;
  const height = maxY - minY;
  const offsetX = -minX;
  const offsetY = -minY;

  const getHexAtPixel = (x: number, y: number): HexCoord | null => {
    for (const coord of HEX_GRID) {
      const { x: hexX, y: hexY } = hexToPixel(coord.q, coord.r);
      const dx = x - (hexX + offsetX);
      const dy = y - (hexY + offsetY);
      const distance = Math.sqrt(dx * dx + dy * dy);
      if (distance <= RADIUS) {
        return coord;
      }
    }
    return null;
  };

  const clientToViewBox = (clientX: number, clientY: number): { x: number; y: number } | null => {
    const svg = svgRef.current;
    if (!svg) return null;
    const ctm = svg.getScreenCTM();
    if (!ctm) return null;
    const pt = svg.createSVGPoint();
    pt.x = clientX;
    pt.y = clientY;
    const local = pt.matrixTransform(ctm.inverse());
    return { x: local.x, y: local.y };
  };

  const pulseHex = (coord: HexCoord) => {
    const key = coordKey(coord);
    setPulseKeys((prev) => {
      const next = new Set(prev);
      next.add(key);
      return next;
    });
    window.setTimeout(() => {
      setPulseKeys((prev) => {
        const next = new Set(prev);
        next.delete(key);
        return next;
      });
    }, 320);
  };

  const triggerShake = () => {
    setBoardShake(true);
    window.setTimeout(() => setBoardShake(false), 420);
  };

  // Only empty cells accept a drop. Same-type merge pulls from *neighbors*, not pile-on-cell.
  const isValidPlacement = (tile: HexTile | undefined, dragType: string | undefined): boolean => {
    if (!tile || !dragType) return false;
    return tile.type === 'empty';
  };

  // Collect pulse: UI-observe clear (tile → empty). Skip mass resets (new game).
  useEffect(() => {
    const prev = prevTilesRef.current;
    prevTilesRef.current = tiles;
    if (!prev) return;

    const cleared: HexCoord[] = [];
    for (const before of prev) {
      if (before.type === 'empty') continue;
      const after = tiles.find((t) => t.coord.q === before.coord.q && t.coord.r === before.coord.r);
      if (after && after.type === 'empty') {
        cleared.push(before.coord);
      }
    }
    if (cleared.length === 0 || cleared.length > 6) return;
    for (const coord of cleared) pulseHex(coord);
  }, [tiles]);

  const [{ isOver, dragItem }, dropRef] = useDrop(
    () => ({
      accept: 'pancake',
      hover: (_item: DragPancake, monitor) => {
        const clientOffset = monitor.getClientOffset();
        if (!clientOffset) {
          setHoverCoord(null);
          return;
        }
        const view = clientToViewBox(clientOffset.x, clientOffset.y);
        if (!view) {
          setHoverCoord(null);
          return;
        }
        setHoverCoord(getHexAtPixel(view.x, view.y));
      },
      drop: (item: DragPancake, monitor) => {
        const clientOffset = monitor.getClientOffset();
        if (!clientOffset) return;

        const view = clientToViewBox(clientOffset.x, clientOffset.y);
        if (!view) return;

        const coord = getHexAtPixel(view.x, view.y);
        if (!coord) return;

        const tile = tiles.find((t) => t.coord.q === coord.q && t.coord.r === coord.r);
        const valid = isValidPlacement(tile, item.type);

        onDrop(item, coord);

        if (!valid) {
          triggerShake();
          return;
        }
        pulseHex(coord);
      },
      collect: (monitor) => ({
        isOver: !!monitor.isOver(),
        dragItem: monitor.getItem() as DragPancake | null,
      }),
    }),
    [tiles, onDrop],
  );

  useEffect(() => {
    if (!isOver) setHoverCoord(null);
  }, [isOver]);

  const getPancakeEmoji = (type: string) => {
    if (skinMode === 'space') {
      return spaceEmojis[type] || spaceEmojis.default || '🌌';
    }
    return pancakeEmoji[type as keyof typeof pancakeEmoji] || '🥞';
  };

  const getPancakeColorFn = (type: string) => {
    if (skinMode === 'space') {
      return spaceColors[type] || spaceColors.default || '#6C5CE7';
    }
    return pancakeColor[type as keyof typeof pancakeColor] || '#DEB887';
  };

  const dragging = !!dragItem;

  return (
    <div
      ref={dropRef as any}
      className={['hex-grid-container', boardShake ? 'hex-grid-container--shake' : '']
        .filter(Boolean)
        .join(' ')}
    >
      <svg
        ref={svgRef}
        className="hex-grid-svg"
        width="100%"
        height="100%"
        viewBox={`${-50} ${-50} ${width + 100} ${height + 100}`}
        style={{ maxWidth: '900px', margin: '0 auto', cursor: isOver ? 'copy' : 'default' }}
      >
        {HEX_GRID.map((coord) => {
          const tile = tiles.find((t) => t.coord.q === coord.q && t.coord.r === coord.r);
          const { x, y } = hexToPixel(coord.q, coord.r);
          const type = tile?.type || 'empty';
          const count = tile?.count || 0;
          const fillColor = type === 'empty' ? EMPTY_COLOR : getPancakeColorFn(type);
          const key = coordKey(coord);
          const validTarget = dragging && isValidPlacement(tile, dragItem?.type);
          const hovered = sameCoord(hoverCoord, coord);
          const cellClass = [
            'hex-cell',
            validTarget ? 'hex-cell--valid' : '',
            validTarget && boostHighlight ? 'hex-cell--valid-boost' : '',
            dragging && hovered && validTarget ? 'hex-cell--hover-valid' : '',
            dragging && hovered && !validTarget ? 'hex-cell--hover-invalid' : '',
            pulseKeys.has(key) ? 'hex-cell--pulse' : '',
          ]
            .filter(Boolean)
            .join(' ');

          return (
            <g key={key} transform={`translate(${x + offsetX}, ${y + offsetY})`}>
              <g className={cellClass}>
                <polygon
                  points={points}
                  fill={fillColor}
                  stroke={HEX_STROKE}
                  strokeWidth="2"
                  opacity={isOver && !dragging ? 0.9 : 1}
                />
                {type !== 'empty' && (
                  <>
                    <text
                      x="0"
                      y="-6"
                      textAnchor="middle"
                      fill="#fff"
                      fontSize="18"
                      fontWeight="bold"
                      style={{ textShadow: '1px 1px 0 #000' }}
                    >
                      {getPancakeEmoji(type)}
                    </text>
                    <text
                      x="0"
                      y="18"
                      textAnchor="middle"
                      fill="#ffd700"
                      fontSize="12"
                      fontWeight="bold"
                      style={{ textShadow: '1px 1px 0 #000' }}
                    >
                      x{count}
                    </text>
                  </>
                )}
              </g>
            </g>
          );
        })}
      </svg>
    </div>
  );
}
