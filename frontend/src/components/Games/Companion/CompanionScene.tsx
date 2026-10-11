import { useState } from 'react';
import { Button } from '../../ui/Button';
import {
  CompanionPet,
  MEMENTOS,
  MementoIcon,
  mementoById,
  type Growth,
  type Species,
} from './CompanionArt';
import './companion.css';

type Mood = 'happy' | 'ok' | 'sad' | 'critical';
type Pose = 'idle' | 'rest' | 'play' | 'sad';

const ITEM_POS: Record<string, { left: string; bottom: string }> = {
  lamp: { left: '72%', bottom: '30%' },
  plant: { left: '84%', bottom: '8%' },
  books: { left: '64%', bottom: '32%' },
  pillow: { left: '14%', bottom: '36%' },
  gear: { left: '48%', bottom: '8%' },
  ticket: { left: '68%', bottom: '40%' },
  spark: { left: '60%', bottom: '42%' },
  bowl: { left: '46%', bottom: '8%' },
};

export function CompanionDevice({
  name,
  species,
  color,
  levelLine,
  growth,
  hunger,
  energy,
  joy,
  mood,
  moodLine,
  pose,
  playTick,
  mementos,
  showBowl,
  albumOpen,
  onToggleAlbum,
  revealId,
  revealFull,
  onCloseReveal,
}: {
  name: string;
  species: Species;
  color: string;
  levelLine: string;
  growth: Growth;
  hunger: number;
  energy: number;
  joy: number;
  mood: Mood;
  moodLine: string;
  pose: Pose;
  playTick: number;
  mementos: readonly string[];
  showBowl: boolean;
  albumOpen: boolean;
  onToggleAlbum: () => void;
  revealId: string | null;
  revealFull: boolean;
  onCloseReveal: () => void;
}) {
  const owned = new Set(mementos);
  const items = MEMENTOS.filter((m) => owned.has(m.id) && m.id !== 'bowl');

  return (
    <div className="companion-device">
      <div className="companion-bezel">
        <div
          className="companion-screen"
          data-mood={mood}
          data-asleep={pose === 'rest' ? 'true' : 'false'}
        >
          {albumOpen ? (
            <Album owned={owned} />
          ) : (
            <Room
              name={name}
              species={species}
              color={color}
              levelLine={levelLine}
              growth={growth}
              hunger={hunger}
              energy={energy}
              joy={joy}
              moodLine={moodLine}
              pose={pose}
              playTick={playTick}
              owned={owned}
              items={items}
              showBowl={showBowl && pose !== 'rest'}
            />
          )}
          {(revealId || revealFull) && (
            <GiftCard id={revealId} full={revealFull} onClose={onCloseReveal} />
          )}
        </div>
      </div>
      <button type="button" className="companion-key" onClick={onToggleAlbum}>
        {albumOpen ? 'Комната' : 'Подарки'}
      </button>
    </div>
  );
}

