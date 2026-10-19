import { test, expect, type APIRequestContext, type Page } from '@playwright/test';
import { readFileSync } from 'node:fs';
import type { GameView, Move } from '../src/game/types';

// The M1 done-gate at phone size: a whole game against the bot, played only
// with gestures on the board. Setup by dragging Basics and pressing Ready;
// each of your turns: evolve and bench by drag, drag the zone's energy onto a
// Pokémon, tap your Active and pick an attack on its card (End Turn when none
// is usable), answer the new-Active choice by tapping, let the bot's turn play
// back (Skip every other turn). Turn 1 allows only a cost-free attack (T5). Nothing for your seat goes through the API or
// the store; /api/state is only read, to decide and to check. Runs once from a
// finished draft (▶ Play deck) and once per seed on /?battle.

const SLOTS = ['left', 'behind', 'right'];
const RESULT = { win: 'Win', loss: 'Loss', tie: 'Tie' } as const;
const REASON = { points: '3 points', no_pokemon: 'No Pokémon left', concede: 'Concede', turn_cap: 'Turn cap' } as const;
const SEEDS = [3, 11, 42, 77];

type Pt = { x: number; y: number };
type Ctx = { page: Page; api: APIRequestContext; label: string; step: string };

const state = async (api: APIRequestContext): Promise<GameView> => (await api.get('/api/state')).json();
const actor = (v: GameView) => (v.pending ? v.pending.player : v.active);
const movesOf = <T extends Move['type']>(v: GameView, type: T) =>
  v.players[0].legalMoves.filter((m): m is Extract<Move, { type: T }> => m.type === type);

// A stuck point is a done-gate failure: keep a screenshot and name the step.
async function stuck(ctx: Ctx, why: string): Promise<never> {
  const path = `test-results/full-game-${ctx.label}-stuck.png`;
  await ctx.page.screenshot({ path });
  const said = await ctx.page.locator('[data-testid="play-reason"], [data-testid="play-hint"]').allTextContents();
  why += `; on screen: ${JSON.stringify(said)}`;
  throw new Error(`[${ctx.label}] stuck at "${ctx.step}": ${why} (screenshot ${path})`);
}

// --- gestures ---------------------------------------------------------------

async function touchDrag(page: Page, from: Pt, to: Pt) {
  const cdp = await page.context().newCDPSession(page);
  const send = (type: 'touchStart' | 'touchMove' | 'touchEnd', p: Pt) => cdp.send('Input.dispatchTouchEvent', { type, touchPoints: type === 'touchEnd' ? [] : [p] });
  await send('touchStart', from);
  for (let i = 1; i <= 8; i++) await send('touchMove', { x: from.x + ((to.x - from.x) * i) / 8, y: from.y + ((to.y - from.y) * i) / 8 });
  await send('touchEnd', to);
  await cdp.detach();
}

async function centre(page: Page, selector: string): Promise<Pt> {
  const b = await page.locator(selector).first().boundingBox();
  if (!b) throw new Error(`no box for ${selector}`);
  return { x: b.x + b.width / 2, y: b.y + b.height / 2 };
}

// Hand cards overlap and tilt, so aim at the middle of the strip where this
// card is on top. Whole pixels only: touch points are truncated to them.
async function handPoint(page: Page, index: number): Promise<Pt> {
  const p = await page.evaluate((i) => {
    const el = document.querySelector(`[data-testid="my-hand"] [data-pick="hand:${i}"]`);
    if (!el) return null;
    const r = el.getBoundingClientRect();
    for (const y of [0.5, 0.3, 0.7].map((f) => Math.round(r.top + r.height * f))) {
      const xs: number[] = [];
      for (let x = Math.ceil(r.left); x < r.right; x++) if (document.elementFromPoint(x, y)?.closest('[data-pick]') === el) xs.push(x);
      if (xs.length > 0) return { x: xs[Math.floor(xs.length / 2)], y };
    }
    return null;
  }, index);
  if (!p) throw new Error(`hand card ${index} has no visible point`);
  return p;
}

