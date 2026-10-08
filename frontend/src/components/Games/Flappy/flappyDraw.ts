/** Canvas draw helpers for Flappy — cosmetic only; hitboxes stay in the store. */

export const FLAPPY_W = 800;
export const FLAPPY_H = 500;

export type PipeDraw = {
  x: number;
  topHeight: number;
  bottomY: number;
};

export function drawSky(ctx: CanvasRenderingContext2D, w: number, h: number) {
  const gradient = ctx.createLinearGradient(0, 0, 0, h);
  gradient.addColorStop(0, '#5BA3D9');
  gradient.addColorStop(0.55, '#87CEEB');
  gradient.addColorStop(1, '#E8F6FF');
  ctx.fillStyle = gradient;
  ctx.fillRect(0, 0, w, h);
}

/** Far parallax: sparse stars drifting slowly. */
export function drawStars(
  ctx: CanvasRenderingContext2D,
  w: number,
  _h: number,
  scrollX: number,
) {
  const stars = [
    [40, 36, 1.2],
    [120, 70, 0.9],
    [210, 28, 1.4],
    [310, 90, 0.8],
    [400, 44, 1.1],
    [490, 78, 1.0],
    [580, 22, 1.3],
    [670, 64, 0.9],
    [750, 48, 1.2],
    [90, 110, 0.7],
    [350, 16, 1.0],
    [620, 100, 0.8],
  ] as const;

  ctx.save();
  for (const [bx, y, r] of stars) {
    const x = ((bx - scrollX) % (w + 40) + (w + 40)) % (w + 40) - 20;
    ctx.fillStyle = 'rgba(255, 255, 255, 0.55)';
    ctx.beginPath();
    ctx.arc(x, y, r, 0, Math.PI * 2);
    ctx.fill();
  }
  ctx.restore();
}

export function drawCloud(
  ctx: CanvasRenderingContext2D,
  cx: number,
  cy: number,
  scale = 1,
) {
  ctx.save();
  ctx.fillStyle = 'rgba(255, 255, 255, 0.78)';
  ctx.beginPath();
  ctx.ellipse(cx, cy, 40 * scale, 28 * scale, 0, 0, Math.PI * 2);
  ctx.ellipse(cx + 32 * scale, cy - 8 * scale, 48 * scale, 32 * scale, 0, 0, Math.PI * 2);
  ctx.ellipse(cx - 28 * scale, cy - 6 * scale, 34 * scale, 24 * scale, 0, 0, Math.PI * 2);
  ctx.ellipse(cx + 10 * scale, cy + 10 * scale, 36 * scale, 22 * scale, 0, 0, Math.PI * 2);
  ctx.fill();
  ctx.restore();
}

/** Mid parallax: repeating cloud band at 0.3× pipe scroll. */
export function drawClouds(
  ctx: CanvasRenderingContext2D,
  w: number,
  scrollX: number,
) {
  const band = w + 200;
  const bases = [
    { x: 150, y: 80, s: 1 },
    { x: 420, y: 55, s: 0.85 },
    { x: 680, y: 110, s: 1.1 },
    { x: 920, y: 70, s: 0.9 },
  ];
  for (const c of bases) {
    const x = ((c.x - scrollX) % band + band) % band - 100;
    drawCloud(ctx, x, c.y, c.s);
  }
}

function pipeBodyGradient(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  cosmic: boolean,
) {
  if (cosmic) {
    // Void → cyan → indigo → gold (no rainbow spectrum).
    const g = ctx.createLinearGradient(x, y, x, y + h);
    g.addColorStop(0, '#0B1020');
    g.addColorStop(0.35, '#1A3A5C');
    g.addColorStop(0.6, '#4F7CAC');
    g.addColorStop(0.82, '#6366F1');
    g.addColorStop(1, '#C9A227');
    return g;
  }
  const g = ctx.createLinearGradient(x, y, x + w, y);
  g.addColorStop(0, '#1B5E20');
  g.addColorStop(0.35, '#2E7D32');
  g.addColorStop(0.55, '#43A047');
  g.addColorStop(1, '#1B5E20');
  return g;
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

  const paintSegment = (y: number, h: number, isTop: boolean) => {
    if (h <= 0) return;
    ctx.fillStyle = pipeBodyGradient(ctx, x, y, w, h, cosmic);
    ctx.fillRect(x, y, w, h);

    // Rim / lip
    const lipH = 30;
    const lipY = isTop ? y + h - lipH : y;
    ctx.fillStyle = pipeBodyGradient(ctx, x - 5, lipY, w + 10, lipH, cosmic);
    ctx.fillRect(x - 5, lipY, w + 10, lipH);

    // Highlight edge
    ctx.fillStyle = cosmic ? 'rgba(125, 211, 252, 0.35)' : 'rgba(165, 214, 167, 0.45)';
    ctx.fillRect(x + 4, y, 6, h);

    // Shadow edge
    ctx.fillStyle = 'rgba(0,0,0,0.18)';
    ctx.fillRect(x + w - 8, y, 6, h);

    ctx.strokeStyle = cosmic ? 'rgba(201, 162, 39, 0.45)' : 'rgba(0,0,0,0.25)';
    ctx.lineWidth = 1.5;
    ctx.strokeRect(x, y, w, h);
    ctx.strokeRect(x - 5, lipY, w + 10, lipH);

    if (!cosmic) {
      ctx.fillStyle = 'rgba(27, 94, 32, 0.55)';
      for (let i = 0; i < 3; i++) {
        const ry = isTop ? topHeight - 20 + i * 10 : bottomY + 10 + i * 10;
        ctx.fillRect(x + 10, ry, 40, 5);
      }
    }
  };

  paintSegment(0, topHeight, true);
  paintSegment(bottomY, gameHeight - bottomY, false);
}

