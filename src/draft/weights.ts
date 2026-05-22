import type { Card, EnergyType } from '../types/card';

export type TypeTally = Partial<Record<EnergyType, number>>;

// Each drafted on-type Pokémon nudges that type's weight by this much.
// Higher = more direction, lower = more chaos. Rules.md: "a nudge, not a lock".
export const TYPE_WEIGHT_COEF = 0.5;

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
