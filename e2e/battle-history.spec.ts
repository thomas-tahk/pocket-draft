import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

// F15: a small notebook icon opens a scrollable list of the moves so far,
// newest first, each with the card used and what it did. No text log shows
// until it is opened. Desktop shows more of it at once, with the same entries.

type Card = { id: string; name: string; stage: string };
type Action = { player: number; kind: string; card: Card | null; result: string[] };
type View = { phase: string; active: number; pending: { player: number } | null; players: { hand: Card[] }[]; actions: Action[] };

// Pinned so draft-pool changes can't shift the bot's deck.
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

// Setup, end your first turn, let the bot take its turn, attach an energy.
async function playAFewMoves(api: APIRequestContext): Promise<View> {
  const start = await post(api, '/api/new', { seed: 7, bot: BOT, preset: 'fire' });
  const basics = start.players[0].hand.filter((c) => c.stage === 'Basic');
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basics[0].id });
  await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: basics[1].id, slot: 0 });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  await post(api, '/api/move', { type: 'EndTurn', player: 0 });
  for (let i = 0; i < 40; i++) {
    const v = await post(api, '/api/bot');
    if (v.phase !== 'main' || (v.pending ? v.pending.player : v.active) === 0) break;
  }
  return post(api, '/api/move', { type: 'AttachEnergy', player: 0, target: 0 });
}

async function openHistory(page: Page, view: View) {
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  // No text log on the board until the icon is tapped.
  await expect(page.getByTestId('history-panel')).toHaveCount(0);
  for (const line of view.actions.flatMap((a) => a.result)) await expect(page.getByText(line, { exact: true })).toHaveCount(0);

  await page.getByTestId('history-button').click();
  const panel = page.getByTestId('history-panel');
  await expect(panel).toBeVisible();
  return panel;
}

// Every move is listed, newest first, with its card name and result lines.
async function expectEntriesInOrder(page: Page, view: View) {
  const entries = page.getByTestId('history-entry');
  const newestFirst = [...view.actions].reverse();
  await expect(entries).toHaveCount(newestFirst.length);
  for (const [i, a] of newestFirst.entries()) {
    const entry = entries.nth(i);
    await expect(entry).toHaveAttribute('data-kind', a.kind);
    await expect(entry).toContainText(a.player === 0 ? 'You' : 'Opponent');
    if (a.card) await expect(entry).toContainText(a.card.name);
    // The Energy type in "attached Fire energy" is drawn as an icon, not text.
    for (const line of a.result) await expect(entry).toContainText(line.replace(/(Grass|Fire|Water|Lightning|Psychic|Fighting|Darkness|Metal|Dragon|Colorless) (?=energy)/, ''));
  }
}

test('history: the notebook icon lists the last moves in order', async ({ page, request }) => {
  const view = await playAFewMoves(request);
  expect(view.actions.length).toBeGreaterThan(5);
  const panel = await openHistory(page, view);

  // The icon is small and on the board.
  const icon = await page.getByTestId('history-button').boundingBox();
  expect(icon!.width).toBeLessThanOrEqual(44);
  expect(icon!.x).toBeGreaterThanOrEqual(0);
  expect(icon!.y).toBeGreaterThanOrEqual(0);

  await expectEntriesInOrder(page, view);
  // The newest move is the energy you just attached.
  await expect(page.getByTestId('history-entry').first()).toHaveAttribute('data-kind', 'attach_energy');

  // Scrollable, and fits the phone.
  const list = page.getByTestId('history-list');
  expect(await list.evaluate((el) => getComputedStyle(el).overflowY)).toBe('auto');
  const b = await panel.boundingBox();
  expect(b!.y + b!.height).toBeLessThanOrEqual(844 + 1);
  await page.waitForFunction(() => [...document.images].every((i) => i.complete));
  await page.screenshot({ path: 'test-results/battle-history-390x844.png' });

  // Closing it returns to the board with no log showing.
  await page.getByTestId('history-close').click();
  await expect(panel).toHaveCount(0);
});

test.describe('desktop', () => {
  test.use({ viewport: { width: 1280, height: 800 }, isMobile: false, hasTouch: false, deviceScaleFactor: 1 });

  test('history: same entries, more of them at once', async ({ page, request }) => {
    const view = await playAFewMoves(request);
    const panel = await openHistory(page, view);
    await expectEntriesInOrder(page, view);

    // Docked beside the board rather than over it, and taller than on the phone.
    const board = await page.getByTestId('battle-board').boundingBox();
    const b = await panel.boundingBox();
    expect(b!.x).toBeGreaterThanOrEqual(board!.x + board!.width - 1);
    expect(b!.height).toBeGreaterThan(600);
    await page.waitForFunction(() => [...document.images].every((i) => i.complete));
    await page.screenshot({ path: 'test-results/battle-history-1280x800.png' });
  });
});
