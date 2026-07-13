import { create } from 'zustand';
import type { GameView, Move } from './types';
import { newGame, getState, sendMove, botMove } from './api';

export type Selection =
  | { kind: 'hand'; cardId: string }
  | { kind: 'bench'; index: number }
  | { kind: 'setupActive'; cardId: string }
  | null;

type State = {
  view: GameView | null;
  prev: GameView | null;
  status: 'idle' | 'loading' | 'ready';
  error: string | null;
  selection: Selection;
  deck: string[] | null; // the deck this game was started from, for restart
};

type Actions = {
  init: () => Promise<void>;
  startGame: (deck: string[], seed?: number) => Promise<void>;
  restart: (seed?: number) => Promise<void>;
  dispatch: (move: Move) => Promise<void>;
  runBot: () => Promise<void>;
  select: (sel: Selection) => void;
  clearSelection: () => void;
};

export const useGameStore = create<State & Actions>((set, get) => ({
  view: null,
  prev: null,
  status: 'idle',
  error: null,
  selection: null,
  deck: null,

  init: async () => {
    set({ status: 'loading', error: null });
    try {
      const view = await getState();
      set({ view, prev: null, status: 'ready' });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ status: 'idle', error: msg });
    }
  },

  startGame: async (deck, seed) => {
    set({ status: 'loading', error: null, selection: null, deck });
    try {
      const view = await newGame({ deck, seed });
      set({ view, prev: null, status: 'ready' });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ status: 'idle', error: msg });
    }
  },

  restart: async (seed) => {
    set({ status: 'loading', error: null, selection: null });
    try {
      const view = await newGame({ deck: get().deck ?? undefined, seed });
      set({ view, prev: null, status: 'ready' });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ status: 'idle', error: msg });
    }
  },

  dispatch: async (move) => {
    const before = get().view;
    try {
      const res = await sendMove(move);
      if (!res.ok) {
        // Illegal move: server state is unchanged; surface the reason, keep selection.
        set({ error: res.error ?? 'illegal move', view: res.state });
        return;
      }
      set({ prev: before, view: res.state, error: null, selection: null });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ error: msg });
    }
  },

  runBot: async () => {
    const before = get().view;
    try {
      const res = await botMove();
      set({ prev: before, view: res.state, error: null });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ error: msg });
    }
  },

  select: (sel) => set({ selection: sel }),
  clearSelection: () => set({ selection: null }),
}));
