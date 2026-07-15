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

// Космические цвета для подноса
const spaceColors: Record<string, string> = {
  nutella: '#4A2C6B',
  strawberry: '#A29BFE',
  fish: '#6C5CE7',
  sausage: '#74B9FF',
  chicken: '#E17055',
  caesar: '#FDA7DF',
  cranberry: '#00B894',
  pancake: '#FDCB6E',
  default: '#6C5CE7',
};

export function Tray({ stacks, skinMode = 'default' }: TrayProps) {
  return (
    <div className="tray">
      <h3>🍽️ Поднос</h3>
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