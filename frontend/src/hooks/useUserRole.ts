import { useEffect, useState } from 'react';
import api from '../services/api';

export type UserRole = 'user' | 'author' | 'admin';

export function useUserRole() {
  const [role, setRole] = useState<UserRole>(
    () => (localStorage.getItem('role') as UserRole) || 'user'
  );
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (!token) {
      setLoading(false);
      return;
    }

    const fetchRole = async () => {
      try {
        const { data } = await api.get('/auth/whoami');
        const r = (data.role as UserRole) || 'user';
        setRole(r);
        localStorage.setItem('role', r);
      } catch {
        const cached = localStorage.getItem('role') as UserRole;
        if (cached) setRole(cached);
      } finally {
        setLoading(false);
      }
    };

    fetchRole();
  }, []);

  return {
    role,
    loading,
    isAdmin: role === 'admin',
    isAuthor: role === 'author' || role === 'admin',
  };
}
