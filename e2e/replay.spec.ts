import { test, expect, type APIRequestContext, type Page } from '@playwright/test';
import type { GameView, Move } from '../src/game/types';

// #119: a finished battle replays read-only on the board; the last step is the
// game's real final state, a record that no longer replays shows an error, and
// the run screen's Replay button opens it and Exit comes back.

const KEY = 'pocket-draft-v0.4';

async function post<T>(api: APIRequestContext, url: string, data?: object): Promise<T> {
  const body = await (await api.post(url, data ? { data } : undefined)).json();
  return (body.state ?? body) as T;
}
const state = (api: APIRequestContext) => post<GameView>(api, '/api/state');

// What the greedy player does next: answer a prompt, else attack > attach > play > end turn.
const PREFERENCE: Move['type'][] = ['ChooseNewActive', 'ChooseOption', 'RetreatDiscard', 'UseAttack', 'AttachEnergy', 'PlayBasic', 'Evolve', 'EndTurn'];
function nextMove(v: GameView): Move {
  const legal = v.players[0].legalMoves;
  for (const type of PREFERENCE) {
    const m = legal.find((l) => l.type === type);
    if (m) return m;
  }
  return legal[0];
}

// A whole game through the API to a real finish (not a concede); returns the final view.
async function playToTheEnd(api: APIRequestContext): Promise<GameView> {
  const { deck } = await post<{ deck: string[] }>(api, '/api/autodeck');
  let v = await post<GameView>(api, '/api/new', { you: deck, seed: 5150 });
  const basic = v.players[0].hand.find((c) => c.stage === 'Basic')!;
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basic.id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 400; i++) {
    v = await state(api);
    if (v.phase === 'over') return v;
    const mine = v.phase === 'main' && (v.pending ? v.pending.player === 0 : v.active === 0);
    if (mine) await post(api, '/api/move', nextMove(v));
    else await post(api, '/api/bot');
  }
  throw new Error('game did not finish');
}

// A fresh run in the app with a finished battle saved to it (recording is covered by run-record.spec).
async function openWithSavedBattle(page: Page, api: APIRequestContext, tamper?: (b: Record<string, unknown>) => void) {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await expect(page.getByTestId('run-record')).toBeVisible();
  const final = await playToTheEnd(api);
  const { seed, you, bot, actions } = final.record;
  const battle: Record<string, unknown> = { result: final.outcome, seed, yourDeck: you, botDeck: bot, actions, endedAt: new Date().toISOString() };
  tamper?.(battle);
  await page.evaluate(([key, b]) => {
    const saved = JSON.parse(localStorage.getItem(key)!);
    saved.state.run.battles = [b];
    localStorage.setItem(key, JSON.stringify(saved));
  }, [KEY, battle] as const);
  return final;
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

const withoutRecord = ({ record: _record, ...rest }: GameView) => rest;

test('a battle played to a real finish replays to its exact final state; read-only; History reachable', async ({ page, request }) => {
  test.setTimeout(120_000);
  const final = await openWithSavedBattle(page, request);
  expect(final.reason).not.toBe('concede');
  expect(final.record.actions.length).toBeGreaterThan(20);

  const errors: string[] = [];
  page.on('console', (m) => m.type() === 'error' && errors.push(m.text()));
  await page.goto('/?replay=0');
  await expect(page.getByTestId('replay-step')).toHaveText(/^Step 0 \/ \d+$/);
  const total = Number((await page.getByTestId('replay-step').innerText()).split('/')[1]);
  await expect(page.getByTestId('replay-narration')).toHaveText('Battle start');
  await expect(page.getByTestId('replay-back')).toBeDisabled();

  // read-only: no menu, no End Turn, tapping the board changes nothing
  await expect(page.getByTestId('menu')).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'End Turn' })).toBeHidden();
  await page.getByTestId('my-active').click({ force: true });
  await expect(page.getByTestId('replay-step')).toHaveText(`Step 0 / ${total}`);
  // Exit does not cover History
  await page.getByTestId('history-button').click();
  await expect(page.getByTestId('history-panel')).toBeVisible();
  await page.getByTestId('history-close').click();

  await page.getByTestId('replay-forward').click();
  await expect(page.getByTestId('replay-narration')).not.toHaveText(/seed|P1|P2/);

  while (!(await page.getByTestId('replay-forward').isDisabled())) await page.getByTestId('replay-forward').click();
  await expect(page.getByTestId('replay-step')).toHaveText(`Step ${total} / ${total}`);
  await expect(page.getByTestId('replay-narration')).toHaveText(/^(You win|You lose|Tie) \d–\d$/);
  // the rebuilt game on the server equals the live game's final view
  expect(withoutRecord(await state(request))).toEqual(withoutRecord(final));
  await page.screenshot({ path: 'test-results/replay-390-final.png' });
  const facts = await boardFacts(page);
  expect(facts.points).toEqual([final.players[1].points, final.players[0].points]);

  await page.getByTestId('replay-back').click();
  await expect(page.getByTestId('replay-step')).toHaveText(`Step ${total - 1} / ${total}`);
  await expect(page.getByTestId('replay-forward')).toBeEnabled();

  await page.getByTestId('replay-exit').click();
  await expect(page.getByTestId('replay-step')).toHaveCount(0);
  expect(errors).toEqual([]);
});

