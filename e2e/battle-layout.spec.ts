import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

// L1–L4 at phone size: every zone of the new battle screen sits where the
// owner's Pocket screenshot puts it. The game is driven through the API to
// your second turn with energy attached, so piles, bench and the spent energy
// circle all have something to show.

type Card = { id: string; stage: string; image?: string };
type View = { phase: string; active: number; pending: { player: number } | null; players: { hand: Card[] }[] };

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

async function runBotUntilYourTurn(api: APIRequestContext) {
  for (let i = 0; i < 40; i++) {
    const v = await post(api, '/api/bot');
    const actor = v.pending ? v.pending.player : v.active;
    if (v.phase !== 'main' || actor === 0) return;
  }
}

async function playToSecondTurn(api: APIRequestContext) {
  const start = await post(api, '/api/new', { seed: 7, preset: 'fire' });
  const basics = start.players[0].hand.filter((c) => c.stage === 'Basic');
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basics[0].id });
  await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: basics[1].id, slot: 0 });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  // the bot places its setup one card per call, then presses Ready
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  await post(api, '/api/move', { type: 'EndTurn', player: 0 });
  await runBotUntilYourTurn(api);
  await post(api, '/api/move', { type: 'AttachEnergy', player: 0, target: 0 });
}

async function box(page: Page, testId: string) {
  const b = await page.getByTestId(testId).boundingBox();
  expect(b, `${testId} is on screen`).not.toBeNull();
  return { ...b!, right: b!.x + b!.width, bottom: b!.y + b!.height, cx: b!.x + b!.width / 2, cy: b!.y + b!.height / 2 };
}

