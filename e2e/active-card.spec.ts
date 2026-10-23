import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';
import { FIRE } from './decks';

// F9 / T9 / T10 at phone size: tap your Active to open its card. Its attacks
// are listed, the ones you can't use greyed; tapping a usable one attacks.
// Retreat makes your Bench glow and tapping the replacement switches.
// Seed 7, driven through the API: the bot goes first; you have Entei ex (attack
// costs Fire Fire, retreat 2) Active and Heatran on the Bench, and Fire is
// attached to the Active on each of your turns.

type View = { phase: string; active: number; pending: { player: number } | null; players: { active: { remainingHp: number } }[] };

// Pinned so draft-pool changes can't shift the bot's deck.
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

async function runBotUntilYourTurn(api: APIRequestContext) {
  for (let i = 0; i < 40; i++) {
    const v = await post(api, '/api/bot');
    const actor = v.pending ? v.pending.player : v.active;
    if (v.phase !== 'main' || actor === 0) return;
  }
}

const endTurn = (api: APIRequestContext) => post(api, '/api/move', { type: 'EndTurn', player: 0 });
const attachToActive = (api: APIRequestContext) => post(api, '/api/move', { type: 'AttachEnergy', player: 0, target: 0 });

// Your first turn (game turn 2), nothing attached yet.
async function toYourFirstTurn(api: APIRequestContext) {
  await post(api, '/api/new', { seed: 7, bot: BOT, you: FIRE });
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: 'B2a-126' });
  await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: 'A4b-076', slot: 0 });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  await runBotUntilYourTurn(api);
}

// Your second turn with Fire Fire on Entei ex: its attack and Retreat are both legal.
async function toYourSecondTurn(api: APIRequestContext): Promise<View> {
  await toYourFirstTurn(api);
  await attachToActive(api);
  await endTurn(api);
  await runBotUntilYourTurn(api);
  return attachToActive(api);
}

async function tap(page: Page, l: Locator) {
  const b = (await l.boundingBox())!;
  await page.mouse.click(b.x + b.width / 2, b.y + b.height / 2);
}

async function openBoard(page: Page) {
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
}

const sheet = (page: Page) => page.getByTestId('active-sheet');

test('tap your Active: attacks listed, unaffordable greyed; a tap-tap drop on it does not open it', async ({ page, request }) => {
  await toYourFirstTurn(request);
  await openBoard(page);
  const active = page.getByTestId('active-card');

  // Tap-tap: energy held, then a tap on the Active attaches it and opens nothing.
  await tap(page, page.getByTestId('my-energy').getByTestId('energy-current'));
  await expect(page.getByTestId('play-hint')).toBeVisible();
  await tap(page, active);
  await expect(page.getByTestId('my-active').getByTestId('energy-icon')).toHaveCount(1);
  await expect(sheet(page)).toHaveCount(0);

  // One Fire on a Fire Fire attack: listed but greyed; Retreat (cost 2) greyed too.
  await tap(page, active);
  await expect(sheet(page)).toBeVisible();
  const attack = sheet(page).getByTestId('sheet-attack');
  await expect(attack).toHaveCount(1);
  await expect(attack).toContainText('Blazing Beatdown');
  await expect(attack).toHaveAttribute('data-usable', 'false');
  await expect(attack).toBeDisabled();
  await expect(sheet(page).getByTestId('sheet-retreat')).toBeDisabled();
  await page.waitForFunction(() => [...document.images].every((i) => i.complete));
  await page.screenshot({ path: 'test-results/active-card-open.png' });

  // Tapping off the card closes it.
  await page.mouse.click(195, 30);
  await expect(sheet(page)).toHaveCount(0);

  // A long press zooms (F2) and its release doesn't open the card.
  const b = (await active.boundingBox())!;
  await page.mouse.move(b.x + b.width / 2, b.y + b.height / 2);
  await page.mouse.down();
  await expect(page.getByTestId('card-zoom')).toBeVisible();
  await page.mouse.up();
  await expect(page.getByTestId('card-zoom')).toHaveCount(0);
  await expect(sheet(page)).toHaveCount(0);
});

test('tapping a usable attack uses it', async ({ page, request }) => {
  const before = await toYourSecondTurn(request);
  await openBoard(page);
  await tap(page, page.getByTestId('active-card'));
  const attack = sheet(page).getByTestId('sheet-attack');
  await expect(attack).toHaveAttribute('data-usable', 'true');
  await attack.click();
  await expect(sheet(page)).toHaveCount(0);
  const hp = before.players[1].active.remainingHp;
  await expect(page.getByTestId('opp-active').getByTestId('hp')).not.toHaveText(String(hp));
});

test('Retreat makes your Bench glow; tapping the replacement switches', async ({ page, request }) => {
  await toYourSecondTurn(request);
  await openBoard(page);
  await tap(page, page.getByTestId('active-card'));
  await sheet(page).getByTestId('sheet-retreat').click();
  await expect(sheet(page)).toHaveCount(0);

  const bench = page.getByTestId('my-bench');
  await expect(bench.locator('[data-slot="left"] [data-pickable="true"]')).toHaveCount(1);
  await expect(bench.locator('[data-pickable="true"]')).toHaveCount(1); // only the occupied slot
  await page.screenshot({ path: 'test-results/active-card-retreat.png' });
  await tap(page, bench.locator('[data-slot="left"]'));

  await expect(page.getByTestId('my-active').locator('img')).toHaveAttribute('alt', 'Heatran');
  await expect(bench.locator('[data-slot="left"] img')).toHaveAttribute('alt', 'Entei ex');
  await expect(page.getByTestId('choice-dim')).toHaveCount(0);
});
