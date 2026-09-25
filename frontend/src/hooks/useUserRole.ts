import { useEffect, useState } from 'react';
import api from '../services/api';

export type UserRole = 'user' | 'author' | 'admin';

/** Module cache — StrictMode + multi-mount pages were hitting /auth/whoami 8×. */
let cachedRole: UserRole | null = null;
let cachedAt = 0;
let inflight: Promise<UserRole> | null = null;
const CACHE_MS = 60_000;

function roleFromStorage(): UserRole {
  return (localStorage.getItem('role') as UserRole) || 'user';
}

export function invalidateWhoamiCache(): void {
  cachedRole = null;
  cachedAt = 0;
  inflight = null;
}

async function fetchWhoamiOnce(): Promise<UserRole> {
  const now = Date.now();
  if (cachedRole && now - cachedAt < CACHE_MS) return cachedRole;
  if (inflight) return inflight;

  inflight = api
    .get('/auth/whoami')
    .then(({ data }) => {
      const r = (data.role as UserRole) || 'user';
      cachedRole = r;
      cachedAt = Date.now();
      localStorage.setItem('role', r);
      return r;
    })
    .catch(() => {
      const fallback = roleFromStorage();
      cachedRole = fallback;
      cachedAt = Date.now();
      return fallback;
    })
    .finally(() => {
      inflight = null;
    });

  return inflight;
}

export function useUserRole() {
  const [role, setRole] = useState<UserRole>(() => cachedRole ?? roleFromStorage());
  const [loading, setLoading] = useState(!cachedRole);

  useEffect(() => {
    const onAuthChange = () => {
      // Login/logout — drop cache; next fetchWhoamiOnce will refill if token exists.
      invalidateWhoamiCache();
      if (!localStorage.getItem('accessToken')) {
        setRole('user');
        setLoading(false);
      }
    };
    window.addEventListener('authChange', onAuthChange);
    return () => window.removeEventListener('authChange', onAuthChange);
  }, []);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (!token) {
      setLoading(false);
      return;
    }

    let cancelled = false;
    void fetchWhoamiOnce().then((r) => {
      if (!cancelled) {
        setRole(r);
        setLoading(false);
      }
    });

    return () => {
      cancelled = true;
    };
  }, []);

  return {
    role,
    loading,
    isAdmin: role === 'admin',
    isAuthor: role === 'author' || role === 'admin',
  };
}
