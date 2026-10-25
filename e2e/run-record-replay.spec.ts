import { test, expect, type APIRequestContext } from '@playwright/test';
import type { GameView, Move } from '../src/game/types';

// ADR 0005: the record a finished game hands out (seed + both decks + every
// applied move) replays to the same game. Driven through the API, as the app does.

async function post<T>(api: APIRequestContext, url: string, data?: object): Promise<T> {
  const body = await (await api.post(url, data ? { data } : undefined)).json();
  return (body.state ?? body) as T;
}

// bot: false replays without the bot deck, so the server derives it from the seed.
for (const withBot of [true, false]) {
test(`a finished game replays from its record to the same final state (bot deck ${withBot ? 'given' : 'derived from seed'})`, async ({ request }) => {
  const { deck } = await post<{ deck: string[] }>(request, '/api/autodeck');
  let v = await post<GameView>(request, '/api/new', { you: deck, seed: 424242 });
  expect(v.record.actions).toEqual([]); // no record before the game is over

  const basic = v.players[0].hand.find((c) => c.stage === 'Basic')!;
  await post(request, '/api/move', { type: 'SetupActive', player: 0, cardId: basic.id });
  await post(request, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 8 && (await post<GameView>(request, '/api/bot')).phase === 'setup'; i++);

  // a few real turns: the bot plays its side, we end ours
  for (let turn = 0; turn < 3; turn++) {
    v = await post<GameView>(request, '/api/state');
    if (v.phase !== 'main') break;
    if (v.active === 0) await post(request, '/api/move', { type: 'EndTurn', player: 0 });
    for (let i = 0; i < 40; i++) if (!(await post<{ acted: boolean }>(request, '/api/bot')).acted) break;
  }
  const final = await post<GameView>(request, '/api/move', { type: 'Concede', player: 0 });
  expect(final.phase).toBe('over');

  const { seed, you, bot, actions } = final.record;
  expect(seed).toBe(424242);
  expect(you).toEqual(deck);
  expect(bot).toHaveLength(20);
  expect(actions.length).toBeGreaterThan(5);

  await post(request, '/api/new', { you, bot: withBot ? bot : undefined, seed });
  let replay: GameView | undefined;
  for (const move of actions as Move[]) {
    const res = await (await request.post('/api/move', { data: move })).json();
    expect(res.ok, `${move.type}: ${res.error}`).toBe(true);
    replay = res.state;
  }
  expect(replay!.narration).toEqual(final.narration);
  expect(replay!.winner).toBe(final.winner);
  expect(replay!.reason).toBe(final.reason);
  expect(replay!.players.map((p) => [p.points, p.deckCount, p.discardCount])).toEqual(
    final.players.map((p) => [p.points, p.deckCount, p.discardCount]),
  );
  expect(replay!.record.bot).toEqual(bot);
});
}
