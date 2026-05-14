import { useDrop } from 'react-dnd';
import { 
  type HexCoord, 
  hexToPixel, 
  getHexagonPoints, 
  HEX_GRID, 
  pancakeEmoji, 
  pancakeColor, 
  EMPTY_COLOR, 
  HEX_STROKE 
} from '../../utils/hexagon';

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
  const RADIUS = 45;
  const points = getHexagonPoints(RADIUS);
  
  // Вычисляем размеры SVG
  const allPixels = HEX_GRID.map(coord => hexToPixel(coord.q, coord.r));
  const minX = Math.min(...allPixels.map(p => p.x)) - RADIUS;
  const maxX = Math.max(...allPixels.map(p => p.x)) + RADIUS;
  const minY = Math.min(...allPixels.map(p => p.y)) - RADIUS;
  const maxY = Math.max(...allPixels.map(p => p.y)) + RADIUS;
  
  const width = maxX - minX;
  const height = maxY - minY;
  const offsetX = -minX;
  const offsetY = -minY;

  // Функция для поиска гекса по координатам мыши
  const getHexAtPixel = (x: number, y: number): HexCoord | null => {
    // Переводим координаты мыши в систему координат SVG
    const svgX = x - offsetX;
    const svgY = y - offsetY;
    
    for (const coord of HEX_GRID) {
      const { x: hexX, y: hexY } = hexToPixel(coord.q, coord.r);
      const dx = svgX - hexX;
      const dy = svgY - hexY;
      const distance = Math.sqrt(dx * dx + dy * dy);
      if (distance <= RADIUS) {
        return coord;
      }
    }
    return null;
  };

  const [{ isOver }, dropRef] = useDrop(() => ({
    accept: 'pancake',
    drop: (item: any, monitor) => {
      const clientOffset = monitor.getClientOffset();
      if (!clientOffset) return;
      
      // Получаем позицию мыши относительно SVG
      const svgElement = document.querySelector('.hex-grid-svg');
      if (!svgElement) return;
      
      const rect = svgElement.getBoundingClientRect();
      const x = clientOffset.x - rect.left;
      const y = clientOffset.y - rect.top;
      
      const coord = getHexAtPixel(x, y);
      if (coord) {
        onDrop(item, coord);
      }
    },
    collect: (monitor) => ({
      isOver: !!monitor.isOver(),
    }),
  }));

  return (
    <div ref={dropRef as any} className="hex-grid-container">
      <svg 
        className="hex-grid-svg"
        width="100%" 
        height="auto" 
        viewBox={`${-50} ${-50} ${width + 100} ${height + 100}`}
        style={{ maxWidth: '900px', margin: '0 auto', cursor: isOver ? 'copy' : 'default' }}
      >
        {HEX_GRID.map((coord) => {
          const tile = tiles.find(t => t.coord.q === coord.q && t.coord.r === coord.r);
          const { x, y } = hexToPixel(coord.q, coord.r);
          const type = tile?.type || 'empty';
          const count = tile?.count || 0;
          const fillColor = type === 'empty' ? EMPTY_COLOR : (pancakeColor[type as keyof typeof pancakeColor] || '#DEB887');
          
          return (
            <g 
              key={`${coord.q},${coord.r}`} 
              transform={`translate(${x + offsetX}, ${y + offsetY})`}
              style={{ cursor: 'pointer' }}
            >
              <polygon
                points={points}
                fill={fillColor}
                stroke={HEX_STROKE}
                strokeWidth="3"
                opacity={isOver ? 0.9 : 1}
              />
              {type !== 'empty' && (
                <>
                  <text x="0" y="-8" textAnchor="middle" fill="#fff" fontSize="24" fontWeight="bold">
                    {pancakeEmoji[type as keyof typeof pancakeEmoji] || '🥞'}
                  </text>
                  <text x="0" y="24" textAnchor="middle" fill="#ffd700" fontSize="16" fontWeight="bold">
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
