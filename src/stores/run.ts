import type { Move } from '../game/types';

// docs/rules.md §4: a run ends at 5 wins or 2 losses.
export const RUN_WINS_TO_WIN = 5;
export const RUN_LOSSES_TO_LOSE = 2;

export type RunStatus = 'active' | 'won' | 'lost' | 'quit';

// One finished game, with what it takes to replay it (ADR-0005): the seed, both
// decks as card ids in dealt order, and every applied move, bot's included.
export type RunBattle = {
  result: 'win' | 'loss' | 'tie';
  seed: number;
  yourDeck: string[];
  botDeck: string[];
  actions: Move[];
  endedAt: string; // ISO string
};

export type Run = {
  deck: Record<string, number>; // the finalized deck, cardId → count
  battles: RunBattle[];
  status: RunStatus;
  historyId?: string; // the history entry made when this deck was finalized
};

export function newRun(deck: Record<string, number>): Run {
  return { deck: { ...deck }, battles: [], status: 'active' };
}

// A tie counts toward neither wins nor losses.
export function runScore(run: Pick<Run, 'battles'>): { wins: number; losses: number; ties: number } {
  const count = (result: RunBattle['result']) => run.battles.filter((b) => b.result === result).length;
  return { wins: count('win'), losses: count('loss'), ties: count('tie') };
}

export function runLine(run: Run): string {
  const { wins, losses, ties } = runScore(run);
  const tieText = ties > 0 ? `, ${ties} ${ties === 1 ? 'tie' : 'ties'}` : '';
  return `Run: ${wins}–${losses}${tieText}`;
}

// Quitting ends an active run; any other run is unchanged.
export function quit(run: Run): Run {
  return run.status === 'active' ? { ...run, status: 'quit' } : run;
}

// The run after a battle ends. A finished run, or a battle already recorded
// (same seed), comes back unchanged.
export function withBattle(run: Run, battle: RunBattle): Run {
  if (run.status !== 'active' || run.battles.some((b) => b.seed === battle.seed)) return run;
  const next = { ...run, battles: [...run.battles, battle] };
  const { wins, losses } = runScore(next);
  if (wins >= RUN_WINS_TO_WIN) return { ...next, status: 'won' };
  if (losses >= RUN_LOSSES_TO_LOSE) return { ...next, status: 'lost' };
  return next;
}
