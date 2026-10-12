import { getHexagonPoints, pancakeEmoji, type PancakeType } from '../../../utils/hexagon';

export const CHIP_FACE: Record<PancakeType, string> = {
  nutella: '#6b3a24',
  strawberry: '#c45c6a',
  fish: '#3d7a8c',
  sausage: '#a85a4a',
  chicken: '#c9a227',
  caesar: '#5a8f62',
  cranberry: '#a05070',
  pancake: '#d4b896',
};

export const CHIP_SPACE: Record<PancakeType, string> = {
  nutella: '#2a1840',
  strawberry: '#3a2458',
  fish: '#1a2848',
  sausage: '#241830',
  chicken: '#3d2a10',
  caesar: '#143028',
  cranberry: '#2a1830',
  pancake: '#2a2418',
};

export function chipFill(type: string, space: boolean): string {
  const key = type as PancakeType;
  const table = space ? CHIP_SPACE : CHIP_FACE;
  return table[key] ?? (space ? '#1a2030' : '#7a6a58');
}

function shade(hex: string, amt: number): string {
  const n = parseInt(hex.slice(1), 16);
  const r = Math.max(0, Math.min(255, ((n >> 16) & 255) + amt));
  const g = Math.max(0, Math.min(255, ((n >> 8) & 255) + amt));
  const b = Math.max(0, Math.min(255, (n & 255) + amt));
  return `#${((r << 16) | (g << 8) | b).toString(16).padStart(6, '0')}`;
}

function emojiFor(type: string): string {
  return pancakeEmoji[type as PancakeType] ?? '🥞';
}

export function HexPancakeStack({
  type,
  count,
  radius,
  space,
}: {
  type: string;
  count: number;
  radius: number;
  space: boolean;
}) {
  const visible = Math.min(4, Math.max(1, count));
  const fill = chipFill(type, space);
  const chipR = radius * 0.72;
  const gap = 5;
  const lowestCenterY = radius - chipR;
  const points = getHexagonPoints(chipR);
  const topCenterY = lowestCenterY - (visible - 1) * gap;

  return (
    <g>
      {Array.from({ length: visible }, (_, i) => {
        const centerY = lowestCenterY - i * gap;
        const layerFill = i === visible - 1 ? fill : shade(fill, -16 * (visible - 1 - i));
        return (
          <polygon
            key={i}
            points={points}
            fill={layerFill}
            stroke={shade(fill, -40)}
            strokeWidth="1.1"
            transform={`translate(0, ${centerY})`}
          />
        );
      })}
      <text
        x="0"
        y={topCenterY + 5}
        textAnchor="middle"
        fontSize="16"
        style={{ pointerEvents: 'none' }}
      >
        {emojiFor(type)}
      </text>
      {count > 1 && (
        <text
          x="0"
          y={lowestCenterY + chipR - 4}
          textAnchor="middle"
          fill="#f6ead0"
          fontSize="9"
          fontWeight="600"
          style={{ fontFamily: 'ui-rounded, "Trebuchet MS", sans-serif' }}
        >
          ×{count}
        </text>
      )}
    </g>
  );
}

export function TrayPancakeStack({
  type,
  count,
  space,
}: {
  type: string;
  count: number;
  space: boolean;
}) {
  const visible = Math.min(4, Math.max(1, count));
  const fill = chipFill(type, space);
  const x = 4;
  const w = 48;
  const h = 13;
  return (
    <svg viewBox="0 0 56 60" width="56" height="56" aria-hidden>
      {Array.from({ length: visible }, (_, i) => {
        const fromBottom = i;
        const y = 32 - fromBottom * 7;
        const layerFill = i === visible - 1 ? fill : shade(fill, -16 * (visible - 1 - i));
        return (
          <rect
            key={i}
            x={x}
            y={y}
            width={w}
            height={h}
            rx="4"
            fill={layerFill}
            stroke={shade(fill, -36)}
          />
        );
      })}
      <text
        x="28"
        y={32 - (visible - 1) * 7 + 11}
        textAnchor="middle"
        fontSize="14"
      >
        {emojiFor(type)}
      </text>
      <text
        x="28"
        y="56"
        textAnchor="middle"
        fill="#ead9c0"
        fontSize="11"
        fontWeight="600"
        style={{ fontFamily: 'ui-rounded, "Trebuchet MS", sans-serif' }}
      >
        ×{count}
      </text>
    </svg>
  );
}
