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
      console.log('📡 Login response:', response.data);
      
      const { access_token, user_id } = response.data;
      
      if (access_token) {
        localStorage.setItem('accessToken', access_token);
        if (user_id) {
            localStorage.setItem('userId', user_id);
            console.log('✅ Saved userId:', user_id);
        } else {
            // Парсим из токена
            try {
                const payload = JSON.parse(atob(access_token.split('.')[1]));
                const uid = payload.user_id;
                if (uid) {
                    localStorage.setItem('userId', uid);
                    console.log('🔧 Extracted userId from token:', uid);
                }
            } catch (e) {
                console.error('Failed to parse token', e);
            }
        }
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