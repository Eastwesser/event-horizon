/** Canvas paint for Gears. Physics and RADIUS stay in GearsGame. */

const INK = '#2a2118';

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

export function drawWell(ctx: CanvasRenderingContext2D, w: number, h: number, dangerY: number) {
  const sky = ctx.createLinearGradient(0, 0, 0, h);
  sky.addColorStop(0, '#31456e');
  sky.addColorStop(0.55, '#1e2a44');
  sky.addColorStop(1, '#141c2c');
  ctx.fillStyle = sky;
  ctx.fillRect(0, 0, w, h);

  ctx.strokeStyle = 'rgba(198,161,90,0.14)';
  ctx.lineWidth = 1;
  for (let y = h - 10; y > h * 0.42; y -= 16) {
    ctx.beginPath();
    ctx.moveTo(12, y);
    ctx.lineTo(w - 12, y);
    ctx.stroke();
  }

  ctx.strokeStyle = 'rgba(232, 90, 90, 0.72)';
  ctx.setLineDash([6, 5]);
  ctx.lineWidth = 1.6;
  ctx.beginPath();
  ctx.moveTo(12, dangerY);
  ctx.lineTo(w - 12, dangerY);
  ctx.stroke();
  ctx.setLineDash([]);
}

export function drawAim(ctx: CanvasRenderingContext2D, x: number, y0: number, y1: number) {
  ctx.save();
  ctx.strokeStyle = 'rgba(246, 234, 208, 0.5)';
  ctx.setLineDash([5, 6]);
  ctx.lineWidth = 1.5;
  ctx.beginPath();
  ctx.moveTo(x, y0);
  ctx.lineTo(x, y1);
  ctx.stroke();
  ctx.restore();
}

export function drawMergeFlash(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  t: number,
) {
  ctx.save();
  ctx.beginPath();
  ctx.arc(x, y, r * (1 + (1 - t) * 0.85), 0, Math.PI * 2);
  ctx.strokeStyle = `rgba(255, 233, 184, ${t})`;
  ctx.lineWidth = 3.4;
  ctx.stroke();
  ctx.beginPath();
  ctx.arc(x, y, r * (0.35 + (1 - t) * 0.45), 0, Math.PI * 2);
  ctx.fillStyle = `rgba(255, 233, 184, ${t * 0.28})`;
  ctx.fill();
  ctx.restore();
}

export function drawGear(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  color: string,
  level: number,
  alpha = 1,
  flower = false,
  rot = 0,
) {
  ctx.save();
  ctx.globalAlpha = alpha;
  ctx.translate(x, y);
  ctx.rotate(rot);

  if (flower) {
    const petals = 6 + (level % 3);
    for (let i = 0; i < petals; i++) {
      const ang = (i / petals) * Math.PI * 2 - Math.PI / 2;
      ctx.beginPath();
      ctx.ellipse(Math.cos(ang) * r * 0.52, Math.sin(ang) * r * 0.52, r * 0.4, r * 0.26, ang, 0, Math.PI * 2);
      ctx.fillStyle = color;
      ctx.fill();
      ctx.strokeStyle = INK;
      ctx.lineWidth = 1.4;
      ctx.stroke();
    }
    const bore = Math.max(r * 0.46, 8);
    ctx.beginPath();
    ctx.arc(0, 0, bore, 0, Math.PI * 2);
    ctx.fillStyle = '#12141c';
    ctx.fill();
    ctx.strokeStyle = INK;
    ctx.lineWidth = 1.6;
    ctx.stroke();
  } else {
    const teeth = 10 + Math.min(level, 6);
    const tip = r;
    const valley = r * 0.72;
    const root = r * 0.64;
    ctx.beginPath();
    for (let i = 0; i < teeth; i++) {
      const a0 = (i / teeth) * Math.PI * 2 - Math.PI / 2;
      const a1 = ((i + 0.32) / teeth) * Math.PI * 2 - Math.PI / 2;
      const a2 = ((i + 0.5) / teeth) * Math.PI * 2 - Math.PI / 2;
      const a3 = ((i + 0.82) / teeth) * Math.PI * 2 - Math.PI / 2;
      const a4 = ((i + 1) / teeth) * Math.PI * 2 - Math.PI / 2;
      const pts: [number, number][] = [
        [Math.cos(a0) * valley, Math.sin(a0) * valley],
        [Math.cos(a1) * tip, Math.sin(a1) * tip],
        [Math.cos(a2) * tip, Math.sin(a2) * tip],
        [Math.cos(a3) * valley, Math.sin(a3) * valley],
        [Math.cos(a4) * root, Math.sin(a4) * root],
      ];
      for (let p = 0; p < pts.length; p++) {
        if (i === 0 && p === 0) ctx.moveTo(pts[p][0], pts[p][1]);
        else ctx.lineTo(pts[p][0], pts[p][1]);
      }
    }
    ctx.closePath();
    ctx.fillStyle = color;
    ctx.fill();
    ctx.strokeStyle = INK;
    ctx.lineWidth = Math.max(1.6, r * 0.07);
    ctx.stroke();

    ctx.beginPath();
    ctx.arc(0, 0, r * 0.58, 0, Math.PI * 2);
    ctx.fillStyle = shade(color, -36);
    ctx.fill();
    ctx.strokeStyle = INK;
    ctx.lineWidth = 1.5;
    ctx.stroke();

    const bore = Math.max(r * 0.4, 8);
    ctx.beginPath();
    ctx.arc(0, 0, bore, 0, Math.PI * 2);
    ctx.fillStyle = '#12141c';
    ctx.fill();
    ctx.strokeStyle = '#d7c9a4';
    ctx.lineWidth = 1.4;
    ctx.stroke();
  }

  ctx.rotate(-rot);
  ctx.fillStyle = '#f6ead0';
  const bore = Math.max(r * 0.4, 8);
  ctx.font = `bold ${Math.max(11, Math.round(bore * 1.2))}px ui-monospace, monospace`;
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText(String(level), 0, 1);
  ctx.restore();
}
