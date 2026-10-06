// frontend/src/components/Games/Hexagon/Tray.tsx
import { useDrag } from 'react-dnd';
import { pancakeEmoji, pancakeColor, type PancakeType } from '../../../utils/hexagon';

interface TrayStack {
  id: number;
  type: PancakeType;
  count: number;
}

interface TrayProps {
  stacks: TrayStack[];
  skinMode?: 'default' | 'space'; // ← добавить
}

// Космические эмодзи - используем реальные типы
const spaceEmojis: Record<string, string> = {
  nutella: '🌙',
  strawberry: '⭐',
  fish: '🌌',
  sausage: '☄️',
  chicken: '🪐',
  caesar: '🌠',
  cranberry: '✨',
  pancake: '☀️',
  default: '🌌',
};

// Космические цвета — тёмная подложка под эмодзи (не розовый кринж)
const spaceColors: Record<string, string> = {
  nutella: '#1a1228',
  strawberry: '#1c1830',
  fish: '#12182a',
  sausage: '#0f1a2e',
  chicken: '#1c1410',
  caesar: '#1a1220',
  cranberry: '#0e1a18',
  pancake: '#1a1810',
  default: '#12141c',
};

export function Tray({ stacks, skinMode = 'default' }: TrayProps) {
  return (
    <div className="tray">
      <h3>Поднос</h3>
      <div className="tray-stacks">
        {stacks.map((stack) => (
          <TrayStack key={stack.id} stack={stack} skinMode={skinMode} />
        ))}
      </div>
    </div>
  );
}

function TrayStack({ stack, skinMode }: { stack: TrayStack; skinMode?: 'default' | 'space' }) {
  const [{ isDragging }, dragRef] = useDrag(() => ({
    type: 'pancake',
    item: { id: stack.id, type: stack.type, count: stack.count },
    collect: (monitor) => ({
      isDragging: !!monitor.isDragging(),
    }),
  }));

  const getEmoji = () => {
    if (skinMode === 'space') {
      return spaceEmojis[stack.type] || spaceEmojis.default || '🌌';
    }
    return pancakeEmoji[stack.type] || '🥞';
  };

  const getColor = () => {
    if (skinMode === 'space') {
      return spaceColors[stack.type] || spaceColors.default || '#6C5CE7';
    }
    return pancakeColor[stack.type] || '#DEB887';
  };

  return (
    <div
      ref={dragRef as any}
      className="tray-stack"
      style={{
        backgroundColor: getColor(),
        opacity: isDragging ? 0.5 : 1,
        cursor: 'grab',
      }}
    >
      <span className="emoji">{getEmoji()}</span>
      <span className="count">x{stack.count}</span>
    </div>
  );
}