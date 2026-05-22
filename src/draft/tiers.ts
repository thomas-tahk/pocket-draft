import type { Card } from '../types/card';

export type Tier = 1 | 2 | 3 | 4 | 5;

// Tier 1 = ex / mega-ex Pokémon (the pack-1 anchor target)
// Tier 2 = ◊◊◊◊ non-ex Pokémon and Trainers
// Tier 3..5 = ◊◊◊ / ◊◊ / ◊
export function tierOf(card: Card): Tier {
  if (card.draftRarity === '◊◊◊◊') return card.isEx ? 1 : 2;
  if (card.draftRarity === '◊◊◊') return 3;
  if (card.draftRarity === '◊◊') return 4;
  return 5;
}

// Per-slot tier distribution for ordinary slots (packs 2-20, and pack 1 slots 1-4).
// Tunable. Sums to 1.0.
//
// Note: tier 2 is intentionally 0 — Pocket reserves ◊◊◊◊ exclusively for EX/Mega-EX,
// so there are no ◊◊◊◊ non-ex Pokémon in the actual card pool. Rules.md flags a future
// "high-impact Trainers" promotion into tier 2; deferred until we tag those cards.
//
// ~34% of packs offer at least one tier-1 card (1 - 0.92^5), so EX/Mega-EX feels rare
// but actively tempting across a 20-pack draft.
export const SLOT_TIER_WEIGHTS: Record<Tier, number> = {
  5: 0.5,
  4: 0.25,
  3: 0.17,
  2: 0,
  1: 0.08,
};

export function indexByTier(pool: Card[]): Record<Tier, Card[]> {
  const idx: Record<Tier, Card[]> = { 1: [], 2: [], 3: [], 4: [], 5: [] };
  for (const c of pool) idx[tierOf(c)].push(c);
  return idx;
}
