import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Card } from '../types/card';
import { generateOffer } from '../draft/offer';
import { indexByTier } from '../draft/tiers';

export const TOTAL_PACKS = 20;
export const SHOP_TICKETS = 3;

export type Phase = 'unstarted' | 'drafting' | 'shop' | 'review';

type State = {
  pickIds: string[];
  offerIds: string[];
  packOffers: string[][]; // history of full 5-card offers per completed pack
  shopPurchasedIds: string[];
  shopFinalized: boolean;
};

type Actions = {
  start: (pool: Card[]) => void;
  pick: (cardId: string, pool: Card[]) => void;
  purchase: (cardId: string) => void;
  unpurchase: (cardId: string) => void;
  finalizeShop: () => void;
  reset: () => void;
};

const initialState: State = {
  pickIds: [],
  offerIds: [],
  packOffers: [],
  shopPurchasedIds: [],
  shopFinalized: false,
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

      purchase: (cardId) => {
        const { shopPurchasedIds } = get();
        if (shopPurchasedIds.includes(cardId)) return;
        if (shopPurchasedIds.length >= SHOP_TICKETS) return;
        set({ shopPurchasedIds: [...shopPurchasedIds, cardId] });
      },

      unpurchase: (cardId) => {
        const { shopPurchasedIds } = get();
        set({ shopPurchasedIds: shopPurchasedIds.filter((id) => id !== cardId) });
      },

      finalizeShop: () => set({ shopFinalized: true }),

      reset: () => set({ ...initialState }),
    }),
    {
      name: 'pocket-draft-v0.2',
      // If the schema drifts again later, bump the name above and add a migrate fn.
    },
  ),
);

export function derivePhase(s: Pick<State, 'pickIds' | 'offerIds' | 'shopFinalized'>): Phase {
  if (s.pickIds.length === 0 && s.offerIds.length === 0) return 'unstarted';
  if (s.pickIds.length < TOTAL_PACKS) return 'drafting';
  if (!s.shopFinalized) return 'shop';
  return 'review';
}
