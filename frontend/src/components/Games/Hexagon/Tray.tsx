import { useDrag } from 'react-dnd';
import { type PancakeType } from '../../../utils/hexagon';
import { TrayPancakeStack } from './pancakeChip';
import { cn } from '../../../lib/cn';

interface TrayStack {
  id: number;
  type: PancakeType;
  count: number;
}

interface TrayProps {
  stacks: TrayStack[];
  skinMode?: 'default' | 'space';
  pickedId?: number | null;
  onPick?: (stack: TrayStack) => void;
}

export function Tray({ stacks, skinMode = 'default', pickedId = null, onPick }: TrayProps) {
  return (
    <div className="tray">
      <h3>Поднос</h3>
      <div className="tray-stacks">
        {stacks.map((stack) => (
          <TrayStackItem
            key={stack.id}
            stack={stack}
            skinMode={skinMode}
            picked={pickedId === stack.id}
            onPick={onPick}
          />
        ))}
      </div>
    </div>
  );
}

function TrayStackItem({
  stack,
  skinMode,
  picked,
  onPick,
}: {
  stack: TrayStack;
  skinMode?: 'default' | 'space';
  picked: boolean;
  onPick?: (stack: TrayStack) => void;
}) {
  const [{ isDragging }, dragRef] = useDrag(() => ({
    type: 'pancake',
    item: { id: stack.id, type: stack.type, count: stack.count },
    collect: (monitor) => ({
      isDragging: !!monitor.isDragging(),
    }),
  }));

  return (
    <button
      type="button"
      ref={dragRef as any}
      className={cn('tray-stack', picked && 'tray-stack--picked')}
      style={{ opacity: isDragging ? 0.5 : 1 }}
      onClick={() => onPick?.(stack)}
    >
      <TrayPancakeStack type={stack.type} count={stack.count} space={skinMode === 'space'} />
    </button>
  );
}
