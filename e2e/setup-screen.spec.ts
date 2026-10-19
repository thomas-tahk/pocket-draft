import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';

// S4, S5, F11 at phone size: a new game opens on the coin flip (it lands by
// itself whoever flips, 2.11), then "You go first / second",
// then setup on the board: drag a Basic to the Active spot and one to the
// Bench, press Ready, and the game starts on turn 1. Both first energies show
// during setup; the opponent's picks don't.

type Card = { id: string; name: string; stage: string };
type View = { phase: string; turn: number; firstPlayer: number; openingFlip: { flipper: number; heads: boolean }; players: { hand: Card[] }[] };

const basicsOf = (v: View) => v.players[0].hand.filter((c) => c.stage === 'Basic');

// The first seed whose opening flip is made by `flipper` and deals you two Basics.
async function newGameFlippedBy(api: APIRequestContext, flipper: number): Promise<View> {
  for (let seed = 1; seed < 60; seed++) {
    const v = (await (await api.post('/api/new', { data: { seed } })).json()) as View;
    if (v.openingFlip.flipper === flipper && basicsOf(v).length >= 2) return v;
  }
  throw new Error(`no seed where player ${flipper} flips with two Basics`);
}

async function centre(l: Locator) {
  const b = (await l.boundingBox())!;
  return { x: b.x + b.width / 2, y: b.y + b.height / 2 };
}

// Hand cards overlap to the right, so grab the visible left edge.
async function drag(page: Page, card: Locator, to: Locator) {
  const b = (await card.boundingBox())!;
  const t = await centre(to);
  await page.mouse.move(b.x + Math.min(10, b.width / 2), b.y + b.height / 2);
  await page.mouse.down();
  await page.mouse.move(t.x, t.y, { steps: 8 });
  await page.mouse.up();
}

async function expectCoinAndOrder(page: Page, v: View) {
  await expect(page.getByTestId('setup-coin')).toHaveAttribute('data-face', v.openingFlip.heads ? 'heads' : 'tails');
  await expect(page.getByTestId('setup-order')).toHaveText(v.firstPlayer === 0 ? 'You go first' : 'You go second');
  await page.screenshot({ path: 'test-results/setup-coin-390x844.png' });
  await expect(page.getByTestId('setup-intro')).toHaveCount(0, { timeout: 4000 });
}

test('you flip, place Active and Bench, Ready, game starts on turn 1', async ({ page, request }) => {
  const v = await newGameFlippedBy(request, 0);
  const [a, b] = basicsOf(v);
  await page.goto('/?battle');

  // 2.11: the opening coin never waits, even when it is yours to flip.
  await expect(page.getByTestId('setup-intro')).toContainText('You flip the coin');
  await expect(page.getByTestId('setup-coin')).toBeDisabled();
  await expectCoinAndOrder(page, v);

  // S4: both sides' first energy shows from setup on.
  for (const side of ['my-energy', 'opp-energy']) {
    await expect(page.getByTestId(side).getByTestId('energy-next')).toHaveAttribute('data-spent', 'false');
  }
  const ready = page.getByTestId('setup-ready');
  await expect(ready).toBeDisabled(); // no Active yet

  // Drag a Basic to the Active spot: it glows, the hint names it.
  const active = page.getByTestId('my-active').locator('[data-spot="active"]');
  const hand = page.getByTestId('my-hand');
  await drag(page, hand.locator(`[data-card-id="${a.id}"]`).first(), active);
  await expect(page.getByTestId('my-active').locator('img')).toHaveAttribute('alt', a.name);

  // Then one to the Bench.
  const left = page.getByTestId('my-bench').locator('[data-slot="left"]');
  await drag(page, hand.locator(`[data-card-id="${b.id}"]`).first(), left);
  await expect(left.locator('img')).toHaveAttribute('alt', b.name);
  await expect(page.getByTestId('opp-active').locator('img')).toHaveCount(0); // S5: hidden until both Ready
  await page.screenshot({ path: 'test-results/setup-placed-390x844.png' });

  await ready.click();

  // The screen runs the bot's setup; the game starts on turn 1 with both Actives shown.
  await expect(page.getByTestId('turn-banner')).toBeVisible({ timeout: 15_000 });
  await expect(page.getByTestId('turn-banner')).toHaveAttribute('data-side', v.firstPlayer === 0 ? 'you' : 'opponent');
  await expect(page.getByTestId('opp-active').locator('img')).toHaveCount(1);
  const state = (await (await request.get('/api/state')).json()) as View;
  expect(state.phase).toBe('main');
  expect(state.turn).toBeGreaterThanOrEqual(1);
});

test("the opponent's flip lands on its own", async ({ page, request }) => {
  const v = await newGameFlippedBy(request, 1);
  await page.goto('/?battle');
  await expect(page.getByTestId('setup-intro')).toContainText('Opponent flips the coin');
  await expect(page.getByTestId('setup-coin')).toBeDisabled();
  await expectCoinAndOrder(page, v);
  await expect(page.getByTestId('setup-ready')).toBeVisible();
});
