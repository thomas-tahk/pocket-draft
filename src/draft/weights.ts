import type { Card, EnergyType } from '../types/card';
import { nicheGroupsOf, pokemonBaseName } from './nicheSupport';
import { NOT_TYPE_SUPPORT, TYPE_SUPPORT_TABLE } from './typeSupport';
import { EEVEE_EX, eeveeBoost, isEeveelution } from './eeveeBoost';

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

// Picks of one type past this many add no more lean, for Pokémon or Trainers,
// so off-type cards never become near-impossible over a mono-type draft (owner
// 2026-10-28, decision 1). The lean at 2-3 picks is unchanged; the weight tops
// out at 1 + 2 × 4 = 9. Measured odds are in server/draft_type_lean_test.go
// (TestTypeLeanIsCapped). Mirrored by typeLeanCap in server/draft_weights.go.
export const TYPE_LEAN_CAP = 4;

const leanCount = (n: number) => Math.min(n, TYPE_LEAN_CAP);

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
//   3. a candidate Pokémon's type is named by the text of one of your picks;
//   4. the other direction of rule 1 (#126): a candidate Pokémon is named in the
//      text of a Trainer you picked, or a Trainer and a Pokémon share a group in
//      nicheSupport.ts and you picked the other one.
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

// Types a Trainer/Item supports, by Trainer name, as the Go engine worked them
// out from the card text (server/draft_type_support.go): the types the text
// names plus the card types of the Pokémon it names, else the fallback table.
// Loaded from public/data/draft-pool.json by src/data/index.ts.
let registeredSupport: Readonly<Record<string, EnergyType[]>> | null = null;
const supportTypesCache = new Map<string, EnergyType[]>();

// Registers Go's support types (null clears them).
export function setSupportTypes(byName: Readonly<Record<string, EnergyType[]>> | null): void {
  registeredSupport = byName;
  supportTypesCache.clear();
}

// Types a Trainer/Item supports; Pokémon and NOT_TYPE_SUPPORT support nothing.
// A Trainer Go registered gets Go's answer. Any other (a card outside the pool
// file, or nothing registered) gets the part TS can read itself: the types its
// text names (a Trainer has no type of its own, so the own-type trap cannot
// apply), else TYPE_SUPPORT_TABLE. Mirrored by supportTypes in
// server/draft_type_support.go. Cached per card id.
export function supportedTypes(card: Card): EnergyType[] {
  const name = card.name ?? '';
  if (card.cardType !== 'Trainer' || NOT_TYPE_SUPPORT.has(name)) return [];
  let types = supportTypesCache.get(card.id);
  if (!types) {
    const named = namedTypes(card);
    types = registeredSupport?.[name] ?? (named.length > 0 ? named : (TYPE_SUPPORT_TABLE[name] ?? []));
    supportTypesCache.set(card.id, types);
  }
  return types;
}

export type Synergy = {
  pickedTypes: Set<EnergyType>; // types of picked Pokémon (not Colorless)
  textTypes: Set<EnergyType>; // types named in the text of any pick
  pickedNames: string[]; // names of picked Pokémon, matched in Trainer text by textNames
  trainerTexts: string[]; // text of picked Trainers, for the Pokémon they name (rule 4)
  nicheOfTrainers: Set<number>; // niche groups of picked Trainers (rule 4)
  nicheOfPokemon: Set<number>; // niche groups of picked Pokémon (rule 4)
  eeveelution?: boolean; // an Eeveelution is picked (eeveeBoost.ts)
  eeveeEx?: boolean; // Eevee ex is picked (eeveeBoost.ts)
};

export const NO_SYNERGY: Synergy = {
  pickedTypes: new Set(),
  textTypes: new Set(),
  pickedNames: [],
  trainerTexts: [],
  nicheOfTrainers: new Set(),
  nicheOfPokemon: new Set(),
};


