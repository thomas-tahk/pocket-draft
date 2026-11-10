import type { EnergyType } from '../types/card';

// Type support for draft offers: which Trainers/Items help one type's Pokémon,
// so leaning a type raises their odds too (owner, 2026-10-25, point 3).
//
// The card text decides (owner 2026-10-28, decision 2): a Trainer supports the
// types its text names plus the card types of the Pokémon its text names. Go
// works that out (server/draft_type_support.go, the source of truth) and writes
// it into public/data/draft-pool.json as `supportTypes`; the SPA registers it
// with setSupportTypes in weights.ts. This file holds only the fallback for text
// that names neither, and the exclusions; Go's TestTypeSupportTSTableMatchesGo
// keeps both equal to server/draft_type_support.go.
//
// It is NOT the niche Trainer <-> Pokémon table (#126): that one pairs a Trainer
// with specific Pokémon; this one only says which type it favours.

// A Trainer's exact name -> the types it helps, for Trainers whose text names no
// type and no typed Pokémon. Hand-made.
export const TYPE_SUPPORT_TABLE: Record<string, EnergyType[]> = {
  Drayden: ['Dragon'], // names the Draco Meteor attack, not a Pokémon
};

// Trainers whose text names a type but works against it, so the text rule must
// not count them.
export const NOT_TYPE_SUPPORT: ReadonlySet<string> = new Set([
  'Squirt Bottle', // discards a [R] Energy from the OPPONENT's Pokémon
]);
