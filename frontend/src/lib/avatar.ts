/** Client-persisted profile avatar URL (upload via POST /api/v1/uploads). */

export function avatarStorageKey(userId?: string | null): string {
  const id = userId || localStorage.getItem('userId') || '';
  return id ? `avatar_${id}` : 'avatar';
}

export function getAvatarUrl(userId?: string | null): string {
  return localStorage.getItem(avatarStorageKey(userId)) || '';
}

export function setAvatarUrl(url: string, userId?: string | null): void {
  const key = avatarStorageKey(userId);
  const trimmed = url.trim();
  if (!trimmed) localStorage.removeItem(key);
  else localStorage.setItem(key, trimmed);
}
