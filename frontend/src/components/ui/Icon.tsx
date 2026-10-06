import type { ReactNode, SVGProps } from 'react';
import { cn } from '../../lib/cn';

/** Chrome + Berserk + game-tile icons (Wave 1 Emoji → SVG). In-game UI keeps emoji for polish wave. */
export type IconName =
  | 'user'
  | 'trophy'
  | 'cart'
  | 'package'
  | 'scroll'
  | 'pen'
  | 'credit-card'
  | 'wrench'
  | 'chart'
  | 'gift'
  | 'palette'
  | 'ticket'
  | 'backpack'
  | 'lock'
  | 'key'
  | 'frame'
  | 'sparkle'
  | 'bell'
  | 'check'
  | 'x'
  | 'info'
  | 'warning'
  | 'bird'
  | 'hex'
  | 'tower'
  | 'cards'
  | 'armor'
  | 'counter'
  | 'uchr'
  | 'tap'
  | 'strike'
  | 'instant'
  | 'zov'
  | 'zoz'
  | 'zot'
  | 'zor'
  | 'zoal'
  | 'zom'
  | 'zoo'
  | 'regen'
  | 'stamina'
  | 'direct'
  | 'ova'
  | 'ovz'
  | 'ovs'
  | 'diamond'
  | 'hanoi'
  | 'twenty48'
  | 'gears'
  | 'star'
  | 'lamp'
  | 'undo'
  | 'crown'
  | 'medal';

type IconProps = {
  name: IconName;
  className?: string;
  title?: string;
} & Omit<SVGProps<SVGSVGElement>, 'name'>;

const stroke = {
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 1.75,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
};

