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
};

type Actions = {
  init: () => Promise<void>;
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

  init: async () => {
    set({ status: 'loading', error: null });
    const view = await getState();
    set({ view, prev: null, status: 'ready' });
  },

  restart: async (seed) => {
    set({ status: 'loading', error: null, selection: null });
    const view = await newGame(seed);
    set({ view, prev: null, status: 'ready' });
  },

  dispatch: async (move) => {
    const before = get().view;
    const res = await sendMove(move);
    if (!res.ok) {
      // Illegal move: server state is unchanged; surface the reason, keep selection.
      set({ error: res.error ?? 'illegal move', view: res.state });
      return;
    }
    set({ prev: before, view: res.state, error: null, selection: null });
  },

  runBot: async () => {
    const before = get().view;
    const res = await botMove();
    set({ prev: before, view: res.state, error: null });
  },

  select: (sel) => set({ selection: sel }),
  clearSelection: () => set({ selection: null }),
}));
