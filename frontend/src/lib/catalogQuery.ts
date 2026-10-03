/** Shared shop/inventory catalog paging + sorting + filters (URL-driven). */

export const CATALOG_PAGE_SIZE = 100;

export type CatalogSort =
  | 'newest'
  | 'name_asc'
  | 'name_desc'
  | 'rarity_asc'
  | 'rarity_desc'
  | 'price_asc'
  | 'price_desc'
  | 'artist_asc'
  | 'element_asc'
  | 'set_asc'
  | 'set_desc';

export const CATALOG_SORT_OPTIONS: { value: CatalogSort; label: string }[] = [
  { value: 'newest', label: 'Новые' },
  { value: 'name_asc', label: 'Имя A–Z' },
  { value: 'name_desc', label: 'Имя Z–A' },
  { value: 'rarity_asc', label: 'Редкость ↑' },
  { value: 'rarity_desc', label: 'Редкость ↓' },
  { value: 'price_asc', label: 'Цена ↑' },
  { value: 'price_desc', label: 'Цена ↓' },
  { value: 'artist_asc', label: 'Художник A–Z' },
  { value: 'element_asc', label: 'Стихия A–Z' },
  { value: 'set_asc', label: 'Выпуск ↑' },
  { value: 'set_desc', label: 'Выпуск ↓' },
];

export type StatOp = 'eq' | 'gt' | 'lt' | 'gte' | 'lte';

export const STAT_OP_OPTIONS: { value: StatOp; label: string }[] = [
  { value: 'eq', label: '=' },
  { value: 'gt', label: '>' },
  { value: 'lt', label: '<' },
  { value: 'gte', label: '≥' },
  { value: 'lte', label: '≤' },
];

const RARITY_RANK: Record<string, number> = {
  common: 0,
  uncommon: 1,
  rare: 2,
  ultra: 3,
};

export const ELEMENT_OPTIONS: { value: string; label: string }[] = [
  { value: 'mountains', label: 'Горы' },
  { value: 'woods', label: 'Лес' },
  { value: 'steppes', label: 'Степи' },
  { value: 'swamps', label: 'Болота' },
  { value: 'darkness', label: 'Тьма' },
  { value: 'neutral', label: 'Нейтралы' },
];

export const RARITY_OPTIONS: { value: string; label: string }[] = [
  { value: 'common', label: 'Обычная' },
  { value: 'uncommon', label: 'Необычная' },
  { value: 'rare', label: 'Редкая' },
  { value: 'ultra', label: 'Ультра' },
];

export const COST_TIER_OPTIONS: { value: string; label: string }[] = [
  { value: 'rank_and_file', label: 'Рядовая' },
  { value: 'elite', label: 'Элитная' },
];

export const TYPE_MAIN_OPTIONS: { value: string; label: string }[] = [
  { value: 'creature', label: 'Существо' },
  { value: 'artifact', label: 'Артефакт' },
  { value: 'land', label: 'Местность' },
];

/** Only icon types present in seeded catalog (curl-verified). */
/** Distinct attributes.icons[].type from live catalog (280 cards). No counter/uchr/tap/instant. */
export const ICON_FILTER_OPTIONS: { value: string; label: string }[] = [
  { value: 'armor', label: 'Доспех' },
  { value: 'zoal', label: 'ЗОЛ' },
  { value: 'zov', label: 'ЗОВ' },
  { value: 'zoz', label: 'ЗОЗ' },
  { value: 'zot', label: 'ЗОТ' },
  { value: 'zor', label: 'ЗОР' },
  { value: 'zoo', label: 'ЗОО' },
  { value: 'zom', label: 'ЗОМ' },
  { value: 'regen', label: 'Реген' },
  { value: 'stamina', label: 'Стойкость' },
  { value: 'direct', label: 'Направл.' },
  { value: 'ova', label: 'ОВА' },
  { value: 'ovz', label: 'ОВЗ' },
  { value: 'ovs', label: 'ОВС' },
];

export const FLAG_FILTER_KEYS = [
  'foil',
  'noir',
  'flying',
  'companion',
  'symbiont',
  'parasite',
  'unique',
  'in_stock',
] as const;

