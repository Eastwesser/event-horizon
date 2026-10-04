/** Dev-only: `?motion=force` ignores prefers-reduced-motion for visual QA. */
export function isMotionForced(): boolean {
  if (!import.meta.env.DEV) return false;
  try {
    return new URLSearchParams(window.location.search).get('motion') === 'force';
  } catch {
    return false;
  }
}

/** True when OS/browser wants less motion AND no DEV force override. */
export function prefersReducedMotion(): boolean {
  if (isMotionForced()) return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/** Apply/remove `eh-motion-force` on <html> so CSS can bypass reduce rules. */
export function syncMotionForceClass(): boolean {
  const forced = isMotionForced();
  document.documentElement.classList.toggle('eh-motion-force', forced);
  return forced;
}
