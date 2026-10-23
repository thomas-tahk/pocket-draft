import { test, expect, type APIRequestContext, type Page } from '@playwright/test';
import { FIRE } from './decks';

// M3-2: when the board changes, the cards that moved slide from their old
// place (an Animation on the card, at most 400 ms); reduced motion turns it
// off; Skip on the bot's turn still reaches the end of it.

type Card = { id: string; name: string; stage: string };
type View = { phase: string; active: number; turn: number; pending: { player: number } | null; players: { hand: Card[] }[] };

const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

// Seed 7: you go first; your second turn has a Basic in hand and an empty Bench.
async function toYourSecondTurn(api: APIRequestContext): Promise<Card> {
  const start = await post(api, '/api/new', { seed: 7, bot: BOT, you: FIRE });
  const basics = start.players[0].hand.filter((c) => c.stage === 'Basic');
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basics[0].id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  await post(api, '/api/move', { type: 'EndTurn', player: 0 });
  for (let i = 0; i < 40; i++) {
    const v = await post(api, '/api/bot');
    if (v.phase !== 'main' || (v.pending ? v.pending.player : v.active) === 0) break;
  }
  return basics[1];
}

async function centre(page: Page, testId: string) {
  const b = (await page.getByTestId(testId).boundingBox())!;
  return { x: b.x + b.width / 2, y: b.y + b.height / 2 };
}

// Drags the first Basic in hand onto the left Bench slot.
async function dropBasic(page: Page, card: Card) {
  const hand = (await page.getByTestId('my-hand').locator(`[data-card-id="${card.id}"]`).first().boundingBox())!;
  const slot = (await page.getByTestId('my-bench').locator('[data-slot="left"]').boundingBox())!;
  await page.mouse.move(hand.x + 10, hand.y + hand.height / 2);
  await page.mouse.down();
  await page.mouse.move(slot.x + slot.width / 2, slot.y + slot.height / 2, { steps: 8 });
  await page.mouse.up();
}

// Milliseconds the left Bench card has an animation running, polled every frame; -1 if it never did.
const benchCardMotion = (page: Page) => page.evaluate(() => new Promise<number>((done) => {
  const running = () => [...document.querySelectorAll('[data-testid="my-bench"] [data-slot="left"] [data-fx]')]
    .some((el) => el.getAnimations().some((a) => a.playState === 'running'));
  const t0 = performance.now();
  let start = -1;
  const tick = () => {
    const now = performance.now();
    if (start < 0 && running()) start = now;
    if (start >= 0 && !running()) return done(now - start);
    if (now - t0 > 3000) return done(start < 0 ? -1 : 9999);
    requestAnimationFrame(tick);
  };
  tick();
}));

test('a Basic dropped on the Bench slides there from the hand, then settles', async ({ page, request }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const basic = await toYourSecondTurn(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
  const motion = benchCardMotion(page);
  await dropBasic(page, basic);
  const ms = await motion;
  expect(ms, 'an animation ran on the Bench card').toBeGreaterThan(0);
  expect(ms, 'and finished quickly').toBeLessThan(450);
});

test.describe('reduced motion', () => {
  test('nothing slides', async ({ page, request }) => {
    const basic = await toYourSecondTurn(request);
    await page.goto('/?battle');
    await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
    const motion = benchCardMotion(page);
    await dropBasic(page, basic);
    expect(await motion).toBe(-1);
    await expect(page.getByTestId('my-bench').locator('[data-slot="left"] img')).toHaveAttribute('alt', basic.name);
  });
});

test('Skip on the bot turn still reaches the end of it', async ({ page, request }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await toYourSecondTurn(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
  await page.getByTestId('end-turn').getByRole('button').click();
  const confirm = page.getByTestId('confirm-ok');
  if (await confirm.isVisible().catch(() => false)) await confirm.click();
  await page.locator('.bt-skip').click({ timeout: 3000 }).catch(() => {}); // the bot may have finished already
  await expect(page.getByTestId('end-turn').getByRole('button')).toBeEnabled({ timeout: 5000 });
  await expect(page.locator('.bt-skip')).toHaveCount(0);
  expect(await page.evaluate(() => document.getAnimations().filter((a) => a.id === 'fx' && a.playState === 'running').length)).toBe(0);
});
