import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

// F8 / F13: when your Active is knocked out and you have more than one Pokémon
// on the Bench, everything dims except your Bench, which glows; you are told to
// "Choose your new Active", and tapping a Bench Pokémon promotes it.
// The game is driven through the API: you bench every Basic you can and pass,
// the bot attacks until it knocks out your Active (seed 1 gets there on turn 8).

type Move = { type: string; player: number; [k: string]: unknown };
type Mon = { card: { id: string; name: string } } | null;
type View = {
  phase: string;
  active: number;
  pending: { player: number; kind: string } | null;
  players: { hand: { id: string; stage: string }[]; bench: Mon[]; active: Mon; legalMoves: Move[] }[];
};

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

async function setUp(api: APIRequestContext) {
  const start = await post(api, '/api/new', { seed: 1 });
  const basics = start.players[0].hand.filter((c) => c.stage === 'Basic');
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basics[0].id });
  for (let i = 1; i < basics.length && i <= 3; i++) await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: basics[i].id, slot: i - 1 });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
}

// Your turn: bench a Basic if you can, else pass. Their turn: let the bot act.
async function stepUntilYouPromote(api: APIRequestContext): Promise<View> {
  let v = await (await api.get('/api/state')).json() as View;
  for (let i = 0; i < 200; i++) {
    if (v.phase !== 'main') break;
    if (v.pending?.player === 0 && v.pending.kind === 'new_active') return v;
    const actor = v.pending ? v.pending.player : v.active;
    if (actor !== 0) v = await post(api, '/api/bot');
    else {
      const bench = v.players[0].legalMoves.find((m) => m.type === 'PlayBasic');
      v = await post(api, '/api/move', bench ?? { type: 'EndTurn', player: 0 });
    }
  }
  throw new Error('never reached a knockout promotion with a choice');
}

async function topAt(page: Page, testId: string) {
  const b = (await page.getByTestId(testId).first().boundingBox())!;
  return page.evaluate(([x, y]) => document.elementFromPoint(x, y)?.closest('[data-testid]')?.getAttribute('data-testid'), [b.x + b.width / 2, b.y + b.height / 2]);
}

test('knockout promotion: overlay dims everything except your Bench', async ({ page, request }) => {
  await setUp(request);
  const v = await stepUntilYouPromote(request);
  const options = v.players[0].legalMoves.filter((m) => m.type === 'ChooseNewActive');
  expect(options.length).toBeGreaterThan(1);

  await page.goto('/?battle');
  await expect(page.getByTestId('choice-prompt')).toHaveText(/Choose your new Active/);
  await expect(page.getByTestId('choice-dim')).toBeVisible();

  // Lit: exactly the Bench slots the engine offers; nothing else.
  const lit = page.locator('[data-lit="true"]');
  await expect(lit).toHaveCount(options.length);
  for (const o of options) await expect(page.getByTestId('my-bench').locator(`[data-pos="${(o.benchIndex as number) + 1}"] [data-lit="true"]`)).toHaveCount(1);

  // Dimmed: the dim layer is on top of every other zone, under the lit slots.
  for (const zone of ['opp-active', 'opp-bench', 'opp-hand', 'my-hand', 'my-deck', 'end-turn', 'my-energy']) expect(await topAt(page, zone), zone).toBe('choice-dim');
  expect(await topAt(page, 'my-bench')).not.toBe('choice-dim');

  await page.waitForFunction(() => [...document.images].every((i) => i.complete));
  await page.screenshot({ path: 'test-results/choice-overlay-390x844.png' });

  // Tapping a lit Bench Pokémon promotes it and the overlay goes away.
  const pick = options[0].benchIndex as number;
  const chosen = v.players[0].bench[pick]!.card.name;
  await page.getByTestId('my-bench').locator(`[data-pos="${pick + 1}"]`).click();
  await expect(page.getByTestId('choice-dim')).toHaveCount(0);
  await expect(page.getByTestId('my-active').locator('[title]').first()).toHaveAttribute('title', chosen);
});

// 2.14: concede is immediate from any state, including while a choice is up.
test('concede from the menu while the choice dim is up ends the game', async ({ page, request }) => {
  await setUp(request);
  await stepUntilYouPromote(request);

  await page.goto('/?battle');
  await expect(page.getByTestId('choice-dim')).toBeVisible();
  expect(await topAt(page, 'menu')).toBe('menu');
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();

  await expect(page.getByTestId('end-screen')).toBeVisible();
  await expect(page.getByTestId('end-reason')).toHaveText('Concede');
  await expect(page.getByTestId('choice-dim')).toHaveCount(0);
});
