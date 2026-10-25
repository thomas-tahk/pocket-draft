import { test, expect, type Page } from '@playwright/test';
import { RUN_LOSSES_TO_LOSE, RUN_WINS_TO_WIN, newRun, withBattle, type RunBattle } from '../src/stores/run';

// #114: a run starts when the deck is finalized, and each game that ends is
// saved to it with its replayable record (ADR 0005).

type Saved = { run: { status: string; battles: { result: string; seed: number; yourDeck: string[]; botDeck: string[]; actions: { type: string }[] }[] } };

const savedRun = (page: Page) =>
  page.evaluate(() => (JSON.parse(localStorage.getItem('pocket-draft-v0.4') ?? '{}').state as Saved).run);

async function concedeGame(page: Page) {
  await page.getByRole('button', { name: '▶ Play deck' }).click();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();
  await expect(page.getByTestId('end-result')).toHaveText('Loss');
  await page.getByRole('button', { name: 'Back' }).click();
}

test('conceding twice records 0-1 then ends the run as lost, with a log per battle', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–0 (active)');

  await concedeGame(page);
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–1 (active)');
  let run = await savedRun(page);
  expect(run.battles).toHaveLength(1);
  expect(run.battles[0].result).toBe('loss');
  expect(run.battles[0].actions.length).toBeGreaterThan(0);
  expect(run.battles[0].actions.at(-1)?.type).toBe('Concede');
  expect(run.battles[0].yourDeck).toHaveLength(20);
  expect(run.battles[0].botDeck).toHaveLength(20);

  await concedeGame(page);
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–2 (lost)');
  run = await savedRun(page);
  expect(run.status).toBe('lost');
  expect(run.battles).toHaveLength(RUN_LOSSES_TO_LOSE);
  expect(run.battles[0].seed).not.toBe(run.battles[1].seed);
});

const battle = (result: 'win' | 'loss', seed: number): RunBattle => ({
  result, seed, yourDeck: [], botDeck: [], actions: [], endedAt: '2026-10-25T00:00:00.000Z',
});

test('run logic: 5 wins win, 2 losses lose, a repeated seed is ignored, a finished run is frozen', () => {
  let run = newRun({ a: 2 });
  for (let i = 0; i < RUN_WINS_TO_WIN - 1; i++) run = withBattle(run, battle('win', i));
  expect(run.status).toBe('active');
  expect(withBattle(run, battle('win', 0))).toBe(run); // same seed: not recorded twice
  run = withBattle(run, battle('win', 99));
  expect(run.status).toBe('won');
  expect(withBattle(run, battle('loss', 100))).toBe(run);

  let lost = withBattle(newRun({}), battle('loss', 1));
  lost = withBattle(withBattle(lost, battle('win', 2)), battle('loss', 3));
  expect(lost.status).toBe('lost');
});
