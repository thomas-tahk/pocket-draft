import type { Card } from '../types/card';
import { STAPLE_SHELF_IDS } from './staples';

// SWUCUBE-style shop inventory: 1 ticket grants ACCESS to a card (adds it to
// the deckbuild collection at the standard 2x copy cap). Two shelves:
//
//   - Staples: a fixed list of 12 generic Trainers (rules.md §2 backstop).
//   - Same-type alts: every EX/Mega-EX Pokémon whose type matches a drafted
//     Pokémon's type, excluding names you've already drafted (those are free
//     to interchange via the in-collection 2x cap).
//
// Pokémon-specific Trainers (Blaine → Ninetales/Rapidash/Magmar, etc.) are
// deferred until a hand-curated trainer→target table exists.

export type ShopInventory = {
  staples: Card[];
  alts: Card[];
};

export function buildShopInventory(picks: Card[], fullPool: Card[]): ShopInventory {
  const draftedNames = new Set(picks.map((c) => c.name));
  const draftedTypes = new Set(
    picks
      .filter((p) => p.cardType !== 'Trainer' && p.cardType !== 'Colorless')
      .map((p) => p.cardType),
  );

  const byId = new Map(fullPool.map((c) => [c.id, c]));

  const staples = STAPLE_SHELF_IDS.map((id) => byId.get(id))
    .filter((c): c is Card => Boolean(c))
    .filter((c) => !draftedNames.has(c.name));

  const alts = fullPool
    .filter((c) => c.cardType !== 'Trainer')
    .filter((c) => c.isEx) // EX or Mega-EX
    .filter((c) => draftedTypes.has(c.cardType as never))
    .filter((c) => !draftedNames.has(c.name));

  // Dedupe alts by name (multiple set variants of the same card collapse).
  const altsByName = new Map<string, Card>();
  for (const c of alts) {
    const existing = altsByName.get(c.name);
    if (!existing || c.id < existing.id) altsByName.set(c.name, c);
  }

  return {
    staples,
    alts: [...altsByName.values()].sort((a, b) => a.name.localeCompare(b.name)),
  };
}
