import type { Card } from '../types/card';

// The Energy Zone types a deck may choose (engine ValidateEnergyTypes: Grass to Metal;
// no Colorless or Dragon Energy Zone energy exists), in the order the pickers list them.
export const ZONE_TYPES = ['Grass', 'Fire', 'Water', 'Lightning', 'Psychic', 'Fighting', 'Darkness', 'Metal'] as const;
export const MAX_ZONE_TYPES = 3;

// A printed cost letter and the type it means. C (Colorless) and N (Dragon) are not
// zone types, so they never count, exactly as in the engine.
const COST_LETTER: Record<string, string> = {
  G: 'Grass', R: 'Fire', W: 'Water', L: 'Lightning', P: 'Psychic', F: 'Fighting', D: 'Darkness', M: 'Metal',
};

// Mirrors engine/energy.go DefaultEnergyTypes: the typed energy the deck's attack costs
// need, one count per cost symbol, most-needed first (ties in ZONE_TYPES order, which is
// the engine's energyOrder), at most 3. Empty means no typed cost: the engine then uses
// Colorless, which is not choosable (see ZONE_FALLBACK).
export function defaultZoneTypes(deckCards: Card[]): string[] {
  const need: Record<string, number> = {};
  for (const c of deckCards) {
    for (const a of c.attacks) {
      for (const ch of a.cost) {
        const t = COST_LETTER[ch];
        if (t) need[t] = (need[t] ?? 0) + 1;
      }
    }
  }
  return ZONE_TYPES.filter((t) => (need[t] ?? 0) > 0)
    .sort((a, b) => need[b] - need[a]) // stable, like the engine's SliceStable
    .slice(0, MAX_ZONE_TYPES);
}

// What the engine uses when a deck has no typed attack cost and nothing was chosen.
export const ZONE_FALLBACK = 'Colorless';
