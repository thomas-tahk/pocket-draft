import { create } from 'zustand';
import type { GameView, Move } from './types';
import { newGame, getState, sendMove, botMove } from './api';

type State = {
  view: GameView | null;
  status: 'idle' | 'loading' | 'ready';
  error: string | null;
  deck: string[] | null; // the deck this game was started from, for restart
  botStuck: boolean; // last bot attempt made no progress — stop auto-driving it
  leftOut: number; // cards of `deck` the server left out of this game (it can't play them yet)
};

type Actions = {
  init: () => Promise<void>;
  startGame: (deck: string[], seed?: number) => Promise<void>;
  restart: (seed?: number) => Promise<void>;
  dispatch: (move: Move) => Promise<void>;
  runBot: () => Promise<void>;
};

// A fresh game holds your whole deck in hand and draw pile, so any shortfall is
// cards the server skipped (e.g. Trainers, whose effects aren't built yet).
function leftOut(deck: string[] | null, view: GameView): number {
  if (!deck) return 0;
  const you = view.players[0];
  return deck.length - (you.handCount + you.deckCount);
}

export const useGameStore = create<State & Actions>((set, get) => ({
  view: null,
  status: 'idle',
  error: null,
  deck: null,
  botStuck: false,
  leftOut: 0,

  init: async () => {
    set({ status: 'loading', error: null });
    try {
      const view = await getState();
      set({ view, status: 'ready', botStuck: false });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ status: 'idle', error: msg });
    }
  },

  startGame: async (deck, seed) => {
    set({ status: 'loading', error: null, deck });
    try {
      const view = await newGame({ deck, seed });
      set({ view, status: 'ready', botStuck: false, leftOut: leftOut(deck, view) });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ status: 'idle', error: msg });
    }
  },

  restart: async (seed) => {
    set({ status: 'loading', error: null });
    try {
      const deck = get().deck;
      const view = await newGame({ deck: deck ?? undefined, seed });
      set({ view, status: 'ready', botStuck: false, leftOut: leftOut(deck, view) });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ status: 'idle', error: msg });
    }
  },

  dispatch: async (move) => {
    try {
      const res = await sendMove(move);
      if (!res.ok) {
        // Illegal move: server state is unchanged; surface the reason.
        set({ error: res.error ?? 'illegal move', view: res.state });
        return;
      }
      // A human/dispatched move can re-open the bot's turn, so clear any stuck flag.
      set({ view: res.state, error: null, botStuck: false });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ error: msg });
    }
  },

  runBot: async () => {
    try {
      const res = await botMove();
      // If the bot couldn't act, flag it so the auto-driver stops instead of spinning.
      set({ view: res.state, error: null, botStuck: !res.acted });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ error: msg });
    }
  },
}));
