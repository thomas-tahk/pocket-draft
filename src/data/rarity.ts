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

// Rarity as words, for screen readers: the glyphs (diamonds, stars, crown) are read as nothing or as
// their Unicode names. One place for every accessible name that includes a rarity.
export function rarityWords(rarity: string): string {
  const count = [...rarity].length;
  const plural = (n: number, one: string) => `${n} ${one}${n === 1 ? '' : 's'}`;
  if (/^◊+$/.test(rarity)) return plural(count, 'diamond');
  if (/^☆+$/.test(rarity)) return plural(count, 'star');
  if (rarity === '♕' || rarity === '♛') return 'crown';
  return rarity.replace(/[^\p{L}\p{N} -]/gu, '').trim() || 'unknown rarity';
}

// "Pikachu, 2 diamonds, EX": the card's name and the key facts the tile shows, in words.
export function cardFacts(card: Pick<Card, 'name' | 'draftRarity' | 'isEx' | 'isMegaEx'>): string {
  return `${card.name}, ${rarityWords(card.draftRarity)}${card.isMegaEx ? ', Mega-EX' : card.isEx ? ', EX' : ''}`;
}
