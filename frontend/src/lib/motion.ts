function devSearchParam(name: string): string | null {
  if (!import.meta.env.DEV) return null;
  try {
    return new URLSearchParams(window.location.search).get(name);
  } catch {
    return null;
  }
}

/** Dev-only: `?motion=force` ignores prefers-reduced-motion for visual QA. */
export function isMotionForced(): boolean {
  return devSearchParam('motion') === 'force';
}

/** Dev-only: `?debug-void=1` outlines each VOID layer. */
export function isVoidDebug(): boolean {
  return devSearchParam('debug-void') === '1';
}

/** True when OS/browser wants less motion AND no DEV force override. */
export function prefersReducedMotion(): boolean {
  if (isMotionForced()) return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

const VOID_DEBUG_ANC = 'data-void-debug-anc';
const VOID_DEBUG_SIB = 'data-void-debug-sib';
const VOID_DEBUG_BG = 'data-void-debug-bg';

function clearVoidDebugMarks(): void {
  document.querySelectorAll(`[${VOID_DEBUG_ANC}], [${VOID_DEBUG_SIB}], [${VOID_DEBUG_BG}]`).forEach((el) => {
    el.removeAttribute(VOID_DEBUG_ANC);
    el.removeAttribute(VOID_DEBUG_SIB);
    el.removeAttribute(VOID_DEBUG_BG);
  });
}

/** Tag ancestors / siblings / nearby painted nodes for ?debug-void=1 outlines. */
export function syncVoidDebugMarks(): void {
  clearVoidDebugMarks();
  if (!isVoidDebug()) return;

  const disk = document.querySelector('.eh-disk');
  if (!disk) return;

  let anc = 0;
  let el: Element | null = disk.parentElement;
  while (el && el !== document.documentElement) {
    el.setAttribute(VOID_DEBUG_ANC, String(anc++));
    el = el.parentElement;
  }

  const markSiblings = (node: Element | null) => {
    if (!node?.parentElement) return;
    [...node.parentElement.children].forEach((sib, i) => {
      if (sib === node) return;
      sib.setAttribute(VOID_DEBUG_SIB, String(i));
    });
  };
  markSiblings(disk);
  markSiblings(document.querySelector('.eh-disk-hit'));

  // Any painted background within ~3 levels of the disk tree.
  const scope = disk.parentElement?.parentElement?.parentElement;
  if (!scope) return;
  let bgI = 0;
  scope.querySelectorAll('*').forEach((node) => {
    const cs = getComputedStyle(node);
    const painted =
      (cs.backgroundColor !== 'rgba(0, 0, 0, 0)' && cs.backgroundColor !== 'transparent') ||
      cs.backgroundImage !== 'none' ||
      (cs.backdropFilter !== 'none' && cs.backdropFilter !== '') ||
      (cs.filter !== 'none' && cs.filter !== '');
    if (!painted) return;
    if (node.classList.contains('eh-disk') || node.closest('.eh-disk') === disk) {
      // skip internal VOID layers — they already have named outlines
      if (
        node.classList.contains('eh-disk-glow') ||
        node.classList.contains('eh-disk-pull') ||
        node.classList.contains('eh-disk-rings') ||
        node.classList.contains('eh-disk-core') ||
        node.classList.contains('eh-disk-particles') ||
        node.classList.contains('eh-disk')
      ) {
        return;
      }
    }
    node.setAttribute(VOID_DEBUG_BG, String(bgI++));
  });
}

/** Apply/remove DEV html classes for motion force + void layer outlines. */
export function syncMotionForceClass(): boolean {
  const forced = isMotionForced();
  const voidDebug = isVoidDebug();
  document.documentElement.classList.toggle('eh-motion-force', forced);
  document.documentElement.classList.toggle('eh-void-debug', voidDebug);
  syncVoidDebugMarks();
  return forced;
}
