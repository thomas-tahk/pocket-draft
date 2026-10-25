import { test, expect } from '@playwright/test';

test('fresh app shows Draft, no Continue run; Practice disabled', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('button', { name: 'Start draft' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Continue run' })).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Practice' })).toBeDisabled();
  await page.screenshot({ path: 'test-results/opening-fresh.png' });
});

test('active run: reload shows Continue run, leads to deck view', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await expect(page.getByRole('button', { name: '▶ Play deck' })).toBeVisible();
  await page.reload();
  const cont = page.getByRole('button', { name: 'Continue run' });
  await expect(cont).toBeVisible();
  await expect(page.getByRole('button', { name: 'Practice' })).toBeEnabled();
  await page.screenshot({ path: 'test-results/opening-active.png' });
  await cont.click();
  await expect(page.getByRole('button', { name: '▶ Play deck' })).toBeVisible();
});

test('Past runs opens history', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Past runs' }).click();
  await expect(page.getByRole('heading', { name: 'Draft history' })).toBeVisible();
});

type Saved = { state: { run: { battles: unknown[]; status: string } | null; history: { run?: { battles: unknown[]; status: string } }[] } };
const saved = (page: import('@playwright/test').Page) =>
  page.evaluate(() => JSON.parse(localStorage.getItem('pocket-draft-v0.4')!) as Saved);

test('Practice does not record into the run', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await page.getByRole('button', { name: '▶ Play deck' }).waitFor();
  await page.reload();
  await page.getByRole('button', { name: 'Practice' }).click();
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();
  await expect(page.getByTestId('end-screen')).toBeVisible();
  expect((await saved(page)).state.run?.battles).toHaveLength(0);
});

test('Start a new draft instead archives the run as quit with its battle', async ({ page }) => {
  const deck = { a: 20 };
  const battle = { result: 'win', seed: 1, yourDeck: [], botDeck: [], actions: [], endedAt: '2026-01-01T00:00:00Z' };
  await page.addInitScript(
    ([d, b]) => {
      if (localStorage.getItem('pocket-draft-v0.4')) return;
      const entry = { id: 'h1', completedAt: '2026-01-01T00:00:00Z', pickIds: [], shopPurchasedIds: [], deck: d, notes: '', wins: 0, losses: 0 };
      const run = { deck: d, battles: [b], status: 'active' };
      localStorage.setItem(
        'pocket-draft-v0.4',
        JSON.stringify({ version: 2, state: { pickIds: Array(20).fill('a'), offerIds: [], packOffers: [], shopPurchasedIds: [], shopFinalized: true, deck: d, deckFinalized: true, history: [entry], run } }),
      );
    },
    [deck, battle] as const,
  );
  await page.goto('/');
  await page.getByRole('button', { name: 'Start a new draft instead' }).click();
  await expect(page.getByText('This ends your current run (1–0).')).toBeVisible();
  await page.getByRole('button', { name: 'Keep run' }).click();
  await expect(page.getByRole('button', { name: 'Continue run' })).toBeVisible();
  await page.getByRole('button', { name: 'Start a new draft instead' }).click();
  await page.getByRole('button', { name: 'End run and draft' }).click();
  await expect(page.getByRole('button', { name: 'Continue run' })).toHaveCount(0);
  const { state } = await saved(page);
  expect(state.history[0].run?.status).toBe('quit');
  expect(state.history[0].run?.battles).toHaveLength(1);
});
