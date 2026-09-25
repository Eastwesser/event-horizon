import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { register } from '../../services/api';
import { Button } from '../ui/Button';

const inputClass =
  'rounded-md border border-white/10 bg-void px-4 py-3 text-base text-text-primary placeholder:text-text-muted focus:border-photon-cyan/60 focus-visible:outline-none';

export function Register() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.SyntheticEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setSuccess('');

    try {
      const response = await register(email, password);
      const data = response.data;
      if (data.success) {
        const uid = data.user_id || data.userId;
        if (uid) localStorage.setItem('userId', uid);
        setSuccess('Регистрация прошла успешно! Перенаправляем на вход...');
        setTimeout(() => navigate('/login'), 1500);
      } else {
        const msg = data.message || data.error;
        if (msg?.includes('already exists')) {
          setError('Пользователь с таким email уже существует');
        } else {
          setError(msg || 'Ошибка регистрации');
        }
      }
    } catch (err: any) {
      const status = err.response?.status;
      const msg = err.response?.data?.error || err.response?.data?.message || err.message;
      if (status === 409 || msg?.includes('already exists')) {
        setError('Пользователь с таким email уже существует');
      } else if (status === 400 || msg?.includes('password') || msg?.includes('validation')) {
        setError('Пароль должен быть от 8 до 128 символов');
      } else {
        setError(msg || 'Ошибка соединения. Попробуйте позже.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-void px-6 py-12 sm:px-8">
      <div className="eh-glow" aria-hidden="true" />
      <div className="eh-ring relative w-full max-w-md rounded-lg border border-white/10 bg-nebula/90 px-8 py-10 shadow-elevated backdrop-blur-sm sm:px-10">
        <h1 className="mb-2 text-center font-display text-2xl font-semibold text-text-primary sm:text-3xl">
          Регистрация
        </h1>
        <p className="mb-8 text-center text-sm text-text-secondary">Создайте аккаунт в Event Horizon</p>

        <form onSubmit={handleSubmit} className="flex flex-col gap-5" noValidate>
          <label className="flex flex-col gap-2 text-left text-sm font-medium text-text-secondary">
            Email
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="example@mail.com"
              required
              className={inputClass}
            />
          </label>
          <label className="flex flex-col gap-2 text-left text-sm font-medium text-text-secondary">
            Пароль (мин. 8 символов)
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
              minLength={8}
              maxLength={128}
              className={inputClass}
            />
          </label>

          {error && (
            <div role="alert" className="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">
              {error}
            </div>
          )}
          {success && (
            <div role="status" className="rounded-sm border border-success/30 bg-success/10 px-3 py-2 text-sm text-success">
              {success}
            </div>
          )}

          <Button type="submit" disabled={loading} className="mt-1 w-full">
            {loading ? 'Создаём аккаунт…' : 'Зарегистрироваться'}
          </Button>
        </form>

        <p className="mt-8 text-center text-sm text-text-secondary">
          Уже есть аккаунт?{' '}
          <Link to="/login" className="font-medium text-indigo-soft hover:underline">
            Войти
          </Link>
        </p>
      </div>
    </div>
  );
}
