import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';
import { draftTwenty } from './helpers/draft';
import { expandDeck } from '../src/game/deck';
import type { Move } from '../src/game/types';

// #117: the player can leave. Leaving a run's battle starts a 90 s countdown on the run
// screen; back in time resumes the same game, past it the battle is a loss, even when the
// tab was closed. A reload lands back in the battle. The draft can be abandoned on screen.
// Every resume here first clobbers the server's game, so a pass proves the rebuild from the
// saved log, not luck with the game the server still held.

const KEY = 'pocket-draft-v0.4';

type Card = { id: string; name: string; stage: string };
type View = { phase: string; active: number; pending: unknown; players: { hand: Card[] }[]; record?: unknown };

async function state(api: APIRequestContext): Promise<View> {
  return (await (await api.get('/api/state')).json()) as View;
}

// The board as the player sees it, without the record (a resume rebuilds it the same).
async function board(api: APIRequestContext) {
  const { record: _record, ...rest } = await state(api);
  return rest;
}

async function centre(l: Locator) {
  const b = (await l.boundingBox())!;
  return { x: b.x + b.width / 2, y: b.y + b.height / 2 };
}

async function drag(page: Page, card: Locator, to: Locator) {
  const b = (await card.boundingBox())!;
  const t = await centre(to);
  await page.mouse.move(b.x + Math.min(10, b.width / 2), b.y + b.height / 2);
  await page.mouse.down();
  await page.mouse.move(t.x, t.y, { steps: 8 });
  await page.mouse.up();
}

// A run battle played through setup, waiting on your first turn so the board is still.
async function startRun(page: Page) {
  await page.clock.install();
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await expect(page.getByTestId('run-record')).toBeVisible();
}

async function intoBattle(page: Page, api: APIRequestContext) {
  await startRun(page);
  await enterBattle(page, api);
}

async function enterBattle(page: Page, api: APIRequestContext) {
  await page.getByRole('button', { name: 'Next battle' }).click();
  const ready = page.getByTestId('setup-ready');
  await expect(ready).toBeVisible({ timeout: 15_000 });
  await expect(page.getByTestId('setup-intro')).toHaveCount(0, { timeout: 10_000 });
  const basic = (await state(api)).players[0].hand.find((c) => c.stage === 'Basic')!;
  await drag(page, page.getByTestId('my-hand').locator(`[data-card-id="${basic.id}"]`).first(), page.getByTestId('my-active').locator('[data-spot="active"]'));
  await ready.click();
  await expect.poll(async () => {
    const v = await state(api);
    return v.phase === 'main' && !v.pending && v.active === 0;
  }, { timeout: 30_000 }).toBe(true);
}

async function leave(page: Page) {
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-leave').click();
  await expect(page.getByTestId('leave-countdown')).toBeVisible();
}

// What a replay opened during the countdown does to the server.
const clobberServerGame = (api: APIRequestContext) => api.post('/api/new', { data: { seed: 1 } });

test('leave, return at 60 s: the same game resumes', async ({ page, request }) => {
  await intoBattle(page, request);
  const before = await board(request);
  await leave(page);
  await expect(page.getByRole('button', { name: 'Next battle' })).toHaveCount(0);
  await page.screenshot({ path: 'test-results/leave-countdown-390.png' });
  await clobberServerGame(request);

  await page.clock.fastForward(60_000);
  expect(Number(await page.getByTestId('leave-seconds').textContent())).toBeLessThanOrEqual(30);
  await page.getByRole('button', { name: 'Return to battle' }).click();

  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect.poll(() => board(request)).toEqual(before);
  await expect(page.getByTestId('leave-countdown')).toHaveCount(0);
});

