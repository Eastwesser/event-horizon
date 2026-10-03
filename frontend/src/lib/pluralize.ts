/** Russian plural: 1 карта / 2–4 карты / 5+ карт */
export function pluralCards(n: number): string {
  const abs = Math.abs(Math.trunc(n));
  const mod10 = abs % 10;
  const mod100 = abs % 100;
  if (mod10 === 1 && mod100 !== 11) return `${abs} карта`;
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return `${abs} карты`;
  return `${abs} карт`;
}
