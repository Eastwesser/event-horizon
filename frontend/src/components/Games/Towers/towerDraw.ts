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

export function drawBackground(ctx: CanvasRenderingContext2D, w: number, h: number) {
  const g = ctx.createLinearGradient(0, 0, 0, h);
  g.addColorStop(0, '#12162a');
  g.addColorStop(0.55, '#1a1a2e');
  g.addColorStop(1, '#0e1220');
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, w, h);

  // Soft floor band
  const floor = ctx.createLinearGradient(0, h - 56, 0, h);
  floor.addColorStop(0, 'rgba(232, 184, 74, 0)');
  floor.addColorStop(1, 'rgba(232, 184, 74, 0.08)');
  ctx.fillStyle = floor;
  ctx.fillRect(0, h - 56, w, 56);
}

export function drawBlock(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  color: string,
) {
  if (w <= 0 || h <= 0) return;

  ctx.save();
  const body = ctx.createLinearGradient(x, y, x + w, y);
  body.addColorStop(0, shade(color, -28));
  body.addColorStop(0.35, color);
  body.addColorStop(0.65, shade(color, 18));
  body.addColorStop(1, shade(color, -22));
  ctx.fillStyle = body;
  ctx.fillRect(x, y, w, h);

  // Top bevel
  ctx.fillStyle = 'rgba(255,255,255,0.22)';
  ctx.fillRect(x, y, w, 3);
  // Bottom edge
  ctx.fillStyle = 'rgba(0,0,0,0.28)';
  ctx.fillRect(x, y + h - 3, w, 3);
  // Left highlight / right shadow
  ctx.fillStyle = 'rgba(255,255,255,0.12)';
  ctx.fillRect(x, y, 3, h);
  ctx.fillStyle = 'rgba(0,0,0,0.2)';
  ctx.fillRect(x + w - 3, y, 3, h);

  // Grain lines
  ctx.fillStyle = 'rgba(255,255,255,0.12)';
  const slots = Math.min(4, Math.max(1, Math.floor(w / 28)));
  for (let j = 0; j < slots; j++) {
    const lx = x + 6 + ((j + 0.5) * (w - 12)) / slots;
    ctx.fillRect(lx, y + 5, 2, h - 10);
  }

  ctx.strokeStyle = 'rgba(255,255,255,0.35)';
  ctx.lineWidth = 1;
  ctx.strokeRect(x + 0.5, y + 0.5, w - 1, h - 1);
  ctx.restore();
}

export function drawMovingBlock(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  color: string,
) {
  ctx.save();
  ctx.shadowBlur = 14;
  ctx.shadowColor = 'rgba(0,0,0,0.55)';
  drawBlock(ctx, x, y, w, h, color);
  ctx.shadowBlur = 0;
  // Outer glow ring
  ctx.strokeStyle = 'rgba(255,255,255,0.55)';
  ctx.lineWidth = 1.5;
  ctx.strokeRect(x - 1, y - 1, w + 2, h + 2);
  ctx.restore();
}

export function drawDirectionChevron(
  ctx: CanvasRenderingContext2D,
  direction: 1 | -1,
  gameWidth: number,
  blockY: number,
  blockH: number,
) {
  const cy = blockY + blockH / 2;
  const size = 10;
  ctx.save();
  ctx.fillStyle = 'rgba(255,255,255,0.55)';
  ctx.beginPath();
  if (direction === 1) {
    const tipX = gameWidth - 18;
    ctx.moveTo(tipX, cy);
    ctx.lineTo(tipX - size, cy - size);
    ctx.lineTo(tipX - size, cy + size);
  } else {
    const tipX = 18;
    ctx.moveTo(tipX, cy);
    ctx.lineTo(tipX + size, cy - size);
    ctx.lineTo(tipX + size, cy + size);
  }
  ctx.closePath();
  ctx.fill();
  ctx.restore();
}

export function stackStartY(gameHeight: number): number {
  return gameHeight - 50;
}