function Room({
  name,
  species,
  color,
  levelLine,
  growth,
  hunger,
  energy,
  joy,
  moodLine,
  pose,
  playTick,
  owned,
  items,
  showBowl,
}: {
  name: string;
  species: Species;
  color: string;
  levelLine: string;
  growth: Growth;
  hunger: number;
  energy: number;
  joy: number;
  moodLine: string;
  pose: Pose;
  playTick: number;
  owned: Set<string>;
  items: { id: string }[];
  showBowl: boolean;
}) {
  const poseClass =
    pose === 'rest' ? 'companion-pet--rest' : pose === 'play' ? 'companion-pet--play' : pose === 'sad' ? 'companion-pet--sad' : '';
  const shelf = MEMENTOS.filter((m) => owned.has(m.id));

  return (
    <>
      <div className="companion-room">
        <RoomSvg />
        {owned.has('lamp') && pose !== 'rest' && (
          <div className="companion-glow" style={{ left: '68%', bottom: '34%' }} />
        )}
        {items.map((item) => {
          const pos = ITEM_POS[item.id];
          if (!pos) return null;
          return (
            <div key={item.id} className="companion-item" style={pos}>
              <MementoIcon id={item.id} />
            </div>
          );
        })}
        {showBowl && (
          <div className="companion-item" style={ITEM_POS.bowl}>
            <MementoIcon id="bowl" />
          </div>
        )}
        <div key={playTick} className={['companion-pet', poseClass].filter(Boolean).join(' ')}>
          <CompanionPet species={species} color={color} asleep={pose === 'rest'} growth={growth} />
        </div>
        <div className="companion-plate">
          <p className="companion-name">{name}</p>
          <p className="companion-lv">{levelLine}</p>
          <div className="companion-meters">
            {(
              [
                ['Сытость', hunger],
                ['Энергия', energy],
                ['Радость', joy],
              ] as const
            ).map(([label, value]) => (
              <div key={label} className="companion-meter">
                <span>{label}</span>
                <span>
                  <i style={{ width: `${Math.round(value)}%` }} />
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
      <footer className="companion-footer">
        <p className="companion-mood">{moodLine}</p>
        <div className="companion-shelf" aria-label="На полке">
          {shelf.length === 0 ? (
            <span className="companion-mood">Полка пустая</span>
          ) : (
            shelf.map((m) => <MementoIcon key={m.id} id={m.id} />)
          )}
        </div>
      </footer>
    </>
  );
}

function RoomSvg() {
  const ink = { stroke: '#2a2118', strokeWidth: 2, strokeLinejoin: 'round' as const };
  return (
    <svg className="companion-backdrop" viewBox="0 0 360 240" preserveAspectRatio="xMidYMax meet" aria-hidden>
      <rect width="360" height="156" fill="#31456e" />
      <rect y="156" width="360" height="8" fill="#c6a15a" stroke="#2a2118" strokeWidth="2" />
      <rect y="164" width="360" height="76" fill="#8a5a3c" />
      <g stroke="#6b4128" strokeWidth="1.5" opacity="0.55">
        <line x1="0" y1="184" x2="360" y2="184" />
        <line x1="0" y1="204" x2="360" y2="204" />
        <line x1="0" y1="224" x2="360" y2="224" />
      </g>

      <rect x="214" y="22" width="108" height="78" fill="#c6a15a" stroke="#2a2118" strokeWidth="3" />
      <rect x="222" y="30" width="92" height="62" fill="#10182e" />
      {/* Night sky stays inside the window only. */}
      <g>
        <circle cx="244" cy="46" r="8" fill="#f6ead0" />
        <circle cx="268" cy="40" r="1.2" fill="#d7e4ff" />
        <circle cx="290" cy="52" r="1.1" fill="#d7e4ff" />
        <circle cx="300" cy="38" r="1.3" fill="#d7e4ff" />
        <circle cx="278" cy="72" r="1" fill="#d7e4ff" />
        <circle cx="236" cy="70" r="1.1" fill="#d7e4ff" />
      </g>
      <line x1="268" y1="30" x2="268" y2="92" stroke="#c6a15a" strokeWidth="4" />
      <line x1="222" y1="61" x2="314" y2="61" stroke="#c6a15a" strokeWidth="4" />
      <rect x="208" y="98" width="120" height="8" fill="#c6a15a" stroke="#2a2118" strokeWidth="2" />

      <rect x="24" y="112" width="16" height="52" rx="2" fill="#6b4630" stroke="#2a2118" strokeWidth="3" />
      <rect x="38" y="128" width="124" height="24" rx="2" fill="#efe4cf" stroke="#2a2118" strokeWidth="3" />
      <rect x="86" y="128" width="76" height="24" fill="#3d5d86" stroke="#2a2118" strokeWidth="3" />
      <rect x="44" y="116" width="32" height="18" rx="4" fill="#f7efe2" stroke="#2a2118" strokeWidth="2.5" />
      <rect x="46" y="152" width="8" height="16" fill="#4a3428" {...ink} />
      <rect x="146" y="152" width="8" height="16" fill="#4a3428" {...ink} />

      <rect x="248" y="136" width="88" height="10" rx="1" fill="#6b4630" stroke="#2a2118" strokeWidth="3" />
      <rect x="256" y="146" width="8" height="22" fill="#4a3428" {...ink} />
      <rect x="320" y="146" width="8" height="22" fill="#4a3428" {...ink} />

      <rect x="176" y="176" width="64" height="12" rx="2" fill="#4a3560" stroke="#2a2118" strokeWidth="2" />
    </svg>
  );
}

function Album({ owned }: { owned: Set<string> }) {
  const first = MEMENTOS.find((m) => owned.has(m.id))?.id ?? MEMENTOS[0].id;
  const [selected, setSelected] = useState(first);
  const detail = mementoById(selected);

  return (
    <div className="companion-album">
      <h2>Подарки</h2>
      <div className="companion-grid">
        {MEMENTOS.map((m, index) => {
          const has = owned.has(m.id);
          return (
            <button
              key={m.id}
              type="button"
              className={[
                'companion-slot',
                has ? 'companion-slot--owned' : '',
                selected === m.id ? 'companion-slot--on' : '',
              ]
                .filter(Boolean)
                .join(' ')}
              onClick={() => has && setSelected(m.id)}
              aria-label={has ? m.name : `Пусто ${index + 1}`}
            >
              {has ? <MementoIcon id={m.id} /> : '?'}
            </button>
          );
        })}
      </div>
      {detail && owned.has(detail.id) ? (
        <div className="companion-detail">
          <strong>{detail.name}</strong>
          <p>{detail.line}</p>
        </div>
      ) : (
        <div className="companion-detail">
          <p>Подарок дня кладёт сюда одну вещь.</p>
        </div>
      )}
    </div>
  );
}

function GiftCard({
  id,
  full,
  onClose,
}: {
  id: string | null;
  full: boolean;
  onClose: () => void;
}) {
  const item = id ? mementoById(id) : undefined;
  return (
    <div className="companion-gift" role="dialog" aria-label="Подарок дня">
      <p>{full ? 'Полка полная' : 'Подарок дня'}</p>
      {item ? <MementoIcon id={item.id} /> : null}
      <p>{item ? item.name : 'Новых вещей нет'}</p>
      <p>{item ? item.line : 'Билетики всё равно пришли.'}</p>
      <Button size="sm" onClick={onClose}>
        В комнату
      </Button>
    </div>
  );
}
