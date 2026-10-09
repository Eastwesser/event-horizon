/** Shared nickname for profile + game submits + LB (avoids Admin/Nimda split). */

export function nicknameStorageKey(userId?: string | null): string {
  const id = userId || localStorage.getItem('userId') || '';
  return id ? `nickname_${id}` : 'nickname';
}

export function getNickname(): string {
  const userId = localStorage.getItem('userId') || '';
  const keyed = userId ? localStorage.getItem(`nickname_${userId}`) : null;
  if (keyed?.trim()) return keyed.trim();
  const bare = localStorage.getItem('nickname');
  if (bare?.trim()) return bare.trim();
  const email = localStorage.getItem('userEmail') || '';
  return email.includes('@') ? email.split('@')[0] : 'Игрок';
}

export function setNickname(nick: string, userId?: string | null): void {
  const trimmed = nick.trim();
  if (!trimmed) return;
  const id = userId || localStorage.getItem('userId') || '';
  localStorage.setItem('nickname', trimmed);
  if (id) localStorage.setItem(`nickname_${id}`, trimmed);
}
