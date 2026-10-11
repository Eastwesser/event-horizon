// Original companion drawings. Species silhouettes stay distinct; color is the player's.
import type { ReactNode } from 'react';

export type Species = 'звезда' | 'кот' | 'дракон' | 'лисичка' | 'зайчик';

/** Care-streak days → growth stage. This is the evolution players see. */
export type Growth = 1 | 2 | 3;

export function growthFromStreak(streak: number): Growth {
  if (streak >= 21) return 3;
  if (streak >= 7) return 2;
  return 1;
}

export function growthLabel(g: Growth): string {
  if (g === 3) return 'вырос';
  if (g === 2) return 'растёт';
  return 'малыш';
}

export type MementoId =
  | 'lamp'
  | 'plant'
  | 'books'
  | 'pillow'
  | 'gear'
  | 'ticket'
  | 'spark'
  | 'bowl';

export const MEMENTOS: { id: MementoId; name: string; line: string }[] = [
  { id: 'lamp', name: 'Лампа', line: 'Свет в углу, когда за окном темно.' },
  { id: 'plant', name: 'Росток', line: 'Маленький, но уже тянется к окну.' },
  { id: 'books', name: 'Три книги', line: 'Их можно не дочитывать. Они всё равно тут.' },
  { id: 'pillow', name: 'Лунная подушка', line: 'Для сна, не для счёта.' },
  { id: 'gear', name: 'Игрушечная шестерня', line: 'Крутится, если толкнуть лапой.' },
  { id: 'ticket', name: 'Билетик', line: 'На память. В магазин с ним не ходят.' },
  { id: 'spark', name: 'Осколок звезды', line: 'Тёплый. Лежит на тумбе.' },
  { id: 'bowl', name: 'Миска', line: 'Пустая миска тоже забота.' },
];

export function nextMemento(owned: readonly string[]): MementoId | null {
  return MEMENTOS.find((m) => !owned.includes(m.id))?.id ?? null;
}

export function mementoById(id: string) {
  return MEMENTOS.find((m) => m.id === id);
}

const INK = '#2a2118';

function Face({
  asleep,
  lx,
  ly,
  rx,
  ry,
  mouth,
}: {
  asleep: boolean;
  lx: number;
  ly: number;
  rx: number;
  ry: number;
  mouth: string;
}) {
  return (
    <g stroke={INK} strokeLinecap="round">
      {asleep ? (
        <path
          d={`M${lx - 5} ${ly} q5 4 10 0 M${rx - 5} ${ry} q5 4 10 0`}
          fill="none"
          strokeWidth="2.2"
        />
      ) : (
        <>
          <ellipse cx={lx} cy={ly} rx="5.2" ry="4.6" fill="#fffaf2" strokeWidth="2" />
          <ellipse cx={rx} cy={ry} rx="5.2" ry="4.6" fill="#fffaf2" strokeWidth="2" />
          <circle cx={lx + 0.8} cy={ly + 0.3} r="1.8" fill={INK} stroke="none" />
          <circle cx={rx + 0.8} cy={ry + 0.3} r="1.8" fill={INK} stroke="none" />
        </>
      )}
      <path d={mouth} fill="none" strokeWidth="2" />
    </g>
  );
}

/** Same drawing for all ages — grow the whole group from the feet. */
function scaleForGrowth(g: Growth): number {
  if (g === 1) return 0.72;
  if (g === 3) return 1.22;
  return 1;
}

