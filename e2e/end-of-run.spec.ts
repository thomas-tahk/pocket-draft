import { test, expect, type Page } from '@playwright/test';

// #118: a run has a real ending, kept in Past runs.

const KEY = 'pocket-draft-v0.4';
const entry = (id: string, extra: object = {}) => ({
  id, completedAt: '2026-01-01T00:00:00Z', pickIds: [], shopPurchasedIds: [], deck: { a: 20 }, notes: '', wins: 3, losses: 1, ...extra,
});

async function seed(page: Page, state: object) {
  await page.addInitScript(
    ([k, s]) => { if (!localStorage.getItem(k as string)) localStorage.setItem(k as string, JSON.stringify({ version: 2, state: s })); },
    [KEY, state],
  );
}

test('two concessions end the run on a summary; Draft again keeps it in Past runs as 0–2, read-only', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–0 (active)');
  // Auto-draft makes no history entry; give the run one so it can be archived.
  await page.evaluate((k) => {
    const saved = JSON.parse(localStorage.getItem(k)!);
    const s = saved.state;
    s.history = [{ id: 'mine', completedAt: '2026-01-01T00:00:00Z', pickIds: [], shopPurchasedIds: [], deck: s.deck, notes: '', wins: 0, losses: 0 }];
    s.run.historyId = 'mine';
    localStorage.setItem(k, JSON.stringify(saved));
  }, KEY);
  await page.reload();
  await page.getByRole('button', { name: 'Continue run' }).click();
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–0 (active)');
  for (let i = 0; i < 2; i++) {
    await page.getByRole('button', { name: 'Next battle' }).click();
    await expect(page.getByTestId('battle-board')).toBeVisible();
    await page.getByTestId('menu').click();
    await page.getByTestId('menu-concede').click();
    await page.getByTestId('confirm-ok').click();
    await expect(page.getByTestId('end-result')).toHaveText('Loss');
    await page.getByRole('button', { name: 'Back' }).click();
  }
  await expect(page.getByTestId('run-status')).toHaveText('Run lost');
  const rows = page.getByTestId('run-battles').locator('li');
  await expect(rows).toHaveCount(2);
  await expect(rows.nth(0)).toContainText('Battle 1: Loss');
  await expect(rows.nth(1)).toContainText('Battle 2: Loss');
  await page.screenshot({ path: 'test-results/end-of-run-summary.png', fullPage: true });

  await rows.nth(0).getByRole('button', { name: 'Replay' }).click();
  await page.getByTestId('replay-exit').click();
  await expect(page.getByTestId('run-status')).toHaveText('Run lost');

  await page.getByRole('button', { name: 'Draft again' }).click();
  page.once('dialog', (d) => d.accept());
  await page.getByRole('button', { name: 'Cancel draft' }).click();
  await page.getByRole('button', { name: /Past runs/ }).click();
  await expect(page.getByText('0W')).toBeVisible();
  await expect(page.getByTestId('history-run-status')).toHaveText('Run lost');
  await page.getByRole('button', { name: 'View' }).click();
  await expect(page.getByText('Record 0–2, from the run.')).toBeVisible();
  await expect(page.getByRole('spinbutton')).toHaveCount(0);
});

test('an old history entry without a run keeps editable W/L', async ({ page }) => {
  await seed(page, {
    pickIds: [], offerIds: [], packOffers: [], shopPurchasedIds: [], shopFinalized: false,
    deck: {}, deckFinalized: false, history: [entry('old')], run: null,
  });
  await page.goto('/');
  await page.getByRole('button', { name: /Past runs/ }).click();
  await page.getByRole('button', { name: 'View' }).click();
  const wins = page.getByLabel('Wins:');
  await expect(wins).toHaveValue('3');
  await wins.fill('4');
  await expect(page.getByText('4W')).toBeVisible();
});
