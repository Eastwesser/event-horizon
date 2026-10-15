/** Canvas draw helpers for Flappy — cosmetic only; hitboxes stay in the store. */

export const FLAPPY_W = 800;
export const FLAPPY_H = 500;
const GROUND_H = 36;

export type PipeDraw = {
  x: number;
  topHeight: number;
  bottomY: number;
};

const SKY = '#24344f';
const CITY = '#121820';
const CITY_WIN = 'rgba(201, 162, 39, 0.35)';
const GROUND = '#3a3228';
const GROUND_TOP = '#c9a227';

function px(ctx: CanvasRenderingContext2D, x: number, y: number, s: number, color: string) {
  ctx.fillStyle = color;
  ctx.fillRect(Math.round(x) * s, Math.round(y) * s, s, s);
}

export function drawSky(ctx: CanvasRenderingContext2D, w: number, h: number) {
  ctx.fillStyle = SKY;
  ctx.fillRect(0, 0, w, h);
}

export function drawStars(
  ctx: CanvasRenderingContext2D,
  w: number,
  _h: number,
  scrollX: number,
) {
  const stars = [
    [40, 36, 1],
    [120, 70, 1],
    [210, 28, 2],
    [310, 90, 1],
    [400, 44, 1],
    [490, 78, 1],
    [580, 22, 2],
    [670, 64, 1],
    [750, 48, 1],
    [90, 110, 1],
    [350, 16, 1],
    [620, 100, 1],
  ] as const;

  ctx.fillStyle = 'rgba(246, 234, 208, 0.7)';
  for (const [bx, y, r] of stars) {
    const x = ((bx - scrollX) % (w + 40) + (w + 40)) % (w + 40) - 20;
    ctx.fillRect(x, y, r, r);
  }
}

/** Stepped pixel clouds, not ellipses. */
export function drawClouds(ctx: CanvasRenderingContext2D, w: number, scrollX: number) {
  const band = w + 220;
  const bases = [
    { x: 140, y: 64 },
    { x: 400, y: 40 },
    { x: 660, y: 88 },
    { x: 900, y: 52 },
  ];
  ctx.fillStyle = '#6a7a94';
  for (const c of bases) {
    const x = ((c.x - scrollX) % band + band) % band - 80;
    const y = c.y;
    ctx.fillRect(x + 16, y, 48, 8);
    ctx.fillRect(x + 8, y + 8, 72, 8);
    ctx.fillRect(x, y + 16, 88, 8);
  }
}

export function drawCity(ctx: CanvasRenderingContext2D, w: number, h: number, scrollX: number) {
  const band = w + 160;
  const groundY = h - GROUND_H;
  const heights = [48, 72, 40, 88, 56, 64, 36, 80, 52, 70];
  ctx.fillStyle = CITY;
  for (let i = 0; i < 14; i++) {
    const bw = 44;
    const bh = heights[i % heights.length];
    const x = ((i * 58 - scrollX) % band + band) % band - 50;
    ctx.fillRect(x, groundY - bh, bw, bh);
    ctx.fillStyle = CITY_WIN;
    for (let wy = groundY - bh + 8; wy < groundY - 8; wy += 12) {
      ctx.fillRect(x + 8, wy, 4, 4);
      ctx.fillRect(x + 20, wy, 4, 4);
      ctx.fillRect(x + 32, wy, 4, 4);
    }
    ctx.fillStyle = CITY;
  }
}

export function drawGround(ctx: CanvasRenderingContext2D, w: number, h: number, scrollX: number) {
  const y = h - GROUND_H;
  ctx.fillStyle = GROUND;
  ctx.fillRect(0, y, w, GROUND_H);
  ctx.fillStyle = GROUND_TOP;
  ctx.fillRect(0, y, w, 3);
  ctx.fillStyle = '#2a241c';
  const step = 18;
  const off = Math.floor(scrollX) % step;
  for (let x = -off; x < w; x += step) {
    ctx.fillRect(x, y + 10, 10, 4);
    ctx.fillRect(x + 6, y + 22, 8, 4);
  }
}

function pipeFill(cosmic: boolean, dark: boolean) {
  if (cosmic) return dark ? '#152238' : '#3d5a80';
  return dark ? '#3d2a18' : '#8a6a3a';
}

