import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';

// #140, Active slot: while aiming a targeted Ability, tapping the lit Active must only
// make the choice, never also reopen the Active's card. Espeon ex's Psychic Healing
// (heal 30 from 1 of your Pokémon) aims at the Active among several targets.
// Both decks are pinned, so the draft pool can't shift them. Real game through the API.

type Mon = { damage: number; card: { id: string } } | null;
type Card = { id: string; stage: string };
type Move = { type: string; target?: number; handCardId?: string; spot?: { player: number; pos: number } };
type View = {
  phase: string;
  active: number;
  pending: { player: number } | null;
  players: { hand: Card[]; active: Mon; bench: Mon[]; legalMoves: Move[] }[];
};

const EEVEE = 'A1-206';
const ESPEON = 'A4-083';
const MATE = 'A2b-041'; // Hitmonchan: retreat cost 1
const FILLER = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150'];
const DECK = [EEVEE, ESPEON, MATE, ...FILLER].flatMap((id) => [id, id]);
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = JSON.parse(await res.text());
  if (body.error) throw new Error(url + ': ' + body.error);
  return body.state ?? body;
}

const actor = (v: View) => (v.pending ? v.pending.player : v.active);

async function runBotUntilYourTurn(api: APIRequestContext): Promise<View> {
  let v = await post(api, '/api/bot');
  for (let i = 0; i < 40 && v.phase === 'main' && actor(v) !== 0; i++) v = await post(api, '/api/bot');
  return v;
}

const aimsAtActive = (v: View) =>
  v.players[0].legalMoves.filter((m) => m.type === 'UseAbility' && m.target === 0 && m.spot).length >= 2;

// Posts the first legal move of a type, as the board would; null when there is none.
async function playLegal(api: APIRequestContext, v: View, type: string, pick?: (m: Move) => boolean): Promise<View | null> {
  const move = v.players[0].legalMoves.find((m) => m.type === type && (!pick || pick(m)));
  if (!move) return null;
  const body = await (await api.post('/api/move', { data: move })).json();
  return body.ok ? body.state : null;
}

// Heal targets are damaged Pokémon only, so two must be hurt: Hitmonchan starts Active and takes the
// bot's hit, retreats behind Eevee, which takes the next hit and evolves into Espeon ex.
async function toEspeonAiming(api: APIRequestContext) {
  for (let seed = 1; seed < 120; seed++) {
    const start = await post(api, '/api/new', { seed, you: DECK, bot: BOT });
    const hand = start.players[0].hand;
    const eevee = hand.find((c) => c.id === EEVEE);
    const mate = hand.find((c) => c.id === MATE);
    if (!eevee || !mate) continue;
    await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: mate.id });
    await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: eevee.id, slot: 0 });
    await post(api, '/api/move', { type: 'SetupReady', player: 0 });
    for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
    let v = await runBotUntilYourTurn(api);
    for (let turn = 0; turn < 14 && v.phase === 'main' && v.players[0].active; turn++) {
      const me = v.players[0];
      if (me.active!.card.id === ESPEON && me.active!.damage > 0 && aimsAtActive(v)) return;
      const espeon = me.hand.find((c) => c.id === ESPEON);
      if (me.active!.card.id === MATE) {
        v = (await playLegal(api, v, 'AttachEnergy', (m) => m.target === 0)) ?? v;
        if (me.active!.damage > 0 && espeon) v = (await playLegal(api, v, 'Retreat')) ?? v;
        for (let i = 0; i < 4 && v.pending?.player === 0; i++) v = (await playLegal(api, v, 'RetreatDiscard')) ?? (await playLegal(api, v, 'ChooseOption')) ?? v;
      }
      if (espeon && v.players[0].active?.card.id === EEVEE) v = (await playLegal(api, v, 'Evolve', (m) => m.handCardId === espeon.id)) ?? v;
      await post(api, '/api/move', { type: 'EndTurn', player: 0 });
      v = await runBotUntilYourTurn(api);
    }
  }
  throw new Error('no seed gets a damaged Espeon ex Active beside another damaged Pokémon');
}

async function tap(page: Page, l: Locator) {
  const b = (await l.boundingBox())!;
  await page.mouse.click(b.x + b.width / 2, b.y + b.height / 2);
}

// Card images are blocked in the sandbox: answer them with a 1px gif.
const GIF = Buffer.from('R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7', 'base64');

test('aiming an Ability at the lit Active heals it and does not reopen its card', async ({ page, request }) => {
  await page.route(/digitaloceanspaces/, (r) => r.fulfill({ contentType: 'image/gif', body: GIF }));
  await toEspeonAiming(request);
  const before = ((await (await request.get('/api/state')).json()) as View).players[0].active!.damage;
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });

  await tap(page, page.getByTestId('active-card'));
  const sheet = page.getByTestId('active-sheet');
  await expect(sheet.getByTestId('sheet-ability')).toContainText('Psychic Healing');
  await sheet.getByTestId('sheet-ability').click();

  const lit = page.getByTestId('my-active').locator('[data-lit="true"][data-pickable="true"]');
  await expect(lit).toBeVisible();
  await tap(page, lit);

  await expect(page.getByTestId('choice-prompt')).toHaveCount(0);
  await expect(sheet).toHaveCount(0);
  const after = ((await (await request.get('/api/state')).json()) as View).players[0].active!.damage;
  expect(after).toBe(Math.max(0, before - 30));
});