export function drawBird(
  ctx: CanvasRenderingContext2D,
  birdY: number,
  birdSize: number,
  velocity: number,
  golden: boolean,
) {
  const cx = 100;
  const cy = birdY + birdSize / 2;
  const rx = birdSize / 2;
  const ry = birdSize / 2;
  // Cosmetic wing tilt from velocity — does not affect hitbox
  const wingTilt = Math.max(-0.7, Math.min(0.7, -velocity * 0.08));

  ctx.save();
  ctx.shadowBlur = golden ? 16 : 10;
  ctx.shadowColor = golden ? 'rgba(255, 215, 0, 0.45)' : 'rgba(0,0,0,0.3)';

  // Body
  const body = ctx.createRadialGradient(cx - 4, cy - 4, 2, cx, cy, rx);
  if (golden) {
    body.addColorStop(0, '#FFF3A0');
    body.addColorStop(0.45, '#FFD700');
    body.addColorStop(1, '#C9A000');
  } else {
    body.addColorStop(0, '#7EB8F0');
    body.addColorStop(0.5, '#4A90D9');
    body.addColorStop(1, '#2E5A8C');
  }
  ctx.fillStyle = body;
  ctx.beginPath();
  ctx.ellipse(cx, cy, rx, ry, 0, 0, Math.PI * 2);
  ctx.fill();

  if (golden) {
    ctx.fillStyle = 'rgba(255, 255, 255, 0.35)';
    ctx.beginPath();
    ctx.ellipse(cx - 5, cy - 8, 8, 5, 0, 0, Math.PI * 2);
    ctx.fill();
  }

  // Wing
  ctx.save();
  ctx.translate(cx - 10, cy);
  ctx.rotate(wingTilt);
  const wing = ctx.createLinearGradient(-12, -8, 12, 8);
  if (golden) {
    wing.addColorStop(0, '#FFE566');
    wing.addColorStop(1, '#E6A800');
  } else {
    wing.addColorStop(0, '#FFB347');
    wing.addColorStop(1, '#E06B00');
  }
  ctx.fillStyle = wing;
  ctx.beginPath();
  ctx.ellipse(0, 0, 12, 8, -Math.PI / 4, 0, Math.PI * 2);
  ctx.fill();
  ctx.restore();

  // Eye
  ctx.shadowBlur = 0;
  ctx.fillStyle = '#FFF';
  ctx.beginPath();
  ctx.arc(cx + 8, cy - 5, 5, 0, Math.PI * 2);
  ctx.fill();
  ctx.fillStyle = '#111';
  ctx.beginPath();
  ctx.arc(cx + 10, cy - 5, 2.5, 0, Math.PI * 2);
  ctx.fill();
  ctx.fillStyle = '#FFF';
  ctx.beginPath();
  ctx.arc(cx + 9, cy - 6, 1, 0, Math.PI * 2);
  ctx.fill();

  // Beak
  const beak = ctx.createLinearGradient(cx + 12, cy, cx + 26, cy);
  beak.addColorStop(0, '#FF8A65');
  beak.addColorStop(1, '#E53935');
  ctx.fillStyle = beak;
  ctx.beginPath();
  ctx.moveTo(cx + 14, cy - 3);
  ctx.lineTo(cx + 26, cy);
  ctx.lineTo(cx + 14, cy + 3);
  ctx.closePath();
  ctx.fill();

  ctx.restore();
}

export function drawScore(ctx: CanvasRenderingContext2D, score: number, w: number) {
  ctx.save();
  ctx.font = 'bold 36px "Press Start 2P", monospace';
  ctx.fillStyle = '#FFF';
  ctx.strokeStyle = 'rgba(0,0,0,0.35)';
  ctx.lineWidth = 4;
  const text = String(score);
  const tw = ctx.measureText(text).width;
  const x = w / 2 - tw / 2;
  ctx.strokeText(text, x, 60);
  ctx.fillText(text, x, 60);
  ctx.restore();
}

export function drawStartHint(ctx: CanvasRenderingContext2D, w: number, _h: number) {
  ctx.save();
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  // Directly under the score (score baseline ~y=60).
  const cx = w / 2;
  const cy = 108;
  ctx.font = 'bold 20px "Press Start 2P", monospace';
  ctx.fillStyle = '#FFF';
  ctx.shadowColor = '#000';
  ctx.shadowBlur = 6;
  ctx.fillText('НАЖМИТЕ ПРОБЕЛ', cx, cy);
  ctx.font = '16px monospace';
  ctx.shadowBlur = 0;
  ctx.fillText('или кликните мышкой', cx, cy + 32);
  ctx.restore();
}
