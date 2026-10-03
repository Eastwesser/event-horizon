import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { Badge } from '../ui/Badge';
import { Icon, type IconName } from '../ui/Icon';

export type CardIcon = { type?: string; value?: number | null };

/** Berserk icon chips — SVG glyphs (Wave 1 Emoji → SVG). */
const ICON_META: Record<string, { label: string; icon: IconName }> = {
  armor: { label: 'Доспех', icon: 'armor' },
  counter: { label: 'Фишка', icon: 'counter' },
  uchr: { label: 'Удар через ряд', icon: 'uchr' },
  tap: { label: 'Поворот', icon: 'tap' },
  strike: { label: 'Атака', icon: 'strike' },
  instant: { label: 'Мгновенно', icon: 'instant' },
  zov: { label: 'Защита от выстрелов', icon: 'zov' },
  zoz: { label: 'Защита от заклинаний', icon: 'zoz' },
  zot: { label: 'Защита от метаний', icon: 'zot' },
  zor: { label: 'Защита от разрядов', icon: 'zor' },
  zoal: { label: 'Защита от летающих', icon: 'zoal' },
  zom: { label: 'Защита от магии', icon: 'zom' },
  zoo: { label: 'Защита от отравления', icon: 'zoo' },
  regen: { label: 'Регенерация', icon: 'regen' },
  stamina: { label: 'Стойкость', icon: 'stamina' },
  direct: { label: 'Направленный удар', icon: 'direct' },
  ova: { label: 'Опыт в атаке', icon: 'ova' },
  ovz: { label: 'Опыт в защите', icon: 'ovz' },
  ovs: { label: 'Опыт в стрельбе', icon: 'ovs' },
};

const ELEMENT_RU: Record<string, string> = {
  mountains: 'Горы',
  woods: 'Лес',
  steppes: 'Степи',
  swamps: 'Болота',
  darkness: 'Тьма',
  neutral: 'Нейтралы',
};

const RARITY_RU: Record<string, string> = {
  common: 'Обычная',
  uncommon: 'Необычная',
  rare: 'Редкая',
  ultra: 'Ультра',
};

const COST_TIER_RU: Record<string, string> = {
  rank_and_file: 'Рядовая',
  elite: 'Элитная',
};

const TYPE_MAIN_RU: Record<string, string> = {
  creature: 'Существо',
  artifact: 'Артефакт',
  land: 'Местность',
};

function asBool(v: unknown): boolean {
  return v === true || v === 'true' || v === 1;
}

function asString(v: unknown): string {
  if (v === null || v === undefined) return '';
  return String(v);
}

function asNumber(v: unknown): number | null {
  if (typeof v === 'number' && Number.isFinite(v)) return v;
  if (typeof v === 'string' && v.trim() !== '' && !Number.isNaN(Number(v))) return Number(v);
  return null;
}

export function parseIcons(raw: unknown): CardIcon[] {
  if (!Array.isArray(raw)) return [];
  return raw
    .map((entry) => {
      if (!entry || typeof entry !== 'object') return null;
      const o = entry as { type?: string; value?: number | null };
      if (!o.type) return null;
      return { type: String(o.type).toLowerCase(), value: o.value ?? null };
    })
    .filter(Boolean) as CardIcon[];
}

export function CardIconChips({ icons }: { icons: CardIcon[] }) {
  if (!icons.length) {
    return <span className="text-sm text-text-muted">—</span>;
  }
  return (
    <div className="flex flex-wrap gap-1.5">
      {icons.map((icon, i) => {
        const meta = ICON_META[icon.type || ''] || {
          label: icon.type || 'icon',
          icon: 'diamond' as IconName,
        };
        const text =
          icon.value != null ? `${meta.label} ${icon.value}` : meta.label;
        return (
          <span
            key={`${icon.type}-${i}`}
            title={text}
            className="inline-flex items-center gap-1 rounded-sm border border-white/15 bg-white/5 px-2 py-1 text-xs text-text-secondary"
          >
            <Icon name={meta.icon} className="h-3.5 w-3.5 text-horizon-gold" title={meta.label} />
            <span className="font-medium text-text-primary">
              {icon.type}
              {icon.value != null ? `:${icon.value}` : ''}
            </span>
          </span>
        );
      })}
    </div>
  );
}