export function CompanionPet({
  species,
  color,
  asleep = false,
  growth = 2,
}: {
  species: Species;
  color: string;
  asleep?: boolean;
  growth?: Growth;
}) {
  const ink = {
    fill: color,
    stroke: INK,
    strokeWidth: 3,
    strokeLinejoin: 'round' as const,
    strokeLinecap: 'round' as const,
  };
  const s = scaleForGrowth(growth);

  let body: ReactNode;
  if (species === 'дракон') {
    body = (
      <g>
        <path {...ink} d="M92 96c16-2 28 6 26 18-10-2-18-6-22-12z" />
        <path {...ink} d="M114 108l8 10-10 4-4-12z" />
        <path {...ink} d="M46 80L18 52c-2 18 6 32 16 38 6-2 10-4 12-10z" />
        <path {...ink} d="M94 80L122 52c2 18-6 32-16 38-6-2-10-4-12-10z" />
        <path {...ink} d="M42 88c0-16 12-28 28-28s28 12 28 28c0 18-12 30-28 30s-28-12-28-30z" />
        <ellipse cx="58" cy="116" rx="8" ry="5" fill={color} stroke={INK} strokeWidth="2.4" />
        <ellipse cx="82" cy="116" rx="8" ry="5" fill={color} stroke={INK} strokeWidth="2.4" />
        <path {...ink} d="M46 52c0-16 10-28 24-28s24 12 24 28c0 12-8 20-24 20s-24-8-24-20z" />
        <path {...ink} d="M58 32c-1-14 8-16 10-4 0 6-4 8-8 6z" />
        <path {...ink} d="M82 32c1-14-8-16-10-4 0 6 4 8 8 6z" />
        <circle cx="56" cy="56" r="3.4" fill="#f3b4c4" />
        <circle cx="84" cy="56" r="3.4" fill="#f3b4c4" />
        <Face asleep={asleep} lx={60} ly={50} rx={80} ry={50} mouth="M66 62 q4 4 8 0" />
      </g>
    );
  } else if (species === 'кот') {
    body = (
      <g>
        <path {...ink} d="M52 56c-4-24 6-38 14-18 3 8 2 14-2 18z" />
        <path {...ink} d="M88 56c4-24-6-38-14-18-3 8-2 14 2 18z" />
        <path d="M56 42c2 8 6 14 8 10" fill="#f7c2d0" />
        <path d="M84 42c-2 8-6 14-8 10" fill="#f7c2d0" />
        <path {...ink} d="M98 88c16 4 20 16 6 18-8-2-14-8-16-14z" />
        <path {...ink} d="M40 78c0-20 14-32 30-32s30 12 30 32c2 22-12 36-30 36s-32-14-30-36z" />
        {asleep ? (
          <path
            d="M57 70 q5 4 10 0 M75 70 q5 4 10 0"
            fill="none"
            stroke={INK}
            strokeWidth="2.2"
            strokeLinecap="round"
          />
        ) : (
          <>
            <ellipse cx="62" cy="70" rx="5.2" ry="4.6" fill="#fffaf2" stroke={INK} strokeWidth="2" />
            <ellipse cx="80" cy="70" rx="5.2" ry="4.6" fill="#fffaf2" stroke={INK} strokeWidth="2" />
            <circle cx="63" cy="70.4" r="1.8" fill={INK} />
            <circle cx="81" cy="70.4" r="1.8" fill={INK} />
          </>
        )}
        <circle cx="54" cy="84" r="5" fill="#f3b4c4" />
        <circle cx="88" cy="84" r="5" fill="#f3b4c4" />
        <path d="M66 77h8l-4 6z" fill="#e07a8a" stroke={INK} strokeWidth="1.4" strokeLinejoin="round" />
        <path
          d="M62 86 q5 8 9 0 q5 8 9 0"
          fill="none"
          stroke={INK}
          strokeWidth="2.4"
          strokeLinecap="round"
        />
        <path
          d="M48 80h-10 M48 86h-8 M92 80h10 M92 86h8"
          fill="none"
          stroke={INK}
          strokeWidth="1.5"
          strokeLinecap="round"
        />
      </g>
    );
  } else if (species === 'лисичка') {
    body = (
      <g>
        {/* Хвост-кисточка справа, сужается вверх */}
        <path
          {...ink}
          d="M82 108c22-2 46-20 44-44-2-16-20-14-28 4-10 20-18 34-20 40-1 4 0 4 4 0z"
        />

        {/* Сидячее тело */}
        <path
          {...ink}
          d="M48 90c-12 2-16 24-2 32 8 6 24 8 36 4 14-4 20-22 10-32-8-8-28-8-44-4z"
        />
        <ellipse cx="56" cy="122" rx="8" ry="5" fill={color} stroke={INK} strokeWidth="2.4" />
        <ellipse cx="84" cy="122" rx="8" ry="5" fill={color} stroke={INK} strokeWidth="2.4" />

        {/* Голова: широкие щёки, острый подбородок-морда, уши в одном контуре */}
        <path
          {...ink}
          d="
            M70 114
            C46 102 28 84 34 64
            Q20 28 28 12
            Q48 18 54 50
            Q70 40 86 50
            Q92 18 112 12
            Q120 28 106 64
            C112 84 94 102 70 114
            Z
          "
        />
        {asleep ? (
          <>
            <path
              d="M50 70 q8 10 16 0"
              fill="none"
              stroke={INK}
              strokeWidth="2.6"
              strokeLinecap="round"
            />
            <path
              d="M74 70 q8 10 16 0"
              fill="none"
              stroke={INK}
              strokeWidth="2.6"
              strokeLinecap="round"
            />
          </>
        ) : (
          <>
            <path
              d="M50 76 l8-10 8 10"
              fill="none"
              stroke={INK}
              strokeWidth="2.6"
              strokeLinecap="round"
              strokeLinejoin="miter"
            />
            <path
              d="M74 76 l8-10 8 10"
              fill="none"
              stroke={INK}
              strokeWidth="2.6"
              strokeLinecap="round"
              strokeLinejoin="miter"
            />
          </>
        )}
        <path d="M66 106h8l-4 10z" fill={INK} />
        <circle cx="48" cy="86" r="3.4" fill="#f3b4c4" />
        <circle cx="92" cy="86" r="3.4" fill="#f3b4c4" />

        <g fill="#fff6e4" stroke={INK} strokeWidth="1.3" strokeLinejoin="round">
          <ellipse cx="70" cy="48" rx="3.2" ry="5" />
          <ellipse cx="76" cy="52" rx="3.2" ry="5" transform="rotate(72 76 52)" />
          <ellipse cx="73" cy="58" rx="3.2" ry="5" transform="rotate(144 73 58)" />
          <ellipse cx="65" cy="56" rx="3.2" ry="5" transform="rotate(216 65 56)" />
          <ellipse cx="64" cy="50" rx="3.2" ry="5" transform="rotate(288 64 50)" />
          <circle cx="70" cy="53" r="2" fill="#f3b4c4" stroke="none" />
        </g>
      </g>
    );
  } else if (species === 'зайчик') {
    body = (
      <g>
        <path {...ink} d="M52 58c-2-28 6-46 14-46 6 0 10 10 8 28z" />
        <path {...ink} d="M78 58c2-28 8-46 14-46 8 0 12 16 8 28z" />
        <path d="M56 40c2 10 6 16 8 16" fill="#f7c2d0" />
        <path d="M86 40c-2 10-6 16-8 16" fill="#f7c2d0" />
        <path {...ink} d="M40 78c0-20 14-32 30-32s30 12 30 32c2 22-12 36-30 36s-32-14-30-36z" />
        {asleep ? (
          <path
            d="M57 68 q5 4 10 0 M75 68 q5 4 10 0"
            fill="none"
            stroke={INK}
            strokeWidth="2.2"
            strokeLinecap="round"
          />
        ) : (
          <>
            <ellipse cx="62" cy="68" rx="5.2" ry="4.6" fill="#fffaf2" stroke={INK} strokeWidth="2" />
            <ellipse cx="80" cy="68" rx="5.2" ry="4.6" fill="#fffaf2" stroke={INK} strokeWidth="2" />
            <circle cx="63" cy="68.4" r="1.8" fill={INK} />
            <circle cx="81" cy="68.4" r="1.8" fill={INK} />
          </>
        )}
        <circle cx="56" cy="78" r="4.2" fill="#f3b4c4" />
        <circle cx="86" cy="78" r="4.2" fill="#f3b4c4" />
        <ellipse cx="71" cy="82" rx="12" ry="8" fill="#fff6e4" stroke={INK} strokeWidth="2" />
        <ellipse cx="71" cy="80" rx="3" ry="2.2" fill="#e07a8a" stroke={INK} strokeWidth="1.2" />
        <rect
          x="66"
          y="88"
          width="4.5"
          height="8"
          rx="1"
          fill="#fffaf2"
          stroke={INK}
          strokeWidth="1.4"
        />
        <rect
          x="71.5"
          y="88"
          width="4.5"
          height="8"
          rx="1"
          fill="#fffaf2"
          stroke={INK}
          strokeWidth="1.4"
        />
      </g>
    );
  } else {
    body = (
      <g>
        <path {...ink} d="M70 8l14 28 32 4-22 20 8 30-32-16-32 16 8-30L24 40l32-4z" />
        <path
          d="M70 28l8 16 18 2-12 12 4 16-18-10-18 10 4-16-12-12 18-2z"
          fill="#fff6e4"
          stroke={INK}
          strokeWidth="2.2"
          strokeLinejoin="round"
        />
        <circle cx="58" cy="50" r="3" fill="#f3b4c4" opacity="0.9" />
        <circle cx="84" cy="50" r="3" fill="#f3b4c4" opacity="0.9" />
        <Face asleep={asleep} lx={64} ly={44} rx={78} ry={44} mouth="M66 54 q6 4 12 0" />
      </g>
    );
  }

  return (
    <svg viewBox="0 0 140 130" aria-hidden overflow="visible">
      <g transform={`translate(${70 * (1 - s)} ${122 * (1 - s)}) scale(${s})`}>{body}</g>
    </svg>
  );
}

