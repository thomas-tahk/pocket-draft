import { test, expect, type Page } from '@playwright/test';
import { draftRealDeck } from './helpers/draft';

// #118: a run has a real ending, kept in Past runs.

const entry = (id: string) => ({
  id, completedAt: '2026-01-01T00:00:00Z', pickIds: [], shopPurchasedIds: [], deck: { a: 20 }, notes: '', wins: 3, losses: 1,
});

// A real draft, then two conceded battles through the UI: the run is lost.
async function loseARun(page: Page) {
  await page.goto('/');
  await draftRealDeck(page);
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
}

async function expectLostRowReadOnly(page: Page) {
  await page.getByRole('button', { name: /Past runs/ }).click();
  await expect(page.getByText('0W')).toBeVisible();
  await expect(page.getByText('2L')).toBeVisible();
  await expect(page.getByTestId('history-run-status')).toHaveText('Run lost');
  await page.getByRole('button', { name: 'View' }).click();
  await expect(page.getByText('Record 0–2, from the run.')).toBeVisible();
  await expect(page.getByRole('spinbutton')).toHaveCount(0);
}

test('two concessions end the run on a summary; Draft again keeps it in Past runs as 0W / 2L, read-only', async ({ page }) => {
  test.setTimeout(90_000);
  await loseARun(page);
  await expect(page.getByTestId('run-status')).toHaveText('Run lost');
  const rows = page.getByTestId('run-battles').locator('li');
  await expect(rows).toHaveCount(2);
  await expect(rows.nth(0)).toContainText('Battle 1: Loss');
  await expect(rows.nth(1)).toContainText('Battle 2: Loss');
  for (const name of ['Copy code', 'Copy link', 'Export JSON']) {
    await expect(page.getByRole('button', { name })).toBeVisible();
  }
  await page.screenshot({ path: 'test-results/end-of-run-summary.png', fullPage: true });

  await rows.nth(0).getByRole('button', { name: 'Replay' }).click();
  await page.getByTestId('replay-exit').click();
  await expect(page.getByTestId('run-status')).toHaveText('Run lost');

  await page.getByRole('button', { name: 'Draft again' }).click();
  await page.getByRole('button', { name: 'Abandon draft' }).click();
  await page.getByTestId('abandon-ok').click();
  await expectLostRowReadOnly(page);
});

test('a lost run not yet archived already shows 0W / 2L, read-only, in Past runs after a reload', async ({ page }) => {
  test.setTimeout(90_000);
  await loseARun(page);
  await page.reload();
  await expectLostRowReadOnly(page);
});

test('an old history entry without a run keeps editable W/L', async ({ page }) => {
  await page.addInitScript(
    (s) => { if (!localStorage.getItem('pocket-draft-v0.4')) localStorage.setItem('pocket-draft-v0.4', JSON.stringify({ version: 2, state: s })); },
    {
      pickIds: [], offerIds: [], packOffers: [], shopPurchasedIds: [], shopFinalized: false,
      deck: {}, deckFinalized: false, history: [entry('old')], run: null,
    },
  );
  await page.goto('/');
  await page.getByRole('button', { name: /Past runs/ }).click();
  await page.getByRole('button', { name: 'View' }).click();
  const wins = page.getByLabel('Wins:');
  await expect(wins).toHaveValue('3');
  await wins.fill('4');
  await expect(page.getByText('4W')).toBeVisible();
});
