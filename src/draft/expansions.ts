import type { Card } from '../types/card';

// SWUCUBE-style free expansions at deckbuild:
//  - For each drafted Pokémon: every under-evolution in its chain becomes
//    freely deckbuild-eligible (the drafted Charizard ex grants Charmander
//    and Charmeleon, which are not draftable themselves).
//  - 2nd copies are implicit via the 2x copy cap applied uniformly.
//  - Free universals (Pokeball, Prof's Research, X Speed, Potion, Red Card)
//    are always present.
//
// Same-name variants are not unified here — each card stays as its own variant
// id, and the 2x cap is enforced by *name* at the deck level (so two A1/A2b
// Charizard ex collapse to a single name with two slots).

export type CollectionSource = 'drafted' | 'shop' | 'under-evo' | 'universal';

export type CollectionEntry = {
  card: Card;
  source: CollectionSource;
};

const COPY_CAP_BY_NAME = 2;
export { COPY_CAP_BY_NAME };

export function buildCollection(
  picks: Card[],
  shopPurchases: Card[],
  fullPool: Card[],
  universals: Card[],
): CollectionEntry[] {
  const byName = new Map<string, Card[]>();
  for (const c of fullPool) {
    const list = byName.get(c.name) ?? [];
    list.push(c);
    byName.set(c.name, list);
  }

  const entries = new Map<string, CollectionEntry>(); // keyed by name

  // 1. Drafted picks — keep the actual variant the player drafted.
  for (const p of picks) {
    if (entries.has(p.name)) continue;
    entries.set(p.name, { card: p, source: 'drafted' });
  }

  // 1b. Shop purchases — additions to the collection (1 ticket each).
  for (const s of shopPurchases) {
    if (entries.has(s.name)) continue;
    entries.set(s.name, { card: s, source: 'shop' });
  }

  // 2. Under-evolutions of each drafted Pokémon and each shop-purchased
  //    Pokémon. Walk the evolvesFrom chain backwards.
  const seedNames = [...picks, ...shopPurchases]
    .filter((p) => p.cardType !== 'Trainer')
    .map((p) => p.name);
  const draftedPokemonNames = seedNames;

  const visited = new Set<string>(draftedPokemonNames);
  const queue: string[] = [...draftedPokemonNames];

  while (queue.length > 0) {
    const current = queue.shift()!;
    const variantsOfCurrent = byName.get(current) ?? [];
    // Find every distinct predecessor name (evolvesFrom field across variants).
    const predecessorNames = new Set<string>();
    for (const v of variantsOfCurrent) {
      if (v.evolvesFrom) predecessorNames.add(v.evolvesFrom);
    }

    for (const predName of predecessorNames) {
      if (visited.has(predName)) continue;
      visited.add(predName);
      queue.push(predName);

      // Skip if already drafted (we'd lose the original drafted variant).
      if (entries.has(predName)) continue;
      const predVariants = byName.get(predName) ?? [];
      if (predVariants.length === 0) continue;
      // Lowest-id variant wins as the default. UI may let the user swap variants later.
      const chosen = [...predVariants].sort((a, b) => a.id.localeCompare(b.id))[0];
      entries.set(predName, { card: chosen, source: 'under-evo' });
    }
  }

  // 3. Free universals — always available.
  for (const u of universals) {
    if (entries.has(u.name)) continue;
    entries.set(u.name, { card: u, source: 'universal' });
  }

  return [...entries.values()];
}

// Helper: count how many cards in the deck share a given name.
export function deckCountByName(deck: Record<string, number>, pool: Map<string, Card>, name: string): number {
  let n = 0;
  for (const [cardId, count] of Object.entries(deck)) {
    const card = pool.get(cardId);
    if (card?.name === name) n += count;
  }
  return n;
}

export function deckSize(deck: Record<string, number>): number {
  return Object.values(deck).reduce((s, n) => s + n, 0);
}