test('leave, 91 s later: a loss is recorded and the countdown is gone', async ({ page, request }) => {
  await intoBattle(page, request);
  await leave(page);
  await page.clock.fastForward(91_000);

  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–1 (active)');
  await expect(page.getByTestId('leave-countdown')).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Next battle' })).toBeVisible();
  await expect(page.getByTestId('run-battles').locator('li')).toContainText('Battle 1: Loss');

  // The lost battle replays like any other.
  await page.getByRole('button', { name: 'Replay' }).click();
  await expect(page.getByTestId('replay-step')).toBeVisible();
  await expect(page.getByTestId('replay-error')).toHaveCount(0);
});

test('reload mid-battle lands in the same battle', async ({ page, request }) => {
  await intoBattle(page, request);
  const before = await board(request);
  await clobberServerGame(request);
  await page.reload();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect.poll(() => board(request)).toEqual(before);
});

test('closed and reopened after 90 s: the battle is a loss', async ({ page, request, context }) => {
  await intoBattle(page, request);
  await leave(page);
  await page.evaluate((key) => {
    const saved = JSON.parse(localStorage.getItem(key)!);
    saved.state.run.inProgress.leftAt = new Date(Date.now() - 200_000).toISOString();
    localStorage.setItem(key, JSON.stringify(saved));
  }, KEY);
  await page.close();

  const reopened = await context.newPage();
  await reopened.goto('/');
  await expect(reopened.getByTestId('run-record')).toHaveText('Run: 0–1 (active)', { timeout: 15_000 });
  await expect(reopened.getByTestId('leave-countdown')).toHaveCount(0);
});

test('reopened inside the 90 s: the countdown is still running', async ({ page, request, context }) => {
  await intoBattle(page, request);
  await leave(page);
  await page.close();
  const reopened = await context.newPage();
  await reopened.goto('/');
  await expect(reopened.getByTestId('leave-countdown')).toBeVisible();
  await reopened.getByRole('button', { name: 'Return to battle' }).click();
  await expect(reopened.getByTestId('battle-board')).toBeVisible();
});

async function pickFirstOffers(page: Page, count: number) {
  for (let pack = 1; pack <= count; pack++) {
    await expect(page.getByText(`Pack ${pack} / 20`)).toBeVisible();
    await page.locator('main button[title]').first().click();
  }
}

async function expectAbandonFlow(page: Page, stillThere: Locator) {
  await page.getByRole('button', { name: 'Abandon draft' }).click();
  await expect(page.getByTestId('abandon-confirm')).toBeVisible();
  await page.getByTestId('abandon-cancel').click();
  await expect(page.getByTestId('abandon-confirm')).toHaveCount(0);
  await expect(stillThere).toBeVisible();
  await page.getByRole('button', { name: 'Abandon draft' }).click();
  await page.getByTestId('abandon-ok').click();
  await expect(page.getByRole('button', { name: 'Start draft' })).toBeVisible();
}

test('abandon the draft from picks: confirm, Cancel keeps it, confirming goes Home', async ({ page }) => {
  page.on('dialog', () => { throw new Error('window.confirm used'); });
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  await pickFirstOffers(page, 2);
  await expectAbandonFlow(page, page.getByText('Pack 3 / 20'));
});

test('abandon the draft from deckbuild: confirm, Cancel keeps it, confirming goes Home', async ({ page }) => {
  page.on('dialog', () => { throw new Error('window.confirm used'); });
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  await pickFirstOffers(page, 20);
  await page.getByRole('button', { name: /Skip shop|Continue to deckbuild/ }).click();
  const heading = page.getByRole('heading', { name: 'Deckbuilding' });
  await expect(heading).toBeVisible();
  await expectAbandonFlow(page, heading);
});

test('abandon the draft from the shop: confirm, Cancel keeps it, confirming goes Home', async ({ page }) => {
  page.on('dialog', () => { throw new Error('window.confirm used'); });
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  await draftTwenty(page);
  const shop = page.getByRole('button', { name: 'Skip shop' });
  await expect(shop).toBeVisible();
  await expectAbandonFlow(page, shop);
});

