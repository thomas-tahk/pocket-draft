import type { GameView } from '../game/types';
import { KIND_LABEL } from './History';

// One plain line for the step on show, worded like the History list (You / Opponent).
// Placeholder wording, the owner's to change (docs/questions/m4-app-flow.md).
export function captionFor(view: GameView, step: number): string {
  if (step === 0) return 'Battle start';
  const action = view.actions[view.actions.length - 1];
  if (view.phase === 'over') return endLine(view, action?.player ?? 0, action?.kind === 'concede');
  if (!action) return 'Battle start';
  const who = action.player === 0 ? 'You' : 'Opponent';
  return `${who} ${KIND_LABEL[action.kind] ?? action.kind}${action.card ? ` · ${action.card.name}` : ''}`;
}

function endLine(view: GameView, last: number, conceded: boolean): string {
  const [you, opp] = view.players;
  if (conceded) return last === 0 ? 'You conceded — Opponent wins' : 'Opponent conceded — You win';
  const score = `${you.points}–${opp.points}`;
  if (view.outcome === 'win') return `You win ${score}`;
  if (view.outcome === 'loss') return `You lose ${score}`;
  return `Tie ${score}`;
}
