import type { Card, RawCard, SetInfo } from '../types/card';
import {
  getFreeUniversals,
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
  universals: Card[]; // the 5 free universal Trainers
  sets: SetInfo[];
  packs: Record<PackId, PackComposition>;
};

export async function loadCardPool(): Promise<CardPool> {
  const [allCards, sets] = await Promise.all([loadRawCards(), loadSets()]);
  const draftableCards = toDraftablePool(allCards);
  const fullCards = toFullCardPool(allCards);
  const universals = getFreeUniversals(allCards);
  const packs = buildPackComposition(draftableCards, sets);
  return { allCards, draftableCards, fullCards, universals, sets, packs };
}

export type { PackComposition, PackId } from './packs';
