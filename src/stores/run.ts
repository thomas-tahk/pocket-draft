import type { Move } from '../game/types';

// docs/rules.md §4: a run ends at 5 wins or 2 losses.
export const RUN_WINS_TO_WIN = 5;
export const RUN_LOSSES_TO_LOSE = 2;

export type RunStatus = 'active' | 'won' | 'lost' | 'quit';

// One finished game, with what it takes to replay it (ADR-0005): the seed, both
// decks as card ids in dealt order, and every applied move, bot's included.
export const RESULT_LABEL = { win: 'Win', loss: 'Loss', tie: 'Tie' } as const;
export const STATUS_LABEL = { active: 'In progress', won: 'Run won', lost: 'Run lost', quit: 'Run quit' } as const;

export type RunBattle = {
  result: 'win' | 'loss' | 'tie';
  seed: number;
  yourDeck: string[];
  botDeck: string[];
  energy?: string[]; // your deck's Energy Zone types; absent in older records (the default applies)
  actions: Move[];
  endedAt: string; // ISO string
};

// What the app did to the server's game, in order: a move the player made, or one bot step.
// The bot is deterministic, so replaying these from the seed rebuilds the same game.
// A bot step remembers how many actions the game had after it, to catch a rebuild that drifts.
export type BotStep = { bot: number };
export type LogStep = Move | BotStep;

// The battle being played right now, saved so a reload or a closed tab can resume it.
// The bot's deck is not here: it comes from the seed, and is only revealed at game over.
export type InProgressBattle = {
  seed: number;
  yourDeck: string[];
  energy?: string[];
  log: LogStep[];
  leftAt: string | null; // ISO time the player left to the run screen; null while playing
  closedAt?: string | null; // ISO time the page was hidden or closed mid-battle, without Leave battle
};

// A left battle is lost this long after leaving (the owner's "about one full turn").
export const LEAVE_GRACE_MS = 90_000;

export function leaveMsLeft(battle: InProgressBattle, now: number): number | null {
  const since = battle.leftAt ?? battle.closedAt;
  if (!since) return null;
  return Math.max(0, Date.parse(since) + LEAVE_GRACE_MS - now);
}

export type Run = {
  deck: Record<string, number>; // the finalized deck, cardId → count
  energy?: string[]; // its chosen Energy Zone types, sent to every battle of the run
  battles: RunBattle[];
  status: RunStatus;
  historyId?: string; // the history entry made when this deck was finalized
  inProgress?: InProgressBattle;
};

export function newRun(deck: Record<string, number>, energy?: string[]): Run {
  return { deck: { ...deck }, ...(energy?.length ? { energy: [...energy] } : {}), battles: [], status: 'active' };
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
  return run.status === 'active' ? { ...run, status: 'quit', inProgress: undefined } : run;
}

// The run after a battle ends. A finished run, or a battle already recorded
// (same seed), comes back unchanged.
export function withBattle(run: Run, battle: RunBattle): Run {
  if (run.status !== 'active' || run.battles.some((b) => b.seed === battle.seed)) return run;
  const next = { ...run, battles: [...run.battles, battle], inProgress: undefined };
  const { wins, losses } = runScore(next);
  if (wins >= RUN_WINS_TO_WIN) return { ...next, status: 'won' };
  if (losses >= RUN_LOSSES_TO_LOSE) return { ...next, status: 'lost' };
  return next;
}
