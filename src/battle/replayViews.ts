import type { GameView, MoveResp } from '../game/types';
import type { RunBattle } from '../stores/run';

// Rebuild a finished battle (ADR-0005): start the seed + decks on the server, apply
// every saved move in order, and keep the view after each one. Index 0 is the
// opening position; the last is the final view. The server holds one global game,
// so this replaces whatever game it had: open a replay only when no live game matters.
// Views are made read-only here: no legal moves, so nothing on the board can be played.
export function buildReplayViews(battle: RunBattle): Promise<GameView[]> {
  // One build per battle: a second caller (React's double effect) shares it, since two
  // builds would interleave their moves on the server's single game.
  const running = inFlight.get(battle) ?? replayOnServer(battle);
  inFlight.set(battle, running);
  void running.finally(() => inFlight.delete(battle)).catch(() => {});
  return running;
}

const inFlight = new WeakMap<RunBattle, Promise<GameView[]>>();

async function replayOnServer(battle: RunBattle): Promise<GameView[]> {
  const first = await post<GameView>('/api/new', { you: battle.yourDeck, bot: battle.botDeck, seed: battle.seed });
  const views = [first];
  for (const move of battle.actions) {
    const res = await post<MoveResp>('/api/move', move);
    if (!res.ok) throw new Error(`replay diverged at ${move.type}: ${res.error ?? 'rejected'}`);
    views.push(res.state);
  }
  return views.map(readOnly);
}

async function post<T>(url: string, body: object): Promise<T> {
  const res = await fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return (await res.json()) as T;
}

function readOnly(view: GameView): GameView {
  return { ...view, pending: null, players: [{ ...view.players[0], legalMoves: [] }, { ...view.players[1], legalMoves: [] }] };
}
