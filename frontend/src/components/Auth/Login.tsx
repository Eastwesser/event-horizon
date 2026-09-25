import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { login } from '../../services/api';
import { setTokens } from '../../lib/auth';
import { Button } from '../ui/Button';

const inputClass =
  'rounded-md border border-white/10 bg-void px-4 py-3 text-base text-text-primary placeholder:text-text-muted focus:border-photon-cyan/60 focus-visible:outline-none';

export function Login() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.SyntheticEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await login(email, password);
      const { access_token, refresh_token, user_id } = response.data;

      if (access_token) {
        setTokens(access_token, refresh_token);
        if (user_id) {
          localStorage.setItem('userId', user_id);
          localStorage.setItem('userEmail', email);
        } else {
          try {
            const payload = JSON.parse(atob(access_token.split('.')[1]));
            if (payload.user_id) localStorage.setItem('userId', payload.user_id);
          } catch {
            /* ignore parse errors */
          }
        }
        navigate('/');
      } else {
        setError('Не удалось получить токен');
      }
    } catch (err: any) {
      const status = err.response?.status;
      const message = err.response?.data?.error || err.response?.data?.message;

      if (status === 401 || message?.includes('invalid credentials')) {
        setError('Неверный email или пароль. Попробуйте ещё раз.');
      } else if (status === 400 || message?.includes('password')) {
        setError('Пароль должен быть от 8 до 128 символов');
      } else if (status === 404) {
        setError('Пользователь с таким email не найден.');
      } else {
        setError(message || 'Ошибка соединения. Попробуйте позже.');
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
          Вход в <span className="text-horizon-gold">Event Horizon</span>
        </h1>
        <p className="mb-8 text-center text-sm text-text-secondary">Рады видеть вас снова</p>

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
            Пароль
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
              className={inputClass}
            />
          </label>

          {error && (
            <div role="alert" className="rounded-sm border border-error/30 bg-error/10 px-3 py-2 text-sm text-error">
              {error}
            </div>
          )}

          <Button type="submit" disabled={loading} className="mt-1 w-full">
            {loading ? 'Загрузка…' : 'Войти'}
          </Button>
        </form>

        <p className="mt-8 text-center text-sm text-text-secondary">
          Нет аккаунта?{' '}
          <Link to="/register" className="font-medium text-indigo-soft hover:underline">
            Зарегистрироваться
          </Link>
        </p>
      </div>
    </div>
  );
}
