// frontend/src/components/Games/Companion/CompanionGame.tsx
// Soft tamagotchi MVP: customize, feed/play/rest. No death — critical loneliness after neglect.
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../../services/api';
import { useGameBoost } from '../../../hooks/useGameBoost';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import Notification from '../../Common/Notification/Notification';
import { BoostCheckbox, boostUnrankedToast } from '../BoostCheckbox';

type Species = 'звезда' | 'кот' | 'дракон' | 'кактус';
type Mood = 'happy' | 'ok' | 'sad' | 'critical';

interface Pet {
  name: string;
  species: Species;
  color: string;
  hunger: number; // 0..100 (100 = full)
  energy: number;
  joy: number;
  lastCareAt: number; // ms
  careStreak: number;
}

const STORAGE_KEY = 'eh_companion_pet_v1';
const SPECIES: { id: Species; emoji: string; label: string }[] = [
  { id: 'звезда', emoji: '⭐', label: 'Звезда' },
  { id: 'кот', emoji: '🐱', label: 'Кот' },
  { id: 'дракон', emoji: '🐉', label: 'Дракон' },
  { id: 'кактус', emoji: '🌵', label: 'Кактус' },
];
const COLORS = ['#c9a227', '#7c9cff', '#5ec8b8', '#e07a5f', '#b388ff'];

const WEEK_MS = 7 * 24 * 60 * 60 * 1000;

function loadPet(): Pet | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as Pet;
  } catch {
    return null;
  }
}

function savePet(p: Pet) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(p));
}

function decay(p: Pet, now = Date.now()): Pet {
  const hours = Math.max(0, (now - p.lastCareAt) / 3_600_000);
  const hunger = Math.max(0, p.hunger - hours * 1.2);
  const energy = Math.max(0, p.energy - hours * 0.8);
  const joy = Math.max(0, p.joy - hours * 1.0);
  return { ...p, hunger, energy, joy };
}

function moodOf(p: Pet, now = Date.now()): Mood {
  if (now - p.lastCareAt >= WEEK_MS) return 'critical';
  const avg = (p.hunger + p.energy + p.joy) / 3;
  if (avg >= 70) return 'happy';
  if (avg >= 35) return 'ok';
  return 'sad';
}

function careScore(p: Pet): number {
  return Math.round((p.hunger + p.energy + p.joy) / 3 + p.careStreak * 5);
}

