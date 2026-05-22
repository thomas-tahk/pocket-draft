import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Card } from '../types/card';
import { generateOffer } from '../draft/offer';
import { indexByTier } from '../draft/tiers';

export const TOTAL_PACKS = 20;

export type Phase = 'unstarted' | 'drafting' | 'review';

type State = {
  pickIds: string[];
  offerIds: string[];
};

type Actions = {
  start: (pool: Card[]) => void;
  pick: (cardId: string, pool: Card[]) => void;
  reset: () => void;
};

// Persists only id strings; full Card objects are looked up from the runtime pool.
// This keeps localStorage tiny and decouples saved state from the card-data schema.
export const useDraftStore = create<State & Actions>()(
  persist(
    (set, get) => ({
      pickIds: [],
      offerIds: [],

      start: (pool) => {
        const byTier = indexByTier(pool);
        const firstOffer = generateOffer(byTier, { isPack1: true, picks: [] });
        set({ pickIds: [], offerIds: firstOffer.map((c) => c.id) });
      },

      pick: (cardId, pool) => {
        const { pickIds } = get();
        const picked = pool.find((c) => c.id === cardId);
        if (!picked) return;
        const newPickIds = [...pickIds, cardId];

        if (newPickIds.length >= TOTAL_PACKS) {
          set({ pickIds: newPickIds, offerIds: [] });
          return;
        }

        const byId = new Map(pool.map((c) => [c.id, c]));
        const picksSoFar = newPickIds
          .map((id) => byId.get(id))
          .filter((c): c is Card => Boolean(c));
        const byTier = indexByTier(pool);
        const nextOffer = generateOffer(byTier, { isPack1: false, picks: picksSoFar });
        set({ pickIds: newPickIds, offerIds: nextOffer.map((c) => c.id) });
      },

      reset: () => set({ pickIds: [], offerIds: [] }),
    }),
    { name: 'pocket-draft-v0.1' },
  ),
);

export function derivePhase(pickIds: string[], offerIds: string[]): Phase {
  if (pickIds.length === 0 && offerIds.length === 0) return 'unstarted';
  if (pickIds.length >= TOTAL_PACKS) return 'review';
  return 'drafting';
}
