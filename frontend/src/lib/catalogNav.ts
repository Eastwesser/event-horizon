/** Persist ordered id list so detail pages can prev/next within context. */

const KEY = 'eh_catalog_nav_v1';

export interface CatalogNavContext {
  ids: string[];
  /** Where «back» / list lives, e.g. /shop?type=карточка */
  listPath: string;
}

export function saveCatalogNav(ctx: CatalogNavContext): void {
  try {
    sessionStorage.setItem(KEY, JSON.stringify(ctx));
  } catch {
    /* ignore quota */
  }
}

export function loadCatalogNav(): CatalogNavContext | null {
  try {
    const raw = sessionStorage.getItem(KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as CatalogNavContext;
    if (!Array.isArray(parsed.ids) || typeof parsed.listPath !== 'string') return null;
    return parsed;
  } catch {
    return null;
  }
}

export function navNeighbors(
  id: string,
  ctx: CatalogNavContext | null
): { prevId: string | null; nextId: string | null; index: number; total: number } {
  if (!ctx?.ids?.length) {
    return { prevId: null, nextId: null, index: -1, total: 0 };
  }
  const index = ctx.ids.indexOf(id);
  if (index < 0) {
    return { prevId: null, nextId: null, index: -1, total: ctx.ids.length };
  }
  return {
    prevId: index > 0 ? ctx.ids[index - 1] : null,
    nextId: index < ctx.ids.length - 1 ? ctx.ids[index + 1] : null,
    index,
    total: ctx.ids.length,
  };
}
