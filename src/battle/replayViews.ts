import { asJson } from '../game/api';
import type { BotResp, GameView, MoveResp } from '../game/types';
import type { LogStep, RunBattle } from '../stores/run';

// Rebuild a finished battle (ADR-0005): start the seed + decks on the server, apply
// every saved move in order, and keep the view after each one. Index 0 is the
// opening position; the last is the final view. The server holds one global game,
// so this replaces whatever game it had: open a replay only when no live game matters.
// Views are made read-only here: no legal moves, so nothing on the board can be played.

// Anything that stops a battle from rebuilding to its recorded end reads the same to the player.
export class ReplayMismatch extends Error {}

// Builds run one at a time: two interleaving their moves would corrupt the one server game.
let queue: Promise<unknown> = Promise.resolve();

// Stops between moves once `signal` aborts (the replay was left).
export function buildReplayViews(battle: RunBattle, signal: AbortSignal): Promise<GameView[]> {
  return inQueue(() => rebuild(battle, signal));
}

export function inQueue<T>(job: () => Promise<T>): Promise<T> {
  const run = queue.then(job);
  queue = run.catch(() => {});
  return run;
}

type Start = { you: string[]; bot?: string[]; seed: number };

// Start the game and apply each step in order, keeping the view after each one.
// The same rebuild serves a finished replay, a resumed battle and a timed-out one.
export async function applySteps(start: Start, steps: LogStep[], signal: AbortSignal): Promise<GameView[]> {
  const first = await post<GameView>('/api/new', start);
  if (!first.players) throw new Error((first as { error?: string }).error ?? 'game did not start');
  const views = [first];
  for (const step of steps) {
    if (signal.aborted) throw new DOMException('replay left', 'AbortError');
    views.push(await applyStep(step));
  }
  return views;
}

async function applyStep(step: LogStep): Promise<GameView> {
  if (step === 'bot') {
    const res = await post<BotResp>('/api/bot', {});
    if (!res.acted) throw new Error('replay diverged: the bot had no move');
    return res.state;
  }
  const res = await post<MoveResp>('/api/move', step);
  if (!res.ok) throw new Error(`replay diverged at ${step.type}: ${res.error ?? 'rejected'}`);
  return res.state;
}

async function rebuild(battle: RunBattle, signal: AbortSignal): Promise<GameView[]> {
  try {
    return await replayOnServer(battle, signal);
  } catch (e) {
    if (signal.aborted) throw e;
    throw new ReplayMismatch(e instanceof Error ? e.message : String(e));
  }
}

async function replayOnServer(battle: RunBattle, signal: AbortSignal): Promise<GameView[]> {
  const views = await applySteps({ you: battle.yourDeck, bot: battle.botDeck, seed: battle.seed }, battle.actions, signal);
  const last = views[views.length - 1];
  if (last.phase !== 'over' || last.outcome !== battle.result) throw new Error('replay does not end where the battle did');
  return views.map(readOnly);
}

function post<T>(url: string, body: object): Promise<T> {
  return fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }).then((r) => asJson<T>(r));
}

function readOnly(view: GameView): GameView {
  return { ...view, pending: null, players: [{ ...view.players[0], legalMoves: [] }, { ...view.players[1], legalMoves: [] }] };
}