export function buildSynergy(picks: Card[]): Synergy {
  const pickedTypes = new Set<EnergyType>();
  const textTypes = new Set<EnergyType>();
  const pickedNames: string[] = [];
  const trainerTexts: string[] = [];
  const nicheOfTrainers = new Set<number>();
  const nicheOfPokemon = new Set<number>();
  for (const p of picks) {
    for (const t of namedTypes(p)) textTypes.add(t);
    if (p.cardType === 'Trainer') {
      trainerTexts.push(p.trainerText ?? '');
      nicheGroupsOf(p).forEach((g) => nicheOfTrainers.add(g));
      continue;
    }
    nicheGroupsOf(p).forEach((g) => nicheOfPokemon.add(g));
    if (p.cardType !== 'Colorless') pickedTypes.add(p.cardType);
    pickedNames.push(p.name ?? '');
  }
  return {
    pickedTypes,
    textTypes,
    pickedNames,
    trainerTexts,
    nicheOfTrainers,
    nicheOfPokemon,
    eeveelution: picks.some(isEeveelution),
    eeveeEx: picks.some((p) => p.name === EEVEE_EX),
  };
}

// Text saying "Alolan Marowak" does not name "Marowak": these precede a Pokémon
// name to make a different, longer card name.
const LONGER_NAME_PREFIXES = ['Alolan ', 'Galarian ', 'Hisuian ', 'Paldean ', "Team Rocket's "];
const isLetter = (ch: string) => /\p{L}/u.test(ch);

// Whether a Trainer's text names the Pokémon card called cardName, as a whole
// word, used the same way in both directions (rule 1 and rule 4). A card named
// "X ex" is named by text saying "X" or "X ex"; a card named "X" is not named by
// text saying "X ex", a different card. An occurrence does not count when it is
// part of a longer card name (preceded by a regional or Team Rocket's prefix) or
// when it only says what something evolves from ("evolves from Eevee" names no
// Eevee ex). Mirrored by textNames in server/draft_synergy.go.
export function textNames(text: string, cardName: string): boolean {
  const word = pokemonBaseName(cardName);
  const isEx = word !== cardName;
  for (let from = 0; from <= text.length; ) {
    const start = text.indexOf(word, from);
    if (start < 0) return false;
    const end = start + word.length;
    from = start + 1;
    if (start > 0 && isLetter([...text.slice(0, start)].pop()!)) continue;
    if (end < text.length && isLetter(String.fromCodePoint(text.codePointAt(end)!))) continue;
    if (!isEx && text.startsWith(' ex', end)) continue;
    const head = text.slice(0, start);
    if (head.endsWith('evolve from ') || head.endsWith('evolves from ')) continue;
    if (LONGER_NAME_PREFIXES.some((p) => head.endsWith(p) && !word.startsWith(p))) continue;
    return true;
  }
  return false;
}

const inSet = (set: Set<number>, groups: number[]) => groups.some((g) => set.has(g));

function matchesSynergy(card: Card, s: Synergy): boolean {
  // A Trainer naming a picked type is type support, see weightOf.
  if (card.cardType === 'Trainer')
    return s.pickedNames.some((n) => textNames(card.trainerText ?? '', n)) || inSet(s.nicheOfPokemon, nicheGroupsOf(card));
  if (s.textTypes.has(card.cardType as EnergyType)) return true;
  if (s.trainerTexts.some((t) => textNames(t, card.name ?? '')) || inSet(s.nicheOfTrainers, nicheGroupsOf(card))) return true;
  return namedTypes(card).some((t) => s.pickedTypes.has(t));
}

// Sampling weight for one candidate given the running tally.
// Colorless Pokémon: flat 1 (always equally eligible per rules.md).
// Other Pokémon: 1 + TYPE_WEIGHT_COEF × picks of their type, capped at TYPE_LEAN_CAP.
// Trainers: 1 + TYPE_SUPPORT_COEF × picks of the most-picked type they support, capped.
// A synergy match then multiplies the result once.
export function weightOf(card: Card, tally: TypeTally, synergy: Synergy = NO_SYNERGY): number {
  let base = 1;
  if (card.cardType === 'Trainer') {
    const most = Math.max(0, ...supportedTypes(card).map((t) => tally[t] ?? 0));
    base = 1 + TYPE_SUPPORT_COEF * leanCount(most);
  } else if (card.cardType !== 'Colorless') {
    base = 1 + TYPE_WEIGHT_COEF * leanCount(tally[card.cardType] ?? 0);
  }
  const matched = matchesSynergy(card, synergy);
  return (matched ? base * SYNERGY_MULT : base) * eeveeBoost(card, synergy, matched);
}
