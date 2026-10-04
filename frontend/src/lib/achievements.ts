import api from '../services/api';
import type { IconName } from '../components/ui/Icon';

const SEEN_KEY = 'eh_seen_achievements';

export type ProfileAchievement = {
  code: string;
  title: string;
  description: string;
  icon: string;
  unlocked_at: string;
};

export type ProfilePayload = {
  nickname?: string;
  total_score?: number;
  best_scores?: Record<string, number>;
  achievements?: ProfileAchievement[];
};

const KNOWN_ICONS = new Set<IconName>([
  'trophy',
  'sparkle',
  'star',
  'bird',
  'hex',
  'tower',
  'cards',
  'hanoi',
  'twenty48',
  'gears',
  'crown',
  'medal',
  'gift',
  'check',
]);

export function achievementIcon(name: string | undefined): IconName {
  if (name && KNOWN_ICONS.has(name as IconName)) {
    return name as IconName;
  }
  return 'trophy';
}

function readSeen(): string[] | null {
  const raw = localStorage.getItem(SEEN_KEY);
  if (raw == null) return null;
  try {
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.map(String) : [];
  } catch {
    return [];
  }
}

function writeSeen(codes: string[]) {
  localStorage.setItem(SEEN_KEY, JSON.stringify([...new Set(codes)]));
}

export type SyncAchievementsResult = {
  profile: ProfilePayload;
  achievements: ProfileAchievement[];
  /** Newly unlocked since last client-seen set (empty on first silent seed). */
  fresh: ProfileAchievement[];
  toastMessage: string | null;
};

/**
 * Fetch profile achievements.
 * Option (a): empty seen-set → seed silently (no toast flood for legacy unlocks).
 * Later unlocks → toast for new codes only.
 */
export async function syncAchievements(opts?: {
  toast?: boolean;
  delayMs?: number;
}): Promise<SyncAchievementsResult> {
  const toast = opts?.toast ?? false;
  const delayMs = opts?.delayMs ?? 0;
  if (delayMs > 0) {
    await new Promise((r) => setTimeout(r, delayMs));
  }

  const token = localStorage.getItem('accessToken');
  if (!token) {
    return { profile: {}, achievements: [], fresh: [], toastMessage: null };
  }

  const res = await api.get('/api/profile', {
    headers: { Authorization: `Bearer ${token}` },
  });
  const profile: ProfilePayload = res.data || {};
  const achievements: ProfileAchievement[] = Array.isArray(profile.achievements)
    ? profile.achievements
    : [];
  const codes = achievements.map((a) => a.code).filter(Boolean);

  const seen = readSeen();
  if (seen == null) {
    writeSeen(codes);
    return { profile, achievements, fresh: [], toastMessage: null };
  }

  const seenSet = new Set(seen);
  const fresh = achievements.filter((a) => a.code && !seenSet.has(a.code));
  writeSeen([...seenSet, ...codes]);

  let toastMessage: string | null = null;
  if (toast && fresh.length === 1) {
    toastMessage = `Достижение: ${fresh[0].title}`;
  } else if (toast && fresh.length > 1) {
    toastMessage = `Получено достижений: ${fresh.length}`;
  }

  if (toastMessage) {
    window.dispatchEvent(
      new CustomEvent('eh:achievement', { detail: { message: toastMessage } }),
    );
  }

  return { profile, achievements, fresh, toastMessage };
}

/** Call after a ranked (non-boost) score submit. */
export function afterRankedSubmit() {
  void syncAchievements({ toast: true, delayMs: 1200 });
}
