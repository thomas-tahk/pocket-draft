import type { Card } from '../types/card';
import { distinctFunctionalVariants } from '../data/signature';

// SWUCUBE-style free expansions at deckbuild:
//   - For each drafted/purchased Pokémon: every under-evolution in the
//     evolves-from chain becomes deckbuild-eligible. All distinct functional
//     variants surface (so multiple mechanical Eevee printings can all be
//     chosen between, capped together at the 2x by-name limit).
//   - The free universals (Poké Ball, Professor's Research) are always present.
//   - 2nd copies are implicit via the uniform 2x copy cap.

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

  const entries: CollectionEntry[] = [];
  const seenIds = new Set<string>();

  const tryAdd = (card: Card, source: CollectionSource) => {
    if (seenIds.has(card.id)) return;
    seenIds.add(card.id);
    entries.push({ card, source });
  };

  // 1. Drafted picks — keep the actual variant the player drafted.
  for (const p of picks) tryAdd(p, 'drafted');

  // 1b. Shop purchases.
  for (const s of shopPurchases) tryAdd(s, 'shop');

  // 2. Under-evolutions of each drafted/purchased Pokémon. Walk the
  //    evolvesFrom chain backwards; at each predecessor name, surface ALL
  //    distinct functional variants.
  const seedNames = [...picks, ...shopPurchases]
    .filter((p) => p.cardType !== 'Trainer')
    .map((p) => p.name);

  const visited = new Set<string>(seedNames);
  const queue: string[] = [...seedNames];

  while (queue.length > 0) {
    const current = queue.shift()!;
    const variantsOfCurrent = byName.get(current) ?? [];
    const predecessorNames = new Set<string>();
    for (const v of variantsOfCurrent) {
      if (v.evolvesFrom) predecessorNames.add(v.evolvesFrom);
    }
    for (const predName of predecessorNames) {
      if (visited.has(predName)) continue;
      visited.add(predName);
      queue.push(predName);
      const predVariants = byName.get(predName) ?? [];
      if (predVariants.length === 0) continue;
      const distinct = distinctFunctionalVariants(predVariants);
      for (const v of distinct) tryAdd(v, 'under-evo');
    }
  }

  // 3. Free universals — always available.
  for (const u of universals) tryAdd(u, 'universal');

  return entries;
}

// Helper: count how many cards in the deck share a given name.
export function deckCountByName(
  deck: Record<string, number>,
  pool: Map<string, Card>,
  name: string,
): number {
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
