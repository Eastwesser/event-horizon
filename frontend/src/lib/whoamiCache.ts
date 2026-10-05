export type UserRole = 'user' | 'author' | 'admin';

let cachedRole: UserRole | null = null;
let cachedAt = 0;
let inflight: Promise<UserRole> | null = null;

const CACHE_MS = 60_000;

export function getWhoamiCache(): {
  role: UserRole | null;
  at: number;
  inflight: Promise<UserRole> | null;
  cacheMs: number;
} {
  return { role: cachedRole, at: cachedAt, inflight, cacheMs: CACHE_MS };
}

export function setWhoamiCache(role: UserRole): void {
  cachedRole = role;
  cachedAt = Date.now();
}

export function setWhoamiInflight(p: Promise<UserRole> | null): void {
  inflight = p;
}

export function invalidateWhoamiCache(): void {
  cachedRole = null;
  cachedAt = 0;
  inflight = null;
}
