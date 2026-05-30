import { useDrag } from 'react-dnd';
import { pancakeEmoji, pancakeColor, type PancakeType } from '../../../utils/hexagon';


interface TrayStack {
  id: number;
  type: PancakeType;
  count: number;
}

interface TrayProps {
  stacks: TrayStack[];
}

export function Tray({ stacks }: TrayProps) {
  return (
    <div className="tray">
      <h3>🍽️ Поднос</h3>
      <div className="tray-stacks">
        {stacks.map((stack) => (
          <TrayStack key={stack.id} stack={stack} />
        ))}
      </div>
    </div>
  );
}

function TrayStack({ stack }: { stack: TrayStack }) {
  const [{ isDragging }, dragRef] = useDrag(() => ({
    type: 'pancake',
    item: { id: stack.id, type: stack.type, count: stack.count },
    collect: (monitor) => ({
      isDragging: !!monitor.isDragging(),
    }),
  }));

  return (
    <div
      ref={dragRef as any}
      className="tray-stack"
      style={{
        backgroundColor: pancakeColor[stack.type],
        opacity: isDragging ? 0.5 : 1,
        cursor: 'grab',
      }}
    >
      <span className="emoji">{pancakeEmoji[stack.type]}</span>
      <span className="count">x{stack.count}</span>
    </div>
  );
}