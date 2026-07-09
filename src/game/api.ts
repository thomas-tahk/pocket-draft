import type { GameView, Move, MoveResp, BotResp } from './types';

// Thin wrappers over the Go server's /api endpoints. Relative URLs resolve via
// the Vite proxy in dev (see vite.config.ts).

async function asJson<T>(res: Response): Promise<T> {
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return (await res.json()) as T;
}

export function newGame(seed?: number): Promise<GameView> {
  const q = seed === undefined ? '' : `?seed=${seed}`;
  return fetch(`/api/new${q}`, { method: 'POST' }).then((r) => asJson<GameView>(r));
}

export function getState(): Promise<GameView> {
  return fetch('/api/state').then((r) => asJson<GameView>(r));
}

export function sendMove(move: Move): Promise<MoveResp> {
  return fetch('/api/move', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(move),
  }).then((r) => asJson<MoveResp>(r));
}

export function botMove(): Promise<BotResp> {
  return fetch('/api/bot', { method: 'POST' }).then((r) => asJson<BotResp>(r));
}
