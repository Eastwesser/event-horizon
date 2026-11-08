import { Navigate } from 'react-router-dom';
import { useUserRole } from '../../hooks/useUserRole';

export function RequireRole({ roles, children }: { roles: string[]; children: React.ReactNode }) {
  const { role, loading } = useUserRole();
  if (loading) return null;
  if (!role || !roles.includes(role)) return <Navigate to="/" replace />;
  return <>{children}</>;
}