export type FlagBadge = {
  key: string;
  label: string;
  tone: 'gold' | 'neutral' | 'cyan' | 'success' | 'warning' | 'indigo';
};

export function cardFlagBadges(attrs: Record<string, unknown>): FlagBadge[] {
  const flags: FlagBadge[] = [];
  if (asBool(attrs.foil)) flags.push({ key: 'foil', label: '✦ ФОЙЛ', tone: 'gold' });
  if (asBool(attrs.noir)) flags.push({ key: 'noir', label: '◐ НУАР', tone: 'neutral' });
  if (asBool(attrs.flying)) flags.push({ key: 'flying', label: 'ЛЕТАЮЩИЙ', tone: 'cyan' });
  if (asBool(attrs.companion)) flags.push({ key: 'companion', label: 'C Компаньон', tone: 'success' });
  if (asBool(attrs.unique)) flags.push({ key: 'unique', label: '♛ УНИКАЛЬНАЯ', tone: 'warning' });
  if (asBool(attrs.symbiont)) flags.push({ key: 'symbiont', label: 'Симбионт', tone: 'indigo' });
  if (asBool(attrs.parasite)) flags.push({ key: 'parasite', label: 'Паразит', tone: 'indigo' });
  return flags;
}

function Row({ label, children }: { label: string; children: ReactNode }) {
  if (children === null || children === undefined || children === '') return null;
  return (
    <div className="grid grid-cols-[6.5rem_1fr] gap-x-2 gap-y-0.5 text-sm sm:grid-cols-[7.5rem_1fr]">
      <dt className="text-text-muted">{label}</dt>
      <dd className="min-w-0 text-text-primary">{children}</dd>
    </div>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="space-y-1.5">
      <h3 className="font-display text-sm font-semibold uppercase tracking-wide text-text-secondary">
        {title}
      </h3>
      <dl className="space-y-1">{children}</dl>
    </section>
  );
}

export function CardFlagBadges({ attrs }: { attrs: Record<string, unknown> }) {
  const flags = cardFlagBadges(attrs);
  if (!flags.length) return null;
  return (
    <div className="flex flex-wrap gap-1.5">
      {flags.map((f) => (
        <Badge key={f.key} tone={f.tone}>
          {f.label}
        </Badge>
      ))}
    </div>
  );
}

/** Compact attr lines for purchase modal — only filled fields. */
export function CardPurchaseSummary({ attrs }: { attrs?: Record<string, unknown> | null }) {
  if (!attrs || Object.keys(attrs).length === 0) return null;

  const rows: { label: string; value: string }[] = [];
  const rarity = asString(attrs.rarity);
  if (rarity) rows.push({ label: 'Редкость', value: RARITY_RU[rarity] || rarity });

  const element = asString(attrs.element);
  if (element) rows.push({ label: 'Стихия', value: ELEMENT_RU[element] || element });

  if (asBool(attrs.foil)) rows.push({ label: 'Фойл', value: 'Да' });
  if (asBool(attrs.noir)) rows.push({ label: 'Нуар', value: 'Да' });
  if (asBool(attrs.flying)) rows.push({ label: 'Летающий', value: 'Да' });

  const hp = asNumber(attrs.hp);
  if (hp != null) rows.push({ label: 'HP', value: String(hp) });

  const attackDice = asString(attrs.attack_dice);
  const attackType = asString(attrs.attack_type);
  const attack = [attackType, attackDice].filter(Boolean).join(' ');
  if (attack) rows.push({ label: 'Удар', value: attack });

  const cost = asNumber(attrs.cost);
  const costTier = asString(attrs.cost_tier);
  if (cost != null) {
    rows.push({
      label: 'Стоимость',
      value: `${cost}${costTier ? ` · ${COST_TIER_RU[costTier] || costTier}` : ''}`,
    });
  }

  const classList = Array.isArray(attrs.class)
    ? (attrs.class as unknown[]).map(asString).filter(Boolean)
    : asString(attrs.class)
      ? [asString(attrs.class)]
      : [];
  if (classList.length) rows.push({ label: 'Класс', value: classList.join(', ') });

  if (!rows.length) return null;

  return (
    <dl className="mt-3 w-full space-y-1 rounded-sm border border-white/10 bg-nebula px-3 py-2 text-left text-sm">
      {rows.map((r) => (
        <div key={r.label} className="grid grid-cols-[6.5rem_1fr] gap-2">
          <dt className="text-text-muted">{r.label}</dt>
          <dd className="text-text-primary">{r.value}</dd>
        </div>
      ))}
    </dl>
  );
}

