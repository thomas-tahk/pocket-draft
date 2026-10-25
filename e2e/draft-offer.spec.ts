import { test, expect } from '@playwright/test';
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
const mon = (name: string, cardType: string, evolvesFrom: string | null = null) =>
  ({ id: name + cardType, name, cardType, evolvesFrom }) as Card;
const trainer = (name: string, trainerText: string) =>
  ({ id: name, name, cardType: 'Trainer', trainerText }) as Card;

test('evolution line gets the synergy bonus in both directions, once', () => {
  const syn = buildSynergy([mon('Charmeleon', 'Fire', 'Charmander')]);
  const tally = { Fire: 1 };
  expect(weightOf(mon('Charizard', 'Fire', 'Charmeleon'), tally, syn)).toBe(2 * SYNERGY_MULT);
  expect(weightOf(mon('Charmander', 'Fire'), tally, syn)).toBe(2 * SYNERGY_MULT);
  expect(weightOf(mon('Charmander', 'Colorless'), tally, syn)).toBe(SYNERGY_MULT);
  expect(weightOf(mon('Squirtle', 'Fire'), tally, syn)).toBe(2);
  const both = buildSynergy([mon('Charmeleon', 'Fire', 'Charmander'), mon('Charizard', 'Fire', 'Charmeleon')]);
  expect(weightOf(mon('Charmeleon', 'Fire', 'Charmander'), { Fire: 2 }, both)).toBe(3 * SYNERGY_MULT);
});

test('a Trainer naming a picked Pokemon gets the bonus, one that does not does not', () => {
  const syn = buildSynergy([mon('Onix', 'Fighting')]);
  expect(weightOf(trainer('Brock', 'Attach a [F] Energy to Golem or Onix.'), {}, syn)).toBe(SYNERGY_MULT);
  expect(weightOf(trainer('Misty', 'Attach Water Energy.'), {}, syn)).toBe(1);
  expect(weightOf(trainer('Other', 'Attach to Onixia.'), {}, syn)).toBe(1);
  const ex = buildSynergy([mon('Mewtwo ex', 'Psychic')]);
  expect(weightOf(trainer('X', 'Heal your Mewtwo.'), {}, ex)).toBe(SYNERGY_MULT);
});

test('offers contain the line of a pick measurably more often', () => {
  const raichu = [0, 1, 2].map(
    (i) => ({ ...mon('Raichu', 'Fire', 'Pika'), id: `r${i}`, draftRarity: '◊', isEx: false }) as Card,
  );
  const lineTier = indexByTier([...pool, ...raichu]);
  const count = (picks: Card[]) => {
    let n = 0;
    for (let seed = 1; seed <= 20; seed++) {
      const rng = seeded(seed);
      for (let i = 0; i < 100; i++)
        for (const c of generateOffer(lineTier, { isPack1: false, picks, rng })) if (c.name === 'Raichu') n++;
    }
    return n;
  };
  const withPick = count([mon('Pika', 'Fire')]);
  const control = count([mon('Other', 'Fire')]);
  expect(control).toBeGreaterThan(0);
  expect(withPick).toBeGreaterThan(control * 1.5);
});
