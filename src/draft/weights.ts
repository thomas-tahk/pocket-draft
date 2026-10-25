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

// A candidate that works with the picks (evolution line, or a Trainer naming a
// picked Pokémon) has its weight multiplied by this, once, never stacked.
// Energy match is not a synergy rule; type weighting already covers it.
// Mirrored by synergyMult in server/draft_weights.go.
export const SYNERGY_MULT = 3;

export type Synergy = {
  names: Set<string>; // picked Pokémon names
  parents: Set<string>; // names the picked Pokémon evolve from
  text: RegExp | null; // a picked Pokémon's base name inside Trainer text
};

export const NO_SYNERGY: Synergy = { names: new Set(), parents: new Set(), text: null };

const escapeRegExp = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

export function buildSynergy(picks: Card[]): Synergy {
  const names = new Set<string>();
  const parents = new Set<string>();
  const needles: string[] = [];
  for (const p of picks) {
    if (p.cardType === 'Trainer') continue;
    names.add(p.name);
    if (p.evolvesFrom) parents.add(p.evolvesFrom);
    needles.push(escapeRegExp(p.name.replace(/ ex$/, '')));
  }
  const text =
    needles.length > 0 ? new RegExp(`(^|[^\\p{L}])(${needles.join('|')})($|[^\\p{L}])`, 'u') : null;
  return { names, parents, text };
}

function matchesSynergy(card: Card, s: Synergy): boolean {
  if (card.cardType === 'Trainer') return s.text !== null && s.text.test(card.trainerText ?? '');
  return (!!card.evolvesFrom && s.names.has(card.evolvesFrom)) || s.parents.has(card.name);
}

// Sampling weight for one candidate given the running tally.
// Trainers and Colorless Pokémon: flat 1 (always equally eligible per rules.md).
// A synergy match multiplies the result once.
export function weightOf(card: Card, tally: TypeTally, synergy: Synergy = NO_SYNERGY): number {
  const base =
    card.cardType === 'Trainer' || card.cardType === 'Colorless'
      ? 1
      : 1 + TYPE_WEIGHT_COEF * (tally[card.cardType] ?? 0);
  return matchesSynergy(card, synergy) ? base * SYNERGY_MULT : base;
}
