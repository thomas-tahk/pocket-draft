import type { GameView } from '../game/types';
import type { InProgressBattle, RunBattle } from '../stores/run';
import { applySteps, inQueue } from './replayViews';

// Leaving a battle (#117). The saved log is the only truth: a resume or a timeout
// rebuilds the game from seed + deck + log, never from whatever game the server
// holds now (a replay opened during the countdown replaces it).

const never = new AbortController().signal;

// The battle as the player left it: the view after the last logged step.
export async function resumeView(saved: InProgressBattle): Promise<GameView> {
  const views = await inQueue(() => applySteps({ you: saved.yourDeck, seed: saved.seed }, saved.log, never));
  return views[views.length - 1];
}

// A battle left past the grace period: finish it as a concede and take the game's own record,
// which carries the bot's deck and every move, so the loss replays like any other.
export async function timedOutBattle(saved: InProgressBattle, now: Date): Promise<RunBattle> {
  try {
    const log = [...saved.log, { type: 'Concede' as const, player: 0 }];
    const views = await inQueue(() => applySteps({ you: saved.yourDeck, seed: saved.seed }, log, never));
    const { record } = views[views.length - 1];
    return { result: 'loss', seed: record.seed, yourDeck: record.you, botDeck: record.bot, actions: record.actions, endedAt: now.toISOString() };
  } catch {
    // The engine could not rebuild it: the loss still counts; its replay will say it no longer matches.
    return { result: 'loss', seed: saved.seed, yourDeck: saved.yourDeck, botDeck: [], actions: [], endedAt: now.toISOString() };
  }
}
