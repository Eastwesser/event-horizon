import React, { useEffect, useRef, useState } from 'react';
import { Button } from '../ui/Button';
import { Icon } from '../ui/Icon';
import { uploadInventoryImage } from '../../services/inventoryApi';

const inputClass =
  'w-full rounded-sm border border-white/10 bg-nebula px-3 py-2 text-sm text-text-primary focus:border-photon-cyan/50';
const labelClass = 'mb-1 block text-sm text-text-secondary';

interface InventoryImageUrlFieldProps {
  value: string;
  onChange: (url: string) => void;
}

/** URL input + optional local file upload + 96×96 preview for create/edit modals. */
export const InventoryImageUrlField: React.FC<InventoryImageUrlFieldProps> = ({
  value,
  onChange,
}) => {
  const [imgFailed, setImgFailed] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);
  const trimmed = value.trim();

  useEffect(() => {
    setImgFailed(false);
  }, [trimmed]);

  const handleFile = async (file: File | undefined) => {
    if (!file) return;
    setUploadError(null);
    setUploading(true);
    try {
      const url = await uploadInventoryImage(file);
      onChange(url);
    } catch (e) {
      setUploadError(e instanceof Error ? e.message : 'Не удалось загрузить файл');
    } finally {
      setUploading(false);
      if (fileRef.current) fileRef.current.value = '';
    }
  };

  return (
    <div>
      <label className={labelClass}>Изображение (URL или файл)</label>
      <div className="flex items-center gap-2">
        <input
          type="url"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder="https://… или загрузите файл"
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
      <div className="mt-2 flex flex-wrap items-center gap-2">
        <input
          ref={fileRef}
          type="file"
          accept="image/jpeg,image/png,image/webp,.jpg,.jpeg,.png,.webp"
          className="sr-only"
          onChange={(e) => void handleFile(e.target.files?.[0])}
        />
        <Button
          type="button"
          variant="ghost"
          size="sm"
          disabled={uploading}
          onClick={() => fileRef.current?.click()}
        >
          {uploading ? 'Загрузка…' : 'Выбрать файл'}
        </Button>
        <span className="text-xs text-text-muted">JPG, PNG, WebP · до 2 МБ</span>
      </div>
      {uploadError ? <p className="mt-1 text-sm text-rose-400">{uploadError}</p> : null}
      {trimmed ? (
        <div className="mt-2 flex h-24 w-24 items-center justify-center overflow-hidden rounded-sm border border-white/10 bg-white/5">
          {imgFailed ? (
            <Icon name="package" className="h-8 w-8 text-text-muted" />
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
