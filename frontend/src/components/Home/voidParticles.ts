/** Static particle field for VOID disk (Level 1 SVG). */
export type VoidParticle = {
  cx: number;
  cy: number;
  r: number;
  tone: 'gold' | 'cyan' | 'white';
  dur: number;
  delay: number;
  spin: 1 | -1;
};

const TONES: VoidParticle['tone'][] = ['gold', 'cyan', 'white'];
const RADII = [0.55, 0.85, 1.25];

export const VOID_PARTICLES: VoidParticle[] = Array.from({ length: 30 }, (_, i) => {
  const angle = (i / 30) * Math.PI * 2 + (i % 5) * 0.17;
  const orbit = 36 + (i % 6) * 3.2;
  const cx = 50 + Math.cos(angle) * orbit;
  const cy = 50 + Math.sin(angle) * orbit * 0.52;
  return {
    cx: Math.round(cx * 10) / 10,
    cy: Math.round(cy * 10) / 10,
    r: RADII[i % 3],
    tone: TONES[i % 3],
    dur: 0.85 + (i % 9) * 0.18,
    delay: (i % 12) * 0.11,
    spin: i % 2 === 0 ? 1 : -1,
  };
});
