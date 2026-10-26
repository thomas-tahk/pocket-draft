import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { idOf } from './helpers/draft';
import { generateOffer } from '../src/draft/offer';
import { indexByTier } from '../src/draft/tiers';
import { buildSynergy, weightOf, SYNERGY_MULT } from '../src/draft/weights';
import type { Card } from '../src/types/card';

// The TS offer generator (what the SPA's manual draft uses), over 20 seeded
// drafts of a synthetic pool. Mirrors the Go tests in server/draft_offer_test.go.

const TYPES = ['Fire', 'Water', 'Grass', 'Lightning'] as const;
const RARITIES = ['◊◊◊◊', '◊◊◊', '◊◊', '◊'] as const;

const pool: Card[] = TYPES.flatMap((type) =>
  RARITIES.flatMap((draftRarity) =>
    [0, 1, 2, 3, 4, 5].map(
      (i) => ({ id: `${type}-${draftRarity}-${i}`, name: `${type}-${draftRarity}-${i}`, cardType: type, draftRarity, isEx: draftRarity === '◊◊◊◊' }) as Card,
    ),
  ),
);
const byTier = indexByTier(pool);

// mulberry32: small seeded rng.
function seeded(seed: number): () => number {
  let a = seed;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function fireShare(picks: Card[]): number {
  let onType = 0;
  let total = 0;
  for (let seed = 1; seed <= 20; seed++) {
    const rng = seeded(seed);
    for (let i = 0; i < 10; i++) {
      for (const c of generateOffer(byTier, { isPack1: false, picks, rng })) {
        total++;
        if (c.cardType === 'Fire') onType++;
      }
    }
  }
  return onType / total;
}

test('pack 1 always offers an ex over 20 seeded drafts', () => {
  for (let seed = 1; seed <= 20; seed++) {
    const offer = generateOffer(byTier, { isPack1: true, picks: [], rng: seeded(seed) });
    expect(offer.some((c) => c.isEx), `seed ${seed}`).toBe(true);
  }
});

test('later offers favour the types already picked', () => {
  const fire = pool.filter((c) => c.cardType === 'Fire').slice(0, 8);
  const weighted = fireShare(fire);
  const baseline = fireShare([]);
  expect(weighted).toBeGreaterThan(baseline * 1.5);
});

// Synergy (#109). Mirrors server/draft_synergy_test.go.
const mon = (name: string, cardType: string) => ({ id: name + cardType, name, cardType, attacks: [] }) as unknown as Card;
const trainer = (name: string, trainerText: string) =>
  ({ id: name + trainerText, name, cardType: 'Trainer', trainerText, attacks: [] }) as unknown as Card;

test('a Trainer naming a picked Pokemon gets the bonus, one that does not does not', () => {
  const syn = buildSynergy([mon('Onix', 'Fighting')]);
  expect(weightOf(trainer('Brock', 'Attach an Energy to Golem or Onix.'), {}, syn)).toBe(SYNERGY_MULT);
  expect(weightOf(trainer('Misty', 'Attach an Energy to your Pokémon.'), {}, syn)).toBe(1);
  expect(weightOf(trainer('Other', 'Attach to Onixia.'), {}, syn)).toBe(1);
  const ex = buildSynergy([mon('Mewtwo ex', 'Psychic')]);
  expect(weightOf(trainer('X', 'Heal your Mewtwo.'), {}, ex)).toBe(SYNERGY_MULT);
});

test('text naming a picked type gets the bonus; costs and Colorless do not', () => {
  const picks = [mon('Squirtle', 'Water')];
  const syn = buildSynergy(picks);
  const tally = { Water: 1 };
  for (const text of ['Choose 1 of your [W] Pokémon.', 'Attach a {W} Energy.', 'Heal your Water Pokémon.', 'Discard a Water Energy.'])
    expect(weightOf(trainer('T', text), tally, syn), text).toBe(SYNERGY_MULT);
  expect(weightOf(trainer('Blaine', 'Choose 1 of your [R] Pokémon.'), tally, syn)).toBe(1);
  expect(weightOf(trainer('C', 'Attach a [C] Energy.'), tally, syn)).toBe(1);
  const psyduck = { ...mon('Psyduck', 'Psychic'), attacks: [{ cost: '[W]', name: '[W] Splash', damage: '', effect: '' }] } as unknown as Card;
  expect(weightOf(psyduck, tally, syn)).toBe(1);
  const withEffect = { ...psyduck, id: 'psyduck-effect', attacks: [{ cost: 'W', name: 'X', damage: '', effect: 'Take a [W] Energy.' }] } as unknown as Card;
  expect(weightOf(withEffect, tally, syn)).toBe(SYNERGY_MULT);
});

test('a card naming its own type never counts, as pick or candidate', () => {
  const withText = (name: string, cardType: string, effect: string) =>
    ({ ...mon(name, cardType), id: name, attacks: [{ cost: 'W', name: 'X', damage: '', effect }] }) as unknown as Card;
  // 3b: a Water pick whose text says [W] does not boost other Water Pokémon.
  const waterPick = buildSynergy([withText('Vaporeon', 'Water', 'Take a [W] Energy.')]);
  expect(weightOf(mon('Squirtle', 'Water'), {}, waterPick)).toBe(1);
  const firePick = buildSynergy([withText('Flareon', 'Fire', 'Take a [W] Energy.')]);
  expect(weightOf(mon('Squirtle', 'Water'), {}, firePick)).toBe(SYNERGY_MULT);
  // 3a: a Water candidate whose own text says [W] is not boosted by Water picks.
  const candidate = withText('Kyogre', 'Water', 'Take a [W] Energy.');
  expect(weightOf(candidate, {}, buildSynergy([mon('Squirtle', 'Water')]))).toBe(1);
});

test('a [N] Trainer gets the bonus after a Dragon pick', () => {
  const syn = buildSynergy([mon('Dratini', 'Dragon')]);
  expect(weightOf(trainer('Lance', 'Choose 1 of your [N] Pokémon.'), {}, syn)).toBe(SYNERGY_MULT);
  expect(weightOf(trainer('Blaine', 'Choose 1 of your [R] Pokémon.'), {}, syn)).toBe(1);
});

test('a pick whose text names a type boosts candidates of that type, once', () => {
  const syn = buildSynergy([trainer('Misty', 'Choose 1 of your [W] Pokémon.')]);
  expect(weightOf(mon('Squirtle', 'Water'), {}, syn)).toBe(SYNERGY_MULT);
  expect(weightOf(mon('Charmander', 'Fire'), {}, syn)).toBe(1);
  expect(weightOf(trainer('Other', 'Draw a card.'), {}, syn)).toBe(1);
  const colorless = buildSynergy([trainer('Poké Ball', 'Put a [C] Pokémon in hand.')]);
  expect(weightOf(mon('Meowth', 'Colorless'), {}, colorless)).toBe(1);
  const two = buildSynergy([mon('Squirtle', 'Water')]);
  expect(weightOf(trainer('X', 'Heal your Squirtle and your [W] Pokémon.'), {}, two)).toBe(SYNERGY_MULT);
});

// A pick is secured for the whole draft (#124). Mirrors
// TestBotDraftNeverReoffersAPickedName in server/draft_no_repeat_test.go.
test('generateOffer never offers a picked name, other printings included', () => {
  const picks = pool.slice(0, 20).map((c, i) => ({ ...c, id: `reprint-${i}` }) as Card);
  const names = new Set(picks.map((c) => c.name));
  for (let seed = 1; seed <= 20; seed++) {
    for (const c of generateOffer(byTier, { isPack1: false, picks, rng: seeded(seed) })) {
      expect(names.has(c.name), `seed ${seed}: ${c.name}`).toBe(false);
    }
  }
});

test('a full 20-pick draft in the UI never shows an already-picked name', async ({ page }) => {
  const raw = JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as { id: string; name: string }[];
  const nameById = new Map(raw.map((c) => [c.id, c.name]));
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  const pickedNames = new Set<string>();
  for (let pack = 1; pack <= 20; pack++) {
    await expect(page.getByText(`Pack ${pack} / 20`)).toBeVisible();
    const offer = page.locator('main button[title]');
    await expect(offer).toHaveCount(5);
    const ids = (await offer.evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''))).map(idOf);
    const names = ids.map((id) => nameById.get(id)!);
    for (const name of names) expect(pickedNames.has(name), `pack ${pack} offered ${name}`).toBe(false);
    // Always take the first card: nothing here steers around repeats.
    pickedNames.add(names[0]);
    await offer.nth(0).click();
  }
});

test('Trainers naming the picked type are offered measurably more often', () => {
  const mistys = [0, 1, 2].map(
    (i) => ({ ...trainer('Misty', 'Choose 1 of your [W] Pokémon.'), id: `m${i}`, draftRarity: '◊', isEx: false }) as unknown as Card,
  );
  const tiers = indexByTier([...pool, ...mistys]);
  const count = (picks: Card[]) => {
    let n = 0;
    for (let seed = 1; seed <= 20; seed++) {
      const rng = seeded(seed);
      for (let i = 0; i < 100; i++)
        for (const c of generateOffer(tiers, { isPack1: false, picks, rng })) if (c.name === 'Misty') n++;
    }
    return n;
  };
  const withWater = count([mon('W', 'Water')]);
  const withFire = count([mon('F', 'Fire')]);
  expect(withFire).toBeGreaterThan(0);
  expect(withWater).toBeGreaterThan(withFire * 1.5);
});