test('a record whose seed was tampered with shows the mismatch error, not a board', async ({ page, request }) => {
  test.setTimeout(120_000);
  await openWithSavedBattle(page, request, (b) => { b.seed = (b.seed as number) + 1; });
  await page.goto('/?replay=0');
  await expect(page.getByTestId('replay-error')).toHaveText(/This replay no longer matches its battle\./);
  await expect(page.getByTestId('battle-board')).toHaveCount(0);
  await page.screenshot({ path: 'test-results/replay-390-error.png' });
  await page.getByTestId('replay-exit').click();
  await expect(page.getByTestId('replay-error')).toHaveCount(0);
});

test('a shortened record shows the same error', async ({ page, request }) => {
  test.setTimeout(120_000);
  await openWithSavedBattle(page, request, (b) => { b.actions = (b.actions as unknown[]).slice(0, 12); });
  await page.goto('/?replay=0');
  await expect(page.getByTestId('replay-error')).toBeVisible();
});

test('a record with no actions shows the same friendly error', async ({ page, request }) => {
  test.setTimeout(120_000);
  await openWithSavedBattle(page, request, (b) => { delete b.actions; });
  await page.goto('/?replay=0');
  await expect(page.getByTestId('replay-error')).toBeVisible();
  await expect(page.getByText(/TypeError|400/)).toHaveCount(0);
});

test('a bad ?replay index lands on the run screen and drops the param', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await expect(page.getByTestId('run-record')).toBeVisible();
  await page.goto('/?replay=7');
  await expect(page.getByTestId('run-record')).toBeVisible();
  expect(page.url()).not.toContain('replay');
});

test('run screen: Replay opens the battle, stepping works, Exit returns to the run screen', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await page.getByRole('button', { name: 'Next battle' }).click();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();
  await expect(page.getByTestId('end-result')).toHaveText('Loss');
  await page.getByRole('button', { name: 'Back' }).click();

  await page.getByTestId('run-battles').getByRole('button', { name: 'Replay' }).first().click();
  await expect(page.getByTestId('replay-step')).toHaveText(/^Step 0 \/ \d+$/);
  await page.getByTestId('replay-forward').click();
  await expect(page.getByTestId('replay-step')).toHaveText(/^Step 1 \/ \d+$/);
  while (!(await page.getByTestId('replay-forward').isDisabled())) await page.getByTestId('replay-forward').click();
  await expect(page.getByTestId('replay-narration')).toHaveText('You conceded — Opponent wins');
  await page.getByTestId('replay-exit').click();
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–1 (active)');
});
