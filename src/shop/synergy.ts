import type { Card } from '../types/card';

// Score how well a candidate card synergizes with the drafted picks.
// Used to pick 1 representative from each pack's 4 unchosen cards.
// Tunable. Higher = more relevant to the drafted deck's direction.
export function synergyScore(candidate: Card, picks: Card[]): number {
  let score = 0;

  // Type alignment: +1 per picked Pokémon of matching energy type.
  if (candidate.cardType !== 'Trainer' && candidate.cardType !== 'Colorless') {
    score += picks.filter((p) => p.cardType === candidate.cardType).length;
  }

  // Trainers are scarce in the draft (no type bias works on them), so the shop
  // is the obvious place to backfill. The base bonus scales inversely with how
  // few Trainers the player drafted — a player with 0 drafted Supporters gets a
  // large nudge toward shop Supporters, while a player who already drafted plenty
  // gets only a small one. Without this, on-type Pokémon outscore every Trainer
  // on a type-committed draft and Supporters never surface — defeating the shop's
  // "ensure access to at least 1–2 Supporters" purpose (rules.md §2).
  if (candidate.cardType === 'Trainer') {
    const trainerCount = picks.filter((p) => p.cardType === 'Trainer').length;
    const scarcityBonus = Math.max(0, 5 - trainerCount);
    const kindBase = candidate.trainerKind === 'Supporter' ? 3 : 2;
    score += kindBase + scarcityBonus;
  }

  // Evolution chain: this card evolves from something already drafted.
  if (candidate.evolvesFrom && picks.some((p) => p.name === candidate.evolvesFrom)) {
    score += 3;
  }
  // Reverse: a drafted card evolves into this one (this is a pre-evo for the deck).
  if (picks.some((p) => p.evolvesFrom === candidate.name)) {
    score += 2;
  }

  // Rarity tilt: higher-tier cards are slightly more valuable as shop pickups.
  const RARITY_BONUS: Record<string, number> = {
    '◊◊◊◊': 1.5,
    '◊◊◊': 1.0,
    '◊◊': 0.5,
    '◊': 0,
  };
  score += RARITY_BONUS[candidate.draftRarity] ?? 0;
  if (candidate.isEx) score += 1;

  return score;
}

// Per rules.md: from each pack's 4 unchosen cards, surface the best-synergy one
// for the player. Returns up to 20 candidates (one per pack), deduped by id.
export function buildUndraftedShelf(
  packOffers: string[][],
  pickIds: string[],
  pool: Map<string, Card>,
): Card[] {
  const picks = pickIds
    .map((id) => pool.get(id))
    .filter((c): c is Card => Boolean(c));

  const seen = new Set<string>(pickIds);
  const shelf: Card[] = [];

  for (let i = 0; i < packOffers.length; i++) {
    const pickedId = pickIds[i];
    const candidates = packOffers[i]
      .filter((id) => id !== pickedId)
      .map((id) => pool.get(id))
      .filter((c): c is Card => Boolean(c));

    if (candidates.length === 0) continue;

    let best: Card | null = null;
    let bestScore = -Infinity;
    for (const c of candidates) {
      if (seen.has(c.id)) continue;
      const s = synergyScore(c, picks);
      if (s > bestScore) {
        best = c;
        bestScore = s;
      }
    }
    if (best) {
      shelf.push(best);
      seen.add(best.id);
    }
  }

  return shelf;
}
