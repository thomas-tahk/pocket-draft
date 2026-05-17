import type { Card, SetInfo } from '../types/card';

export type PackId = string;

export type PackComposition = {
  packId: PackId;
  packName: string;
  setId: string;
  cards: Card[];
};

// Pseudo-pack catch-all for draftable promos (Decision 5.A).
const PROMO_POOL_ID = 'promo-pool';

// Pack-tail values that aren't actual booster packs — these cards stay in the
// raw pool but don't form their own draft pack.
const NON_BOOSTER_PACKS = new Set([
  'Wonder Pick',
  'Shop',
  'Missions',
  'Premium Missions',
  'Campaign',
]);

function packIdFor(setId: string, packName: string): PackId {
  const slug = packName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
  return `${setId}-${slug}`;
}

export function buildPackComposition(cards: Card[], sets: SetInfo[]): Record<PackId, PackComposition> {
  const setNameById = new Map(sets.map((s) => [s.id, s.name]));
  const packs: Record<PackId, PackComposition> = {};

  for (const card of cards) {
    // Promos → single pseudo-pack regardless of upstream pack hint.
    if (card.setId.startsWith('P-')) {
      if (!packs[PROMO_POOL_ID]) {
        packs[PROMO_POOL_ID] = {
          packId: PROMO_POOL_ID,
          packName: 'Promo pool',
          setId: 'promo',
          cards: [],
        };
      }
      packs[PROMO_POOL_ID].cards.push(card);
      continue;
    }
    if (!card.pack || NON_BOOSTER_PACKS.has(card.pack)) continue;
    const packId = packIdFor(card.setId, card.pack);
    if (!packs[packId]) {
      const setName = setNameById.get(card.setId) ?? card.setId;
      packs[packId] = {
        packId,
        packName: `${setName} — ${card.pack}`,
        setId: card.setId,
        cards: [],
      };
    }
    packs[packId].cards.push(card);
  }

  return packs;
}