/** URL keys cleared by «Сбросить фильтры» (sort / type stay). */
export const FILTER_URL_KEYS = [
  'q',
  'element',
  'rarity',
  'cost_tier',
  'set',
  'year',
  'class',
  'artist',
  'type_main',
  'icons',
  'attack',
  'cost_op',
  'cost',
  'hp_op',
  'hp',
  'move_op',
  'move',
  ...FLAG_FILTER_KEYS,
] as const;

const ELEMENT_RU: Record<string, string> = Object.fromEntries(
  ELEMENT_OPTIONS.map((o) => [o.value, o.label.toLowerCase()])
);

export interface CatalogSortable {
  id?: string;
  name: string;
  price?: number;
  price_tickets?: number;
  stock?: number | null;
  created_at?: string;
  attributes?: Record<string, unknown> | null;
}

export interface StatFilter {
  op: StatOp;
  value: number;
}

export interface CatalogFilterFlags {
  inStock?: boolean;
  foil?: boolean;
  noir?: boolean;
  flying?: boolean;
  companion?: boolean;
  symbiont?: boolean;
  parasite?: boolean;
  unique?: boolean;
  query?: string;
  elements?: string[];
  rarities?: string[];
  costTiers?: string[];
  sets?: number[];
  years?: number[];
  classes?: string[];
  artistId?: string;
  typeMains?: string[];
  icons?: string[];
  attackDice?: string;
  cost?: StatFilter | null;
  hp?: StatFilter | null;
  move?: StatFilter | null;
}

export interface CatalogFacets {
  classes: string[];
  years: number[];
  sets: number[];
  artists: { id: string; name: string }[];
}

function attrStr(item: CatalogSortable, key: string): string {
  const v = item.attributes?.[key];
  return typeof v === 'string' ? v : v != null ? String(v) : '';
}

function attrNum(item: CatalogSortable, key: string): number {
  const v = item.attributes?.[key];
  return typeof v === 'number' && Number.isFinite(v) ? v : 0;
}

function attrNumOrNull(item: CatalogSortable, key: string): number | null {
  const v = item.attributes?.[key];
  if (typeof v === 'number' && Number.isFinite(v)) return v;
  if (typeof v === 'string' && v.trim() !== '' && !Number.isNaN(Number(v))) return Number(v);
  return null;
}

function attrBool(item: CatalogSortable, key: string): boolean {
  const v = item.attributes?.[key];
  return v === true || v === 'true' || v === 1;
}

function priceOf(item: CatalogSortable): number {
  if (typeof item.price_tickets === 'number') return item.price_tickets;
  if (typeof item.price === 'number') return item.price;
  return 0;
}

function elementLabel(item: CatalogSortable): string {
  const el = attrStr(item, 'element').toLowerCase();
  return ELEMENT_RU[el] || el;
}

function itemClasses(item: CatalogSortable): string[] {
  const raw = item.attributes?.class;
  if (Array.isArray(raw)) {
    return raw.map((c) => String(c).trim()).filter(Boolean);
  }
  const s = attrStr(item, 'class').trim();
  return s ? [s] : [];
}

function itemIcons(item: CatalogSortable): string[] {
  const raw = item.attributes?.icons;
  if (!Array.isArray(raw)) return [];
  return raw
    .map((entry) => {
      if (!entry || typeof entry !== 'object') return '';
      const t = (entry as { type?: string }).type;
      return t ? String(t).toLowerCase() : '';
    })
    .filter(Boolean);
}

function matchStat(actual: number | null, filter: StatFilter | null | undefined): boolean {
  if (!filter) return true;
  if (actual == null) return false;
  switch (filter.op) {
    case 'eq':
      return actual === filter.value;
    case 'gt':
      return actual > filter.value;
    case 'lt':
      return actual < filter.value;
    case 'gte':
      return actual >= filter.value;
    case 'lte':
      return actual <= filter.value;
    default:
      return true;
  }
}