test('battle screen: every zone where L1 puts it', async ({ page, request }) => {
  await playToSecondTurn(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  const W = 390;
  const H = 844;

  const oppPoints = await box(page, 'opp-points');
  const oppHand = await box(page, 'opp-hand');
  const oppEnergy = await box(page, 'opp-energy');
  const oppBench = await box(page, 'opp-bench');
  const oppActive = await box(page, 'opp-active');
  const oppDeck = await box(page, 'opp-deck');
  const oppDiscard = await box(page, 'opp-discard');
  const endTurn = await box(page, 'end-turn');
  const stadium = await box(page, 'stadium');
  const myActive = await box(page, 'my-active');
  const myDeck = await box(page, 'my-deck');
  const myDiscard = await box(page, 'my-discard');
  const myBench = await box(page, 'my-bench');
  const myHand = await box(page, 'my-hand');
  const myEnergy = await box(page, 'my-energy');
  const myPoints = await box(page, 'my-points');

  // Everything fits the phone screen.
  for (const b of [oppPoints, oppHand, oppEnergy, oppBench, oppActive, oppDeck, oppDiscard, endTurn, stadium, myActive, myDeck, myDiscard, myBench, myHand, myEnergy, myPoints]) {
    expect(b.x).toBeGreaterThanOrEqual(0);
    expect(b.right).toBeLessThanOrEqual(W + 1);
    expect(b.y).toBeGreaterThanOrEqual(0);
    expect(b.bottom).toBeLessThanOrEqual(H + 1);
  }

  // Opponent on top, mirrored: hand backs, then Bench, then Active; you below: Active, Bench, hand.
  expect(oppHand.bottom).toBeLessThanOrEqual(oppBench.y + 12);
  expect(oppBench.bottom).toBeLessThanOrEqual(oppActive.y + 12);
  expect(oppActive.bottom).toBeLessThanOrEqual(myActive.y);
  expect(myActive.bottom).toBeLessThanOrEqual(myBench.y + 12);
  expect(myBench.bottom).toBeLessThanOrEqual(myHand.y + 24);
  expect(myHand.bottom).toBeGreaterThan(H * 0.9);
  // Actives face each other in the middle.
  for (const a of [oppActive, myActive]) expect(Math.abs(a.cx - W / 2)).toBeLessThan(30);
  expect(oppActive.bottom).toBeGreaterThan(H * 0.35);
  expect(myActive.y).toBeLessThan(H * 0.6);

  // Points: opponent top right, you bottom left.
  expect(oppPoints.x).toBeGreaterThan(W / 2);
  expect(oppPoints.y).toBeLessThan(H * 0.1);
  expect(myPoints.right).toBeLessThan(W / 2);
  expect(myPoints.bottom).toBeGreaterThan(H * 0.95);
  await expect(page.getByTestId('opp-points').getByTestId('pip')).toHaveCount(3);
  await expect(page.getByTestId('my-points').getByTestId('pip')).toHaveCount(3);

  // Your deck, discard under it, right of your Active and Bench; opponent's mirrored on the left.
  expect(myDeck.x).toBeGreaterThan(myActive.right);
  expect(myDeck.x).toBeGreaterThan(myBench.right - 70); // beside the bench, off to the right
  expect(myDiscard.y).toBeGreaterThanOrEqual(myDeck.bottom - 1);
  expect(oppDeck.right).toBeLessThan(oppActive.x);
  expect(oppDiscard.bottom).toBeLessThanOrEqual(oppDeck.y + 1);

  // Energy Zone: yours right of the hand, a little below the Bench; the opponent's top left.
  expect(myEnergy.x).toBeGreaterThan(myHand.right - 40);
  expect(myEnergy.cy).toBeGreaterThan(myBench.bottom - 10);
  expect(oppEnergy.cx).toBeLessThan(W / 2);
  expect(oppEnergy.cy).toBeLessThan(H * 0.25);
  // Energy was attached this turn: the current-energy spot is an empty grey circle.
  await expect(page.getByTestId('my-energy').getByTestId('energy-current')).toHaveAttribute('data-spent', 'true');
  await expect(page.getByTestId('my-energy').getByTestId('energy-next')).toHaveAttribute('data-spent', 'false');

  // End Turn at the middle of the screen, right of centre.
  expect(endTurn.cy).toBeGreaterThan(H * 0.4);
  expect(endTurn.cy).toBeLessThan(H * 0.6);
  expect(endTurn.cx).toBeGreaterThan(W / 2);

  // Stadium at the left edge.
  expect(stadium.x).toBeLessThan(W * 0.1);

  // Benches are three slots each; the behind slot sits directly behind the Active.
  await expect(page.getByTestId('my-bench').getByTestId('bench-slot')).toHaveCount(3);
  await expect(page.getByTestId('opp-bench').getByTestId('bench-slot')).toHaveCount(3);
  const behind = await page.getByTestId('my-bench').locator('[data-slot="behind"]').boundingBox();
  expect(Math.abs(behind!.x + behind!.width / 2 - myActive.cx)).toBeLessThan(8);

  // Opponent's hand is card backs, one per card; yours shows real faces.
  const oppHandCount = Number(await page.getByTestId('opp-hand').getAttribute('data-count'));
  expect(oppHandCount).toBeGreaterThan(0);
  await expect(page.getByTestId('opp-hand').getByTestId('card-back')).toHaveCount(oppHandCount);
  await expect(page.getByTestId('my-hand').locator('img').first()).toHaveAttribute('src', /limitlesstcg/);

  // L2: the Active shows its real card image, HP number, and attached energy icons.
  const active = page.getByTestId('my-active');
  await expect(active.locator('img')).toHaveAttribute('src', /limitlesstcg/);
  await expect(active.getByTestId('hp')).toHaveText(/^\d+$/);
  await expect(active.getByTestId('energy-icon')).toHaveCount(1);

  // L4: counts visible.
  await expect(page.getByTestId('my-deck')).toContainText(/\d+/);
  await expect(page.getByTestId('opp-deck')).toContainText(/\d+/);
  await expect(page.getByTestId('opp-hand')).toContainText(String(oppHandCount));

  // Card images come from the card CDN; wait for them (a failed load also counts as complete).
  await page.waitForFunction(() => [...document.images].every((i) => i.complete));
  await page.screenshot({ path: 'test-results/battle-board-390x844.png' });
});