const spotSelector = (target: number) =>
  target === 0 ? '[data-testid="my-active"] [data-spot="active"]' : `[data-testid="my-bench"] [data-slot="${SLOTS[target - 1]}"]`;

const handIndex = (v: GameView, id: string) => v.players[0].hand.findIndex((c) => c.id === id);

// Do one gesture and wait for the server to record it as a new action.
async function act(ctx: Ctx, step: string, gesture: () => Promise<void>) {
  ctx.step = step;
  const before = (await state(ctx.api)).actions.length;
  await gesture();
  try {
    await expect.poll(async () => (await state(ctx.api)).actions.length, { timeout: 5000 }).toBeGreaterThan(before);
  } catch {
    await stuck(ctx, 'the gesture made no move on the server');
  }
}

const dragCard = (ctx: Ctx, cardId: string, target: string) =>
  act(ctx, `drag ${cardId} to ${target}`, async () => {
    const i = handIndex(await state(ctx.api), cardId);
    const from = await handPoint(ctx.page, i);
    const to = await centre(ctx.page, target);
    ctx.step += ` (hand index ${i}, from ${JSON.stringify(from)} [${await topAt(ctx.page, from)}] to ${JSON.stringify(to)} [${await topAt(ctx.page, to)}])`;
    await touchDrag(ctx.page, from, to);
  });

// What is on top at a point, for stuck reports.
const topAt = (page: Page, p: Pt) =>
  page.evaluate(({ x, y }) => {
    const el = document.elementFromPoint(x, y);
    return [el?.closest('[data-pick]')?.getAttribute('data-pick'), el?.closest('[data-spot]')?.getAttribute('data-spot'), el?.className].join('|');
  }, p);

// --- invariants -------------------------------------------------------------

// Split the server's action log into turns by its "--- Turn N" markers.
function actionsByTurn(v: GameView): Map<number, GameView['actions']> {
  const out = new Map<number, GameView['actions']>();
  let turn = 0;
  for (const a of v.actions) {
    out.set(turn, [...(out.get(turn) ?? []), a]);
    const marker = a.result.map((l) => /^--- Turn (\d+)/.exec(l)).find(Boolean);
    if (marker) turn = Number(marker[1]);
  }
  return out;
}

// T5: the first player gets no energy on turn 1, so only an attack that costs
// nothing can be used then.
function checkTurnOne(acts: GameView['actions'], where: string) {
  expect(acts.filter((a) => a.kind === 'attach_energy'), `${where}: energy attached on turn 1`).toHaveLength(0);
  for (const a of acts.filter((x) => x.kind === 'attack')) {
    const used = a.card!.attacks.find((atk) => a.result.some((l) => l.includes(`used ${atk.name}`)));
    expect(used, `${where}: turn-1 attack ${a.result[0]}`).toBeTruthy();
    expect(used!.cost, `${where}: turn-1 attack ${used!.name} has a cost`).toHaveLength(0);
  }
}

async function checkInvariants(ctx: Ctx) {
  const { page } = ctx;
  const v = await state(ctx.api);
  const [you, opp] = v.players;
  const where = `[${ctx.label}] turn ${v.turn}`;
  expect(you.hand.length, `${where} hand ≤ 10`).toBeLessThanOrEqual(10);
  expect(opp.handCount, `${where} opponent hand ≤ 10`).toBeLessThanOrEqual(10);
  await expect(page.getByTestId('hand-card'), `${where} hand on screen`).toHaveCount(you.hand.length);
  for (const [zone, p] of [['my-points', you], ['opp-points', opp]] as const) {
    await expect(page.getByTestId(zone).locator('[data-filled="true"]'), `${where} ${zone}`).toHaveCount(p.points);
  }
  for (const [zone, p] of [['my-bench', you], ['opp-bench', opp]] as const) {
    const occupied = p.bench.filter(Boolean).length;
    expect(occupied, `${where} ${zone} ≤ 3`).toBeLessThanOrEqual(3);
    await expect(page.getByTestId(zone).locator('[data-zoom]'), `${where} ${zone} on screen`).toHaveCount(occupied);
  }
  for (const [turn, acts] of actionsByTurn(v)) {
    for (const seat of [0, 1]) {
      const attaches = acts.filter((a) => a.player === seat && a.kind === 'attach_energy').length;
      expect(attaches, `${where}: seat ${seat} attached ${attaches} energy on turn ${turn}`).toBeLessThanOrEqual(1);
    }
    if (turn === 1) checkTurnOne(acts, where);
  }
}

