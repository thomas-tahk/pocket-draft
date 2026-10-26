import type { Card } from '../types/card';

// Niche support for draft offers (#126): archetype Trainers whose text does not
// name the Pokémon they work with, paired with those Pokémon, so picking either
// raises the other's odds (owner, 2026-10-25, point 2). Mirrored by
// server/draft_niche_support.go; edit both together.
//
// A Trainer that names a Pokémon in its text needs no entry here: the synergy
// rules read the text (rule 1 / rule 4 in weights.ts). This file is only for
// what the text cannot say. Hand-made from public/data/cards.json; the owner
// has not reviewed the groups, so they are guesses.
//
// Names are exact, without " ex": a Pokémon matches by its name minus a trailing
// " ex" (the same stripping rule 1 applies to picked names), so "Aerodactyl"
// covers "Aerodactyl ex" but never a longer name. Pokémon not in today's draft
// pool are listed too: the pool grows as card text lands, and an entry with no
// card in the pool does nothing. Eevee ex is not listed (#128).

// Pairs Trainers with the Pokémon they support. Cross-kind only: a picked Trainer
// raises the group's Pokémon and a picked Pokémon raises the group's Trainers;
// two Pokémon, or two Trainers, of one group do not raise each other.
export type NicheGroup = {
  label: string;
  trainers: string[]; // exact Trainer names
  pokemon?: string[]; // Pokémon names without " ex"
  evolvesFrom?: string[]; // any Pokémon whose card says it evolves from one of these
};

export const NICHE_GROUPS: NicheGroup[] = [
  // Fossils: each Fossil/Old Amber supports the final evolution of its own line.
  { label: 'Helix Fossil', trainers: ['Helix Fossil'], pokemon: ['Omastar'] },
  { label: 'Dome Fossil', trainers: ['Dome Fossil'], pokemon: ['Kabutops'] },
  { label: 'Old Amber', trainers: ['Old Amber'], pokemon: ['Aerodactyl'] },
  { label: 'Skull Fossil', trainers: ['Skull Fossil'], pokemon: ['Rampardos'] },
  { label: 'Armor Fossil', trainers: ['Armor Fossil'], pokemon: ['Bastiodon'] },
  { label: 'Plume Fossil', trainers: ['Plume Fossil'], pokemon: ['Archeops'] },
  { label: 'Cover Fossil', trainers: ['Cover Fossil'], pokemon: ['Carracosta'] },
  { label: 'Root Fossil', trainers: ['Root Fossil'], pokemon: ['Cradily'] },
  { label: 'Claw Fossil', trainers: ['Claw Fossil'], pokemon: ['Armaldo'] },
  { label: 'Jaw Fossil', trainers: ['Jaw Fossil'], pokemon: ['Tyrantrum'] },
  { label: 'Sail Fossil', trainers: ['Sail Fossil'], pokemon: ['Aurorus'] },
  // Eevee Bag's text says "Eevee"; the Pokémon it helps are the ones that evolve
  // from it (Eevee ex itself is #128's).
  { label: 'Eeveelutions', trainers: ['Eevee Bag'], evolvesFrom: ['Eevee'] },
  // Ancient / Future / Ultra Beast: the printed tag is not in cards.json, so the
  // members are the real-TCG ones, by hand.
  {
    label: 'Ancient',
    trainers: ['Ancient Booster Energy Capsule', 'Professor Sada'],
    pokemon: [
      'Great Tusk', 'Scream Tail', 'Brute Bonnet', 'Flutter Mane', 'Slither Wing',
      'Sandy Shocks', 'Roaring Moon', 'Walking Wake', 'Gouging Fire', 'Raging Bolt', 'Koraidon',
    ],
  },
  {
    label: 'Future',
    trainers: ['Future Booster Energy Capsule', 'Professor Turo'],
    pokemon: [
      'Iron Treads', 'Iron Bundle', 'Iron Hands', 'Iron Jugulis', 'Iron Moth',
      'Iron Thorns', 'Iron Valiant', 'Iron Leaves', 'Iron Crown', 'Iron Boulder',
      'Miraidon',
    ],
  },
  {
    label: 'Ultra Beast',
    trainers: ['Lusamine', 'Beast Wall', 'Beastite'],
    pokemon: [
      'Nihilego', 'Buzzwole', 'Pheromosa', 'Xurkitree', 'Celesteela', 'Kartana',
      'Guzzlord', 'Poipole', 'Naganadel', 'Stakataka', 'Blacephalon',
    ],
  },
];

// Lookup tables over NICHE_GROUPS: name -> indexes of the groups it is in.
const byTrainer = new Map<string, number[]>();
const byPokemon = new Map<string, number[]>();
const byEvolvesFrom = new Map<string, number[]>();

function add(map: Map<string, number[]>, name: string, group: number) {
  map.set(name, [...(map.get(name) ?? []), group]);
}

NICHE_GROUPS.forEach((g, i) => {
  g.trainers.forEach((n) => add(byTrainer, n, i));
  (g.pokemon ?? []).forEach((n) => add(byPokemon, n, i));
  (g.evolvesFrom ?? []).forEach((n) => add(byEvolvesFrom, n, i));
});

// A Pokémon's name without a trailing " ex": the form a Trainer's text uses.
export const pokemonBaseName = (name: string): string => name.replace(/ ex$/, '');

// The niche groups a card is in: a Trainer by exact name, a Pokémon by name
// minus " ex" or by what it evolves from. Mirrored by nicheGroupsOf in Go.
export function nicheGroupsOf(card: Card): number[] {
  if (card.cardType === 'Trainer') return byTrainer.get(card.name ?? '') ?? [];
  return [
    ...(byPokemon.get(pokemonBaseName(card.name ?? '')) ?? []),
    ...(card.evolvesFrom ? (byEvolvesFrom.get(card.evolvesFrom) ?? []) : []),
  ];
}
