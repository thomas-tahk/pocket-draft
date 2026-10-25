import { test, expect } from '@playwright/test';

// A draft saved before the 16 -> 20 pick rework (persist version 0) cannot
// resume; it must land on the start screen with past drafts intact.

const ids16 = Array.from({ length: 16 }, (_, i) => `old-pick-${i}`);

const v0State = {
  pickIds: ids16,
  offerIds: [],
  packOffers: ids16.map(() => ids16.slice(0, 5)),
  shopPurchasedIds: [],
  shopFinalized: true,
  deck: { 'old-pick-0': 2 },
  deckFinalized: true,
  history: [
    { id: 'kept-1', completedAt: '2026-09-01T00:00:00.000Z', pickIds: ids16, shopPurchasedIds: [], deck: {}, notes: '', wins: 0, losses: 0 },
  ],
};

test('a saved 16-pick review state resets to the start screen and keeps history', async ({ page }) => {
  await page.addInitScript((state) => {
    localStorage.setItem('pocket-draft-v0.4', JSON.stringify({ state, version: 0 }));
  }, v0State);
  await page.goto('/');

  await expect(page.getByRole('button', { name: 'Start draft' })).toBeVisible();
  await expect(page.getByText('View past drafts (1 /')).toBeVisible();
});

// A v1 save (before the run record) with a finalized deck keeps its deck, its
// history, and gains an empty active run.
test('a saved v1 review state gains an empty run and keeps deck and history', async ({ page }) => {
  const v1 = { ...v0State, pickIds: Array.from({ length: 20 }, (_, i) => `p${i}`) };
  await page.addInitScript((state) => {
    localStorage.setItem('pocket-draft-v0.4', JSON.stringify({ state, version: 1 }));
  }, v1);
  await page.goto('/');

  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–0 (active)');
  const saved = await page.evaluate(() => JSON.parse(localStorage.getItem('pocket-draft-v0.4')!));
  expect(saved.version).toBe(2);
  expect(saved.state.history).toHaveLength(1);
  expect(saved.state.deck).toEqual({ 'old-pick-0': 2 });
});
