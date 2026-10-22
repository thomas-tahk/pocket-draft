import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';

// M3-1: energy is a colored symbol, never a letter. Every place the board shows
// a type (Energy Zone, attached Energy, attack costs, retreat cost, card text,
// the move log) draws an <svg> and no letter; "[L]" in card text is an icon.
// Zeraora's Ability text names [L] and its attack costs Energy.

type Card = { id: string; name: string };
type View = { phase: string; active: number; pending: { player: number } | null; players: { hand: Card[] }[] };

const ZERAORA = 'A3a-021';
const FILLER = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112'];
const DECK = [ZERAORA, ...FILLER, 'A1-150', 'B4-036', 'P-A-001'].flatMap((id) => [id, id]);
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  if (!res.ok()) throw new Error(`${url}: ${JSON.stringify(body)}`);
  return body.state ?? body;
}

async function runBotUntilYourTurn(api: APIRequestContext) {
  // /api/bot moves for whoever is to act; it no longer just ends your turn.
  const now = (await (await api.get('/api/state')).json()) as View;
  if ((now.pending ? now.pending.player : now.active) === 0) await post(api, '/api/move', { type: 'EndTurn', player: 0 });
  for (let i = 0; i < 40; i++) {
    const v = await post(api, '/api/bot');
    if (v.phase !== 'main' || (v.pending ? v.pending.player : v.active) === 0) return;
  }
}

// Zeraora Active, one Energy attached on your first turn.
async function toZeraoraWithEnergy(api: APIRequestContext) {
  for (let seed = 1; seed < 200; seed++) {
    const start = await post(api, '/api/new', { seed, you: DECK, bot: BOT });
    if (!start.players[0].hand.some((c) => c.id === ZERAORA)) continue;
    await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: ZERAORA });
    await post(api, '/api/move', { type: 'SetupReady', player: 0 });
    for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
    await runBotUntilYourTurn(api);
    await post(api, '/api/move', { type: 'AttachEnergy', player: 0, target: 0 });
    return;
  }
  throw new Error('no seed deals Zeraora');
}

async function tap(page: Page, l: Locator) {
  const b = (await l.boundingBox())!;
  await page.mouse.click(b.x + b.width / 2, b.y + b.height / 2);
}

// Every icon inside `scope` has an <svg> (unless used up) and no text at all.
async function expectOnlyIcons(scope: Locator, min: number) {
  const icons = scope.locator('.bt-energy');
  expect(await icons.count()).toBeGreaterThanOrEqual(min);
  for (const icon of await icons.all()) {
    // A used Energy Zone icon is a bare grey disc by design.
    await expect(icon.locator('svg')).toHaveCount((await icon.getAttribute('data-spent')) === 'true' ? 0 : 1);
    expect((await icon.textContent())?.trim()).toBe('');
  }
}

test('board, attached Energy, attack costs, retreat cost and card text use symbols, not letters', async ({ page, request }) => {
  await toZeraoraWithEnergy(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();

  await expectOnlyIcons(page.getByTestId('my-energy'), 1);
  await expectOnlyIcons(page.getByTestId('my-active'), 1);
  await expect(page.getByTestId('my-energy')).not.toContainText(/^[A-Z]$/);

  await tap(page, page.getByTestId('active-card'));
  const sheet = page.getByTestId('active-sheet');
  await expect(sheet).toBeVisible();
  await expectOnlyIcons(sheet.getByTestId('sheet-attack'), 1);
  await expectOnlyIcons(sheet.getByTestId('sheet-retreat'), 0);

  // "[L]" in the Ability text is drawn as a Lightning icon.
  const ability = sheet.getByTestId('sheet-ability');
  await expect(ability).not.toContainText('[');
  await expect(ability.locator('.bt-energy[data-type="Lightning"] svg')).toHaveCount(1);
  await expect(ability.locator('.bt-energy[data-type="Lightning"]')).toHaveAttribute('aria-label', 'Lightning energy');
});

test('the move log draws "attached Lightning energy" with an icon', async ({ page, request }) => {
  await toZeraoraWithEnergy(request);
  await page.goto('/?battle');
  await page.getByTestId('history-button').click();
  const attach = page.locator('[data-testid="history-entry"][data-kind="attach_energy"]').first();
  await expect(attach.locator('.bt-energy svg')).toHaveCount(1);
  await expect(attach).not.toContainText(/(Lightning|Fire|Water|Grass|Psychic|Fighting|Darkness|Metal|Colorless) energy/);
});
