import { test, expect } from '@playwright/test';
import { generateOffer } from '../src/draft/offer';
import { indexByTier } from '../src/draft/tiers';
import { buildSynergy, weightOf, SYNERGY_MULT, TYPE_SUPPORT_COEF, TYPE_WEIGHT_COEF } from '../src/draft/weights';
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

test('text naming a picked type leans a Trainer or boosts a Pokémon; costs and Colorless do not', () => {
  const picks = [mon('Squirtle', 'Water')];
  const syn = buildSynergy(picks);
  const tally = { Water: 1 };
  for (const text of ['Choose 1 of your [W] Pokémon.', 'Attach a {W} Energy.', 'Heal your Water Pokémon.', 'Discard a Water Energy.'])
    expect(weightOf(trainer('T', text), tally, syn), text).toBe(1 + TYPE_SUPPORT_COEF);
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
  const tally = { Dragon: 1 };
  expect(weightOf(trainer('Lance', 'Choose 1 of your [N] Pokémon.'), tally, syn)).toBe(1 + TYPE_SUPPORT_COEF);
  expect(weightOf(trainer('Blaine', 'Choose 1 of your [R] Pokémon.'), tally, syn)).toBe(1);
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

// Type lean (#125). Mirrors server/draft_type_lean_test.go.
test('type support Trainers lean with the picks, by text or by the table; opposing ones do not', () => {
  const marlon = trainer('Marlon', 'Heal 70 damage from 1 of your Carracosta or Jellicent.');
  expect(weightOf(marlon, { Water: 3 }, buildSynergy([]))).toBe(1 + TYPE_SUPPORT_COEF * 3);
  expect(weightOf(marlon, { Fire: 3 }, buildSynergy([]))).toBe(1);
  const bottle = trainer('Squirt Bottle', "Discard a [R] Energy from your opponent's Active Pokémon.");
  expect(weightOf(bottle, { Fire: 3 }, buildSynergy([]))).toBe(1);
  const misty = trainer('Misty', 'Choose 1 of your [W] Pokémon.');
  expect(weightOf(misty, { Water: 3 }, buildSynergy([]))).toBe(1 + TYPE_SUPPORT_COEF * 3);
  expect(weightOf(mon('Squirtle', 'Water'), { Water: 3 }, buildSynergy([]))).toBe(1 + TYPE_WEIGHT_COEF * 3);
});

test('3 Water picks lean Water Pokémon but do not force them', () => {
  const water = pool.filter((c) => c.cardType === 'Water').slice(0, 3);
  const share = (picks: Card[], type: string) => {
    let n = 0;
    let total = 0;
    for (let seed = 1; seed <= 20; seed++) {
      const rng = seeded(seed);
      for (let i = 0; i < 50; i++)
        for (const c of generateOffer(byTier, { isPack1: false, picks, rng })) {
          total++;
          if (c.cardType === type) n++;
        }
    }
    return n / total;
  };
  expect(share(water, 'Water')).toBeGreaterThan(share([], 'Water') * 1.5);
  expect(share(water, 'Fire')).toBeGreaterThan(0);
});
