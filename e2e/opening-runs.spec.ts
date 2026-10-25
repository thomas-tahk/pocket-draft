import { test, expect, type Page } from '@playwright/test';

type Saved = { state: { history: { run?: { status: string } }[] } };
const saved = (page: Page) =>
  page.evaluate(() => JSON.parse(localStorage.getItem('pocket-draft-v0.4')!) as Saved);

test('a lost run is reachable after reload via View last run', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  for (let i = 0; i < 2; i++) {
    await page.getByRole('button', { name: '▶ Play deck' }).click();
    await expect(page.getByTestId('battle-board')).toBeVisible();
    await page.getByTestId('menu').click();
    await page.getByTestId('menu-concede').click();
    await page.getByTestId('confirm-ok').click();
    await expect(page.getByTestId('end-result')).toHaveText('Loss');
    await page.getByRole('button', { name: 'Back' }).click();
  }
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–2 (lost)');
  await page.reload();
  await expect(page.getByRole('button', { name: 'Continue run' })).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Start draft' })).toBeVisible();
  await page.getByRole('button', { name: 'View last run' }).click();
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–2 (lost)');
});

test('a run is archived by its history id even if another entry has the same deck', async ({ page }) => {
  const deck = { a: 20 };
  await page.addInitScript(
    (d) => {
      if (localStorage.getItem('pocket-draft-v0.4')) return;
      const mk = (id: string) => ({ id, completedAt: '2026-01-01T00:00:00Z', pickIds: [], shopPurchasedIds: [], deck: d, notes: '', wins: 0, losses: 0 });
      const run = { deck: d, battles: [], status: 'active', historyId: 'older' };
      localStorage.setItem(
        'pocket-draft-v0.4',
        JSON.stringify({ version: 2, state: { pickIds: Array(20).fill('a'), offerIds: [], packOffers: [], shopPurchasedIds: [], shopFinalized: true, deck: d, deckFinalized: true, history: [mk('newer'), mk('older')], run } }),
      );
    },
    deck,
  );
  await page.goto('/');
  await page.getByRole('button', { name: 'Start a new draft instead' }).click();
  await page.getByRole('button', { name: 'End run and draft' }).click();
  const { state } = await saved(page);
  expect(state.history[0].run).toBeUndefined();
  expect(state.history[1].run?.status).toBe('quit');
});
