import type { Card, RawCard, SetInfo } from '../types/card';
import { enrich, isCosmeticVariant } from './rarity';

const CARDS_URL = '/data/cards.json';
const SETS_URL = '/data/sets.json';

// Free universals: excluded from the draft pool (Decision 1.A) but freely
// available at deckbuild via getFreeUniversals(), subject to the 2x copy cap.
export const FREE_UNIVERSAL_NAMES = new Set([
  'Potion',
  'X Speed',
  'Red Card',
  'Poké Ball',
  "Professor's Research",
]);

// Excluded set: a4b "Deluxe Pack: ex" is reprint-only with glossy treatments;
// every base-art card has a functional duplicate in another set.
const EXCLUDED_SETS = new Set(['A4b']);

async function fetchJson<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`${url} → ${res.status}`);
  return (await res.json()) as T;
}

export async function loadRawCards(): Promise<RawCard[]> {
  return fetchJson<RawCard[]>(CARDS_URL);
}

export async function loadSets(): Promise<SetInfo[]> {
  return fetchJson<SetInfo[]>(SETS_URL);
}

function isDraftCandidate(raw: RawCard): boolean {
  if (EXCLUDED_SETS.has(raw.setId)) return false;
  if (isCosmeticVariant(raw.rarity)) return false;
  if (raw.rarity === null) return false;
  if (raw.cardType === 'Trainer' && FREE_UNIVERSAL_NAMES.has(raw.name)) return false;
  return true;
}

// Pool-then-pick model (2026-05-27): the draft pool contains only "final" Pokémon —
// cards no other Pokémon evolves from. Under-evolutions (Charmander, Charmeleon,
// Eevee, etc.) are excluded from drafting and instead granted FREE at deckbuild
// for each drafted Pokémon whose evolution chain they belong to. This avoids
// awkward draws where a player gets stuck with a base form they can't actually
// evolve, while keeping evolution lines accessible via SWUCUBE-style expansions.
function isFinalEvolution(card: Card, hasSuccessor: Set<string>): boolean {
  if (card.cardType === 'Trainer') return true;
  return !hasSuccessor.has(card.name);
}

function filterToFinalEvolutions(cards: Card[]): Card[] {
  const hasSuccessor = new Set<string>();
  for (const c of cards) {
    if (c.evolvesFrom) hasSuccessor.add(c.evolvesFrom);
  }
  return cards.filter((c) => isFinalEvolution(c, hasSuccessor));
}

// Decision 4.A: dedupe pa-064 / pa-065 (both "Rayquaza ex", same signature) — keep lowest id.
function dedupePromoRayquaza(cards: Card[]): Card[] {
  const seen = new Map<string, Card>();
  for (const c of cards) {
    const isPromo = c.setId.startsWith('P-');
    if (!isPromo) {
      seen.set(c.id, c);
      continue;
    }
    const key = `${c.name}|${c.cardType}|${c.hp ?? ''}|${c.exKind}`;
    const existing = seen.get(key);
    if (!existing) {
      seen.set(key, c);
    } else if (c.id < existing.id) {
      seen.delete(existing.id);
      seen.set(key, c);
    }
  }
  return [...seen.values()];
}

export function toDraftablePool(raw: RawCard[]): Card[] {
  const enriched = raw
    .filter(isDraftCandidate)
    .map(enrich)
    .filter((c): c is Card => c !== null);
  const deduped = dedupePromoRayquaza(enriched);
  const finalsOnly = filterToFinalEvolutions(deduped);
  return finalsOnly.sort((a, b) => (a.id < b.id ? -1 : 1));
}

// The full enriched-but-unfiltered pool — used at deckbuild to grant free access
// to under-evolutions of drafted Pokémon (which are excluded from the draft pool).
export function toFullCardPool(raw: RawCard[]): Card[] {
  const enriched = raw
    .filter(isDraftCandidate)
    .map(enrich)
    .filter((c): c is Card => c !== null);
  return dedupePromoRayquaza(enriched).sort((a, b) => (a.id < b.id ? -1 : 1));
}

// Free universals are filtered out of the draft pool but always available at
// deckbuild. One enriched entry per name (lowest-id variant wins on dedup).
export function getFreeUniversals(raw: RawCard[]): Card[] {
  const candidates = raw
    .filter((c) => FREE_UNIVERSAL_NAMES.has(c.name))
    .filter((c) => c.setId !== 'A4b')
    .filter((c) => !isCosmeticVariant(c.rarity))
    .map(enrich)
    .filter((c): c is Card => c !== null);
  const byName = new Map<string, Card>();
  for (const c of candidates) {
    const existing = byName.get(c.name);
    if (!existing || c.id < existing.id) byName.set(c.name, c);
  }
  return [...byName.values()].sort((a, b) => a.name.localeCompare(b.name));
}
