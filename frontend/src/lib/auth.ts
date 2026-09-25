/**
 * Auth hydration + token storage.
 * Access TTL is ~15 minutes; refresh token keeps the session alive.
 */
let ready = false;
const waiters: Array<() => void> = [];

function resolveReady() {
  if (ready) return;
  ready = true;
  waiters.splice(0).forEach((w) => w());
}

export function hydrateAuth(): void {
  resolveReady();
}

export function isAuthReady(): boolean {
  return ready;
}

export function whenAuthReady(): Promise<void> {
  if (ready) return Promise.resolve();
  return new Promise((resolve) => waiters.push(resolve));
}

export function getAccessToken(): string | null {
  return localStorage.getItem('accessToken');
}

export function getRefreshToken(): string | null {
  return localStorage.getItem('refreshToken');
}

export function setTokens(access: string, refresh?: string | null): void {
  localStorage.setItem('accessToken', access);
  if (refresh) localStorage.setItem('refreshToken', refresh);
  resolveReady();
  window.dispatchEvent(new Event('authChange'));
}

/** @deprecated prefer setTokens — kept for call sites that only have access */
export function setAccessToken(token: string): void {
  setTokens(token);
}

export function clearAuth(): void {
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
  localStorage.removeItem('userId');
  localStorage.removeItem('userEmail');
  localStorage.removeItem('role');
  window.dispatchEvent(new Event('authChange'));
}

export function canFetchProtected(): boolean {
  return ready && !!getAccessToken();
}