test('a page closed without the menu starts the clock: past 90 s it is a loss', async ({ page, request }) => {
  await intoBattle(page, request);
  await page.evaluate(() => window.dispatchEvent(new Event('pagehide')));
  await expect.poll(() => page.evaluate((key) => JSON.parse(localStorage.getItem(key)!).state.run.inProgress.closedAt, KEY)).toBeTruthy();
  // A reload restarts the page's fake clock, so age the stamp instead of advancing time.
  await page.evaluate((key) => {
    const saved = JSON.parse(localStorage.getItem(key)!);
    saved.state.run.inProgress.closedAt = new Date(Date.now() - 91_000).toISOString();
    localStorage.setItem(key, JSON.stringify(saved));
  }, KEY);
  await page.reload();
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–1 (active)', { timeout: 15_000 });
  await expect(page.getByTestId('battle-board')).toHaveCount(0);
});

test('a page closed without the menu, back inside 90 s: still the same battle', async ({ page, request }) => {
  await intoBattle(page, request);
  const before = await board(request);
  await page.evaluate(() => window.dispatchEvent(new Event('pagehide')));
  await page.clock.fastForward(60_000);
  await clobberServerGame(request);
  await page.reload();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect.poll(() => board(request)).toEqual(before);
});

test('a timeout that is the 2nd loss ends the run on its summary', async ({ page, request }) => {
  await startRun(page);
  await page.getByRole('button', { name: 'Next battle' }).click();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();
  await expect(page.getByTestId('end-result')).toHaveText('Loss');
  await page.getByRole('button', { name: 'Back' }).click();
  await enterBattle(page, request);
  await leave(page);
  await page.clock.fastForward(91_000);
  await expect(page.getByTestId('run-status')).toHaveText('Run lost');
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–2 (lost)');
  await expect(page.getByTestId('leave-countdown')).toHaveCount(0);
});

// A fixed seed played through the API: setup, a full turn each, ending on your next turn.
async function fixedSeedLog(api: APIRequestContext, you: string[]) {
  const post = async (url: string, data?: object) => (await (await api.post(url, data ? { data } : undefined)).json()) as Record<string, any>;
  const log: Array<Move | { bot: number }> = [];
  const bot = async () => {
    const r = await post('/api/bot');
    if (r.acted) log.push({ bot: r.state.actions.length });
    return r.state as View & { turn: number };
  };
  const send = async (m: Move) => { log.push(m); return (await post('/api/move', m)).state as View; };
  const v = (await post('/api/new', { you, seed: 424242 })) as unknown as View;
  const basic = v.players[0].hand.find((c) => c.stage === 'Basic')!;
  await send({ type: 'SetupActive', player: 0, cardId: basic.id });
  await send({ type: 'SetupReady', player: 0 });
  const mine = (s: View) => s.phase === 'main' && !s.pending && s.active === 0;
  let s = await bot();
  for (let i = 0; i < 60 && !mine(s); i++) s = await bot();
  s = await send({ type: 'EndTurn', player: 0 });
  for (let i = 0; i < 60 && !mine(s); i++) s = await bot();
  return log;
}

test('a fixed seed resumes after a complete bot turn', async ({ page, request }) => {
  await startRun(page);
  const you = await page.evaluate((key) => JSON.parse(localStorage.getItem(key)!).state.run.deck as Record<string, number>, KEY).then(expandDeck);
  const log = await fixedSeedLog(request, you);
  const before = await board(request);
  await page.evaluate(([key, saved]) => {
    const all = JSON.parse(localStorage.getItem(key)!);
    all.state.run.inProgress = saved;
    localStorage.setItem(key, JSON.stringify(all));
  }, [KEY, { seed: 424242, yourDeck: you, log, leftAt: null }] as const);
  await clobberServerGame(request);
  await page.reload();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect.poll(() => board(request)).toEqual(before);
});
