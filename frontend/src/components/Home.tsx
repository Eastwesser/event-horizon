import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

export function Home() {
  const navigate = useNavigate();
  const token = localStorage.getItem('accessToken');

  useEffect(() => {
    console.log('Home page loaded');
    console.log('Token exists:', !!token);
    if (token) {
      console.log('Token preview:', token.substring(0, 30) + '...');
    }
    if (!token) {
      console.log('No token, redirecting to login');
      navigate('/login');
    }
  }, [token, navigate]);

  const handleLogout = () => {
    localStorage.removeItem('accessToken');
    navigate('/login');
  };

  alert('Token: ' + localStorage.getItem('accessToken'));

  return (
    <div className="home-container">
      <h2>🥞 Добро пожаловать в Блинопёк! 🥞</h2>
      <p>Игровое поле скоро будет готово...</p>
      <p>А пока можно выпекать идеи! 🍳</p>
      <button onClick={handleLogout} className="logout-btn">
        🚪 Выйти
      </button>
    </div>
  );
}
