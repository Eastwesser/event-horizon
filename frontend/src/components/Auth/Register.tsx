import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { register } from '../../services/api';

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
    <div className="auth-container">
      <h2>📝 Регистрация</h2>
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
          <label>🔒 Пароль (мин. 8 символов):</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
            required
            minLength={8}
            maxLength={128}
          />
        </div>
        {error && <div className="error">{error}</div>}
        {success && <div className="success">{success}</div>}
        <button type="submit" disabled={loading}>
          {loading ? '🍳 Создаём аккаунт...' : '🥞 Зарегистрироваться'}
        </button>
      </form>
      <p>
        👋 Уже есть аккаунт? <a href="/login">Войти</a>
      </p>
    </div>
  );
}
