import { useDrop } from 'react-dnd';
import { type HexCoord, hexToPixel, HEX_GRID, pancakeEmoji, pancakeColor } from '../../utils/hexagon';

interface HexTile {
  coord: HexCoord;
  type: string;
  count: number;
}

interface HexGridProps {
  tiles: HexTile[];
  onDrop: (item: any, coord: HexCoord) => void;
}

export function HexGrid({ tiles, onDrop }: HexGridProps) {
  const [{ isOver }, dropRef] = useDrop(() => ({
    accept: 'pancake',
    drop: (item: any, monitor) => {
      // Для простоты пока используем координаты из item
      // TODO: нормальное определение координат по позиции мыши
      const coord = { q: 0, r: 0 };
      onDrop(item, coord);
    },
    collect: (monitor) => ({
      isOver: !!monitor.isOver(),
    }),
  }));

  return (
    <div ref={dropRef as any}>
      <svg viewBox="-300 -300 600 600" style={{ width: '100%', maxWidth: '800px', margin: '0 auto' }}>  
        {HEX_GRID.map((coord) => {
          const tile = tiles.find(t => t.coord.q === coord.q && t.coord.r === coord.r);
          const { x, y } = hexToPixel(coord.q, coord.r);
          const type = tile?.type || 'empty';
          const count = tile?.count || 0;
          
          return (
            <g key={`${coord.q},${coord.r}`} transform={`translate(${x}, ${y})`}>
              <polygon
                points="0,-35 30,-17 30,17 0,35 -30,17 -30,-17"
                fill={type === 'empty' ? '#2a2a4a' : (pancakeColor[type as keyof typeof pancakeColor] || '#DEB887')}
                stroke="#ffd700"
                strokeWidth="2"
                style={{ cursor: 'pointer', opacity: isOver ? 0.8 : 1 }}
              />
              {type !== 'empty' && (
                <>
                  <text x="0" y="-8" textAnchor="middle" fill="#fff" fontSize="20">
                    {pancakeEmoji[type as keyof typeof pancakeEmoji] || '🥞'}
                  </text>
                  <text x="0" y="20" textAnchor="middle" fill="#ffd700" fontSize="12" fontWeight="bold">
                    x{count}
                  </text>
                </>
              )}
            </g>
          );
        })}
      </svg>
    </div>
  );
}