function Glyph({ name }: { name: IconName }) {
  switch (name) {
    case 'user':
      return (
        <g {...stroke}>
          <circle cx="12" cy="8" r="3.5" />
          <path d="M5 19.5c1.5-3.5 3.8-5 7-5s5.5 1.5 7 5" />
        </g>
      );
    case 'trophy':
      return (
        <g {...stroke}>
          <path d="M8 4h8v3a4 4 0 0 1-8 0V4Z" />
          <path d="M8 6H5.5A2.5 2.5 0 0 0 5 11c1.2 1.2 2.5 1.5 3 1.5" />
          <path d="M16 6h2.5A2.5 2.5 0 0 1 19 11c-1.2 1.2-2.5 1.5-3 1.5" />
          <path d="M12 13v3M9 20h6M10 20v-2a2 2 0 0 1 2-2 2 2 0 0 1 2 2v2" />
        </g>
      );
    case 'cart':
      return (
        <g {...stroke}>
          <path d="M3.5 5h1.8l1.4 10.2a1.5 1.5 0 0 0 1.5 1.3h8.3a1.5 1.5 0 0 0 1.5-1.2L19.5 8H7" />
          <circle cx="9.5" cy="19.5" r="1.2" />
          <circle cx="16.5" cy="19.5" r="1.2" />
        </g>
      );
    case 'package':
      return (
        <g {...stroke}>
          <path d="M12 3.5 20 8v8l-8 4.5L4 16V8l8-4.5Z" />
          <path d="M12 12v8.5M12 12 20 8M12 12 4 8" />
        </g>
      );
    case 'scroll':
      return (
        <g {...stroke}>
          <path d="M7 5.5h8.5A2.5 2.5 0 0 1 18 8v10.5H8.5A2.5 2.5 0 0 0 6 21" />
          <path d="M6 5.5A2.5 2.5 0 0 0 6 10.5h1" />
          <path d="M10 10h5M10 13.5h5M10 17h3" />
        </g>
      );
    case 'pen':
      return (
        <g {...stroke}>
          <path d="M14.5 5.5 18.5 9.5 9 19H5v-4L14.5 5.5Z" />
          <path d="M12.5 7.5 16.5 11.5" />
        </g>
      );
    case 'credit-card':
      return (
        <g {...stroke}>
          <rect x="3.5" y="6" width="17" height="12" rx="2" />
          <path d="M3.5 10.5h17M7 15h3" />
        </g>
      );
    case 'wrench':
      return (
        <g {...stroke}>
          <path d="M14.5 6.5a3.5 3.5 0 0 0-4.7 4.7L4.5 16.5 7.5 19.5l5.3-5.3a3.5 3.5 0 0 0 4.7-4.7l-2.5 2.5-2.5-2.5 2.5-2.5Z" />
        </g>
      );
    case 'chart':
      return (
        <g {...stroke}>
          <path d="M4.5 19.5h15" />
          <path d="M7 16v-4.5M12 16V7.5M17 16v-7" />
        </g>
      );
    case 'gift':
      return (
        <g {...stroke}>
          <rect x="4.5" y="10" width="15" height="10" rx="1.5" />
          <path d="M12 10v10M4.5 14h15" />
          <path d="M12 10c-2-3.5-5.5-3-5.5-1S9 11 12 10c2-3.5 5.5-3 5.5-1S15 11 12 10Z" />
        </g>
      );
    case 'palette':
      return (
        <g {...stroke}>
          <path d="M12 4.5a7.5 7.5 0 1 0 0 15h1.5a2.5 2.5 0 0 0 0-5H13a1.5 1.5 0 0 1 0-3h.5A7.5 7.5 0 0 0 12 4.5Z" />
          <circle cx="8" cy="10" r="0.9" fill="currentColor" stroke="none" />
          <circle cx="10.5" cy="7.5" r="0.9" fill="currentColor" stroke="none" />
          <circle cx="14" cy="7.8" r="0.9" fill="currentColor" stroke="none" />
        </g>
      );
    case 'ticket':
      return (
        <g {...stroke}>
          <path d="M4.5 8.5A1.5 1.5 0 0 0 6 10v0a1.5 1.5 0 0 1 0 3v0a1.5 1.5 0 0 0-1.5 1.5v0A1.5 1.5 0 0 0 6 16h12a1.5 1.5 0 0 0 1.5-1.5v0a1.5 1.5 0 0 0-1.5-1.5v0a1.5 1.5 0 0 1 0-3v0A1.5 1.5 0 0 0 18 8.5H6A1.5 1.5 0 0 0 4.5 8.5Z" />
          <path d="M10 9v6" strokeDasharray="1.5 2" />
        </g>
      );
    case 'backpack':
      return (
        <g {...stroke}>
          <path d="M8 8.5V7a4 4 0 0 1 8 0v1.5" />
          <rect x="6" y="8.5" width="12" height="11.5" rx="2.5" />
          <path d="M9.5 13.5h5v3h-5z" />
        </g>
      );
    case 'lock':
      return (
        <g {...stroke}>
          <rect x="6" y="11" width="12" height="9" rx="2" />
          <path d="M8.5 11V8a3.5 3.5 0 0 1 7 0v3" />
        </g>
      );
    case 'key':
      return (
        <g {...stroke}>
          <circle cx="8" cy="12" r="3.5" />
          <path d="M11.5 12H20l-1.5 2H16l-.8 1.5" />
        </g>
      );
    case 'frame':
      return (
        <g {...stroke}>
          <rect x="4.5" y="5.5" width="15" height="13" rx="1.5" />
          <path d="M4.5 15l4-3.5 3 2.5 3.5-4 4.5 5" />
        </g>
      );
    case 'sparkle':
      return (
        <g {...stroke}>
          <path d="M12 4.5v4M12 15.5v4M4.5 12h4M15.5 12h4" />
          <path d="M7.5 7.5 9.5 9.5M14.5 14.5l2 2M16.5 7.5 14.5 9.5M9.5 14.5l-2 2" />
        </g>
      );
    case 'bell':
      return (
        <g {...stroke}>
          <path d="M6.5 16.5h11" />
          <path d="M8 16.5V10a4 4 0 0 1 8 0v6.5" />
          <path d="M10.5 16.5a1.5 1.5 0 0 0 3 0" />
          <path d="M12 5.5V4.5" />
        </g>
      );
    case 'check':
      return (
        <g {...stroke}>
          <path d="M5 12.5 10 17.5 19 7.5" />
        </g>
      );
    case 'x':
      return (
        <g {...stroke}>
          <path d="M6.5 6.5 17.5 17.5M17.5 6.5 6.5 17.5" />
        </g>
      );
    case 'info':
      return (
        <g {...stroke}>
          <circle cx="12" cy="12" r="8.5" />
          <path d="M12 10.5v6M12 7.5h.01" />
        </g>
      );
    case 'warning':
      return (
        <g {...stroke}>
          <path d="M12 4.5 20.5 19H3.5L12 4.5Z" />
          <path d="M12 10v4.5M12 17h.01" />
        </g>
      );
    case 'bird':
      return (
        <g {...stroke}>
          <path d="M4.5 14c2-4 5-6 8.5-6 1.5 0 3 .5 4.5 1.5L20 8l-1 3.5c.5 1 .8 2.2.8 3.5 0 1.5-.5 2.5-1.3 3" />
          <path d="M9 11.5c1.5.5 3 1.5 4 3" />
        </g>
      );
    case 'hex':
      return (
        <g {...stroke}>
          <path d="M12 3.5 19 8v8l-7 4.5L5 16V8l7-4.5Z" />
        </g>
      );
    case 'tower':
      return (
        <g {...stroke}>
          <path d="M8 20V9l2-2V5h1v1h2V5h1v2l2 2v11" />
          <path d="M7 20h10M10 13h4M10 16.5h4" />
        </g>
      );
    case 'cards':
      return (
        <g {...stroke}>
          <rect x="7" y="5" width="10" height="14" rx="1.5" />
          <path d="M5.5 7.5 4 8.5v11a1.5 1.5 0 0 0 1.5 1.5H15" />
          <path d="M10 10h4M10 13h4" />
        </g>
      );
    case 'armor':
      return (
        <g {...stroke}>
          <path d="M12 4.5 18.5 7v5c0 4-2.8 6.8-6.5 8.5C8.3 18.8 5.5 16 5.5 12V7L12 4.5Z" />
          <path d="M9 12.5h6" />
        </g>
      );
    case 'counter':
      return <circle cx="12" cy="12" r="5.5" {...stroke} />;
    case 'uchr':
      return (
        <g {...stroke}>
          <path d="M8 7v4c0 1.5-.8 2.5-2 3.5" />
          <path d="M8 11h3" />
          <path d="M16 17v-4c0-1.5.8-2.5 2-3.5" />
          <path d="M16 13h-3" />
        </g>
      );
    case 'tap':
      return (
        <g {...stroke}>
          <path d="M12 5.5a6.5 6.5 0 1 1-4.6 1.9" />
          <path d="M12 5.5V9l3-1.5" />
        </g>
      );
    case 'strike':
      return (
        <g {...stroke}>
          <path d="M5 19 15.5 8.5" />
          <path d="M14 7l3 3 2-2-3-3-2 2Z" />
          <path d="M5 19l-1.5 1.5" />
        </g>
      );
    case 'instant':
      return (
        <g {...stroke}>
          <path d="M12 4.5v5l2.5 1.5" />
          <path d="M12 19.5v-5l-2.5-1.5" />
          <path d="M8 6.5c-2 1.5-3.5 3.5-3.5 5.5s1.5 4 3.5 5.5" />
          <path d="M16 6.5c2 1.5 3.5 3.5 3.5 5.5s-1.5 4-3.5 5.5" />
        </g>
      );
    case 'zov':
      return (
        <g {...stroke}>
          <path d="M5 12h9l5-3.5" />
          <path d="M14 12l5 3.5" />
        </g>
      );
    case 'zoz':
      return (
        <g {...stroke}>
          <path d="M7 5.5h10v13l-2-1.2-3 1.7-3-1.7-2 1.2v-13Z" />
          <path d="M9.5 10h5M9.5 13.5h5" />
        </g>
      );
    case 'zot':
      return (
        <g {...stroke}>
          <path d="M6 18.5 14 8.5" />
          <path d="M12.5 7l4 4 2-1.5-4.5-4.5L12.5 7Z" />
        </g>
      );
    case 'zor':
      return (
        <g {...stroke}>
          <path d="M13 4.5 8 12.5h4l-1 7 6-9h-4l2-6Z" />
        </g>
      );
    case 'zoal':
      return (
        <g {...stroke}>
          <path d="M12 12c-3-1.5-6-1-8 .5 2.5 1 5.5 1 8-.5Z" />
          <path d="M12 12c3-1.5 6-1 8 .5-2.5 1-5.5 1-8-.5Z" />
          <path d="M12 11.5v3" />
        </g>
      );
    case 'zom':
      return (
        <g {...stroke}>
          <circle cx="12" cy="12" r="7" />
          <path d="M12 5a7 7 0 0 1 0 14" />
          <circle cx="12" cy="8.5" r="1.2" fill="currentColor" stroke="none" />
          <circle cx="12" cy="15.5" r="1.2" fill="none" stroke="currentColor" strokeWidth="1.5" />
        </g>
      );
    case 'zoo':
      return (
        <g {...stroke}>
          <path d="M12 5.5c2.8 0 5 2.4 5 5.4 0 3.4-3.2 6.6-5 8.1-1.8-1.5-5-4.7-5-8.1 0-3 2.2-5.4 5-5.4Z" />
        </g>
      );
    case 'regen':
      return (
        <g {...stroke}>
          <path d="M8 10.5c0-2 1.5-3.5 3.5-3.5S15 8.5 15 10.5c0 3-3 5-3 7.5" />
          <path d="M9 14h6M8.5 17h7" />
          <path d="M7 9.5 5.5 8M17 9.5 18.5 8" />
        </g>
      );
    case 'stamina':
      return (
        <g {...stroke}>
          <path d="M8.5 20V12l-2.5-1.5V8l4-1.5L12 9l2-2.5 4 1.5v2.5L15.5 12V20" />
          <path d="M9.5 12.5h5" />
        </g>
      );
    case 'direct':
      return (
        <g {...stroke}>
          <circle cx="12" cy="12" r="3" />
          <circle cx="12" cy="12" r="7" />
          <path d="M12 3.5v2.5M12 18v2.5M3.5 12h2.5M18 12h2.5" />
        </g>
      );
    case 'ova':
      return (
        <g {...stroke}>
          <path d="M6 18 14.5 9.5" />
          <path d="M13 8l3.5 3.5 2-2L15 6l-2 2Z" />
        </g>
      );
    case 'ovz':
      return (
        <g {...stroke}>
          <path d="M12 5 18 7.5v4.5c0 3.2-2.4 5.5-6 7-3.6-1.5-6-3.8-6-7V7.5L12 5Z" />
        </g>
      );
    case 'ovs':
      return (
        <g {...stroke}>
          <path d="M5 13h8l5-2.5" />
          <path d="M13 13l4 2.5" />
          <circle cx="7.5" cy="13" r="1.2" />
        </g>
      );
    case 'diamond':
      return (
        <g {...stroke}>
          <path d="M12 4.5 19 12l-7 7.5L5 12l7-7.5Z" />
        </g>
      );
    case 'hanoi':
      return (
        <g {...stroke}>
          <path d="M5 19h14" />
          <path d="M7 16h10" />
          <path d="M9 13h6" />
          <path d="M12 5v11" />
          <path d="M10 5h4" />
        </g>
      );
    case 'twenty48':
      return (
        <g {...stroke}>
          <rect x="4.5" y="4.5" width="7" height="7" rx="1" />
          <rect x="12.5" y="4.5" width="7" height="7" rx="1" />
          <rect x="4.5" y="12.5" width="7" height="7" rx="1" />
          <rect x="12.5" y="12.5" width="7" height="7" rx="1" />
          <path d="M6.5 8h3M14.5 8h3M6.5 16h3M14.5 16h3" />
        </g>
      );
    case 'gears':
      return (
        <g {...stroke}>
          <circle cx="9.5" cy="12" r="3" />
          <circle cx="15.5" cy="10" r="2.5" />
          <path d="M9.5 7.5v1M9.5 15.5v1M6 12h1M12 12h1M7.2 9.2l.7.7M11.1 14.1l.7.7M7.2 14.8l.7-.7M11.1 9.9l.7-.7" />
          <path d="M15.5 6.5v1M15.5 12.5v1M12.8 10h1M17.2 10h1" />
        </g>
      );
    case 'star':
      return (
        <g {...stroke}>
          <path d="M12 4.5 13.8 9.2 18.8 9.5 15 12.8 16.2 17.8 12 15.2 7.8 17.8 9 12.8 5.2 9.5 10.2 9.2 12 4.5Z" />
        </g>
      );
    case 'lamp':
      return (
        <g {...stroke}>
          <path d="M9 14.5c0-2.5-2-3.5-2-6a5 5 0 0 1 10 0c0 2.5-2 3.5-2 6" />
          <path d="M10 17h4M10.5 19.5h3" />
        </g>
      );
    case 'undo':
      return (
        <g {...stroke}>
          <path d="M9 8H5V4" />
          <path d="M5 8a7 7 0 1 1-1 4.5" />
        </g>
      );
    case 'crown':
      return (
        <g {...stroke}>
          <path d="M5 16.5h14l-1.5-9-3.5 3.5L12 6l-2 5-3.5-3.5L5 16.5Z" />
          <path d="M6.5 18.5h11" />
        </g>
      );
    case 'medal':
      return (
        <g {...stroke}>
          <circle cx="12" cy="13" r="5" />
          <path d="M9 8.5 7.5 4.5h3L12 7l1.5-2.5h3L15 8.5" />
        </g>
      );
    default:
      return null;
  }
}

export function Icon({ name, className, title, ...rest }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={cn('inline-block shrink-0', className)}
      aria-hidden={title ? undefined : true}
      role={title ? 'img' : undefined}
      {...rest}
    >
      {title ? <title>{title}</title> : null}
      <Glyph name={name} />
    </svg>
  );
}

/** Inline label with a leading icon — nav chips, filter tabs, badges. */
export function IconLabel({
  name,
  children,
  className,
  iconClassName = 'h-3.5 w-3.5',
}: {
  name: IconName;
  children: ReactNode;
  className?: string;
  iconClassName?: string;
}) {
  return (
    <span className={cn('inline-flex items-center gap-1.5', className)}>
      <Icon name={name} className={iconClassName} />
      <span>{children}</span>
    </span>
  );
}
