import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

// L5: no permanent turn indicator. Each turn start flashes YOUR TURN or
// OPPONENT'S TURN in different colours; cards you can use right now pulse, and
// "can use" means exactly: some legal move from the server names that card.

type Card = { id: string; stage: string };
type Move = { type: string; cardId?: string; handCardId?: string; benchIndex?: number };
type View = { phase: string; active: number; turn: number; pending: { player: number } | null; players: { hand: Card[]; legalMoves: Move[] }[] };

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

// Mostly Stage 1s, so the first-turn hand mixes playable Basics with
// Evolutions that can't be used yet. Seed 11: you go first, holding both.
const DECK = [
  ...['B2b-007', 'B4a-095', 'A4b-057'], // Charmander, Vulpix, Charmeleon
  ...['B4a-011', 'B4-021', 'B3-027', 'B3-207', 'B2a-017', 'B2-021', 'B2-025'],
].flatMap((id) => [id, id]);

// Setup through the API, ending on your first main turn.
async function setupToYourFirstTurn(api: APIRequestContext) {
  const start = await post(api, '/api/new', { seed: 11, you: DECK });
  const basics = start.players[0].hand.filter((c) => c.stage === 'Basic');
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basics[0].id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
}

// Turn changes go through the screen's own game store, the path End Turn and
// the bot driver use, so the banner sees a real change rather than a reload.
async function storeCall(page: Page, call: 'endTurn' | 'runBot') {
  return page.evaluate(async (c) => {
    const { useGameStore } = await import('/src/game/store.ts');
    const s = useGameStore.getState();
    if (c === 'endTurn') await s.dispatch({ type: 'EndTurn', player: 0 });
    else await s.runBot();
    const v = useGameStore.getState().view!;
    return { active: v.active, phase: v.phase, pending: v.pending };
  }, call);
}

// Every card of yours pulses iff a legal move names it; returns how many pulse.
async function expectPulseMatchesLegal(page: Page, api: APIRequestContext) {
  const state = (await (await api.get('/api/state')).json()) as View;
  const me = state.players[0];
  const legal = new Set(me.legalMoves.flatMap((m) => [m.cardId, m.handCardId]).filter(Boolean));
  const hand = page.getByTestId('my-hand').getByTestId('hand-card');
  await expect(hand).toHaveCount(me.hand.length);
  let pulsing = 0;
  for (let i = 0; i < me.hand.length; i++) {
    const usable = legal.has(me.hand[i].id);
    if (usable) pulsing++;
    await expect(hand.nth(i)).toHaveAttribute('data-usable', String(usable));
  }
  const activeUsable = me.legalMoves.some((m) => m.type === 'UseAttack' || m.type === 'Retreat');
  await expect(page.getByTestId('my-active').getByTestId('active-card')).toHaveAttribute('data-usable', String(activeUsable));
  return { pulsing, handSize: me.hand.length };
}

async function expectBanner(page: Page, side: 'you' | 'opponent', text: string) {
  const banner = page.getByTestId('turn-banner');
  await expect(banner).toBeVisible();
  await expect(banner).toHaveText(text);
  await expect(banner).toHaveAttribute('data-side', side);
  await expect(banner).toHaveCSS('opacity', '1'); // fully in, before the fade out
  return (await banner.evaluate((el) => getComputedStyle(el).backgroundImage)) as string;
}

test('turn banner flashes on each turn change; only legal cards pulse', async ({ page, request }) => {
  await setupToYourFirstTurn(request);
  const state = (await (await request.get('/api/state')).json()) as View;
  expect(state.phase).toBe('main');
  expect(state.active).toBe(0);
  await page.goto('/?battle');

  const yourColour = await expectBanner(page, 'you', 'YOUR TURN');
  await page.screenshot({ path: 'test-results/turn-banner-yours-390x844.png' });
  // A flash, not a permanent indicator.
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });

  // Pulse: only cards a legal move names; here a mix of both.
  const { pulsing, handSize } = await expectPulseMatchesLegal(page, request);
  expect(pulsing, 'some hand card is playable on your first turn').toBeGreaterThan(0);
  expect(pulsing, 'some hand card is not playable').toBeLessThan(handSize);
  const pulse = await page.locator('[data-usable="true"]').first().evaluate((el) => getComputedStyle(el).animationName);
  expect(pulse).not.toBe('none');
  const still = await page.locator('[data-usable="false"]').first().evaluate((el) => getComputedStyle(el).animationName);
  expect(still).toBe('none');
  // End your turn: the opponent's banner, in a different colour, and nothing of yours pulses.
  await storeCall(page, 'endTurn');
  const theirColour = await expectBanner(page, 'opponent', "OPPONENT'S TURN");
  await page.screenshot({ path: 'test-results/turn-banner-opponent-390x844.png' });
  expect(theirColour).not.toBe(yourColour);
  await expect(page.locator('[data-usable="true"]')).toHaveCount(0);
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });

  // The bot plays out its turn; yours starts again with a fresh YOUR TURN flash.
  for (let i = 0; i < 40; i++) {
    const v = await storeCall(page, 'runBot');
    if (v.phase !== 'main' || (v.active === 0 && !v.pending)) break;
  }
  await expectBanner(page, 'you', 'YOUR TURN');
  await expectPulseMatchesLegal(page, request);
});
