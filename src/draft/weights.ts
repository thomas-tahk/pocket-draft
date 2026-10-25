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

// A candidate that works with the picks has its weight multiplied by this, once,
// never stacked, however many rules match. Rules:
//   1. a Trainer's text names a picked Pokémon;
//   2. a candidate's effect text names a type you have picked Pokémon of;
//   3. a candidate Pokémon's type is named by the text of one of your picks.
// A Colorless reference never counts. Mirrored by synergyMult in
// server/draft_synergy.go.
export const SYNERGY_MULT = 3;

// Energy symbol letters in card text -> card type ([W], {W}).
const TYPE_SYMBOLS: Record<string, EnergyType> = {
  G: 'Grass', R: 'Fire', W: 'Water', L: 'Lightning', P: 'Psychic',
  F: 'Fighting', D: 'Darkness', M: 'Metal', N: 'Dragon',
};

// Trainer text, ability and attack effects. Attack costs are not text.
function effectText(card: Card): string {
  return [card.trainerText ?? '', card.ability?.effect ?? '', ...(card.attacks ?? []).map((a) => a.effect)].join('\n');
}

// Non-Colorless types a text names: "<Type> Pokémon", "<Type> Energy", [W] or {W}.
function typesNamedIn(text: string): EnergyType[] {
  return Object.entries(TYPE_SYMBOLS)
    .filter(
      ([letter, type]) =>
        text.includes(`[${letter}]`) ||
        text.includes(`{${letter}}`) ||
        text.includes(`${type} Pokémon`) ||
        text.includes(`${type} Energy`),
    )
    .map(([, type]) => type);
}

export type Synergy = {
  pickedTypes: Set<EnergyType>; // types of picked Pokémon (not Colorless)
  textTypes: Set<EnergyType>; // types named in the text of any pick
  nameText: RegExp | null; // a picked Pokémon's base name inside Trainer text
};

export const NO_SYNERGY: Synergy = { pickedTypes: new Set(), textTypes: new Set(), nameText: null };

const escapeRegExp = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

export function buildSynergy(picks: Card[]): Synergy {
  const pickedTypes = new Set<EnergyType>();
  const textTypes = new Set<EnergyType>();
  const needles: string[] = [];
  for (const p of picks) {
    for (const t of typesNamedIn(effectText(p))) textTypes.add(t);
    if (p.cardType === 'Trainer') continue;
    if (p.cardType !== 'Colorless') pickedTypes.add(p.cardType);
    needles.push(escapeRegExp((p.name ?? '').replace(/ ex$/, '')));
  }
  const nameText =
    needles.length > 0 ? new RegExp(`(^|[^\\p{L}])(${needles.join('|')})($|[^\\p{L}])`, 'u') : null;
  return { pickedTypes, textTypes, nameText };
}

function matchesSynergy(card: Card, s: Synergy): boolean {
  if (card.cardType === 'Trainer' && s.nameText?.test(card.trainerText ?? '')) return true;
  if (card.cardType !== 'Trainer' && s.textTypes.has(card.cardType as EnergyType)) return true;
  return typesNamedIn(effectText(card)).some((t) => s.pickedTypes.has(t));
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
