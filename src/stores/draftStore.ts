import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Card } from '../types/card';
import { generateOffer } from '../draft/offer';
import { indexByTier } from '../draft/tiers';

export const TOTAL_PACKS = 20;
export const DECK_SIZE = 20;

export type Phase = 'unstarted' | 'drafting' | 'deckbuild' | 'review';

type State = {
  pickIds: string[];
  offerIds: string[];
  packOffers: string[][]; // history of full 5-card offers per completed pack
  deck: Record<string, number>; // cardId → count (1 or 2)
  deckFinalized: boolean;
};

type Actions = {
  start: (pool: Card[]) => void;
  pick: (cardId: string, pool: Card[]) => void;
  addToDeck: (cardId: string) => void;
  removeFromDeck: (cardId: string) => void;
  finalizeDeck: () => void;
  reset: () => void;
};

const initialState: State = {
  pickIds: [],
  offerIds: [],
  packOffers: [],
  deck: {},
  deckFinalized: false,
};

export const useDraftStore = create<State & Actions>()(
  persist(
    (set, get) => ({
      ...initialState,

      start: (pool) => {
        const byTier = indexByTier(pool);
        const firstOffer = generateOffer(byTier, { isPack1: true, picks: [] });
        set({ ...initialState, offerIds: firstOffer.map((c) => c.id) });
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

      // Note: callers must enforce the 2x-by-name cap and deck-size limit before
      // calling addToDeck — the store doesn't know about the card pool.
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

      finalizeDeck: () => set({ deckFinalized: true }),

      reset: () => set({ ...initialState }),
    }),
    { name: 'pocket-draft-v0.3' },
  ),
);

export function derivePhase(s: Pick<State, 'pickIds' | 'offerIds' | 'deckFinalized'>): Phase {
  if (s.pickIds.length === 0 && s.offerIds.length === 0) return 'unstarted';
  if (s.pickIds.length < TOTAL_PACKS) return 'drafting';
  if (!s.deckFinalized) return 'deckbuild';
  return 'review';
}
