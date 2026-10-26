import type { EnergyType } from '../types/card';

// Type support for draft offers: which Trainers/Items help one type's Pokémon,
// so leaning a type raises their odds too (owner, 2026-10-25, point 3).
// Mirrored by server/draft_type_support.go; edit both together.
//
// Most Trainers are found by their text: it names a type ("Water Pokémon",
// "[W]", ...), see supportedTypes in weights.ts. This file holds only what the
// text cannot say, hand-made from public/data/cards.json. It is NOT the niche
// Trainer <-> Pokémon table (#126): that one pairs a Trainer with the specific
// Pokémon it names; this one only says which type a Trainer favours.

// A Trainer's exact name -> the types it helps, for Trainers whose text names
// Pokémon (not a type). The comment lists the Pokémon the text names; the type
// is their card type in cards.json, judged by hand.
export const TYPE_SUPPORT_TABLE: Record<string, EnergyType[]> = {
  Blaine: ['Fire'], // Ninetales, Rapidash, Magmar
  Koga: ['Darkness'], // Muk, Weezing
  Marlon: ['Water', 'Psychic'], // Carracosta (Water), Jellicent (Psychic)
  Lana: ['Water'], // Araquanid
  Hala: ['Fighting'], // Hariyama, Crabominable
  Nemona: ['Lightning'], // Pawmot
  Clemont: ['Lightning'], // Magneton, Heliolisk
  "Clemont's Backpack": ['Lightning'], // Magneton, Heliolisk
  Sophocles: ['Lightning'], // Alolan Golem, Vikavolt, Togedemaru
  Jasmine: ['Metal'], // Steelix, Skarmory ex
  Iris: ['Dragon'], // Haxorus
  Drayden: ['Dragon'], // Draco Meteor attack
  Piers: ['Darkness'], // Galarian Obstagoon
  'Team Galactic Grunt': ['Darkness'], // Stunky, Croagunk (Glameow is Colorless)
  Mallow: ['Grass', 'Psychic'], // Tsareena (Grass), Shiinotic (Psychic)
  Acerola: ['Psychic'], // Mimikyu, Palossand
  'Budding Expeditioner': ['Psychic'], // Mew ex
};

// Trainers whose text names a type but works against it, so the text rule must
// not count them.
export const NOT_TYPE_SUPPORT: ReadonlySet<string> = new Set([
  'Squirt Bottle', // discards a [R] Energy from the OPPONENT's Pokémon
]);
