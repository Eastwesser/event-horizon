// frontend/src/App.tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Login } from './components/Auth/Login';
import { Register } from './components/Auth/Register';
import { Home } from './components/Home/Home';
import { HexagonGame } from './components/Games/Hexagon/HexagonGame';
import { MemoryGame } from './components/Games/Memonia/MemoryGame';
import { FlappyGame } from './components/Games/Flappy/FlappyGame';
import { TowerGame } from './components/Games/Towers/TowerGame';
import { HanoiTower } from './components/Games/Hanoi/HanoiTower';
import { LeaderboardFull } from './components/Leaderboard/LeaderboardFull';
import { Profile } from './components/Profile/Profile';
import { Shop } from './components/Shop/Shop';
import { InventoryPage, InventoryItemDetail } from './components/Inventory';
import { ShopWithInfiniteScroll } from './components/Shop/ShopWithInfiniteScroll';
import { Subscription } from './components/Payment/Subscription';
import { AuthorsPage } from './components/Authors/AuthorsPage';
import { AnalyticsDashboard } from './components/Analytics/AnalyticsDashboard';
import { HistoryPage } from './components/History/HistoryPage';
import { AdminPage } from './components/Admin/AdminPage';
import { useEffect, useState } from 'react';
import { getAccessToken, hydrateAuth } from './lib/auth';
import { ErrorBoundary } from './components/ui/ErrorBoundary';

hydrateAuth();

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(!!getAccessToken());

  useEffect(() => {
    const handleStorageChange = () => {
      setIsAuthenticated(!!getAccessToken());
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
      <ErrorBoundary label="app">
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
            path="/game/hanoi" 
            element={isAuthenticated ? <HanoiTower /> : <Navigate to="/login" />} 
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
          <Route 
            path="/subscription" 
            element={isAuthenticated ? <Subscription /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/authors" 
            element={isAuthenticated ? <AuthorsPage /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/analytics" 
            element={isAuthenticated ? <AnalyticsDashboard /> : <Navigate to="/login" />} 
          />
          <Route 
            path="/history" 
            element={isAuthenticated ? <HistoryPage /> : <Navigate to="/login" />} 
          />
          <Route
            path="/admin"
            element={isAuthenticated ? <AdminPage /> : <Navigate to="/login" />}
          />
        </Routes>
      </ErrorBoundary>
    </BrowserRouter>
  );
}

export default App;