/** Structured card attributes for shop/inventory detail (replaces raw key:value dump). */
export function CardAttributesView({
  attrs,
  trailingAction,
}: {
  attrs: Record<string, unknown>;
  /** Renders bottom-right, aligned with «Художник» (e.g. Купить). */
  trailingAction?: ReactNode;
}) {
  if (!attrs || Object.keys(attrs).length === 0) {
    if (!trailingAction) return null;
    return <div className="flex justify-end">{trailingAction}</div>;
  }

  const icons = parseIcons(attrs.icons);
  const classList = Array.isArray(attrs.class)
    ? (attrs.class as unknown[]).map(asString).filter(Boolean)
    : asString(attrs.class)
      ? [asString(attrs.class)]
      : [];

  const hp = asNumber(attrs.hp);
  const move = asNumber(attrs.move);
  const cost = asNumber(attrs.cost);
  const cardNo = asNumber(attrs.card_no);
  const setNumber = asNumber(attrs.set_number);
  const year = asNumber(attrs.year);

  const element = asString(attrs.element);
  const rarity = asString(attrs.rarity);
  const costTier = asString(attrs.cost_tier);
  const typeMain = asString(attrs.type_main);
  const typeSub = asString(attrs.type_sub);
  const attackDice = asString(attrs.attack_dice);
  const attackType = asString(attrs.attack_type);
  const setName = asString(attrs.set_name);
  const artist =
    asString(attrs.artist_display) ||
    asString(attrs.artist) ||
    asString(attrs.artist_id);

  const artistId = asString(attrs.artist_id);

  return (
    <div className="space-y-4">
      <Section title="Бой">
        <Row label="Здоровье">{hp != null ? String(hp) : '—'}</Row>
        <Row label="Ход">{move != null ? String(move) : '—'}</Row>
        <Row label="Удар">
          {[attackType, attackDice].filter(Boolean).join(' ') || '—'}
        </Row>
        <Row label="Стоимость">
          {cost != null
            ? `${cost}${costTier ? ` · ${COST_TIER_RU[costTier] || costTier}` : ''}`
            : '—'}
        </Row>
        <Row label="Иконки">
          <CardIconChips icons={icons} />
        </Row>
      </Section>

      <Section title="Карта">
        <Row label="Выпуск">
          {[setName, setNumber != null ? `#${setNumber}` : '', year != null ? String(year) : '']
            .filter(Boolean)
            .join(' · ') || '—'}
        </Row>
        <Row label="Номер">{cardNo != null ? String(cardNo) : '—'}</Row>
        <Row label="Редкость">{RARITY_RU[rarity] || rarity || '—'}</Row>
        <Row label="Стихия">{ELEMENT_RU[element] || element || '—'}</Row>
        <Row label="Тип">
          {[TYPE_MAIN_RU[typeMain] || typeMain, typeSub].filter(Boolean).join(' – ') || '—'}
        </Row>
        <Row label="Класс">{classList.length ? classList.join(', ') : '—'}</Row>
      </Section>

      <div className="flex items-end justify-between gap-4">
        <div className="min-w-0 flex-1">
          <Section title="Автор">
            <Row label="Художник">
              {artistId ? (
                <Link
                  to={`/authors/${encodeURIComponent(artistId)}`}
                  className="text-horizon-cyan underline-offset-2 hover:underline"
                >
                  {artist || artistId}
                </Link>
              ) : (
                artist || '—'
              )}
            </Row>
          </Section>
        </div>
        {trailingAction ? (
          <div className="shrink-0 self-end pb-0.5">{trailingAction}</div>
        ) : null}
      </div>
    </div>
  );
}
