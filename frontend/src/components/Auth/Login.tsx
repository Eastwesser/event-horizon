import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { login } from '../../services/api';

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
      const { access_token } = response.data;
      
      if (access_token) {
        localStorage.setItem('accessToken', access_token);
        localStorage.setItem('userId', response.data.user_id);
        window.dispatchEvent(new Event('storage'));
        navigate('/');
      }
    } catch (err: any) {
      const status = err.response?.status;
      const message = err.response?.data?.error || err.response?.data?.message;
      
      if (status === 401 || message?.includes('invalid credentials')) {
        setError('Неверный email или пароль. Попробуйте ещё раз.');
      } else if (status === 404) {
        setError('Пользователь с таким email не найден.');
      } else {
        setError('Ошибка соединения. Попробуйте позже.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <h2>Вход в Блинопёк</h2>
      <form onSubmit={handleSubmit}>
        <div>
          <label>🍳 Email:</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="example@mail.com"
            required
          />
        </div>
        <div>
          <label>🔒 Пароль:</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••"
            required
          />
        </div>
        {error && <div className="error">{error}</div>}
        <button type="submit" disabled={loading}>
          {loading ? '🥞 Загрузка...' : '🍴 Войти'}
        </button>
      </form>
      <p>
        👋 Нет аккаунта? <a href="/register">Зарегистрироваться</a>
      </p>
    </div>
  );
}
