import type { Card } from '../types/card';

// Functional signature for a Pocket card: everything that matters for play.
// Excludes name (so two cards with the same name can be compared) and excludes
// id/set/imagery/rarity-aesthetics (so cosmetic differences across reprints
// don't fragment the pool).
export function functionalSignature(c: Card): string {
  const attacks = (c.attacks ?? [])
    .map((a) => `${a.cost}|${a.name}|${a.damage}|${a.effect}`)
    .join('||');
  const ability = c.ability ? `${c.ability.name}|${c.ability.effect}` : '';
  return [
    c.cardType,
    c.hp ?? '',
    c.stage ?? '',
    c.evolvesFrom ?? '',
    c.weakness ?? '',
    c.retreat ?? '',
    attacks,
    ability,
    c.trainerKind ?? '',
    c.trainerText ?? '',
  ].join('::');
}

// Across a list of cards, collapse name+signature duplicates — keep the
// lowest-id representative. Used at the draft pool (so functionally-identical
// reprints across sets stop showing up as separate options) and at the
// under-evolution expansion (so e.g. all distinct Eevee variants surface
// once each rather than only the lowest-id one).
export function functionalDedupe(cards: Card[]): Card[] {
  const byKey = new Map<string, Card>();
  for (const c of cards) {
    const key = `${c.name}::${functionalSignature(c)}`;
    const existing = byKey.get(key);
    if (!existing || c.id < existing.id) byKey.set(key, c);
  }
  return [...byKey.values()];
}

// Within a single name, surface every distinct functional variant. Use when
// you want all mechanical printings of a Pokémon to be available even if some
// share a name (e.g. multiple Eevee printings with different attacks).
export function distinctFunctionalVariants(cards: Card[]): Card[] {
  const bySig = new Map<string, Card>();
  for (const c of cards) {
    const sig = functionalSignature(c);
    const existing = bySig.get(sig);
    if (!existing || c.id < existing.id) bySig.set(sig, c);
  }
  return [...bySig.values()];
}