export function titleCaseRu(s: string): string {
  if (!s) return s;
  return s
    .split(/([\s-]+)/)
    .map((part) => {
      if (/^[\s-]+$/.test(part)) return part;
      return part.charAt(0).toUpperCase() + part.slice(1).toLowerCase();
    })
    .join('');
}

export function parseCatalogSort(raw: string | null): CatalogSort {
  const v = (raw || 'newest') as CatalogSort;
  return CATALOG_SORT_OPTIONS.some((o) => o.value === v) ? v : 'newest';
}

export function parseCatalogPage(raw: string | null): number {
  const n = Number(raw || '1');
  if (!Number.isFinite(n) || n < 1) return 1;
  return Math.floor(n);
}

export function parseFlag(raw: string | null): boolean {
  return raw === '1' || raw === 'true';
}

export function parseCommaList(raw: string | null): string[] {
  if (!raw) return [];
  return raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
}

export function parseCommaNumbers(raw: string | null): number[] {
  return parseCommaList(raw)
    .map((s) => Number(s))
    .filter((n) => Number.isFinite(n));
}

export function parseStatOp(raw: string | null): StatOp {
  const v = (raw || 'eq') as StatOp;
  return STAT_OP_OPTIONS.some((o) => o.value === v) ? v : 'eq';
}

export function parseStatFilter(
  opRaw: string | null,
  valueRaw: string | null
): StatFilter | null {
  if (valueRaw == null || valueRaw === '') return null;
  const value = Number(valueRaw);
  if (!Number.isFinite(value)) return null;
  return { op: parseStatOp(opRaw), value };
}

export function toggleListValue(list: string[], value: string): string[] {
  const set = new Set(list);
  if (set.has(value)) set.delete(value);
  else set.add(value);
  return [...set];
}

export function listToParam(list: string[] | number[]): string | null {
  if (!list.length) return null;
  return list.join(',');
}

/** Parse all advanced filter fields from URLSearchParams. */
export function parseCatalogFilters(params: URLSearchParams): CatalogFilterFlags {
  return {
    query: params.get('q') || '',
    inStock: parseFlag(params.get('in_stock')),
    foil: parseFlag(params.get('foil')),
    noir: parseFlag(params.get('noir')),
    flying: parseFlag(params.get('flying')),
    companion: parseFlag(params.get('companion')),
    symbiont: parseFlag(params.get('symbiont')),
    parasite: parseFlag(params.get('parasite')),
    unique: parseFlag(params.get('unique')),
    elements: parseCommaList(params.get('element')),
    rarities: parseCommaList(params.get('rarity')),
    costTiers: parseCommaList(params.get('cost_tier')),
    sets: parseCommaNumbers(params.get('set')),
    years: parseCommaNumbers(params.get('year')),
    classes: parseCommaList(params.get('class')),
    artistId: params.get('artist') || '',
    typeMains: parseCommaList(params.get('type_main')),
    icons: parseCommaList(params.get('icons')),
    attackDice: (params.get('attack') || '').trim(),
    cost: parseStatFilter(params.get('cost_op'), params.get('cost')),
    hp: parseStatFilter(params.get('hp_op'), params.get('hp')),
    move: parseStatFilter(params.get('move_op'), params.get('move')),
  };
}

/** Count active filter dimensions (for «Фильтры (N)»). */
export function countActiveFilters(f: CatalogFilterFlags): number {
  let n = 0;
  if ((f.query || '').trim()) n += 1;
  if (f.inStock) n += 1;
  if (f.foil) n += 1;
  if (f.noir) n += 1;
  if (f.flying) n += 1;
  if (f.companion) n += 1;
  if (f.symbiont) n += 1;
  if (f.parasite) n += 1;
  if (f.unique) n += 1;
  if (f.elements?.length) n += 1;
  if (f.rarities?.length) n += 1;
  if (f.costTiers?.length) n += 1;
  if (f.sets?.length) n += 1;
  if (f.years?.length) n += 1;
  if (f.classes?.length) n += 1;
  if (f.artistId) n += 1;
  if (f.typeMains?.length) n += 1;
  if (f.icons?.length) n += 1;
  if (f.attackDice) n += 1;
  if (f.cost) n += 1;
  if (f.hp) n += 1;
  if (f.move) n += 1;
  return n;
}

