// frontend/src/components/Games/Companion/CompanionGame.tsx
// Soft tamagotchi MVP: customize, feed/play/rest. No death — critical loneliness after neglect.
// Not an LB chase: daily care gift (+points) matters more than boost/rank.
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../../services/api';
import { getNickname } from '../../../lib/nickname';
import { Balance } from '../../Billing/Balance';
import { GameShell, ScoreChip } from '../../ui/GameShell';
import { Button } from '../../ui/Button';
import Notification from '../../Common/Notification/Notification';

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
const DAILY_GIFT_KEY = 'eh_companion_daily_gift_v1';
const DAILY_GIFT_POINTS = 1000;

const SPECIES: { id: Species; emoji: string; label: string }[] = [
  { id: 'звезда', emoji: '⭐', label: 'Звезда' },
  { id: 'кот', emoji: '🐱', label: 'Кот' },
  { id: 'дракон', emoji: '🐉', label: 'Дракон' },
  { id: 'кактус', emoji: '🌵', label: 'Кактус' },
];
const COLORS = ['#c9a227', '#7c9cff', '#5ec8b8', '#e07a5f', '#b388ff'];

const WEEK_MS = 7 * 24 * 60 * 60 * 1000;

function todayKey(): string {
  return new Date().toISOString().slice(0, 10);
}

