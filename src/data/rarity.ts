import type { Card, DraftableRarity, ExKind, Rarity, RawCard } from '../types/card';

// Decision 3.A (locked in memory): infer draft rarity for Promo cards from exKind.
//   EX / Mega-EX → ◊◊◊◊, regular → ◊.
export function inferDraftRarity(rarity: Rarity, exKind: ExKind): DraftableRarity | null {
  if (rarity === '◊' || rarity === '◊◊' || rarity === '◊◊◊' || rarity === '◊◊◊◊') {
    return rarity;
  }
  if (rarity === 'Promo') {
    return exKind === 'regular' ? '◊' : '◊◊◊◊';
  }
  return null;
}

export function isCosmeticVariant(rarity: Rarity): boolean {
  return rarity === '☆' || rarity === '☆☆' || rarity === '☆☆☆' || rarity === '♕';
}

export function enrich(raw: RawCard): Card | null {
  const draftRarity = inferDraftRarity(raw.rarity, raw.exKind);
  if (draftRarity === null) return null;
  return {
    ...raw,
    draftRarity,
    isEx: raw.exKind === 'ex' || raw.exKind === 'mega-ex',
    isMegaEx: raw.exKind === 'mega-ex',
  };
}
