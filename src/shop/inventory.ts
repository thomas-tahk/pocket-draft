import type { Card } from '../types/card';
import { STAPLE_SHELF_IDS } from './staples';

// SWUCUBE-style shop inventory: 1 ticket grants ACCESS to a card (adds it to
// the deckbuild collection at the standard 2x copy cap). The shop sells Trainers
// only: the fixed staple shelf (rules.md §2 backstop), minus any you drafted.
//
// Pokémon-specific Trainers (Blaine → Ninetales/Rapidash/Magmar, etc.) are
// deferred until a hand-curated trainer→target table exists.

export type ShopInventory = {
  staples: Card[];
};

export function buildShopInventory(picks: Card[], fullPool: Card[]): ShopInventory {
  const draftedNames = new Set(picks.map((c) => c.name));
  const byId = new Map(fullPool.map((c) => [c.id, c]));

  const staples = STAPLE_SHELF_IDS.map((id) => byId.get(id))
    .filter((c): c is Card => Boolean(c))
    .filter((c) => !draftedNames.has(c.name));

  return { staples };
}
