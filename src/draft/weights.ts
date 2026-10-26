import type { Card, EnergyType } from '../types/card';
import { NOT_TYPE_SUPPORT, TYPE_SUPPORT_TABLE } from './typeSupport';

export type TypeTally = Partial<Record<EnergyType, number>>;

// Each drafted on-type Pokémon nudges that type's weight by this much.
// weight = 1 + COEF × typeCount[type] for energy Pokémon; flat 1 for Colorless.
// Raised from 1 to 2 so the lean shows after 2-3 picks, not ~8 (owner,
// 2026-10-25, point 3); measured odds are in server/draft_type_lean_test.go.
// Mirrored by typeWeightCoef in server/draft_weights.go.
export const TYPE_WEIGHT_COEF = 2;

// The same lean for a Trainer/Item that supports a type: weight = 1 + COEF ×
// (picks of the most-picked type it supports), flat 1 if it supports none.
// Mirrored by typeSupportCoef in server/draft_weights.go.
export const TYPE_SUPPORT_COEF = 2;

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
//   2. a candidate Pokémon's effect text names a type you have picked Pokémon of
//      (a Trainer doing so is type support instead: weightOf's lean);
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

// Types a card's text names, minus its own type: a card naming its own type is
// the type weight's job, so it never counts for rule 2 (as a candidate) or
// rule 3 (as a pick). Trainers have no type, so they are unaffected. Cached per
// card id; a card's text never changes.
const namedTypesCache = new Map<string, EnergyType[]>();

function namedTypes(card: Card): EnergyType[] {
  let types = namedTypesCache.get(card.id);
  if (!types) {
    types = typesNamedIn(effectText(card)).filter((t) => t !== card.cardType);
    namedTypesCache.set(card.id, types);
  }
  return types;
}

// Types a Trainer/Item supports: those its text names (a Trainer has no type of
// its own, so the own-type trap cannot apply) plus TYPE_SUPPORT_TABLE, minus
// NOT_TYPE_SUPPORT. Pokémon support nothing. Mirrored by supportTypes in
// server/draft_weights.go. Cached per card id.
const supportTypesCache = new Map<string, EnergyType[]>();

export function supportedTypes(card: Card): EnergyType[] {
  if (card.cardType !== 'Trainer' || NOT_TYPE_SUPPORT.has(card.name ?? '')) return [];
  let types = supportTypesCache.get(card.id);
  if (!types) {
    types = [...new Set([...namedTypes(card), ...(TYPE_SUPPORT_TABLE[card.name ?? ''] ?? [])])];
    supportTypesCache.set(card.id, types);
  }
  return types;
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
    for (const t of namedTypes(p)) textTypes.add(t);
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
  if (card.cardType === 'Trainer') return false; // naming a picked type is type support, see weightOf
  if (s.textTypes.has(card.cardType as EnergyType)) return true;
  return namedTypes(card).some((t) => s.pickedTypes.has(t));
}

// Sampling weight for one candidate given the running tally.
// Colorless Pokémon: flat 1 (always equally eligible per rules.md).
// Other Pokémon: 1 + TYPE_WEIGHT_COEF × picks of their type.
// Trainers: 1 + TYPE_SUPPORT_COEF × picks of the most-picked type they support.
// A synergy match then multiplies the result once.
export function weightOf(card: Card, tally: TypeTally, synergy: Synergy = NO_SYNERGY): number {
  let base = 1;
  if (card.cardType === 'Trainer') {
    const most = Math.max(0, ...supportedTypes(card).map((t) => tally[t] ?? 0));
    base = 1 + TYPE_SUPPORT_COEF * most;
  } else if (card.cardType !== 'Colorless') {
    base = 1 + TYPE_WEIGHT_COEF * (tally[card.cardType] ?? 0);
  }
  return matchesSynergy(card, synergy) ? base * SYNERGY_MULT : base;
}
