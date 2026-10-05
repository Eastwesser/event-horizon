import { useEffect, useState } from 'react';
import api from '../services/api';
import {
  getWhoamiCache,
  invalidateWhoamiCache,
  setWhoamiCache,
  setWhoamiInflight,
  type UserRole,
} from '../lib/whoamiCache';

export type { UserRole };
export { invalidateWhoamiCache };

function roleFromStorage(): UserRole {
  return (localStorage.getItem('role') as UserRole) || 'user';
}

async function fetchWhoamiOnce(): Promise<UserRole> {
  const { role: cachedRole, at: cachedAt, inflight, cacheMs } = getWhoamiCache();
  const now = Date.now();
  if (cachedRole && now - cachedAt < cacheMs) return cachedRole;
  if (inflight) return inflight;

  const p = api
    .get('/auth/whoami')
    .then(({ data }) => {
      const r = (data.role as UserRole) || 'user';
      setWhoamiCache(r);
      localStorage.setItem('role', r);
      return r;
    })
    .catch(() => {
      const fallback = roleFromStorage();
      setWhoamiCache(fallback);
      return fallback;
    })
    .finally(() => {
      setWhoamiInflight(null);
    });

  setWhoamiInflight(p);
  return p;
}

export function useUserRole() {
  const [role, setRole] = useState<UserRole>(() => getWhoamiCache().role ?? roleFromStorage());
  const [loading, setLoading] = useState(!getWhoamiCache().role);

  useEffect(() => {
    const onAuthChange = () => {
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
