import type { Card, RawCard, SetInfo } from '../types/card';
import { loadRawCards, loadSets, toDraftablePool } from './loader';
import { buildPackComposition, type PackComposition, type PackId } from './packs';

export type CardPool = {
  allCards: RawCard[];
  draftableCards: Card[];
  sets: SetInfo[];
  packs: Record<PackId, PackComposition>;
};

export async function loadCardPool(): Promise<CardPool> {
  const [allCards, sets] = await Promise.all([loadRawCards(), loadSets()]);
  const draftableCards = toDraftablePool(allCards);
  const packs = buildPackComposition(draftableCards, sets);
  return { allCards, draftableCards, sets, packs };
}

export type { PackComposition, PackId } from './packs';
