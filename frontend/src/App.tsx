// frontend/src/App.tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Login } from './components/Auth/Login';
import { Register } from './components/Auth/Register';
import { Home } from './components/Home/Home';
import { HexagonGame } from './components/Games/Hexagon/HexagonGame';
import { MemoryGame } from './components/Games/Memonia/MemoryGame';
import { FlappyGame } from './components/Games/Flappy/FlappyGame';
import { TowerGame } from './components/Games/Towers/TowerGame';
import { LeaderboardFull } from './components/Leaderboard/LeaderboardFull';
import { Profile } from './components/Profile/Profile';
import { Shop } from './components/Shop/Shop';
import { InventoryPage, InventoryItemDetail } from './components/Inventory';
import { ShopWithInfiniteScroll } from './components/Shop/ShopWithInfiniteScroll';
import { useEffect, useState } from 'react';

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(!!localStorage.getItem('accessToken'));

  useEffect(() => {
    const handleStorageChange = () => {
      setIsAuthenticated(!!localStorage.getItem('accessToken'));
    };
    
    window.addEventListener('storage', handleStorageChange);
    window.addEventListener('authChange', handleStorageChange);
    
    return () => {
      window.removeEventListener('storage', handleStorageChange);
      window.removeEventListener('authChange', handleStorageChange);
    };
  }, []);

  return (
    <BrowserRouter>
      <div className="app">
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route 
            path="/" 
            element={isAuthenticated ? <Home /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/game/hexagon" 
            element={isAuthenticated ? <HexagonGame /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/game/flappy" 
            element={isAuthenticated ? <FlappyGame /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/game/memory" 
            element={isAuthenticated ? <MemoryGame /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/game/towers" 
            element={isAuthenticated ? <TowerGame /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/leaderboard" 
            element={isAuthenticated ? <LeaderboardFull /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/profile" 
            element={isAuthenticated ? <Profile /> : <Navigate to="/login" />} 
          />        
          <Route 
            path="/shop" 
            element={isAuthenticated ? <Shop /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/infiniteshop" 
            element={isAuthenticated ? <ShopWithInfiniteScroll /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/inventory" 
            element={isAuthenticated ? <InventoryPage /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/inventory/:id" 
            element={isAuthenticated ? <InventoryItemDetail /> : <Navigate to="/login" />} 
          />
        </Routes>
      </div>
    </BrowserRouter>
  );
}

export default App;