export function drawPipe(
  ctx: CanvasRenderingContext2D,
  pipe: PipeDraw,
  pipeWidth: number,
  gameHeight: number,
  cosmic: boolean,
) {
  const { x, topHeight, bottomY } = pipe;
  const w = pipeWidth;
  const body = pipeFill(cosmic, false);
  const lip = pipeFill(cosmic, true);
  const edge = cosmic ? '#7dd3fc' : '#c9a227';

  const paint = (y: number, h: number, isTop: boolean) => {
    if (h <= 0) return;
    ctx.fillStyle = body;
    ctx.fillRect(x, y, w, h);
    ctx.fillStyle = 'rgba(0,0,0,0.22)';
    ctx.fillRect(x + w - 8, y, 8, h);
    ctx.fillStyle = edge;
    ctx.fillRect(x + 3, y, 3, h);

    const lipH = 22;
    const lipY = isTop ? y + h - lipH : y;
    ctx.fillStyle = lip;
    ctx.fillRect(x - 6, lipY, w + 12, lipH);
    ctx.fillStyle = edge;
    ctx.fillRect(x - 6, lipY, w + 12, 3);
    ctx.fillRect(x - 6, lipY + lipH - 3, w + 12, 3);
  };

  paint(0, topHeight, true);
  paint(bottomY, gameHeight - GROUND_H - bottomY, false);
}

type Wing = 'up' | 'mid' | 'down' | 'fall';

function wingFromVelocity(velocity: number): Wing {
  if (velocity > 5) return 'fall';
  if (velocity < -2.2) return 'up';
  if (velocity > 1.8) return 'down';
  return 'mid';
}

/** Pixel critter. Drawn around the existing 30px hitbox center — hitbox unchanged. */
export function drawBird(
  ctx: CanvasRenderingContext2D,
  birdY: number,
  birdSize: number,
  velocity: number,
  golden: boolean,
) {
  const s = 3;
  const cx = 100;
  const cy = birdY + birdSize / 2;
  const wing = wingFromVelocity(velocity);
  const body = golden ? '#e8c547' : '#6ea0d4';
  const bodyD = golden ? '#b8860b' : '#3d6a9a';
  const wingC = golden ? '#fff3a0' : '#c9a227';
  const beak = '#e07a4a';

  ctx.save();
  ctx.translate(Math.round(cx - 8 * s), Math.round(cy - 6 * s));

  const plot = (x: number, y: number, c: string) => px(ctx, x, y, s, c);

  // Body 12×10-ish
  for (let x = 2; x <= 9; x++) plot(x, 3, body);
  for (let x = 1; x <= 10; x++) plot(x, 4, body);
  for (let x = 1; x <= 10; x++) plot(x, 5, body);
  for (let x = 2; x <= 9; x++) plot(x, 6, bodyD);
  for (let x = 3; x <= 8; x++) plot(x, 7, bodyD);

  if (wing === 'up') {
    plot(3, 1, wingC);
    plot(4, 1, wingC);
    plot(2, 2, wingC);
    plot(3, 2, wingC);
    plot(4, 2, wingC);
  } else if (wing === 'down' || wing === 'fall') {
    plot(2, 7, wingC);
    plot(3, 8, wingC);
    plot(4, 8, wingC);
    plot(3, 9, wingC);
  } else {
    plot(1, 4, wingC);
    plot(2, 4, wingC);
    plot(1, 5, wingC);
  }

  plot(9, 3, '#f6ead0');
  plot(10, 3, '#f6ead0');
  plot(10, 4, '#1a140c');
  plot(11, 4, beak);
  plot(12, 4, beak);
  plot(11, 5, beak);

  if (wing === 'fall') {
    plot(8, 2, '#1a140c');
    plot(9, 2, '#1a140c');
  }

  ctx.restore();
}

export function drawScore(ctx: CanvasRenderingContext2D, score: number, w: number) {
  ctx.save();
  ctx.font = 'bold 28px "Press Start 2P", monospace';
  ctx.fillStyle = '#f6ead0';
  ctx.strokeStyle = 'rgba(0,0,0,0.45)';
  ctx.lineWidth = 4;
  const text = String(score);
  const tw = ctx.measureText(text).width;
  ctx.strokeText(text, w / 2 - tw / 2, 42);
  ctx.fillText(text, w / 2 - tw / 2, 42);
  ctx.restore();
}

export function drawStartHint(ctx: CanvasRenderingContext2D, _w: number) {
  ctx.save();
  ctx.textAlign = 'center';
  ctx.font = 'bold 12px "Press Start 2P", monospace';
  ctx.fillStyle = '#f6ead0';
  ctx.shadowColor = '#000';
  ctx.shadowBlur = 4;
  ctx.fillText('Нажмите пробел', 100, 168);  // было 'Пробел'
  ctx.restore();
}

export function previewPipe(gap: number): PipeDraw {
  const topHeight = 150;
  return { x: 520, topHeight, bottomY: topHeight + gap };
}

