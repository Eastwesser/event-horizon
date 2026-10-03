import { Button } from './Button';

interface CatalogPagerProps {
  page: number;
  pageCount: number;
  total: number;
  disabled?: boolean;
  onPageChange: (page: number) => void;
}

export function CatalogPager({
  page,
  pageCount,
  total,
  disabled,
  onPageChange,
}: CatalogPagerProps) {
  if (total === 0) return null;
  return (
    <div className="mt-6 flex flex-wrap items-center justify-between gap-3 text-sm text-text-secondary">
      <span>
        Всего: {total} · Стр. {page} из {pageCount}
      </span>
      <div className="flex gap-2">
        <Button
          type="button"
          size="sm"
          variant="ghost"
          disabled={disabled || page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          Назад
        </Button>
        <Button
          type="button"
          size="sm"
          variant="ghost"
          disabled={disabled || page >= pageCount}
          onClick={() => onPageChange(page + 1)}
        >
          Далее
        </Button>
      </div>
    </div>
  );
}
