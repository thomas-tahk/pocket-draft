import { test, expect, type APIRequestContext, type Page } from '@playwright/test';
import type { GameView } from '../src/game/types';

// #119: a finished battle replays read-only on the board; stepping to the last
// step shows the game's final view, and Back steps back.

async function post<T>(api: APIRequestContext, url: string, data?: object): Promise<T> {
  const body = await (await api.post(url, data ? { data } : undefined)).json();
  return (body.state ?? body) as T;
}
const state = (api: APIRequestContext) => post<GameView>(api, '/api/state');

// Setup has no screen input in tests: place and drive the bot through the API.
async function playSomeTurns(api: APIRequestContext) {
  const v = await state(api);
  const basic = v.players[0].hand.find((c) => c.stage === 'Basic')!;
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basic.id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 8 && (await post<GameView>(api, '/api/bot')).phase === 'setup'; i++);
  for (let turn = 0; turn < 3; turn++) {
    if ((await state(api)).phase !== 'main') break;
    if ((await state(api)).active === 0) await post(api, '/api/move', { type: 'EndTurn', player: 0 });
    for (let i = 0; i < 40; i++) if (!(await post<{ acted: boolean }>(api, '/api/bot')).acted) break;
  }
}

// What the board shows that must match the game: points, deck and discard counts.
async function boardFacts(page: Page) {
  const pips = (zone: string) => page.getByTestId(zone).locator('[data-testid="pip"][data-filled="true"]').count();
  const label = (zone: string, what: string) => page.getByTestId(zone).locator(`[aria-label^="${what}"]`).getAttribute('aria-label');
  return {
    points: [await pips('opp-points'), await pips('my-points')],
    deck: [await label('opp-deck', 'deck'), await label('my-deck', 'deck')],
    discard: [await label('opp-discard', 'discard'), await label('my-discard', 'discard')],
  };
}

test('replay steps to the final view, Back steps back, and nothing is playable', async ({ page, request }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await page.getByRole('button', { name: '▶ Play deck' }).click();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('hand-card').first()).toBeVisible();
  await page.waitForTimeout(500); // let the page's own game start settle before driving the server
  await playSomeTurns(request);
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();
  await expect(page.getByTestId('end-result')).toHaveText('Loss');
  const final = await state(request);
  const originalFacts = await boardFacts(page);

  const errors: string[] = [];
  page.on('console', (m) => m.type() === 'error' && errors.push(m.text()));
  await page.goto('/?replay=0');
  await expect(page.getByTestId('replay-step')).toHaveText(/^Step 0 \/ \d+$/);
  const total = Number((await page.getByTestId('replay-step').innerText()).split('/')[1]);
  expect(total).toBeGreaterThan(5);
  await expect(page.getByTestId('replay-back')).toBeDisabled();
  await expect(page.getByTestId('end-screen')).toBeHidden();

  // read-only: no move controls, and tapping the board changes nothing
  await expect(page.getByTestId('menu')).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'End Turn' })).toBeHidden();
  await page.getByTestId('my-active').click({ force: true });
  await expect(page.getByTestId('replay-step')).toHaveText(`Step 0 / ${total}`);

  while (!(await page.getByTestId('replay-forward').isDisabled())) await page.getByTestId('replay-forward').click();
  await expect(page.getByTestId('replay-step')).toHaveText(`Step ${total} / ${total}`);
  expect(await boardFacts(page)).toEqual(originalFacts);
  await expect(page.getByTestId('replay-narration')).toHaveText(final.narration.at(-1)!);
  await page.screenshot({ path: 'test-results/replay-390-final.png' });

  await page.getByTestId('replay-back').click();
  await expect(page.getByTestId('replay-step')).toHaveText(`Step ${total - 1} / ${total}`);
  await expect(page.getByTestId('replay-forward')).toBeEnabled();

  await page.getByTestId('replay-exit').click();
  await expect(page.getByTestId('replay-step')).toHaveCount(0);
  expect(errors).toEqual([]);
});
