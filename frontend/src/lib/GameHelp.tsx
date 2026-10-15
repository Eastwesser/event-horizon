import { GAME_HOW_TO } from './gameHowTo';
// import { boostHelpLines } from './gameBoostCopy';

export function GameHelp({ gameId }: { gameId: string }) {
  const rules = GAME_HOW_TO[gameId] ?? [];
  return (
    <>
      {rules.map((line) => (
        <p key={line}>{line}</p>
      ))}
      {/* {boostHelpLines(gameId).map((line) => (
        <p key={line}>{line}</p>
      ))} */}
    </>
  );
}

// Убран импорт boostHelpLines. Информация о бустах теперь только в BoostCheckbox.