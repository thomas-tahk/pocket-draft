import { test, expect, type Page, type Request } from '@playwright/test';
import { draftTwenty } from './helpers/draft';

// #127: at deckbuild the player picks the deck's Energy Zone types (1 to 3, starting
// from the deck's Pokémon types). The pick goes to /api/new with the deck, and the
// battle's Energy Zone then shows only those types.

const isNewGame = (req: Request) => req.url().endsWith('/api/new') && req.method() === 'POST';
const chip = (page: Page, type: string) => page.getByTestId('energy-picker').getByRole('button', { name: type, exact: true });
const ALL = ['Grass', 'Fire', 'Water', 'Lightning', 'Psychic', 'Fighting', 'Darkness', 'Metal'];

// Get the picker to exactly `want`, unpressing extras first so the 3-type cap never blocks a press.
async function choose(page: Page, want: string[]) {
  const pressed = async (t: string) => (await chip(page, t).getAttribute('aria-pressed')) === 'true';
  for (const t of ALL) {
    if (want.includes(t) || !(await pressed(t))) continue;
    const count = (await Promise.all(ALL.map(pressed))).filter(Boolean).length;
    // the last pressed type cannot be unpressed: press a wanted one first
    if (count === 1) await chip(page, want.find((w) => w !== t)!).click();
    await chip(page, t).click();
  }
  for (const t of want) if (!(await pressed(t))) await chip(page, t).click();
}

test('the Energy Zone picker reaches every battle: only Water or Fire show', async ({ page }) => {
  test.setTimeout(150_000);
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  await draftTwenty(page);
  await page.getByRole('button', { name: 'Skip shop' }).click();
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();

  // Fill the deck first so the default (from the deck's Pokémon types) is on show.
  const drafted = page.locator('section', { has: page.getByRole('heading', { name: /^Your draft/ }) });
  const plus = drafted.getByRole('button', { name: '+' });
  for (let i = 0; i < (await plus.count()); i++) {
    for (let c = 0; c < 2 && (await plus.nth(i).isEnabled()); c++) await plus.nth(i).click();
  }
  const picker = page.getByTestId('energy-picker');
  await expect(picker).toBeVisible();
  const pressedCount = await picker.locator('button[aria-pressed="true"]').count();
  expect(pressedCount).toBeGreaterThanOrEqual(1);
  expect(pressedCount).toBeLessThanOrEqual(3);
  for (const b of await picker.getByRole('button').all()) expect((await b.boundingBox())!.height).toBeGreaterThanOrEqual(44);

  await choose(page, ['Water', 'Fire']);
  await expect(picker.locator('button[aria-pressed="true"]')).toHaveCount(2);
  await expect(picker).toContainText('Energy Zone (2 / 3)');

  await picker.scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'test-results/energy-picker-390x844.png' });
  await page.setViewportSize({ width: 1440, height: 900 });
  await picker.scrollIntoViewIfNeeded();
  await page.screenshot({ path: 'test-results/energy-picker-1440x900.png' });
  await page.setViewportSize({ width: 390, height: 844 });

  await expect(page.getByText('Deck: 20 / 20')).toBeVisible();
  await page.getByRole('button', { name: 'Finalize deck' }).click();

  const started = page.waitForRequest(isNewGame);
  await page.getByRole('button', { name: /Next battle/ }).click();
  const sent = (await started).postDataJSON() as { energy?: string[] };
  expect([...(sent.energy ?? [])].sort()).toEqual(['Fire', 'Water']);

  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('setup-intro')).toHaveCount(0, { timeout: 10_000 });

  const setUp = () => page.evaluate(async () => {
    const { useGameStore } = await import('/src/game/store.ts');
    const g = useGameStore.getState();
    const basic = g.view!.players[0].hand.find((c) => c.stage === 'Basic')!;
    await g.dispatch({ type: 'SetupActive', player: 0, cardId: basic.id });
    // Bench every other Basic too, so the opponent cannot end the game on an early knock-out.
    for (let i = 0; i < 3; i++) {
      const move = useGameStore.getState().view!.players[0].legalMoves.find((m) => m.type === 'SetupBench');
      if (move) await useGameStore.getState().dispatch(move);
    }
    await useGameStore.getState().dispatch({ type: 'SetupReady', player: 0 });
  });
  // Set up through the screen's own store, then end turn repeatedly; the screen plays the bot.
  await setUp();


  const seen = new Set<string>();
  let turns = 0;
  for (let round = 0; round < 60 && turns < 6; round++) {
    const state = await page.evaluate(async () => {
      const { useGameStore } = await import('/src/game/store.ts');
      const g = useGameStore.getState();
      const v = g.view!;
      const me = v.players[0];
      if (v.phase === 'over') return { over: true };
      if (v.pending && v.pending.player === 0) {
        const m = me.legalMoves[0];
        if (m) await g.dispatch(m);
        return { acted: true };
      }
      const myTurn = me.legalMoves.some((m) => m.type === 'EndTurn');
      return { myTurn, zone: me.energyZone, next: me.energyNext, types: me.energyTypes };
    });
    if (state.over) {
      // A short game: Rematch keeps the run's types, and play on until enough turns are seen.
      const rematch = page.getByTestId('end-screen').getByRole('button', { name: 'Rematch' });
      if (turns >= 4 || !(await rematch.isVisible())) break;
      await rematch.click();
      await expect(page.getByTestId('setup-intro')).toHaveCount(0, { timeout: 10_000 });
      await setUp();
      continue;
    }
    if (!state.myTurn) {
      await page.waitForTimeout(300);
      continue;
    }

    expect([...(state.types as string[])].sort()).toEqual(['Fire', 'Water']);
    for (const id of ['energy-current', 'energy-next']) {
      const t = await page.getByTestId('my-energy').getByTestId(id).getAttribute('data-type');
      if (t) {
        expect(['Water', 'Fire'], `${id} on turn ${turns + 1}`).toContain(t);
        seen.add(t);
      }
    }
    expect(['Water', 'Fire', '']).toContain(state.zone);
    expect(['Water', 'Fire', '']).toContain(state.next);
    turns++;
    await page.evaluate(async () => {
      const { useGameStore } = await import('/src/game/store.ts');
      await useGameStore.getState().dispatch({ type: 'EndTurn', player: 0 });
    });
  }
  expect(turns).toBeGreaterThanOrEqual(4);
  expect(seen.size).toBeGreaterThan(0);

  // The pick is stored with the run, so every later battle sends it too.
  const saved = await page.evaluate(() => JSON.parse(localStorage.getItem('pocket-draft-v0.4')!).state.run.energy as string[]);
  expect([...saved].sort()).toEqual(['Fire', 'Water']);
});
