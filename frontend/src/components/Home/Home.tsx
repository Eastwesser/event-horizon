// frontend/src/components/Home/Home.tsx
import { useNavigate } from 'react-router-dom';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { AppFooter, AppNavbar, shellInner } from '../Layout/AppNavbar';

const games = [
  {
    id: 'hexagon',
    name: 'Pancaker',
    description: 'Гексагональный пазл с блинчиками',
    icon: '🥞',
    path: '/game/hexagon',
    available: true,
  },
  {
    id: 'flappy',
    name: 'Flappy Bird',
    description: 'Лети и не врезайся в трубы',
    icon: '🐦',
    path: '/game/flappy',
    available: true,
  },
  {
    id: 'towers',
    name: 'Builder',
    description: 'Строй башню из падающих блоков',
    icon: '🗼',
    path: '/game/towers',
    available: true,
  },
  {
    id: 'hanoi',
    name: 'Hanoi',
    description: 'Классическая головоломка с кольцами',
    icon: '🪈',
    path: '/game/hanoi',
    available: true,
  },
  {
    id: 'memory',
    name: 'Memonia',
    description: 'Найди пары фруктов',
    icon: '🎴',
    path: '/game/memory',
    available: true,
  },
];

export function Home() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-void font-body text-text-primary">
      <AppNavbar />

      <main className={shellInner}>
        <section className="relative py-12 sm:py-16">
          <div className="grid items-center gap-12 lg:grid-cols-2 lg:gap-16">
            <div className="text-center lg:text-left">
              <h1 className="font-display text-4xl font-bold leading-tight text-text-primary sm:text-5xl">
                Выбери игру и ставь рекорды
              </h1>
              <p className="mx-auto mt-4 max-w-lg text-lg text-text-secondary lg:mx-0">
                Играй в мини-игры, зарабатывай лампочки и билетики, становись лучшим в лидерборде.
              </p>
              <div className="mt-8 flex flex-wrap justify-center gap-4 lg:justify-start">
                <Button
                  size="md"
                  onClick={() => document.getElementById('games')?.scrollIntoView({ behavior: 'smooth' })}
                >
                  Все игры
                </Button>
                <Button variant="ghost" size="md" onClick={() => navigate('/leaderboard')}>
                  🏆 Лидерборд
                </Button>
              </div>
            </div>

            {/* Decorative accretion disk — brand mark in the core (no flagship game). */}
            <div className="eh-disk mx-auto" aria-hidden="true">
              <div className="eh-disk-rings">
                <div className="eh-disk-ring eh-disk-ring--lensed" />
                <div className="eh-disk-ring eh-disk-ring--mid" />
                <div className="eh-disk-ring eh-disk-ring--main" />
              </div>
              <div className="eh-disk-core" />
              <div className="absolute left-1/2 top-1/2 flex w-[42%] -translate-x-1/2 -translate-y-1/2 flex-col items-center justify-center">
                <img
                  src="/images/brand/logo-minimal.png"
                  alt=""
                  className="h-auto w-full rounded-full object-cover opacity-95"
                />
              </div>
            </div>
          </div>
        </section>

        <section id="games" className="scroll-mt-24 pb-16 pt-4">
          <h2 className="mb-6 font-display text-xl font-semibold text-text-primary">Игры</h2>
          <div className="grid grid-cols-1 items-stretch gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
            {games.map((game, i) => (
              <Card
                key={game.id}
                interactive={game.available}
                className="eh-stagger-in flex h-full min-h-[220px] flex-col items-center gap-2 text-center"
                style={{ animationDelay: `${i * 70}ms` }}
                onClick={() => game.available && navigate(game.path)}
              >
                <div className="text-5xl">{game.icon}</div>
                <h3 className="font-display text-lg font-semibold text-text-primary">{game.name}</h3>
                <p className="min-h-[2.5rem] flex-1 text-sm text-text-secondary">{game.description}</p>
                <Badge
                  tone={game.available ? 'indigo' : 'neutral'}
                  className="mt-auto px-4 py-1.5 text-sm"
                >
                  {game.available ? 'Играть' : 'Скоро'}
                </Badge>
              </Card>
            ))}
          </div>
        </section>
      </main>

      <AppFooter />
    </div>
  );
}
