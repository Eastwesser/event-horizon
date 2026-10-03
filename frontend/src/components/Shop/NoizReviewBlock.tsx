import { Badge } from '../ui/Badge';

export type NoizReview = {
  text?: string;
  rating?: number | null;
  verdict?: string | null;
  author?: string;
};

function asNoizReview(raw: unknown): NoizReview | null {
  if (!raw || typeof raw !== 'object') return null;
  const o = raw as Record<string, unknown>;
  const text = typeof o.text === 'string' ? o.text.trim() : '';
  if (!text) return null;
  const rating =
    typeof o.rating === 'number' && Number.isFinite(o.rating) ? o.rating : null;
  const verdict =
    typeof o.verdict === 'string' && o.verdict.trim() ? o.verdict.trim() : null;
  const author =
    typeof o.author === 'string' && o.author.trim() ? o.author.trim() : 'Noiz';
  return { text, rating, verdict, author };
}

const VERDICT_TONE: Record<string, 'gold' | 'success' | 'warning' | 'error' | 'neutral'> = {
  имба: 'gold',
  норм: 'success',
  филлер: 'warning',
  скип: 'error',
};

function verdictTone(verdict: string) {
  const key = verdict.toLowerCase().split(/[\s/(]/)[0] || '';
  return VERDICT_TONE[key] || 'neutral';
}

/** Prewritten Noiz take — only renders when attributes.noiz_review.text is set. */
export function NoizReviewBlock({ attrs }: { attrs: Record<string, unknown> }) {
  const review = asNoizReview(attrs?.noiz_review);
  if (!review) return null;

  return (
    <section
      className="rounded-sm border border-indigo/25 bg-indigo/5 px-4 py-3"
      aria-label="Мнение Noiz"
    >
      <div className="mb-2 flex flex-wrap items-center gap-2">
        <h3 className="font-display text-sm font-semibold text-text-primary">
          Мнение Noiz
        </h3>
        {review.rating != null ? (
          <Badge tone="cyan" className="font-hud tabular-nums">
            {review.rating}/10
          </Badge>
        ) : null}
        {review.verdict ? (
          <Badge tone={verdictTone(review.verdict)}>{review.verdict}</Badge>
        ) : null}
      </div>
      <blockquote className="text-sm leading-relaxed text-text-secondary">
        {review.text}
      </blockquote>
      <p className="mt-2 text-right text-xs text-text-muted">— {review.author}</p>
    </section>
  );
}

export default NoizReviewBlock;
