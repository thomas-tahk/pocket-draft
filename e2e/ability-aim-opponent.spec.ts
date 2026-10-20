import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';

// F10 at phone size: an Ability aimed at the opponent's Pokémon. Greninja's
// Water Shuriken (A1-089) does 20 damage to 1 of the opponent's Pokémon, so
// the choice overlay must light the opponent's Pokémon, not only yours. Both
// decks are pinned, so the draft pool can't shift them.

type Mon = { damage: number; card: { id: string } } | null;
type Card = { id: string; stage: string };
type View = {
  phase: string;
  active: number;
  pending: { player: number } | null;
  players: { hand: Card[]; active: Mon; bench: Mon[] }[];
};

const GRENINJA = 'A1-089';
const FROAKIE = 'A1-087';
const FROGADIER = 'A1-088';
const FILLER = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150'];
const DECK = [FROAKIE, FROGADIER, GRENINJA, ...FILLER].flatMap((id) => [id, id]);
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = JSON.parse(await res.text());
  return body.state ?? body;
}

const actor = (v: View) => (v.pending ? v.pending.player : v.active);

async function runBotUntilYourTurn(api: APIRequestContext): Promise<View> {
  let v = await post(api, '/api/bot');
  for (let i = 0; i < 40 && v.phase === 'main' && actor(v) !== 0; i++) v = await post(api, '/api/bot');
  return v;
}

// Evolves your Active with whatever the hand allows; refused moves are skipped.
async function evolveActive(api: APIRequestContext, v: View): Promise<View> {
  for (const card of v.players[0].hand.filter((c) => c.stage !== 'Basic')) {
    const res = await api.post('/api/move', { data: { type: 'Evolve', player: 0, handCardId: card.id, target: 0 } });
    const body = await res.json();
    if (body.ok) return body.state;
  }
  return v;
}

// Plays on, evolving each turn, until a Greninja is your Active and the bot's
// Active survives, on the first seed that gets there.
async function toGreninjaActive(api: APIRequestContext) {
  for (let seed = 1; seed < 60; seed++) {
    const start = await post(api, '/api/new', { seed, you: DECK, bot: BOT });
    const froakie = start.players[0].hand.find((c) => c.id === FROAKIE);
    if (!froakie) continue;
    await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: froakie.id });
    await post(api, '/api/move', { type: 'SetupReady', player: 0 });
    for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
    let v = await runBotUntilYourTurn(api);
    for (let turn = 0; turn < 6 && v.phase === 'main' && v.players[0].active; turn++) {
      v = await evolveActive(api, v);
      if (v.players[0].active?.card.id === GRENINJA) return;
      await post(api, '/api/move', { type: 'EndTurn', player: 0 });
      v = await runBotUntilYourTurn(api);
    }
  }
  throw new Error('no seed gets a Greninja into the Active Spot');
}

async function tap(page: Page, l: Locator) {
  const b = (await l.boundingBox())!;
  await page.mouse.click(b.x + b.width / 2, b.y + b.height / 2);
}

test('Water Shuriken can be aimed at an opponent Pokémon', async ({ page, request }) => {
  await toGreninjaActive(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });

  await tap(page, page.getByTestId('active-card'));
  const ability = page.getByTestId('active-sheet').getByTestId('sheet-ability');
  await expect(ability).toContainText('Water Shuriken');
  await ability.click();

  // The overlay lights the opponent's Active (and any Benched Pokémon), not only yours.
  const lit = page.getByTestId('opp-active').locator('[data-lit="true"][data-pickable="true"]');
  await expect(lit).toBeVisible();
  await page.screenshot({ path: 'test-results/ability-aim-opponent.png' });

  await tap(page, lit);
  const state = (await (await request.get('/api/state')).json()) as View;
  expect(state.players[1].active!.damage).toBe(20);
  await expect(page.getByTestId('choice-prompt')).toHaveCount(0);
});
