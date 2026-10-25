import type { Card, EnergyType } from '../types/card';

export type TypeTally = Partial<Record<EnergyType, number>>;

// Each drafted on-type Pokémon nudges that type's weight by this much.
// weight = 1 + COEF × typeCount[type] for energy Pokémon; flat 1 for Trainers
// and Colorless.
//
// Back on (1) per the 2026-10-15 rework answers: offers lean toward the types
// you've picked. Mirrored by typeWeightCoef in server/draft_weights.go.
export const TYPE_WEIGHT_COEF = 1;

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
