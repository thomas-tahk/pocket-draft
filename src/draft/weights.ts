import type { Card, EnergyType } from '../types/card';

export type TypeTally = Partial<Record<EnergyType, number>>;

// Each drafted on-type Pokémon nudges that type's weight by this much.
// weight = 1 + COEF × typeCount[type] for energy Pokémon; flat 1 for Trainers
// and Colorless.
//
// 1.0 means a 4-times-picked type is 5× the weight of an unpicked type. Against
// the actual pool this gives ~38% slot probability and ~91% per-pack probability
// for the leading type — enough to feel "mono-energy commitment" rather than
// "soft nudge." Tunable; 1.5 pushes per-pack rate to ~96%.
export const TYPE_WEIGHT_COEF = 1.0;

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
