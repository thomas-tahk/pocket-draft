import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Card } from '../types/card';
import { generateOffer } from '../draft/offer';
import { indexByTier } from '../draft/tiers';
import { COPY_CAP_BY_NAME } from '../draft/expansions';

export const TOTAL_PACKS = 16;
export const DECK_SIZE = 20;
export const SHOP_TICKETS = 4;
export const HISTORY_CAP = 7;

export type Phase = 'unstarted' | 'drafting' | 'shop' | 'deckbuild' | 'review';

export type DraftHistoryEntry = {
  id: string; // ULID-ish: epoch ms + random suffix
  completedAt: string; // ISO string
  pickIds: string[]; // 16 drafted picks (preserved for reference)
  shopPurchasedIds: string[]; // 0..4 shop purchases
  deck: Record<string, number>; // final 20-card deck (cardId → count)
  notes: string;
  wins: number;
  losses: number;
};

type State = {
  pickIds: string[];
  offerIds: string[];
  packOffers: string[][]; // history of full 5-card offers per completed pack
  shopPurchasedIds: string[];
  shopFinalized: boolean;
  deck: Record<string, number>; // cardId → count (1 or 2)
  deckFinalized: boolean;
  history: DraftHistoryEntry[];
};

type Actions = {
  start: (pool: Card[]) => void;
  pick: (cardId: string, pool: Card[]) => void;
  purchase: (cardId: string) => void;
  unpurchase: (cardId: string) => void;
  finalizeShop: () => void;
  addToDeck: (cardId: string) => void;
  removeFromDeck: (cardId: string) => void;
  finalizeDeck: () => void;
  updateHistoryEntry: (id: string, updates: Partial<Pick<DraftHistoryEntry, 'notes' | 'wins' | 'losses'>>) => void;
  deleteHistoryEntry: (id: string) => void;
  reset: () => void;
};

const activeStateInitial: Omit<State, 'history'> = {
  pickIds: [],
  offerIds: [],
  packOffers: [],
  shopPurchasedIds: [],
  shopFinalized: false,
  deck: {},
  deckFinalized: false,
};

const initialState: State = {
  ...activeStateInitial,
  history: [],
};

function newEntryId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export const useDraftStore = create<State & Actions>()(
  persist(
    (set, get) => ({
      ...initialState,

      start: (pool) => {
        const byTier = indexByTier(pool);
        const firstOffer = generateOffer(byTier, { isPack1: true, picks: [] });
        const { history } = get();
        set({ ...activeStateInitial, history, offerIds: firstOffer.map((c) => c.id) });
      },

      pick: (cardId, pool) => {
        const { pickIds, offerIds, packOffers } = get();
        const picked = pool.find((c) => c.id === cardId);
        if (!picked) return;
        const newPickIds = [...pickIds, cardId];
        const newPackOffers = [...packOffers, offerIds];

        if (newPickIds.length >= TOTAL_PACKS) {
          set({ pickIds: newPickIds, offerIds: [], packOffers: newPackOffers });
          return;
        }

        const byId = new Map(pool.map((c) => [c.id, c]));
        const picksSoFar = newPickIds
          .map((id) => byId.get(id))
          .filter((c): c is Card => Boolean(c));
        const byTier = indexByTier(pool);
        const nextOffer = generateOffer(byTier, { isPack1: false, picks: picksSoFar });
        set({
          pickIds: newPickIds,
          offerIds: nextOffer.map((c) => c.id),
          packOffers: newPackOffers,
        });
      },

      purchase: (cardId) => {
        const { shopPurchasedIds } = get();
        if (shopPurchasedIds.length >= SHOP_TICKETS) return;
        const sameCount = shopPurchasedIds.filter((id) => id === cardId).length;
        if (sameCount >= COPY_CAP_BY_NAME) return;
        set({ shopPurchasedIds: [...shopPurchasedIds, cardId] });
      },

      unpurchase: (cardId) => {
        const { shopPurchasedIds } = get();
        const idx = shopPurchasedIds.lastIndexOf(cardId);
        if (idx === -1) return;
        set({
          shopPurchasedIds: [
            ...shopPurchasedIds.slice(0, idx),
            ...shopPurchasedIds.slice(idx + 1),
          ],
        });
      },

      finalizeShop: () => set({ shopFinalized: true }),

      addToDeck: (cardId) => {
        const { deck } = get();
        const current = deck[cardId] ?? 0;
        set({ deck: { ...deck, [cardId]: current + 1 } });
      },

      removeFromDeck: (cardId) => {
        const { deck } = get();
        const current = deck[cardId] ?? 0;
        if (current <= 1) {
          const next = { ...deck };
          delete next[cardId];
          set({ deck: next });
        } else {
          set({ deck: { ...deck, [cardId]: current - 1 } });
        }
      },

      finalizeDeck: () => {
        const { pickIds, shopPurchasedIds, deck, history } = get();
        const entry: DraftHistoryEntry = {
          id: newEntryId(),
          completedAt: new Date().toISOString(),
          pickIds: [...pickIds],
          shopPurchasedIds: [...shopPurchasedIds],
          deck: { ...deck },
          notes: '',
          wins: 0,
          losses: 0,
        };
        // FIFO cap: newest first, drop oldest beyond HISTORY_CAP.
        const nextHistory = [entry, ...history].slice(0, HISTORY_CAP);
        set({ deckFinalized: true, history: nextHistory });
      },

      updateHistoryEntry: (id, updates) => {
        const { history } = get();
        set({
          history: history.map((e) => (e.id === id ? { ...e, ...updates } : e)),
        });
      },

      deleteHistoryEntry: (id) => {
        const { history } = get();
        set({ history: history.filter((e) => e.id !== id) });
      },

      reset: () => {
        const { history } = get();
        set({ ...activeStateInitial, history });
      },
    }),
    { name: 'pocket-draft-v0.4' },
  ),
);

export function derivePhase(
  s: Pick<State, 'pickIds' | 'offerIds' | 'shopFinalized' | 'deckFinalized'>,
): Phase {
  if (s.pickIds.length === 0 && s.offerIds.length === 0) return 'unstarted';
  if (s.pickIds.length < TOTAL_PACKS) return 'drafting';
  if (!s.shopFinalized) return 'shop';
  if (!s.deckFinalized) return 'deckbuild';
  return 'review';
}
