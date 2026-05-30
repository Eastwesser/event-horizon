import { useDrop } from 'react-dnd';
import { useRef } from 'react';
import { 
  type HexCoord, 
  hexToPixel, 
  getHexagonPoints, 
  HEX_GRID, 
  pancakeEmoji, 
  pancakeColor, 
  EMPTY_COLOR, 
  HEX_STROKE 
} from '../../../utils/hexagon';

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
  const RADIUS = 35;
  const points = getHexagonPoints(RADIUS);
  const svgRef = useRef<SVGSVGElement>(null);
  
  const allPixels = HEX_GRID.map(coord => hexToPixel(coord.q, coord.r));
  const minX = Math.min(...allPixels.map(p => p.x)) - RADIUS;
  const maxX = Math.max(...allPixels.map(p => p.x)) + RADIUS;
  const minY = Math.min(...allPixels.map(p => p.y)) - RADIUS;
  const maxY = Math.max(...allPixels.map(p => p.y)) + RADIUS;
  
  const width = maxX - minX;
  const height = maxY - minY;
  const offsetX = -minX;
  const offsetY = -minY;

  const getHexAtPixel = (x: number, y: number): HexCoord | null => {
    // Координаты уже относительно SVG viewBox
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

  const [{ isOver }, dropRef] = useDrop(() => ({
    accept: 'pancake',
    drop: (item: any, monitor) => {
      const clientOffset = monitor.getClientOffset();
      if (!clientOffset || !svgRef.current) return;
      
      const rect = svgRef.current.getBoundingClientRect();
      const svgX = clientOffset.x - rect.left;
      const svgY = clientOffset.y - rect.top;
      
      // Получаем размеры SVG и viewBox
      const svgRect = svgRef.current.viewBox?.baseVal;
      if (svgRect) {
        const scaleX = svgRect.width / rect.width;
        const scaleY = svgRect.height / rect.height;
        const viewBoxX = svgX * scaleX + svgRect.x;
        const viewBoxY = svgY * scaleY + svgRect.y;
        
        const coord = getHexAtPixel(viewBoxX, viewBoxY);
        if (coord) {
          onDrop(item, coord);
        }
      }
    },
    collect: (monitor) => ({
      isOver: !!monitor.isOver(),
    }),
  }));

  return (
    <div ref={dropRef as any} className="hex-grid-container">
      <svg 
        ref={svgRef}
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
            <g key={`${coord.q},${coord.r}`} transform={`translate(${x + offsetX}, ${y + offsetY})`}>
              <polygon
                points={points}
                fill={fillColor}
                stroke={HEX_STROKE}
                strokeWidth="2"
                opacity={isOver ? 0.9 : 1}
              />
              {type !== 'empty' && (
                <>
                  <text x="0" y="-6" textAnchor="middle" fill="#fff" fontSize="18" fontWeight="bold" style={{ textShadow: '1px 1px 0 #000' }}>
                    {pancakeEmoji[type as keyof typeof pancakeEmoji] || '🥞'}
                  </text>
                  <text x="0" y="18" textAnchor="middle" fill="#ffd700" fontSize="12" fontWeight="bold" style={{ textShadow: '1px 1px 0 #000' }}>
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