export function MementoIcon({ id }: { id: string }) {
  const common = {
    viewBox: '0 0 48 48',
    'aria-hidden': true as const,
    stroke: INK,
    strokeWidth: 2,
    strokeLinejoin: 'round' as const,
  };
  switch (id) {
    case 'lamp':
      return (
        <svg {...common}>
          <path d="M14 20h20l-4 8H18z" fill="#f0d78c" />
          <path d="M23 28v8M16 38h16" fill="none" />
          <path d="M20 36h8v3H20z" fill="#8b5a3c" />
        </svg>
      );
    case 'plant':
      return (
        <svg {...common}>
          <path d="M24 36c-8-6-10-16 0-24 10 8 8 18 0 24z" fill="#3dba8a" />
          <path d="M24 30c-8-2-12-10-6-14" fill="none" stroke="#1f6b48" />
          <path d="M16 36h16l-3 6H19z" fill="#c47a4a" />
        </svg>
      );
    case 'books':
      return (
        <svg {...common}>
          <path d="M10 14h8v22h-8z" fill="#6ec9e8" />
          <path d="M20 10h8v26h-8z" fill="#e8b84a" />
          <path d="M30 16h8v20h-8z" fill="#a78bfa" />
        </svg>
      );
    case 'pillow':
      return (
        <svg {...common}>
          <path d="M8 26c0-8 8-12 16-12s16 4 16 12-8 12-16 12S8 34 8 26z" fill="#efe4cf" />
          <path d="M24 18l2 5 5 .5-4 3 1.5 5L24 28l-4.5 3.5 1.5-5-4-3 5-.5z" fill="#e8b84a" />
        </svg>
      );
    case 'gear':
      return (
        <svg {...common}>
          <path d="M24 6l3 7h7l-5 5 2 8-7-3-7 3 2-8-5-5h7z" fill="#e4dcc8" />
          <circle cx="24" cy="22" r="3" fill="#2a2118" stroke="none" />
        </svg>
      );
    case 'ticket':
      return (
        <svg {...common}>
          <path d="M6 16h36v7a4 4 0 0 0 0 8v7H6v-7a4 4 0 0 0 0-8z" fill="#f2e2b0" />
          <path d="M16 22h4M16 28h10" fill="none" />
        </svg>
      );
    case 'spark':
      return (
        <svg {...common}>
          <path d="M24 4l4 14 14 4-14 4-4 14-4-14-14-4 14-4z" fill="#ffe9b8" />
        </svg>
      );
    case 'bowl':
      return (
        <svg {...common}>
          <path d="M8 18h32l-5 16H13z" fill="#d7e4f2" />
          <path d="M8 18c4 5 28 5 32 0" fill="#e8b84a" />
        </svg>
      );
    default:
      return (
        <svg {...common}>
          <circle cx="24" cy="24" r="8" fill="#6b7394" />
        </svg>
      );
  }
}
