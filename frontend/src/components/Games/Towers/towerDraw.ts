/** Canvas draw helpers for Towers — cosmetic only; geometry stays in towerStore. */

export const BLOCK_HEIGHT = 25;

const COSMIC = [
  '#0B1020',
  '#1A3A5C',
  '#4F7CAC',
  '#6366F1',
  '#7DD3FC',
  '#C9A227',
  '#E8D5A3',
] as const;

const DEFAULT_REDS = [
  '#E74C3C',
  '#C0392B',
  '#A93226',
  '#922B21',
  '#7B241C',
  '#641E16',
  '#4A1A0A',
] as const;

export function blockColor(blockLevel: number, cosmic: boolean): string {
  if (cosmic) {
    return COSMIC[(blockLevel - 1) % COSMIC.length];
  }
  const index = Math.min(Math.floor((blockLevel - 1) / 2), DEFAULT_REDS.length - 1);
  return DEFAULT_REDS[index];
}

function hexToRgb(hex: string): { r: number; g: number; b: number } {
  const h = hex.replace('#', '');
  const n = parseInt(h.length === 3 ? h.split('').map((c) => c + c).join('') : h, 16);
  return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 };
}

function shade(hex: string, amount: number): string {
  const { r, g, b } = hexToRgb(hex);
  const t = (c: number) => Math.max(0, Math.min(255, Math.round(c + amount)));
  return `rgb(${t(r)}, ${t(g)}, ${t(b)})`;
}

function cloud(ctx: CanvasRenderingContext2D, x: number, y: number, s: number) {
  ctx.beginPath();
  ctx.ellipse(x, y, 28 * s, 16 * s, 0, 0, Math.PI * 2);
  ctx.ellipse(x + 22 * s, y - 6 * s, 22 * s, 14 * s, 0, 0, Math.PI * 2);
  ctx.ellipse(x - 20 * s, y - 4 * s, 18 * s, 12 * s, 0, 0, Math.PI * 2);
  ctx.fill();
}

export function drawBackground(ctx: CanvasRenderingContext2D, w: number, h: number) {
  const g = ctx.createLinearGradient(0, 0, 0, h);
  g.addColorStop(0, '#6e9ed4');
  g.addColorStop(0.55, '#b9d7f2');
  g.addColorStop(1, '#efe2c4');
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, w, h);

  ctx.fillStyle = 'rgba(255,255,255,0.82)';
  cloud(ctx, 70, 54, 1);
  cloud(ctx, 250, 86, 0.85);
  cloud(ctx, 340, 40, 0.7);

  ctx.fillStyle = '#e7c98a';
  ctx.fillRect(0, h - 36, w, 36);
  ctx.fillStyle = '#d7b56e';
  ctx.fillRect(0, h - 36, w, 4);
}

export type IsoSlab = { x: number; z: number; w: number; d: number };

const KX = 0.78;
const KZ = 0.46;
const THICK = 13;
const BASE_Y = 418;

function project(x: number, z: number, lift: number) {
  return {
    x: 200 + (x - z) * KX,
    y: BASE_Y + (x + z) * KZ - lift,
  };
}

function quad(
  ctx: CanvasRenderingContext2D,
  a: { x: number; y: number },
  b: { x: number; y: number },
  c: { x: number; y: number },
  d: { x: number; y: number },
) {
  ctx.beginPath();
  ctx.moveTo(a.x, a.y);
  ctx.lineTo(b.x, b.y);
  ctx.lineTo(c.x, c.y);
  ctx.lineTo(d.x, d.y);
  ctx.closePath();
  ctx.fill();
}

export function drawIsoSlab(
  ctx: CanvasRenderingContext2D,
  slab: IsoSlab,
  lift: number,
  color: string,
  dy = 0,
) {
  if (slab.w <= 0.5 || slab.d <= 0.5) return;
  const a = project(slab.x, slab.z, lift);
  const b = project(slab.x + slab.w, slab.z, lift);
  const c = project(slab.x + slab.w, slab.z + slab.d, lift);
  const d = project(slab.x, slab.z + slab.d, lift);
  a.y += dy;
  b.y += dy;
  c.y += dy;
  d.y += dy;
  const down = (p: { x: number; y: number }) => ({ x: p.x, y: p.y + THICK });

  ctx.fillStyle = shade(color, -46);
  quad(ctx, b, c, down(c), down(b));
  ctx.fillStyle = shade(color, -22);
  quad(ctx, d, c, down(c), down(d));
  ctx.fillStyle = shade(color, 36);
  quad(ctx, a, b, c, d);
  ctx.strokeStyle = 'rgba(42, 33, 24, 0.28)';
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(a.x, a.y);
  ctx.lineTo(b.x, b.y);
  ctx.lineTo(c.x, c.y);
  ctx.lineTo(d.x, d.y);
  ctx.closePath();
  ctx.stroke();
}

/** Screen y of the back corner, used to keep the top of the stack in frame. */
export function isoTopY(x: number, z: number, lift: number): number {
  return project(x, z, lift).y;
}