// --- one turn -----------------------------------------------------------------

async function setUp(ctx: Ctx) {
  const { page } = ctx;
  ctx.step = 'opening coin';
  await expect(page.getByTestId('setup-intro')).toBeVisible({ timeout: 10_000 });
  const coin = page.getByTestId('setup-coin');
  if (await coin.isEnabled()) await page.touchscreen.tap(...xy(await centre(page, '[data-testid="setup-coin"]')));
  await expect(page.getByTestId('setup-intro')).toHaveCount(0, { timeout: 6000 });

  const basics = (await state(ctx.api)).players[0].hand.filter((c) => c.stage === 'Basic');
  if (basics.length === 0) await stuck(ctx, 'no Basic in the opening hand');
  await dragCard(ctx, basics[0].id, spotSelector(0));
  for (let i = 1; i < basics.length && i <= 3; i++) await dragCard(ctx, basics[i].id, spotSelector(i));
  const ready = page.getByTestId('setup-ready');
  await expect(ready).toBeEnabled();
  await act(ctx, 'Ready', async () => page.touchscreen.tap(...xy(await centre(page, '[data-testid="setup-ready"]'))));
}

const xy = (p: Pt): [number, number] => [p.x, p.y];

async function chooseNewActive(ctx: Ctx, v: GameView) {
  const pick = movesOf(v, 'ChooseNewActive')[0];
  await expect(ctx.page.getByTestId('choice-prompt')).toHaveText(/Choose your new Active/);
  await act(ctx, 'choose new Active', async () =>
    ctx.page.touchscreen.tap(...xy(await centre(ctx.page, `[data-testid="my-bench"] [data-pos="${pick.benchIndex + 1}"] [data-lit="true"]`))));
}

async function evolveAndBench(ctx: Ctx) {
  for (let i = 0; i < 6; i++) {
    const v = await state(ctx.api);
    const evolve = movesOf(v, 'Evolve')[0];
    if (evolve) { await dragCard(ctx, evolve.handCardId, spotSelector(evolve.target)); continue; }
    const basic = movesOf(v, 'PlayBasic')[0];
    if (!basic) return;
    await dragCard(ctx, basic.cardId, spotSelector(basic.slot + 1));
  }
}

async function attachEnergy(ctx: Ctx) {
  const options = movesOf(await state(ctx.api), 'AttachEnergy');
  if (options.length === 0) return;
  const target = (options.find((m) => m.target === 0) ?? options[0]).target;
  await act(ctx, `drag energy to ${target}`, async () =>
    touchDrag(ctx.page, await centre(ctx.page, '[data-testid="my-energy"] [data-testid="energy-current"]'), await centre(ctx.page, spotSelector(target))));
  const v = await state(ctx.api);
  expect(v.players[0].energyUsed, `[${ctx.label}] energy used`).toBe(true);
  expect(movesOf(v, 'AttachEnergy'), `[${ctx.label}] a second energy offered`).toHaveLength(0);
  await expect(ctx.page.getByTestId('my-energy').getByTestId('energy-current')).toHaveAttribute('data-spent', 'true');
}

