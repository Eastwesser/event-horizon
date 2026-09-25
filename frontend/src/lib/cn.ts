// Tiny className joiner — avoids pulling in clsx/tailwind-merge for one job.
export function cn(...classes: Array<string | false | null | undefined>): string {
  return classes.filter(Boolean).join(' ');
}
