import type { Card } from '../types/card';
import { SLOT_TIER_WEIGHTS, type Tier } from './tiers';
import { buildSynergy, tallyTypes, weightOf, type Synergy } from './weights';

type Rng = () => number;

const SLOTS_PER_PACK = 5;
const TIER_ORDER: Tier[] = [1, 2, 3, 4, 5];

function pickWeighted<T>(items: T[], weights: number[], rng: Rng): T {
  const total = weights.reduce((s, w) => s + w, 0);
  let r = rng() * total;
  for (let i = 0; i < items.length; i++) {
    r -= weights[i];
    if (r <= 0) return items[i];
  }
  return items[items.length - 1];
}

function rollTier(rng: Rng): Tier {
  const weights = TIER_ORDER.map((t) => SLOT_TIER_WEIGHTS[t]);
  return pickWeighted(TIER_ORDER, weights, rng);
}

function pickFromTier(
  pool: Card[],
  tally: ReturnType<typeof tallyTypes>,
  synergy: Synergy,
  exclude: Set<string>,
  rng: Rng,
): Card | null {
  const eligible = pool.filter((c) => !exclude.has(c.id));
  if (eligible.length === 0) return null;
  const weights = eligible.map((c) => weightOf(c, tally, synergy));
  return pickWeighted(eligible, weights, rng);
}

export type OfferOptions = {
  isPack1: boolean; // pack 1's last slot is always tier 1 (ex / mega-ex)
  picks: Card[];
  rng?: Rng;
};

// Generates one 5-card offer. Within an offer, no duplicate ids (you don't see
// the same card twice in a single row of 5); duplicates across packs are allowed.
//
// Pack 1 always offers an ex: its last slot is rolled as tier 1 (falling through
// to other tiers only if tier 1 has nothing eligible).
export function generateOffer(
  byTier: Record<Tier, Card[]>,
  { isPack1, picks, rng = Math.random }: OfferOptions,
): Card[] {
  const tally = tallyTypes(picks);
  const synergy = buildSynergy(picks);
  const offer: Card[] = [];
  const used = new Set<string>();

  for (let slot = 0; slot < SLOTS_PER_PACK; slot++) {
    const targetTier: Tier = isPack1 && slot === SLOTS_PER_PACK - 1 ? 1 : rollTier(rng);

    let card = pickFromTier(byTier[targetTier], tally, synergy, used, rng);

    if (!card) {
      // Tier ran out (e.g. tier 1 has no eligible cards left). Step through other tiers.
      for (const t of TIER_ORDER) {
        if (t === targetTier) continue;
        card = pickFromTier(byTier[t], tally, synergy, used, rng);
        if (card) break;
      }
    }

    if (!card) throw new Error('Card pool exhausted while generating offer');
    offer.push(card);
    used.add(card.id);
  }

  return offer;
}
