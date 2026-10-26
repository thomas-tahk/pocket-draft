import type { Card } from '../types/card';
import { tallyTypes } from '../draft/weights';

// The Energy Zone types a deck may choose (engine ValidateEnergyTypes: Grass to Metal;
// no Colorless or Dragon energy exists), in the order the pickers list them.
export const ZONE_TYPES = ['Grass', 'Fire', 'Water', 'Lightning', 'Psychic', 'Fighting', 'Darkness', 'Metal'] as const;
export const MAX_ZONE_TYPES = 3;

// The default choice: the types of the deck's Pokémon, most copies first (ties in
// ZONE_TYPES order), at most 3. Colorless and Dragon Pokémon add nothing.
export function defaultZoneTypes(deckCards: Card[]): string[] {
  const tally = tallyTypes(deckCards) as Record<string, number>;
  return ZONE_TYPES.filter((t) => (tally[t] ?? 0) > 0)
    .sort((a, b) => tally[b] - tally[a])
    .slice(0, MAX_ZONE_TYPES);
}