export function collectCatalogFacets(items: CatalogSortable[]): CatalogFacets {
  const classMap = new Map<string, string>();
  const years = new Set<number>();
  const sets = new Set<number>();
  const artistMap = new Map<string, string>();

  for (const it of items) {
    for (const c of itemClasses(it)) {
      const key = c.toLowerCase();
      if (!classMap.has(key)) classMap.set(key, titleCaseRu(c));
    }
    const y = attrNumOrNull(it, 'year');
    if (y != null) years.add(y);
    const s = attrNumOrNull(it, 'set_number');
    if (s != null) sets.add(s);
    const aid = attrStr(it, 'artist_id');
    if (aid) {
      const name =
        attrStr(it, 'artist_display') || attrStr(it, 'artist') || aid;
      if (!artistMap.has(aid)) artistMap.set(aid, name);
    }
  }

  return {
    classes: [...classMap.values()].sort((a, b) =>
      a.localeCompare(b, 'ru', { sensitivity: 'base' })
    ),
    years: [...years].sort((a, b) => a - b),
    sets: [...sets].sort((a, b) => a - b),
    artists: [...artistMap.entries()]
      .map(([id, name]) => ({ id, name }))
      .sort((a, b) => a.name.localeCompare(b.name, 'ru', { sensitivity: 'base' })),
  };
}

export function filterCatalogItems<T extends CatalogSortable>(
  items: T[],
  flags: CatalogFilterFlags
): T[] {
  let out = items;
  const q = (flags.query || '').trim().toLowerCase();
  if (q) {
    out = out.filter((it) => it.name.toLowerCase().includes(q));
  }
  if (flags.inStock) {
    out = out.filter((it) => (it.stock ?? 0) > 0);
  }
  if (flags.foil) out = out.filter((it) => attrBool(it, 'foil'));
  if (flags.noir) out = out.filter((it) => attrBool(it, 'noir'));
  if (flags.flying) out = out.filter((it) => attrBool(it, 'flying'));
  if (flags.companion) out = out.filter((it) => attrBool(it, 'companion'));
  if (flags.symbiont) out = out.filter((it) => attrBool(it, 'symbiont'));
  if (flags.parasite) out = out.filter((it) => attrBool(it, 'parasite'));
  if (flags.unique) out = out.filter((it) => attrBool(it, 'unique'));

  if (flags.elements?.length) {
    const want = new Set(flags.elements.map((e) => e.toLowerCase()));
    out = out.filter((it) => want.has(attrStr(it, 'element').toLowerCase()));
  }
  if (flags.rarities?.length) {
    const want = new Set(flags.rarities.map((r) => r.toLowerCase()));
    out = out.filter((it) => want.has(attrStr(it, 'rarity').toLowerCase()));
  }
  if (flags.costTiers?.length) {
    const want = new Set(flags.costTiers.map((c) => c.toLowerCase()));
    out = out.filter((it) => want.has(attrStr(it, 'cost_tier').toLowerCase()));
  }
  if (flags.sets?.length) {
    const want = new Set(flags.sets);
    out = out.filter((it) => {
      const s = attrNumOrNull(it, 'set_number');
      return s != null && want.has(s);
    });
  }
  if (flags.years?.length) {
    const want = new Set(flags.years);
    out = out.filter((it) => {
      const y = attrNumOrNull(it, 'year');
      return y != null && want.has(y);
    });
  }
  if (flags.classes?.length) {
    const want = new Set(flags.classes.map((c) => c.toLowerCase()));
    out = out.filter((it) =>
      itemClasses(it).some((c) => want.has(c.toLowerCase()))
    );
  }
  if (flags.artistId) {
    const id = flags.artistId;
    out = out.filter((it) => attrStr(it, 'artist_id') === id);
  }
  if (flags.typeMains?.length) {
    const want = new Set(flags.typeMains.map((t) => t.toLowerCase()));
    out = out.filter((it) => want.has(attrStr(it, 'type_main').toLowerCase()));
  }
  if (flags.icons?.length) {
    const want = flags.icons.map((i) => i.toLowerCase());
    out = out.filter((it) => {
      const have = itemIcons(it);
      return want.some((w) => have.includes(w));
    });
  }
  if (flags.attackDice) {
    const needle = flags.attackDice.trim().toLowerCase();
    out = out.filter(
      (it) => attrStr(it, 'attack_dice').trim().toLowerCase() === needle
    );
  }
  if (flags.cost) {
    out = out.filter((it) => matchStat(attrNumOrNull(it, 'cost'), flags.cost));
  }
  if (flags.hp) {
    out = out.filter((it) => matchStat(attrNumOrNull(it, 'hp'), flags.hp));
  }
  if (flags.move) {
    out = out.filter((it) => matchStat(attrNumOrNull(it, 'move'), flags.move));
  }

  return out;
}

