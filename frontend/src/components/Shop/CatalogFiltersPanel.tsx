import { useEffect, useMemo, useState, type ReactNode } from 'react';
import {
  COST_TIER_OPTIONS,
  ELEMENT_OPTIONS,
  ICON_FILTER_OPTIONS,
  RARITY_OPTIONS,
  STAT_OP_OPTIONS,
  TYPE_MAIN_OPTIONS,
  type CatalogFacets,
  type CatalogFilterFlags,
  type StatOp,
  countActiveFilters,
  listToParam,
  titleCaseRu,
  toggleListValue,
} from '../../lib/catalogQuery';
import { GAME_LB_TABS } from '../../lib/gameIcons';
import { Button } from '../ui/Button';
import { FilterChip } from '../ui/FilterChip';

const FLAG_CHIPS: { key: keyof CatalogFilterFlags; label: string }[] = [
  { key: 'inStock', label: 'В наличии' },
  { key: 'foil', label: 'Фойл' },
  { key: 'noir', label: 'Нуар' },
  { key: 'flying', label: 'Летающие' },
  { key: 'companion', label: 'Tamagotchi' },
  { key: 'unique', label: 'Уникальная' },
  { key: 'symbiont', label: 'Симбионт' },
  { key: 'parasite', label: 'Паразит' },
];

const FLAG_TO_PARAM: Record<string, string> = {
  inStock: 'in_stock',
  foil: 'foil',
  noir: 'noir',
  flying: 'flying',
  companion: 'companion',
  unique: 'unique',
  symbiont: 'symbiont',
  parasite: 'parasite',
};

interface CatalogFiltersPanelProps {
  filters: CatalogFilterFlags;
  facets: CatalogFacets;
  /** Patch URL filter params; always resets page to 1 in parent. */
  onChange: (patch: Record<string, string | null>) => void;
  onReset: () => void;
  /** Debounced name search draft (parent owns debounce → ?q=). */
  queryDraft: string;
  onQueryDraftChange: (v: string) => void;
  /** Shop type chip — card filters only when `карточка`. */
  productType?: string;
}

function Group({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="space-y-2">
      <h4 className="text-xs font-semibold uppercase tracking-wide text-text-muted">
        {title}
      </h4>
      {children}
    </div>
  );
}

function ChipWrap({ children }: { children: ReactNode }) {
  return <div className="flex flex-wrap gap-1.5">{children}</div>;
}

const COLLAPSE_AT = 8;

/** Show first 8 chips + «Показать все (N)» when the group is large. */
function CollapsibleChips({
  total,
  children,
}: {
  total: number;
  children: ReactNode[];
}) {
  const [expanded, setExpanded] = useState(false);
  const needsCollapse = total > COLLAPSE_AT;
  const visible =
    !needsCollapse || expanded ? children : children.slice(0, COLLAPSE_AT);

  return (
    <div className="space-y-2">
      <ChipWrap>{visible}</ChipWrap>
      {needsCollapse ? (
        <button
          type="button"
          className="text-xs text-horizon-cyan underline-offset-2 hover:underline"
          onClick={() => setExpanded((v) => !v)}
        >
          {expanded
            ? 'Свернуть'
            : `Показать все (${total})`}
        </button>
      ) : null}
    </div>
  );
}

