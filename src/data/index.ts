import type { Card, RawCard, SetInfo } from '../types/card';
import {
  getFreeUniversals,
  keepWorkingText,
  loadDraftPoolFilter,
  loadRawCards,
  loadSets,
  toDraftablePool,
  toFullCardPool,
} from './loader';
import { buildPackComposition, type PackComposition, type PackId } from './packs';

export type CardPool = {
  allCards: RawCard[];
  draftableCards: Card[]; // pool from which packs are drawn (final-evos only)
  fullCards: Card[]; // includes under-evolutions; used at deckbuild
  universals: Card[]; // free universal Trainers (Poké Ball, Professor's Research)
  // Ids of the cards the engine can play. Shop and deckbuild offer only these,
  // so every deck is 20 working cards (owner 2026-10-19 (1)).
  working: ReadonlySet<string>;
  sets: SetInfo[];
  packs: Record<PackId, PackComposition>;
};

export async function loadCardPool(): Promise<CardPool> {
  const [allCards, sets, poolFilter] = await Promise.all([
    loadRawCards(),
    loadSets(),
    loadDraftPoolFilter(),
  ]);
  const draftableCards = keepWorkingText(toDraftablePool(allCards), poolFilter);
  const fullCards = toFullCardPool(allCards);
  const universals = getFreeUniversals(allCards);
  const working = new Set(poolFilter.working);
  const packs = buildPackComposition(draftableCards, sets);
  return { allCards, draftableCards, fullCards, universals, working, sets, packs };
}

export type { PackComposition, PackId } from './packs';
