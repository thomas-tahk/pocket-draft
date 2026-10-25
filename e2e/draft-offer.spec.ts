import { test, expect } from '@playwright/test';
import { generateOffer } from '../src/draft/offer';
import { indexByTier } from '../src/draft/tiers';
import type { Card } from '../src/types/card';

// The TS offer generator (what the SPA's manual draft uses), over 20 seeded
// drafts of a synthetic pool. Mirrors the Go tests in server/draft_offer_test.go.

const TYPES = ['Fire', 'Water', 'Grass', 'Lightning'] as const;
const RARITIES = ['◊◊◊◊', '◊◊◊', '◊◊', '◊'] as const;

const pool: Card[] = TYPES.flatMap((type) =>
  RARITIES.flatMap((draftRarity) =>
    [0, 1, 2, 3, 4, 5].map(
      (i) => ({ id: `${type}-${draftRarity}-${i}`, cardType: type, draftRarity, isEx: draftRarity === '◊◊◊◊' }) as Card,
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
