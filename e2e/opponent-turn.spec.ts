import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

// F14 at phone size: the bot's turn plays out one action at a time, about
// 0.5 s each (the target glows, numbers change), with no panel left over the
// board, and Skip jumps to the end of its whole turn.

type Card = { id: string; stage: string };
type View = { phase: string; active: number; turn: number; pending: { player: number } | null; players: { hand: Card[] }[] };

// Pinned so draft-pool changes can't shift the bot's play.
const BOT = ['A3-005', 'A4-101', 'A1-065', 'A3b-044', 'B1a-044', 'A2b-063', 'A1-155', 'A1-218', 'B4-031', 'A3-145', 'A3-005', 'A1-010', 'B1-123', 'P-A-006', 'B2-047', 'A4-122', 'A4-101', 'A1-010', 'A4-122', 'A1-065'];

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

const actor = (v: View) => (v.pending ? v.pending.player : v.active);

// Seed 1: you go first; on turn 4 the bot attaches energy, then its
// Tyrogue attacks your Active (the attack ends its turn). Everything up to
// your turn 3 is played through the API. On that turn it benches a Basic first.
async function reachTurn3(api: APIRequestContext) {
  const start = await post(api, '/api/new', { seed: 1, bot: BOT, preset: 'fire' });
  const basic = start.players[0].hand.find((c) => c.stage === 'Basic')!;
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basic.id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  let v = await post(api, '/api/move', { type: 'EndTurn', player: 0 });
  for (let i = 0; i < 10 && actor(v) === 1; i++) v = await post(api, '/api/bot');
  expect(v.turn).toBe(3);
}

// Records, in page time, every playback step shown and when End Turn comes back.
async function watchPlayback(page: Page) {
  await page.evaluate(() => {
    const w = window as unknown as { pbLog: { t: number; step: string; kind: string; zone: string }[] };
    w.pbLog = [];
    const root = document.querySelector('.bt-root')!;
    new MutationObserver(() => {
      const step = root.getAttribute('data-bot-step') ?? '';
      const last = w.pbLog[w.pbLog.length - 1];
      if (step && step !== last?.step)
        w.pbLog.push({ t: performance.now(), step, kind: root.getAttribute('data-bot-kind') ?? '', zone: root.getAttribute('data-flash-zone') ?? '' });
    }).observe(root, { subtree: true, attributes: true, childList: true });
  });
}

async function endTurn3(page: Page) {
  await page.getByTestId('end-turn').getByRole('button').click();
  await page.getByTestId('confirm-ok').click(); // energy is waiting on turn 3
  return page.evaluate(() => performance.now());
}

test('the bot turn plays one action about every half second', async ({ page, request }) => {
  await reachTurn3(request);
  await page.goto('/?battle');
  const myHp = page.getByTestId('my-active').getByTestId('hp');
  const hpBefore = Number(await myHp.textContent());
  await watchPlayback(page);

  const t0 = await endTurn3(page);
  const root = page.locator('.bt-root');
  await expect(root).toHaveAttribute('data-bot-kind', 'attack');
  await expect(page.getByTestId('playback')).toHaveCount(0);
  await expect(page.locator('.bt-root')).toHaveAttribute('data-flash-zone', 'my-active');
  await expect(myHp).toHaveText(String(hpBefore - 20));
  await page.screenshot({ path: 'test-results/opponent-turn-390x844.png' });

  // The bot benches a Basic, attaches Energy to its Active, then attacks.
  const log = await page.evaluate(() => (window as unknown as { pbLog: { t: number; kind: string; zone: string }[] }).pbLog);
  expect(log.map((e) => e.kind)).toEqual(['play_basic', 'attach_energy', 'attack']);
  expect(log[1].zone).toBe('opp-active');
  const gaps = [log[0].t - t0, log[1].t - log[0].t, log[2].t - log[1].t];
  for (const gap of gaps) {
    expect(gap).toBeGreaterThan(400);
    expect(gap).toBeLessThan(1500);
  }
  // The last action stays lit for its half second, then the board is yours.
  await expect(root).not.toHaveAttribute('data-bot-step', /./, { timeout: 2000 });
  expect((await (await request.get('/api/state')).json()).turn).toBe(5);
});

test('Skip jumps to the end of the bot turn', async ({ page, request }) => {
  await reachTurn3(request);
  await page.goto('/?battle');
  await watchPlayback(page);

  await endTurn3(page);
  const skip = page.getByRole('button', { name: 'Skip' });
  await skip.click();
  const tSkip = await page.evaluate(() => performance.now());

  await expect.poll(async () => (await (await request.get('/api/state')).json()).turn, { timeout: 2000 }).toBe(5);
  const tDone = await page.evaluate(() => performance.now());
  expect(tDone - tSkip).toBeLessThan(500);
  await expect(page.locator('.bt-root')).not.toHaveAttribute('data-bot-step', /./);
  await expect(skip).toBeHidden();
  await expect(page.getByTestId('end-turn').getByRole('button')).toBeEnabled();
  // Nothing past the moment of Skip was played back.
  const log = await page.evaluate(() => (window as unknown as { pbLog: { kind: string }[] }).pbLog);
  expect(log.map((e) => e.kind)).not.toContain('attack');
});
