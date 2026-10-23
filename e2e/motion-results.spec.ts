import { test, expect, type APIRequestContext, type Page } from '@playwright/test';
import { animationsSoFar, logAnimations, settled, transitionsSoFar } from './fx-log';
import { FIRE } from './decks';

// M3-4: results land where they happen. Attached Energy flies from the Energy
// Zone, an attacker nudges toward its target, a number pops on the Pokémon
// whose HP changed (and matches the HP lost or gained), the HP bar drains.

type Card = { id: string; stage: string };
type Mon = { damage: number; remainingHp: number };
type View = { phase: string; active: number; pending: { player: number } | null; players: { hand: Card[]; active: Mon | null }[] };

const POTION = 'P-A-001';
const RESEARCH = 'A4b-373';
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', POTION, RESEARCH].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}
const state = async (api: APIRequestContext) => (await api.get('/api/state')).json() as Promise<View>;

async function botPlaysUntilYou(api: APIRequestContext) {
  for (let i = 0; i < 40; i++) {
    const v = await post(api, '/api/bot');
    if (v.phase !== 'main' || (v.pending ? v.pending.player : v.active) === 0) return;
  }
}

// Seed 7 (active-card.spec.ts): Entei ex Active, Heatran on the Bench; Fire on the Active each turn.
async function toEnteiTurn(api: APIRequestContext, fireAttached: number) {
  await post(api, '/api/new', { seed: 7, bot: BOT, you: FIRE });
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: 'B2a-126' });
  await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: 'A4b-076', slot: 0 });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  await botPlaysUntilYou(api);
  for (let n = 0; n < fireAttached; n++) {
    await post(api, '/api/move', { type: 'AttachEnergy', player: 0, target: 0 });
    if (n + 1 < fireAttached) {
      await post(api, '/api/move', { type: 'EndTurn', player: 0 });
      await botPlaysUntilYou(api);
    }
  }
}

async function openBoard(page: Page) {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await logAnimations(page);
  await page.goto('/?battle');
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
  await settled(page);
}

test('attached Energy flies from the Energy Zone to the Pokémon', async ({ page, request }) => {
  await toEnteiTurn(request, 0);
  await openBoard(page);
  const from = (await page.getByTestId('my-energy').getByTestId('energy-current').boundingBox())!;
  const to = (await page.getByTestId('my-active').boundingBox())!;
  await page.mouse.move(from.x + from.width / 2, from.y + from.height / 2);
  await page.mouse.down();
  await page.mouse.move(to.x + to.width / 2, to.y + to.height / 2, { steps: 8 });
  await page.mouse.up();
  await expect(page.getByTestId('my-active').getByTestId('energy-icon')).toHaveCount(1);
  const flight = (await animationsSoFar(page)).filter((a) => a.energy === 'Fire' && a.zone === 'my-active');
  expect(flight, 'the attached Fire icon is animated').toHaveLength(1);
  expect(flight[0].props).toContain('translate');
});

test('an attack: the attacker nudges, the damage pops matching the HP lost, the bar drains', async ({ page, request }) => {
  await toEnteiTurn(request, 2);
  const hpBefore = (await state(request)).players[1].active!.remainingHp;
  await openBoard(page);
  await page.getByTestId('active-card').click();
  await page.getByTestId('sheet-attack').click();
  await expect(page.getByTestId('opp-active').getByTestId('hp')).not.toHaveText(String(hpBefore));
  const lost = hpBefore - (await state(request)).players[1].active!.remainingHp;
  expect(lost).toBeGreaterThan(0);
  const log = await animationsSoFar(page);
  expect(log.filter((a) => a.text), 'one number').toEqual([expect.objectContaining({ text: `-${lost}`, copy: true })]);
  expect(log.some((a) => a.zone === 'my-active' && a.key.startsWith('p0:B2a-126#') && a.stops === 3), 'Entei ex nudges').toBe(true);
  expect(await transitionsSoFar(page)).toContain('width@bt-hp-bar');
});

// Seed 5 (trainer-play.spec.ts): end turns until your Active is damaged and a Potion is in hand.
test('a Potion heals: a green number matching the HP gained', async ({ page, request }) => {
  const start = await post(request, '/api/new', { seed: 5, you: BOT, bot: BOT });
  const basic = start.players[0].hand.find((c) => c.stage === 'Basic')!;
  await post(request, '/api/move', { type: 'SetupActive', player: 0, cardId: basic.id });
  await post(request, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(request, '/api/bot')).phase === 'setup'; i++);
  let v = start;
  for (let turn = 0; turn < 6; turn++) {
    await post(request, '/api/move', { type: 'EndTurn', player: 0 });
    await botPlaysUntilYou(request);
    v = await state(request);
    if (v.players[0].active!.damage > 0 && v.players[0].hand.some((c) => c.id === POTION)) break;
  }
  const hp = v.players[0].active!.remainingHp;
  const gained = Math.min(20, v.players[0].active!.damage);
  expect(gained).toBeGreaterThan(0);
  await openBoard(page);
  const potion = (await page.getByTestId('my-hand').locator(`[data-card-id="${POTION}"]`).first().boundingBox())!;
  const active = (await page.getByTestId('my-active').boundingBox())!;
  await page.mouse.move(potion.x + 10, potion.y + potion.height / 2);
  await page.mouse.down();
  await page.mouse.move(active.x + active.width / 2, active.y + active.height / 2, { steps: 8 });
  await page.mouse.up();
  await expect(page.getByTestId('my-active').getByTestId('hp')).toHaveText(String(hp + gained));
  const numbers = (await animationsSoFar(page)).filter((a) => a.text);
  expect(numbers).toEqual([expect.objectContaining({ text: `+${gained}` })]);
});
