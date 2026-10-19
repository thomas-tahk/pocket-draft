import type { GameView } from '../game/types';

// L5: no permanent turn indicator. Each turn start flashes a banner, YOUR TURN
// or OPPONENT'S TURN, in different colours. The key restarts the CSS flash on
// every new turn; the animation hides it again, so nothing stays on the board.
export function TurnBanner({ view }: { view: GameView }) {
  if (view.phase !== 'main') return null;
  const yours = view.active === 0;
  return (
    <div key={view.turn} className="bt-turn-banner" data-testid="turn-banner" data-side={yours ? 'you' : 'opponent'} role="status">
      {yours ? 'YOUR TURN' : "OPPONENT'S TURN"}
    </div>
  );
}