// Tap the Active; use its hardest-hitting usable attack, else close the card and End Turn.
async function attackOrEnd(ctx: Ctx) {
  const { page } = ctx;
  const v = await state(ctx.api);
  const attacks = movesOf(v, 'UseAttack');
  const costed = attacks.filter((m) => v.players[0].active!.card.attacks[m.index].cost.length > 0);
  if (v.turn === 1) expect(costed, `[${ctx.label}] costed attack offered on turn 1`).toHaveLength(0);
  ctx.step = 'open the Active card';
  await page.touchscreen.tap(...xy(await centre(page, '[data-testid="active-card"]')));
  const sheet = page.getByTestId('active-sheet');
  await expect(sheet).toBeVisible().catch(() => stuck(ctx, 'tapping the Active opened no card'));
  await expect(sheet.locator('[data-testid="sheet-attack"][data-usable="true"]')).toHaveCount(attacks.length);
  if (attacks.length > 0) {
    const dmg = v.players[0].active!.card.attacks;
    const best = attacks.reduce((a, b) => (dmg[b.index].damage > dmg[a.index].damage ? b : a));
    await act(ctx, `attack ${dmg[best.index].name}`, async () => page.touchscreen.tap(...xy(await centre(page, `[data-testid="sheet-attack"] >> nth=${best.index}`))));
    return;
  }
  await page.touchscreen.tap(195, 20); // off the card: closes it
  await expect(sheet).toHaveCount(0);
  await act(ctx, 'End Turn', async () => {
    await page.touchscreen.tap(...xy(await centre(page, '[data-testid="end-turn"] button')));
    const confirm = page.getByTestId('confirm');
    if (await confirm.isVisible({ timeout: 500 }).catch(() => false)) await page.getByTestId('confirm-ok').click();
  });
}

async function yourTurn(ctx: Ctx) {
  await expect(ctx.page.getByTestId('end-turn').getByRole('button')).toBeEnabled({ timeout: 10_000 });
  await checkInvariants(ctx);
  await evolveAndBench(ctx);
  await attachEnergy(ctx);
  await attackOrEnd(ctx);
}

// Wait until the game needs you (your turn, your choice) or is over. The bot
// plays itself from the screen; every other bot turn is skipped.
async function waitForYou(ctx: Ctx, skip: boolean): Promise<GameView> {
  ctx.step = 'bot turn';
  const skipButton = ctx.page.getByRole('button', { name: 'Skip' });
  if (skip && (await skipButton.isVisible({ timeout: 1500 }).catch(() => false))) await skipButton.click().catch(() => {});
  let v = await state(ctx.api);
  const deadline = Date.now() + 60_000;
  while (v.phase !== 'over' && actor(v) !== 0) {
    if (Date.now() > deadline) await stuck(ctx, "the bot's turn never handed back");
    await ctx.page.waitForTimeout(200);
    v = await state(ctx.api);
  }
  return v;
}

async function playToEnd(ctx: Ctx) {
  await setUp(ctx);
  for (let round = 0; round < 60; round++) {
    const v = await waitForYou(ctx, round % 2 === 1);
    if (v.phase === 'over') return;
    if (v.pending?.kind === 'new_active') await chooseNewActive(ctx, v);
    else if (v.pending) await stuck(ctx, `a ${v.pending.kind} prompt the plan doesn't answer`);
    else await yourTurn(ctx);
  }
  await stuck(ctx, 'no end after 60 of your moves');
}

async function checkEnd(ctx: Ctx) {
  const v = await state(ctx.api);
  expect(v.phase).toBe('over');
  const end = ctx.page.getByTestId('end-screen');
  await expect(end).toBeVisible();
  await expect(end.getByTestId('end-result')).toHaveText(RESULT[v.outcome as keyof typeof RESULT]);
  await expect(end.getByTestId('end-reason')).toHaveText(REASON[v.reason as keyof typeof REASON]);
  await expect(end.getByTestId('end-score')).toHaveText(`You ${v.players[0].points} – ${v.players[1].points} Opponent`);
  expect(v.reason, `[${ctx.label}] ended by concede`).not.toBe('concede');
  expect(v.outcome).toBe(v.winner === -1 ? 'tie' : v.winner === 0 ? 'win' : 'loss');
  if (v.reason === 'points') expect(v.players[v.winner].points).toBeGreaterThanOrEqual(3);
  if (v.reason === 'turn_cap') expect(v.turn).toBe(50);
  await checkInvariants(ctx);
  await ctx.page.screenshot({ path: `test-results/full-game-${ctx.label}-end.png` });
  const yours: Record<string, number> = {};
  for (const a of v.actions) if (a.player === 0) yours[a.kind] = (yours[a.kind] ?? 0) + 1;
  console.log(`[${ctx.label}] ${v.outcome} by ${v.reason} on turn ${v.turn}, ${v.players[0].points}-${v.players[1].points}; your moves ${JSON.stringify(yours)}`);
}

