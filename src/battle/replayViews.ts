import { asJson } from '../game/api';
import type { GameView, MoveResp } from '../game/types';
import type { RunBattle } from '../stores/run';

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
  const run = queue.then(() => rebuild(battle, signal));
  queue = run.catch(() => {});
  return run;
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
  const first = await post<GameView>('/api/new', { you: battle.yourDeck, bot: battle.botDeck, seed: battle.seed });
  if (!first.players) throw new Error((first as { error?: string }).error ?? 'game did not start');
  const views = [first];
  for (const move of battle.actions) {
    if (signal.aborted) throw new DOMException('replay left', 'AbortError');
    const res = await post<MoveResp>('/api/move', move);
    if (!res.ok) throw new Error(`replay diverged at ${move.type}: ${res.error ?? 'rejected'}`);
    views.push(res.state);
  }
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
