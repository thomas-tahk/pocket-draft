import type { Move } from '../game/types';

// docs/rules.md §4: a run ends at 5 wins or 2 losses.
export const RUN_WINS_TO_WIN = 5;
export const RUN_LOSSES_TO_LOSE = 2;

export type RunStatus = 'active' | 'won' | 'lost' | 'quit';

// One finished game, with what it takes to replay it (ADR-0005): the seed, both
// decks as card ids in dealt order, and every applied move, bot's included.
export type RunBattle = {
  result: 'win' | 'loss';
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
};

export function newRun(deck: Record<string, number>): Run {
  return { deck: { ...deck }, battles: [], status: 'active' };
}

export function runScore(run: Pick<Run, 'battles'>): { wins: number; losses: number } {
  const wins = run.battles.filter((b) => b.result === 'win').length;
  return { wins, losses: run.battles.length - wins };
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
