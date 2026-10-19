import { test, expect, type APIRequestContext } from '@playwright/test';

// F12, F18, F19 at phone size: End Turn asks before throwing away an energy
// attach or an attack, Concede sits in a menu behind a confirm, and the game
// ends on a screen with the result, the reason, the score, Rematch and Back.

type Card = { id: string; stage: string };
type Move = { type: string };
type View = {
  phase: string;
  active: number;
  turn: number;
  pending: { player: number } | null;
  players: { hand: Card[]; legalMoves: Move[] }[];
};

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

// Seed 7: you go first. Setup is placed through the API; the screen has no
// setup input yet.
async function setUpGame(api: APIRequestContext) {
  const start = await post(api, '/api/new', { seed: 7 });
  const basics = start.players[0].hand.filter((c) => c.stage === 'Basic');
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basics[0].id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  // the bot places its setup one card per call, then presses Ready
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
}

test('concede from the menu, confirm, and see Loss with the reason', async ({ page, request }) => {
  await setUpGame(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();

  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  // Cancel first: the game goes on.
  await expect(page.getByTestId('confirm')).toBeVisible();
  await page.getByTestId('confirm-cancel').click();
  await expect(page.getByTestId('confirm')).toBeHidden();
  await expect(page.getByTestId('end-screen')).toBeHidden();

  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();

  const end = page.getByTestId('end-screen');
  await expect(end).toBeVisible();
  await expect(end.getByTestId('end-result')).toHaveText('Loss');
  await expect(end.getByTestId('end-reason')).toHaveText('Concede');
  await expect(end.getByTestId('end-score')).toHaveText(/You 0\D+0 Opponent/);
  await expect(end.getByRole('button', { name: 'Back' })).toBeVisible();
  await page.waitForFunction(() => [...document.images].every((i) => i.complete));
  await page.screenshot({ path: 'test-results/battle-end-390x844.png' });

  // Rematch starts a fresh game.
  await end.getByRole('button', { name: 'Rematch' }).click();
  await expect(end).toBeHidden();
  const fresh = await (await request.get('/api/state')).json();
  expect(fresh.phase).not.toBe('over');
});

test('End Turn asks only when you could still attach energy or attack', async ({ page, request }) => {
  await setUpGame(request);
  await page.goto('/?battle');
  const endTurn = page.getByTestId('end-turn').getByRole('button');

  // Turn 1, first player: no energy and no attack, so End Turn ends at once.
  await expect(endTurn).toBeEnabled();
  await endTurn.click();
  await expect(page.getByTestId('confirm')).toBeHidden();

  // The bot plays its turn on its own and hands it back.
  await expect.poll(async () => (await (await request.get('/api/state')).json()).turn, { timeout: 15_000 }).toBe(3);
  await expect(endTurn).toBeEnabled();

  // Turn 3: energy is waiting, so End Turn asks, naming it.
  const v: View = await (await request.get('/api/state')).json();
  const canAttack = v.players[0].legalMoves.some((m) => m.type === 'UseAttack');
  await endTurn.click();
  const confirm = page.getByTestId('confirm');
  await expect(confirm).toContainText('attach energy');
  if (canAttack) await expect(confirm).toContainText('attack');
  else await expect(confirm).not.toContainText(/\battack\b/);

  await page.getByTestId('confirm-cancel').click();
  expect((await (await request.get('/api/state')).json()).turn).toBe(3);

  await endTurn.click();
  await page.getByTestId('confirm-ok').click();
  await expect.poll(async () => (await (await request.get('/api/state')).json()).turn).toBeGreaterThan(3);
});