for (const seed of SEEDS) {
  test(`a whole game by gestures only, seed ${seed}`, async ({ page, request }) => {
    test.setTimeout(600_000);
    await request.post('/api/new', { data: { seed } });
    await page.goto('/?battle');
    const ctx: Ctx = { page, api: request, label: `seed-${seed}`, step: 'open' };
    await playToEnd(ctx);
    await checkEnd(ctx);
  });
}

// The demo deck is all Basics, so evolving needs a deck with Stage 1 lines.
// Water lines from the draft pool, 2 of each. Seeds found by replaying this
// policy through the API: 20 and 17 evolve, bench and promote by choice; on 83
// the bot's Munchlax draws past a 10-card hand unless the T3 cap holds.
const EVOLUTION_DECK = ['A1-074', 'A1-075', 'A1a-020', 'A2-043', 'A1-062', 'A2b-015', 'B1-048', 'A1-058', 'A4-049', 'A2a-015'].flatMap((id) => [id, id]);

for (const seed of [20, 17, 83]) {
  test(`a whole game by gestures only, evolution deck, seed ${seed}`, async ({ page, request }) => {
    test.setTimeout(600_000);
    expect((await request.post('/api/new', { data: { seed, you: EVOLUTION_DECK } })).ok()).toBe(true);
    await page.goto('/?battle');
    const ctx: Ctx = { page, api: request, label: `evolve-seed-${seed}`, step: 'open' };
    await playToEnd(ctx);
    await checkEnd(ctx);
    const yours = (await state(request)).actions.filter((a) => a.player === 0).map((a) => a.kind);
    expect(yours).toContain('evolve');
    if (seed !== 83) expect(yours).toContain('promote');
  });
}

// --- from a draft ---------------------------------------------------------------

type Raw = { id: string; name: string; stage: string };
const cardById = new Map((JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as Raw[]).map((c) => [c.id, c]));
const idOf = (title: string) => title.split(' · ').pop() ?? '';

// Basics and new names first, deck from drafted cards only, so all 20 play
// (same draft as draft-to-battle.spec.ts).
async function draftAndBuild(page: Page) {
  const names = new Set<string>();
  for (let pack = 1; pack <= 16; pack++) {
    await expect(page.getByText(`Pack ${pack} / 16`)).toBeVisible();
    const offer = page.locator('main button[title]');
    await expect(offer).toHaveCount(5);
    const ids = (await offer.evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''))).map(idOf);
    const fresh = ids.filter((id) => !names.has(cardById.get(id)!.name));
    const id = fresh.find((i) => cardById.get(i)!.stage === 'Basic') ?? fresh[0] ?? ids[0];
    names.add(cardById.get(id)!.name);
    await offer.nth(ids.indexOf(id)).click();
  }
  await page.getByRole('button', { name: 'Skip shop' }).click();
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
  const plus = page.locator('section', { has: page.getByRole('heading', { name: /^Your draft/ }) }).getByRole('button', { name: '+' });
  for (let i = 0; i < (await plus.count()); i++) {
    for (let copy = 0; copy < 2 && (await plus.nth(i).isEnabled()); copy++) await plus.nth(i).click();
  }
  await expect(page.getByText('Deck: 20 / 20')).toBeVisible();
  await page.getByRole('button', { name: 'Finalize deck' }).click();
}

test('a whole game by gestures only, from a finished draft', async ({ page, request }) => {
  test.setTimeout(600_000);
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  await draftAndBuild(page);
  await page.getByRole('button', { name: /Play deck/ }).click();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  const ctx: Ctx = { page, api: request, label: 'draft', step: 'open' };
  await playToEnd(ctx);
  await checkEnd(ctx);
});