export function sortCatalogItems<T extends CatalogSortable>(
  items: T[],
  sort: CatalogSort
): T[] {
  const copy = [...items];
  const byName = (a: T, b: T) =>
    a.name.localeCompare(b.name, 'ru', { sensitivity: 'base' });
  const byArtist = (a: T, b: T) => {
    const aa =
      attrStr(a, 'artist_display') ||
      attrStr(a, 'artist') ||
      attrStr(a, 'artist_id');
    const bb =
      attrStr(b, 'artist_display') ||
      attrStr(b, 'artist') ||
      attrStr(b, 'artist_id');
    return aa.localeCompare(bb, 'ru', { sensitivity: 'base' }) || byName(a, b);
  };
  const byRarity = (a: T, b: T) => {
    const ra = RARITY_RANK[attrStr(a, 'rarity')] ?? 99;
    const rb = RARITY_RANK[attrStr(b, 'rarity')] ?? 99;
    return ra - rb || byName(a, b);
  };
  const bySet = (a: T, b: T) => {
    const sa = attrNum(a, 'set_number');
    const sb = attrNum(b, 'set_number');
    if (sa !== sb) return sa - sb;
    return attrNum(a, 'card_no') - attrNum(b, 'card_no') || byName(a, b);
  };
  const byElement = (a: T, b: T) =>
    elementLabel(a).localeCompare(elementLabel(b), 'ru', { sensitivity: 'base' }) ||
    byName(a, b);
  const byCreated = (a: T, b: T) => {
    const ta = a.created_at ? Date.parse(a.created_at) : 0;
    const tb = b.created_at ? Date.parse(b.created_at) : 0;
    return tb - ta || byName(a, b);
  };

  switch (sort) {
    case 'name_asc':
      copy.sort(byName);
      break;
    case 'name_desc':
      copy.sort((a, b) => -byName(a, b));
      break;
    case 'rarity_asc':
      copy.sort(byRarity);
      break;
    case 'rarity_desc':
      copy.sort((a, b) => -byRarity(a, b));
      break;
    case 'price_asc':
      copy.sort((a, b) => priceOf(a) - priceOf(b) || byName(a, b));
      break;
    case 'price_desc':
      copy.sort((a, b) => priceOf(b) - priceOf(a) || byName(a, b));
      break;
    case 'artist_asc':
      copy.sort(byArtist);
      break;
    case 'element_asc':
      copy.sort(byElement);
      break;
    case 'set_asc':
      copy.sort(bySet);
      break;
    case 'set_desc':
      copy.sort((a, b) => -bySet(a, b));
      break;
    case 'newest':
    default:
      copy.sort(byCreated);
      break;
  }
  return copy;
}

export function paginateItems<T>(
  items: T[],
  page: number,
  pageSize = CATALOG_PAGE_SIZE
): {
  page: number;
  pageCount: number;
  slice: T[];
} {
  const pageCount = Math.max(1, Math.ceil(items.length / pageSize));
  const safePage = Math.min(Math.max(1, page), pageCount);
  const start = (safePage - 1) * pageSize;
  return {
    page: safePage,
    pageCount,
    slice: items.slice(start, start + pageSize),
  };
}