function StatRow({
  label,
  op,
  value,
  onOp,
  onValue,
}: {
  label: string;
  op: StatOp;
  value: string;
  onOp: (op: StatOp) => void;
  onValue: (v: string) => void;
}) {
  return (
    <div className="flex items-center gap-2 text-sm">
      <span className="w-16 shrink-0 text-text-muted">{label}</span>
      <select
        className="rounded-sm border border-white/10 bg-void px-1.5 py-1 text-text-primary"
        value={op}
        onChange={(e) => onOp(e.target.value as StatOp)}
      >
        {STAT_OP_OPTIONS.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
      <input
        type="number"
        className="w-20 rounded-sm border border-white/10 bg-void px-2 py-1 text-text-primary"
        value={value}
        onChange={(e) => onValue(e.target.value)}
        placeholder="—"
      />
    </div>
  );
}

export function CatalogFiltersPanel({
  filters,
  facets,
  onChange,
  onReset,
  queryDraft,
  onQueryDraftChange,
  productType = 'all',
}: CatalogFiltersPanelProps) {
  const cardFilters = productType === 'карточка';
  const skinFilters = productType === 'скин';
  const themeFilters = productType === 'тема';
  const physicalFilters = ['мерч', 'брелок', 'картина', 'фенечка'].includes(productType);
  const activeCount = countActiveFilters(filters);
  const [open, setOpen] = useState(activeCount > 0);

  useEffect(() => {
    if (activeCount > 0) setOpen(true);
  }, [activeCount]);

  const [artistDraft, setArtistDraft] = useState('');
  const selectedArtistName = useMemo(() => {
    if (!filters.artistId) return '';
    return facets.artists.find((a) => a.id === filters.artistId)?.name || filters.artistId;
  }, [filters.artistId, facets.artists]);

  useEffect(() => {
    setArtistDraft(selectedArtistName);
  }, [selectedArtistName]);

  const artistSuggestions = useMemo(() => {
    const q = artistDraft.trim().toLowerCase();
    if (!q || filters.artistId) return [];
    return facets.artists
      .filter((a) => a.name.toLowerCase().includes(q))
      .slice(0, 8);
  }, [artistDraft, facets.artists, filters.artistId]);

  const toggleMulti = (param: string, current: string[], value: string) => {
    onChange({ [param]: listToParam(toggleListValue(current, value)), page: '1' });
  };

  const toggleNum = (param: string, current: number[], value: number) => {
    const asStr = current.map(String);
    const next = toggleListValue(asStr, String(value));
    onChange({ [param]: listToParam(next), page: '1' });
  };

  const setFlag = (key: keyof CatalogFilterFlags, on: boolean) => {
    const param = FLAG_TO_PARAM[key as string];
    if (!param) return;
    onChange({ [param]: on ? null : '1', page: '1' });
  };

  const setStat = (
    opKey: string,
    valKey: string,
    op: StatOp,
    value: string
  ) => {
    if (value === '') {
      onChange({ [opKey]: null, [valKey]: null, page: '1' });
      return;
    }
    onChange({ [opKey]: op, [valKey]: value, page: '1' });
  };

  const inputClass =
    'w-full rounded-sm border border-white/10 bg-void px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:border-photon-cyan/50';

  return (
    <div className="mb-4">
      <div className="mb-2 flex flex-wrap items-center gap-2">
        <Button
          variant={open || activeCount > 0 ? 'secondary' : 'ghost'}
          size="sm"
          onClick={() => setOpen((v) => !v)}
          aria-expanded={open}
        >
          {activeCount > 0 ? `Фильтры (${activeCount})` : 'Фильтры'}
          <span className="ml-1 text-text-muted">{open ? '▲' : '▼'}</span>
        </Button>
        {activeCount > 0 ? (
          <Button variant="ghost" size="sm" onClick={onReset}>
            Сбросить фильтры
          </Button>
        ) : null}
      </div>

      {open && !cardFilters ? (
        <div className="rounded-md border border-white/10 bg-nebula/60 p-4 space-y-4">
          <Group title="Название">
            <input
              type="search"
              className={inputClass}
              placeholder="Поиск по названию…"
              value={queryDraft}
              onChange={(e) => onQueryDraftChange(e.target.value)}
            />
          </Group>
          {skinFilters ? (
            <Group title="Игра">
              <ChipWrap>
                {GAME_LB_TABS.map((g) => (
                  <FilterChip
                    key={g.id}
                    active={filters.games?.includes(g.id)}
                    onClick={() => toggleMulti('game', filters.games || [], g.id)}
                  >
                    {g.label}
                  </FilterChip>
                ))}
              </ChipWrap>
            </Group>
          ) : null}
          {themeFilters ? (
            <Group title="Настроение темы">
              <ChipWrap>
                {[
                  { value: 'light', label: 'Светлая' },
                  { value: 'dark', label: 'Тёмная' },
                  { value: 'cozy', label: 'Уютная' },
                  { value: 'cosmic', label: 'Космос' },
                ].map((o) => (
                  <FilterChip
                    key={o.value}
                    active={filters.themeNiches?.includes(o.value)}
                    onClick={() =>
                      toggleMulti('theme_niche', filters.themeNiches || [], o.value)
                    }
                  >
                    {o.label}
                  </FilterChip>
                ))}
              </ChipWrap>
            </Group>
          ) : null}
          {physicalFilters || productType === 'all' ? (
            <Group title="Цена (билетики)">
              <p className="text-xs text-text-muted">
                Сортировка «Цена ↑/↓» над каталогом. Физ. мерч от 100 000 билетиков.
              </p>
            </Group>
          ) : null}
          <p className="text-xs text-text-muted">
            Фильтры стихии / класса / редкости — только во вкладке «Карточки».
          </p>
        </div>
      ) : null}

      {open && cardFilters ? (
        <div className="rounded-md border border-white/10 bg-nebula/60 p-4">
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* LEFT */}
            <div className="space-y-4">
              <Group title="Название">
                <input
                  type="search"
                  className={inputClass}
                  placeholder="Поиск по названию…"
                  value={queryDraft}
                  onChange={(e) => onQueryDraftChange(e.target.value)}
                />
              </Group>

              <Group title="Стихия">
                <ChipWrap>
                  {ELEMENT_OPTIONS.map((o) => (
                    <FilterChip
                      key={o.value}
                      active={filters.elements?.includes(o.value)}
                      onClick={() =>
                        toggleMulti('element', filters.elements || [], o.value)
                      }
                    >
                      {o.label}
                    </FilterChip>
                  ))}
                </ChipWrap>
              </Group>

              <Group title="Класс">
                {facets.classes.length === 0 ? (
                  <p className="text-xs text-text-muted">Нет данных</p>
                ) : (
                  <CollapsibleChips total={facets.classes.length}>
                    {facets.classes.map((c) => {
                      const key = c.toLowerCase();
                      const active = (filters.classes || []).some(
                        (x) => x.toLowerCase() === key
                      );
                      return (
                        <FilterChip
                          key={key}
                          active={active}
                          onClick={() =>
                            toggleMulti('class', filters.classes || [], key)
                          }
                        >
                          {titleCaseRu(c)}
                        </FilterChip>
                      );
                    })}
                  </CollapsibleChips>
                )}
              </Group>

              <Group title="Редкость">
                <ChipWrap>
                  {RARITY_OPTIONS.map((o) => (
                    <FilterChip
                      key={o.value}
                      active={filters.rarities?.includes(o.value)}
                      onClick={() =>
                        toggleMulti('rarity', filters.rarities || [], o.value)
                      }
                    >
                      {o.label}
                    </FilterChip>
                  ))}
                </ChipWrap>
              </Group>

              <Group title="Тип">
                <ChipWrap>
                  {TYPE_MAIN_OPTIONS.map((o) => (
                    <FilterChip
                      key={o.value}
                      active={filters.typeMains?.includes(o.value)}
                      onClick={() =>
                        toggleMulti('type_main', filters.typeMains || [], o.value)
                      }
                    >
                      {o.label}
                    </FilterChip>
                  ))}
                </ChipWrap>
              </Group>

              <Group title="Флаги">
                <ChipWrap>
                  {FLAG_CHIPS.map((f) => (
                    <FilterChip
                      key={f.key}
                      active={Boolean(filters[f.key])}
                      onClick={() => setFlag(f.key, Boolean(filters[f.key]))}
                    >
                      {f.label}
                    </FilterChip>
                  ))}
                </ChipWrap>
              </Group>
            </div>

            {/* RIGHT */}
            <div className="space-y-4">
              <Group title="Выпуск">
                <ChipWrap>
                  {facets.sets.map((s) => (
                    <FilterChip
                      key={s}
                      active={filters.sets?.includes(s)}
                      onClick={() => toggleNum('set', filters.sets || [], s)}
                    >
                      #{s}
                    </FilterChip>
                  ))}
                </ChipWrap>
              </Group>

              <Group title="Год">
                <ChipWrap>
                  {facets.years.map((y) => (
                    <FilterChip
                      key={y}
                      active={filters.years?.includes(y)}
                      onClick={() => toggleNum('year', filters.years || [], y)}
                    >
                      {y}
                    </FilterChip>
                  ))}
                </ChipWrap>
              </Group>

              <Group title="Стоимость (ряд / элита)">
                <ChipWrap>
                  {COST_TIER_OPTIONS.map((o) => (
                    <FilterChip
                      key={o.value}
                      active={filters.costTiers?.includes(o.value)}
                      onClick={() =>
                        toggleMulti('cost_tier', filters.costTiers || [], o.value)
                      }
                    >
                      {o.label}
                    </FilterChip>
                  ))}
                </ChipWrap>
              </Group>

              <Group title="Характеристики">
                <div className="space-y-2">
                  <StatRow
                    label="Стоим."
                    op={filters.cost?.op || 'eq'}
                    value={filters.cost != null ? String(filters.cost.value) : ''}
                    onOp={(op) =>
                      setStat(
                        'cost_op',
                        'cost',
                        op,
                        filters.cost != null ? String(filters.cost.value) : ''
                      )
                    }
                    onValue={(v) =>
                      setStat('cost_op', 'cost', filters.cost?.op || 'eq', v)
                    }
                  />
                  <StatRow
                    label="HP"
                    op={filters.hp?.op || 'eq'}
                    value={filters.hp != null ? String(filters.hp.value) : ''}
                    onOp={(op) =>
                      setStat(
                        'hp_op',
                        'hp',
                        op,
                        filters.hp != null ? String(filters.hp.value) : ''
                      )
                    }
                    onValue={(v) =>
                      setStat('hp_op', 'hp', filters.hp?.op || 'eq', v)
                    }
                  />
                  <StatRow
                    label="Ход"
                    op={filters.move?.op || 'eq'}
                    value={filters.move != null ? String(filters.move.value) : ''}
                    onOp={(op) =>
                      setStat(
                        'move_op',
                        'move',
                        op,
                        filters.move != null ? String(filters.move.value) : ''
                      )
                    }
                    onValue={(v) =>
                      setStat('move_op', 'move', filters.move?.op || 'eq', v)
                    }
                  />
                </div>
              </Group>

              <Group title="Удар (кубики)">
                <input
                  type="text"
                  className={inputClass}
                  placeholder='Точно, напр. 2-2-3'
                  value={filters.attackDice || ''}
                  onChange={(e) => {
                    const v = e.target.value.trim();
                    onChange({ attack: v || null, page: '1' });
                  }}
                />
              </Group>

              <Group title="Иконки (любая из)">
                <CollapsibleChips total={ICON_FILTER_OPTIONS.length}>
                  {ICON_FILTER_OPTIONS.map((o) => (
                    <FilterChip
                      key={o.value}
                      active={filters.icons?.includes(o.value)}
                      onClick={() =>
                        toggleMulti('icons', filters.icons || [], o.value)
                      }
                    >
                      {o.label}
                    </FilterChip>
                  ))}
                </CollapsibleChips>
              </Group>

              <Group title="Художник">
                <div className="relative">
                  <input
                    type="search"
                    className={inputClass}
                    placeholder="Начните вводить имя…"
                    value={artistDraft}
                    onChange={(e) => {
                      setArtistDraft(e.target.value);
                      if (filters.artistId) {
                        onChange({ artist: null, page: '1' });
                      }
                    }}
                  />
                  {filters.artistId ? (
                    <button
                      type="button"
                      className="mt-1 text-xs text-horizon-cyan underline-offset-2 hover:underline"
                      onClick={() => {
                        setArtistDraft('');
                        onChange({ artist: null, page: '1' });
                      }}
                    >
                      Сбросить художника
                    </button>
                  ) : null}
                  {artistSuggestions.length > 0 ? (
                    <ul className="absolute z-10 mt-1 max-h-48 w-full overflow-auto rounded-sm border border-white/15 bg-void py-1 shadow-lg">
                      {artistSuggestions.map((a) => (
                        <li key={a.id}>
                          <button
                            type="button"
                            className="block w-full px-3 py-1.5 text-left text-sm text-text-primary hover:bg-white/5"
                            onClick={() => {
                              setArtistDraft(a.name);
                              onChange({ artist: a.id, page: '1' });
                            }}
                          >
                            {a.name}
                          </button>
                        </li>
                      ))}
                    </ul>
                  ) : null}
                </div>
              </Group>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
