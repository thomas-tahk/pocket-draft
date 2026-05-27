import type { Card, EnergyType } from '../types/card';

export type TypeTally = Partial<Record<EnergyType, number>>;

// Each drafted on-type Pokémon nudges that type's weight by this much.
// weight = 1 + COEF × typeCount[type] for energy Pokémon; flat 1 for Trainers
// and Colorless.
//
// 0 means true random draft — no type bias. This was the deliberate choice
// 2026-05-27 once the model moved to "pool-then-pick": picks accumulate into a
// collection rather than locking into a deck, so the random pool plus free
// SWUCUBE-style expansions at deckbuild give players enough material to assemble
// a coherent deck without needing mid-draft type pressure.
export const TYPE_WEIGHT_COEF = 0;

export function tallyTypes(picks: Card[]): TypeTally {
  const t: TypeTally = {};
  for (const p of picks) {
    if (p.cardType === 'Trainer') continue;
    if (p.cardType === 'Colorless') continue;
    t[p.cardType] = (t[p.cardType] ?? 0) + 1;
  }
  return t;
}

// Sampling weight for one candidate given the running tally.
// Trainers and Colorless Pokémon: flat 1 (always equally eligible per rules.md).
export function weightOf(card: Card, tally: TypeTally): number {
  if (card.cardType === 'Trainer' || card.cardType === 'Colorless') return 1;
  return 1 + TYPE_WEIGHT_COEF * (tally[card.cardType] ?? 0);
}
