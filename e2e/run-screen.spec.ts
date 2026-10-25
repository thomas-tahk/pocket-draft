import { test, expect, type Page } from '@playwright/test';

// #116: the run screen between battles.

async function start(page: Page) {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–0 (active)');
}

async function concede(page: Page) {
  await page.getByRole('button', { name: 'Next battle' }).click();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();
  await expect(page.getByTestId('end-result')).toHaveText('Loss');
  await page.getByRole('button', { name: 'Back' }).click();
}

test('fresh run shows 0-0, a bot-labelled Next battle, no battles', async ({ page }) => {
  await start(page);
  await expect(page.getByText('Opponent: practice bot')).toBeVisible();
  await expect(page.getByTestId('run-wins')).toContainText('Wins 0 of 5');
  await expect(page.getByTestId('run-losses')).toContainText('Losses 0 of 2');
  await expect(page.getByText('No battles yet.')).toBeVisible();
  await page.screenshot({ path: 'test-results/run-screen-fresh.png', fullPage: true });
});

test('two conceded battles are listed in order as losses; run lost, Next battle gone', async ({ page }) => {
  await start(page);
  await concede(page);
  await concede(page);
  const rows = page.getByTestId('run-battles').locator('li');
  await expect(rows).toHaveCount(2);
  await expect(rows.nth(0)).toContainText('Battle 1: Loss');
  await expect(rows.nth(1)).toContainText('Battle 2: Loss');
  await expect(rows.nth(0).getByRole('button', { name: 'Replay' })).toBeEnabled();
  await expect(page.getByTestId('run-status')).toHaveText('Run lost');
  await expect(page.getByRole('button', { name: 'Next battle' })).toHaveCount(0);
  await page.screenshot({ path: 'test-results/run-screen-lost.png', fullPage: true });
});

test('Quit run asks on-screen, then ends the run as quit', async ({ page }) => {
  await start(page);
  await page.getByRole('button', { name: 'Quit run' }).click();
  await page.getByRole('button', { name: 'Keep playing' }).click();
  await expect(page.getByTestId('run-status')).toHaveText('In progress');
  await page.getByRole('button', { name: 'Quit run' }).click();
  await page.getByTestId('quit-confirm').click();
  await expect(page.getByTestId('run-status')).toHaveText('Run quit');
  await expect(page.getByRole('button', { name: 'Next battle' })).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Quit run' })).toHaveCount(0);
});

for (const opener of ['Start new draft', 'Home']) {
  test(`${opener} mid-run asks on-screen; cancel keeps the run active`, async ({ page }) => {
    await start(page);
    await page.getByRole('button', { name: new RegExp(opener) }).first().click();
    await expect(page.getByTestId('leave-confirm')).toContainText('This quits your run at 0–0.');
    if (opener === 'Home') await page.screenshot({ path: 'test-results/run-leave-confirm-390.png', fullPage: true });
    await page.getByTestId('leave-cancel').click();
    await expect(page.getByTestId('leave-confirm')).toHaveCount(0);
    await expect(page.getByTestId('run-status')).toHaveText('In progress');
  });
}

test('confirming leave quits the run, then leaves', async ({ page }) => {
  await start(page);
  await page.getByRole('button', { name: 'Start new draft' }).click();
  await page.getByTestId('leave-ok').click();
  await expect(page.getByTestId('run-record')).toHaveCount(0);
});

test('lost run: the losing battle end screen has no Rematch', async ({ page }) => {
  await start(page);
  await concede(page);
  await page.getByRole('button', { name: 'Next battle' }).click();
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();
  await expect(page.getByTestId('end-result')).toHaveText('Loss');
  await expect(page.getByRole('button', { name: 'Back' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Rematch' })).toHaveCount(0);
});

test('a won run hides Next battle and Quit run', async ({ page }) => {
  await start(page);
  await page.evaluate(() => {
    const saved = JSON.parse(localStorage.getItem('pocket-draft-v0.4')!);
    const battle = (seed: number) => ({ result: 'win', seed, yourDeck: [], botDeck: [], actions: [], endedAt: new Date().toISOString() });
    saved.state.run.battles = [1, 2, 3, 4, 5].map(battle);
    saved.state.run.status = 'won';
    localStorage.setItem('pocket-draft-v0.4', JSON.stringify(saved));
  });
  await page.reload();
  await page.getByRole('button', { name: 'View last run' }).click();
  await expect(page.getByTestId('run-status')).toHaveText('Run won');
  await expect(page.getByRole('button', { name: 'Next battle' })).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Quit run' })).toHaveCount(0);
});
