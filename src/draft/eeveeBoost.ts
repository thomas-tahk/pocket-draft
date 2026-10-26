import type { Card } from '../types/card';

// Eevee ex ("Veevee ’volve") can evolve into any Eeveelution, so once you have
// picked one, its odds rise by the same ×3 a synergy match gets (owner
// 2026-10-25, "Eevee ex becomes draftable"). Raised odds, never a guarantee.
// Mirrored by eeveeBoost in server/draft_eevee.go.

export const EEVEE_EX = 'Eevee ex';

// Same value as SYNERGY_MULT (weights.ts), kept here so the two files don't import each other.
export const EEVEE_MULT = 3;

// An Eeveelution: a Pokémon that evolves from Eevee.
export const isEeveelution = (card: Card): boolean => card.cardType !== 'Trainer' && card.evolvesFrom === 'Eevee';

// The multiplier on one candidate: ×EEVEE_MULT for Eevee ex once an Eeveelution
// is among the picks, else 1. `alreadyBoosted` (the candidate already got its
// synergy ×3) keeps the boost from stacking.
export function eeveeBoost(card: Card, eeveelutionPicked: boolean, alreadyBoosted: boolean): number {
  return eeveelutionPicked && !alreadyBoosted && card.name === EEVEE_EX ? EEVEE_MULT : 1;
}
