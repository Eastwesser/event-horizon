/** proto3 JSON omits numeric zeros — never render a blank price label. */

export function formatRubPrice(price: number | null | undefined): string {
  const n = typeof price === 'number' && Number.isFinite(price) ? price : 0;
  return n === 0 ? 'Бесплатно' : `${n} ₽`;
}

export function formatTicketPrice(tickets: number | null | undefined): string {
  const n = typeof tickets === 'number' && Number.isFinite(tickets) ? tickets : 0;
  return n === 0 ? 'Бесплатно' : `🎟️ ${n}`;
}
