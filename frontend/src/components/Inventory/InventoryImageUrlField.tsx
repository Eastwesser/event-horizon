import React, { useEffect, useState } from 'react';
import { Button } from '../ui/Button';

const inputClass =
  'w-full rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary focus:border-photon-cyan/50';
const labelClass = 'mb-1 block text-sm text-text-secondary';

interface InventoryImageUrlFieldProps {
  value: string;
  onChange: (url: string) => void;
}

/** URL input + optional 96×96 preview for create/edit inventory modals. */
export const InventoryImageUrlField: React.FC<InventoryImageUrlFieldProps> = ({
  value,
  onChange,
}) => {
  const [imgFailed, setImgFailed] = useState(false);
  const trimmed = value.trim();

  useEffect(() => {
    setImgFailed(false);
  }, [trimmed]);

  return (
    <div>
      <label className={labelClass}>Изображение (URL)</label>
      <div className="flex items-center gap-2">
        <input
          type="url"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder="https://…"
          className={inputClass}
        />
        {trimmed ? (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="shrink-0 px-3"
            onClick={() => onChange('')}
            aria-label="Убрать изображение"
          >
            ×
          </Button>
        ) : null}
      </div>
      {trimmed ? (
        <div className="mt-2 flex h-24 w-24 items-center justify-center overflow-hidden rounded-sm border border-white/10 bg-white/5">
          {imgFailed ? (
            <span className="text-3xl" aria-hidden>
              📦
            </span>
          ) : (
            <img
              src={trimmed}
              alt=""
              className="h-full w-full object-cover"
              onError={() => setImgFailed(true)}
            />
          )}
        </div>
      ) : null}
    </div>
  );
};