export function CompanionGame() {
  const navigate = useNavigate();
  const [pet, setPet] = useState<Pet | null>(() => {
    const p = loadPet();
    return p ? decay(p) : null;
  });
  const [draftName, setDraftName] = useState('');
  const [draftSpecies, setDraftSpecies] = useState<Species>('звезда');
  const [draftColor, setDraftColor] = useState(COLORS[0]);
  const [notif, setNotif] = useState<{ message: string; type: 'success' | 'error' } | null>(null);
  const [saving, setSaving] = useState(false);
  const {
    useBoost,
    setUseBoost,
    boostBusy,
    armBoost,
    boostError,
    setBoostError,
  } = useGameBoost('companion');

  useEffect(() => {
    if (boostError) {
      setNotif({ message: boostError, type: 'error' });
      setBoostError(null);
    }
  }, [boostError, setBoostError]);

  useEffect(() => {
    if (!pet) return;
    savePet(pet);
  }, [pet]);

  useEffect(() => {
    const t = setInterval(() => {
      setPet((p) => (p ? decay(p) : p));
    }, 60_000);
    return () => clearInterval(t);
  }, []);

  const mood = useMemo(() => (pet ? moodOf(pet) : 'ok'), [pet]);
  const score = pet ? careScore(pet) : 0;
  const speciesMeta = SPECIES.find((s) => s.id === pet?.species);

  const create = () => {
    const name = draftName.trim() || 'Странник';
    const p: Pet = {
      name,
      species: draftSpecies,
      color: draftColor,
      hunger: 80,
      energy: 80,
      joy: 80,
      lastCareAt: Date.now(),
      careStreak: 1,
    };
    setPet(p);
  };

  const care = (kind: 'feed' | 'play' | 'rest') => {
    if (!pet) return;
    const now = Date.now();
    let next = decay(pet, now);
    if (kind === 'feed') next = { ...next, hunger: Math.min(100, next.hunger + 28) };
    if (kind === 'play') next = { ...next, joy: Math.min(100, next.joy + 28), energy: Math.max(0, next.energy - 8) };
    if (kind === 'rest') next = { ...next, energy: Math.min(100, next.energy + 32) };
    const dayGap = (now - next.lastCareAt) / 86_400_000;
    const streak = dayGap <= 1.5 ? next.careStreak + (dayGap >= 0.4 ? 1 : 0) : 1;
    next = { ...next, lastCareAt: now, careStreak: Math.min(365, streak) };
    setPet(next);
  };

  const submit = async () => {
    if (!pet) return;
    setSaving(true);
    try {
      const { boostId, boosted } = await armBoost();
      const body: Record<string, unknown> = {
        user_id: localStorage.getItem('userId'),
        game_id: 'companion',
        level: 1,
        score: Math.min(10000, careScore(pet)),
        user_email: localStorage.getItem('userEmail'),
        nickname: localStorage.getItem('nickname') || '',
        seed: `companion_${pet.name}_${Date.now()}`,
        moves: [],
      };
      if (boosted && boostId) body.boost_id = boostId;
      const response = await api.post('/game/submit', body);
      const ranked =
        response.data?.ranked === true &&
        !boosted &&
        !String(response.data?.message || '').includes('not ranked');
      setNotif({
        message: ranked ? 'Забота сохранена в счёт' : boostUnrankedToast(),
        type: 'success',
      });
      if (ranked) {
        void import('../../../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
      }
    } catch {
      setNotif({ message: 'Не удалось сохранить', type: 'error' });
    } finally {
      setSaving(false);
    }
  };

  const moodLine =
    mood === 'critical'
      ? `${pet?.name} скучал(а) по тебе… Критично нужна забота (смерти нет — только тоска).`
      : mood === 'sad'
        ? `${pet?.name} грустит. Покорми, поиграй или дай отдохнуть.`
        : mood === 'happy'
          ? `${pet?.name} светится от счастья!`
          : `${pet?.name} в порядке.`;

  return (
    <GameShell
      title="Компаньон"
      onBack={() => navigate('/')}
      width="narrow"
      stats={pet ? <ScoreChip label="Забота" value={score} /> : undefined}
      controls={
        pet ? (
          <>
            <BoostCheckbox useBoost={useBoost} onChange={setUseBoost} disabled={boostBusy || saving} />
            <Button size="sm" onClick={() => care('feed')}>
              Кормить
            </Button>
            <Button size="sm" onClick={() => care('play')}>
              Играть
            </Button>
            <Button size="sm" onClick={() => care('rest')}>
              Сон
            </Button>
            <Button size="sm" variant="ghost" onClick={() => void submit()} disabled={saving || boostBusy}>
              Сохранить счёт
            </Button>
          </>
        ) : undefined
      }
      help={
        <p>
          Мягкий тамагочи: без смерти. Boost при сохранении — счёт не в лидерборд.
        </p>
      }
    >
      {!pet ? (
        <div className="w-full max-w-sm space-y-4 rounded-md border border-white/10 bg-nebula p-4">
          <p className="font-display text-lg text-text-primary">Создай компаньона</p>
          <label className="block text-sm text-text-secondary">
            Имя
            <input
              className="mt-1 w-full rounded-sm border border-white/10 bg-void px-3 py-2 text-text-primary"
              value={draftName}
              onChange={(e) => setDraftName(e.target.value)}
              placeholder="Имя"
              maxLength={24}
            />
          </label>
          <div className="flex flex-wrap gap-2">
            {SPECIES.map((s) => (
              <button
                key={s.id}
                type="button"
                onClick={() => setDraftSpecies(s.id)}
                className={`rounded-sm border px-3 py-2 text-sm ${
                  draftSpecies === s.id
                    ? 'border-horizon-gold text-horizon-gold'
                    : 'border-white/10 text-text-secondary'
                }`}
              >
                {s.emoji} {s.label}
              </button>
            ))}
          </div>
          <div className="flex gap-2">
            {COLORS.map((c) => (
              <button
                key={c}
                type="button"
                aria-label={c}
                onClick={() => setDraftColor(c)}
                className={`h-8 w-8 rounded-full border-2 ${
                  draftColor === c ? 'border-white' : 'border-transparent'
                }`}
                style={{ background: c }}
              />
            ))}
          </div>
          <Button onClick={create}>Создать</Button>
        </div>
      ) : (
        <div className="flex w-full max-w-sm flex-col items-center gap-4">
          <div
            className="flex h-40 w-40 flex-col items-center justify-center rounded-full border border-white/15 shadow-elevated transition-transform duration-500"
            style={{ background: `radial-gradient(circle at 35% 30%, ${pet.color}55, #0c0e16 70%)` }}
          >
            <span className="text-6xl" aria-hidden>
              {speciesMeta?.emoji}
            </span>
            <span className="mt-1 font-display text-lg text-text-primary">{pet.name}</span>
          </div>
          <p className="text-center text-sm text-text-secondary">{moodLine}</p>
          <div className="grid w-full grid-cols-3 gap-2 text-center text-xs text-text-muted">
            {(
              [
                ['Сытость', pet.hunger],
                ['Энергия', pet.energy],
                ['Радость', pet.joy],
              ] as const
            ).map(([label, v]) => (
              <div key={label} className="rounded-sm border border-white/10 bg-nebula p-2">
                <div className="mb-1 text-text-secondary">{label}</div>
                <div className="h-1.5 overflow-hidden rounded-full bg-white/10">
                  <div
                    className="h-full rounded-full bg-horizon-gold transition-all duration-500"
                    style={{ width: `${v}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
          <p className="text-xs text-text-muted">Серия заботы: {pet.careStreak} дн.</p>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => {
              localStorage.removeItem(STORAGE_KEY);
              setPet(null);
            }}
          >
            Создать нового
          </Button>
        </div>
      )}
      {notif && (
        <Notification message={notif.message} type={notif.type} onClose={() => setNotif(null)} />
      )}
    </GameShell>
  );
}
