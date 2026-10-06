import { BOOST_COST } from '../hooks/useGameBoost';

/** Shared economy line for every game help block. */
export const BOOST_ECONOMY_HELP =
  `Boost (−${BOOST_COST} лампочек) = легальный чит для залипания: забег не в лидерборд и без билетиков/ламп за очки. Честная игра — в топ и в магазин.`;

export const CURRENCY_HELP =
  'Лампочки — для бустов. Билетики — валюта магазина.';

/** Per-game boost effect (one short sentence). */
export const BOOST_EFFECT_BY_GAME: Record<string, string> = {
  hexagon: 'Pancaker: весь поднос одного типа блинов (легче собирать стопки).',
  flappy: 'Flappy: трубы и мир медленнее, птичка с обычной скоростью.',
  towers: 'Builder: блок качается заметно медленнее — проще положить ровно.',
  memory: 'Memonia: в начале подсвечивается одна пара-подсказка.',
  hanoi: 'Hanoi: доступно авто-решение (платный чит лампами).',
  twenty48: '2048: один undo хода за буст (не в лидерборд).',
  gears: 'Gears: следующий дроп копирует текущий номинал.',
  companion: 'Companion: без гонки за топом — важнее ежедневный подарок заботы.',
};

export function boostHelpLines(gameId: string): string[] {
  const effect = BOOST_EFFECT_BY_GAME[gameId];
  return [BOOST_ECONOMY_HELP, CURRENCY_HELP, ...(effect ? [effect] : [])];
}
