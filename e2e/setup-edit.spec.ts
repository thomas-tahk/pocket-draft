import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';

// Owner 2026-10-19 (8) at phone size: before Ready a placed card goes back to
// the hand or moves to another spot, and the opponent's placements show face
// down while you are still setting up (the bot places before your Ready).

type Card = { id: string; name: string; stage: string };
type View = { phase: string; players: { hand: Card[]; ready: boolean }[] };

const basicsOf = (v: View) => v.players[0].hand.filter((c) => c.stage === 'Basic');

async function startWithTwoBasics(api: APIRequestContext): Promise<View> {
  for (let seed = 1; seed < 60; seed++) {
    const v = (await (await api.post('/api/new', { data: { seed, preset: 'fire' } })).json()) as View;
    if (basicsOf(v).length >= 2) return v;
  }
  throw new Error('no seed deals two Basics');
}

async function centre(l: Locator) {
  const b = (await l.boundingBox())!;
  return { x: b.x + b.width / 2, y: b.y + b.height / 2 };
}

async function dragTo(page: Page, from: Locator, to: Locator, edge = false) {
  const b = (await from.boundingBox())!;
  const t = await centre(to);
  await page.mouse.move(edge ? b.x + Math.min(10, b.width / 2) : b.x + b.width / 2, b.y + b.height / 2);
  await page.mouse.down();
  await page.mouse.move(t.x, t.y, { steps: 8 });
  await page.mouse.up();
}

const handCard = (page: Page, id: string) => page.getByTestId('my-hand').locator(`[data-card-id="${id}"]`).first();
const slot = (page: Page, name: string) => page.getByTestId('my-bench').locator(`[data-slot="${name}"]`);

test('a placed card goes back to the hand, and moves between spots, before Ready', async ({ page, request }) => {
  const v = await startWithTwoBasics(request);
  const [a, b] = basicsOf(v);
  await page.goto('/?battle');
  await expect(page.getByTestId('setup-intro')).toBeVisible();
  await expect(page.getByTestId('setup-intro')).toHaveCount(0, { timeout: 6000 });
  const active = page.getByTestId('my-active').locator('[data-spot="active"]');
  await dragTo(page, handCard(page, a.id), active, true);
  await expect(page.getByTestId('my-active').locator('img')).toHaveAttribute('alt', a.name); // the hand has re-fanned
  await dragTo(page, handCard(page, b.id), slot(page, 'left'), true);
  await expect(slot(page, 'left').locator('img')).toHaveAttribute('alt', b.name);

  // Drag the bench card back onto the hand: it leaves the Bench and is playable again.
  await dragTo(page, slot(page, 'left'), page.getByTestId('my-hand'));
  await expect(slot(page, 'left').locator('img')).toHaveCount(0);
  await expect(handCard(page, b.id)).toBeVisible();

  // Place it again, then move it to another Bench slot.
  await dragTo(page, handCard(page, b.id), slot(page, 'left'), true);
  await dragTo(page, slot(page, 'left'), slot(page, 'right'));
  await expect(slot(page, 'right').locator('img')).toHaveAttribute('alt', b.name);
  await expect(slot(page, 'left').locator('img')).toHaveCount(0);

  // Move the Active onto the Bench slot holding b: they swap.
  await dragTo(page, active, slot(page, 'right'));
  await expect(page.getByTestId('my-active').locator('img')).toHaveAttribute('alt', b.name);
  await expect(slot(page, 'right').locator('img')).toHaveAttribute('alt', a.name);
  await page.screenshot({ path: 'test-results/setup-edit-390x844.png' });

  // After Ready nothing can be picked up.
  await page.getByTestId('setup-ready').click();
  const state = (await (await request.get('/api/state')).json()) as View;
  expect(state.players[0].ready).toBe(true);
});

test("the opponent's placed cards show face down before your Ready", async ({ page, request }) => {
  await startWithTwoBasics(request);
  await page.goto('/?battle');

  // The bot has placed an Active (face down, no name) while you have placed nothing.
  await expect(page.getByTestId('opp-active').getByTestId('card-back')).toHaveCount(1, { timeout: 10_000 });
  await expect(page.getByTestId('opp-active').locator('img')).toHaveCount(0);
  const state = (await (await request.get('/api/state')).json()) as View;
  expect(state.phase).toBe('setup');
  expect(state.players[0].ready).toBe(false);
});
