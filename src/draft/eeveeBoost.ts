import type { Card } from '../types/card';

// Eevee ex ("Veevee ’volve") can evolve into any Eeveelution, so the two raise
// each other's odds, both directions (owner 2026-10-25, "Eevee ex becomes
// draftable"; 2026-10-28, decision 4): once you have picked an Eeveelution, Eevee
// ex gets the same ×3 a synergy match gets, and once you have picked Eevee ex,
// every Eeveelution does. Raised odds, never a guarantee, never stacked with
// another synergy match. Mirrored by eeveeBoost in server/draft_eevee.go.

export const EEVEE_EX = 'Eevee ex';

// Same value as SYNERGY_MULT (weights.ts), kept here so the two files don't import each other.
export const EEVEE_MULT = 3;

// An Eeveelution: a Pokémon that evolves from Eevee.
export const isEeveelution = (card: Card): boolean => card.cardType !== 'Trainer' && card.evolvesFrom === 'Eevee';

// What the picks hold, for the boost (a subset of weights.ts's Synergy).
export type EeveePicks = { eeveelution?: boolean; eeveeEx?: boolean };

// The multiplier on one candidate: ×EEVEE_MULT for Eevee ex once an Eeveelution
// is among the picks, or for an Eeveelution once Eevee ex is; else 1.
// `alreadyBoosted` (the candidate already got its synergy ×3) keeps the boost
// from stacking.
export function eeveeBoost(card: Card, picks: EeveePicks, alreadyBoosted: boolean): number {
  if (alreadyBoosted) return 1;
  if (picks.eeveelution && card.name === EEVEE_EX) return EEVEE_MULT;
  if (picks.eeveeEx && isEeveelution(card)) return EEVEE_MULT;
  return 1;
}