function dailyGiftClaimedToday(): boolean {
  return localStorage.getItem(DAILY_GIFT_KEY) === todayKey();
}

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
  const [giftClaimed, setGiftClaimed] = useState(() => dailyGiftClaimedToday());
  /** Points from today's gift — included in save score. */
  const [giftPoints, setGiftPoints] = useState(() => (dailyGiftClaimedToday() ? DAILY_GIFT_POINTS : 0));

  useEffect(() => {
    if (!pet) return;
    savePet(pet);
  }, [pet]);

  useEffect(() => {
    const t = setInterval(() => {
      setPet((p) => (p ? decay(p) : p));
      const claimed = dailyGiftClaimedToday();
      setGiftClaimed(claimed);
      if (!claimed) setGiftPoints(0);
    }, 60_000);
    return () => clearInterval(t);
  }, []);

  const mood = useMemo(() => (pet ? moodOf(pet) : 'ok'), [pet]);
  const baseScore = pet ? careScore(pet) : 0;
  const score = Math.min(10000, baseScore + giftPoints);
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

  const claimDailyGift = async () => {
    if (!pet || giftClaimed || dailyGiftClaimedToday()) {
      setNotif({ message: 'Подарок сегодня уже получен', type: 'error' });
      return;
    }
    try {
      const { data } = await api.post<{
        tickets_granted?: number;
        already_claimed?: boolean;
      }>('/game/companion/daily-gift');
      localStorage.setItem(DAILY_GIFT_KEY, todayKey());
      setGiftClaimed(true);
      setGiftPoints(DAILY_GIFT_POINTS);
      setPet((p) =>
        p
          ? {
              ...p,
              hunger: Math.min(100, p.hunger + 12),
              energy: Math.min(100, p.energy + 12),
              joy: Math.min(100, p.joy + 12),
            }
          : p,
      );
      const tickets = data?.tickets_granted ?? 1000;
      const again = data?.already_claimed
        ? ' (уже было на сервере)'
        : '';
      setNotif({
        message: `Подарок дня: +${DAILY_GIFT_POINTS} к заботе и +${tickets} билетиков${again}.`,
        type: 'success',
      });
    } catch {
      setNotif({ message: 'Не удалось получить подарок дня', type: 'error' });
    }
  };

  const submit = async () => {
    if (!pet) return;
    setSaving(true);
    try {
      const body: Record<string, unknown> = {
        user_id: localStorage.getItem('userId'),
        game_id: 'companion',
        level: 1,
        score: Math.min(10000, careScore(pet) + giftPoints),
        user_email: localStorage.getItem('userEmail'),
        nickname: getNickname(),
        seed: `companion_${pet.name}_${Date.now()}`,
        moves: [],
      };
      await api.post('/game/submit', body);
      setNotif({
        message:
          giftPoints > 0
            ? `Забота сохранена (+${giftPoints} подарок дня)`
            : 'Забота сохранена',
        type: 'success',
      });
      void import('../../../lib/achievements').then(({ afterRankedSubmit }) => afterRankedSubmit());
    } catch {
      setNotif({ message: 'Не удалось сохранить', type: 'error' });
    } finally {
      setSaving(false);
    }
  };

  const moodLine =
    mood === 'critical'
      ? `${pet?.name} скучал(а) по тебе… Нужна забота.`
      : mood === 'sad'
        ? `${pet?.name} грустит. Покорми, поиграй или дай отдохнуть.`
        : mood === 'happy'
          ? `${pet?.name} светится от счастья!`
          : `${pet?.name} в порядке.`;

  const handleBack = () => navigate('/#games');

  return (
    <GameShell
      title="Tamagotchi"
      onBack={handleBack}
      width="narrow"
      actions={<Balance />}
      stats={
        pet ? (
          <>
            <ScoreChip label="Забота" value={score} />
            <ScoreChip label="Серия" value={pet.careStreak} />
          </>
        ) : undefined
      }
      controls={
        pet ? (
          <>
            <Button size="sm" onClick={() => care('feed')}>
              Кормить
            </Button>
            <Button size="sm" onClick={() => care('play')}>
              Играть
            </Button>
            <Button size="sm" onClick={() => care('rest')}>
              Сон
            </Button>
            {!giftClaimed && (
              <Button size="sm" variant="secondary" onClick={claimDailyGift}>
                Подарок дня (+{DAILY_GIFT_POINTS})
              </Button>
            )}
            <Button size="sm" variant="ghost" onClick={() => void submit()} disabled={saving}>
              Сохранить счёт
            </Button>
          </>
        ) : undefined
      }
      help={
        <>
          <p>Играй со своим питомцем: корми, играй, дай отдохнуть. Без давления и дедлайнов.</p>
          <p>
            Это не гонка за лидербордом. Раз в день — подарок: +{DAILY_GIFT_POINTS} к заботе и
            +1000 билетиков на кошелёк.
          </p>
        </>
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
            className="relative flex w-full flex-col items-center rounded-lg border border-white/10 px-4 py-8"
            style={{
              background:
                'linear-gradient(180deg, #1a2238 0%, #12182a 45%, #0c101c 100%), repeating-linear-gradient(90deg, transparent, transparent 18px, rgba(255,255,255,0.03) 18px, rgba(255,255,255,0.03) 19px)',
            }}
          >
            <div
              className="pointer-events-none absolute inset-x-6 top-3 h-10 rounded-full opacity-40 blur-xl"
              style={{ background: pet.color }}
              aria-hidden
            />
            <div
              className="relative flex h-40 w-40 flex-col items-center justify-center rounded-full border border-white/15 shadow-elevated transition-transform duration-500"
              style={{
                background: `radial-gradient(circle at 35% 30%, ${pet.color}55, #0c0e16 70%)`,
              }}
            >
              <span className="text-6xl" aria-hidden>
                {speciesMeta?.emoji}
              </span>
              <span className="mt-1 font-display text-lg text-text-primary">{pet.name}</span>
            </div>
            <p className="relative mt-3 text-xs text-text-muted">Комната питомца</p>
          </div>
          <p className="text-center text-sm text-text-secondary">{moodLine}</p>
          {giftClaimed ? (
            <p className="text-center text-xs text-horizon-gold">
              Подарок дня получен{giftPoints > 0 ? ` (+${giftPoints})` : ''}
            </p>
          ) : (
            <p className="text-center text-xs text-text-muted">
              Можно забрать подарок дня — +{DAILY_GIFT_POINTS} к заботе
            </p>
          )}
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
