import { create } from 'zustand';
import type { GameView, Move } from './types';
import { newGame, getState, sendMove, botMove } from './api';
import { resumeView } from '../battle/leaveBattle';
import type { InProgressBattle, LogStep } from '../stores/run';

type State = {
  view: GameView | null;
  status: 'idle' | 'loading' | 'ready';
  error: string | null;
  deck: string[] | null; // the deck this game was started from, for restart
  seed: number | null; // the seed of the game on show
  log: LogStep[]; // every step applied to it, for resuming (#117)
  botStuck: boolean; // last bot attempt made no progress — stop auto-driving it
};

type Actions = {
  init: () => Promise<void>;
  startGame: (deck: string[], seed?: number) => Promise<void>;
  resumeGame: (saved: InProgressBattle) => Promise<void>;
  restart: (seed?: number) => Promise<void>;
  dispatch: (move: Move) => Promise<void>;
  runBot: () => Promise<void>;
};

// The app picks each game's seed (below 2^53, so it survives JSON) so the record it
// saves replays exactly; the engine's own random seed is a full 64 bits.
const freshSeed = () => Math.floor(Math.random() * Number.MAX_SAFE_INTEGER);

// Moves and bot steps reach the server one at a time, so the log is in the order the server applied
// them. Two in flight at once could be applied in either order, and a rebuild would then diverge.
let inFlight: Promise<unknown> = Promise.resolve();
function serially(job: () => Promise<void>): Promise<void> {
  const run = inFlight.then(job);
  inFlight = run.catch(() => {});
  return run;
}

export const useGameStore = create<State & Actions>((set, get) => ({
  view: null,
  status: 'idle',
  error: null,
  deck: null,
  seed: null,
  log: [],
  botStuck: false,

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

  startGame: async (deck, seed = freshSeed()) => {
    set({ status: 'loading', error: null, deck, view: null, seed, log: [] });
    return serially(async () => {
      try {
        const view = await newGame({ deck, seed });
        set({ view, status: 'ready', botStuck: false });
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e);
        set({ status: 'idle', error: msg });
      }
    });
  },

  resumeGame: async (saved) => {
    set({ status: 'loading', error: null, deck: saved.yourDeck, view: null, seed: saved.seed, log: saved.log });
    return serially(async () => {
      try {
        const view = await resumeView(saved);
        set({ view, status: 'ready', botStuck: false });
      } catch (e) {
        set({ status: 'idle', error: e instanceof Error ? e.message : String(e) });
      }
    });
  },

  restart: async (seed = freshSeed()) => {
    set({ status: 'loading', error: null });
    return serially(async () => {
      try {
        const deck = get().deck;
        const view = await newGame({ deck: deck ?? undefined, seed });
        set({ view, status: 'ready', botStuck: false, seed, log: [] });
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e);
        set({ status: 'idle', error: msg });
      }
    });
  },

  dispatch: (move) => serially(async () => {
    try {
      const res = await sendMove(move);
      if (!res.ok) {
        // Illegal move: server state is unchanged; surface the reason.
        set({ error: res.error ?? 'illegal move', view: res.state });
        return;
      }
      // A human/dispatched move can re-open the bot's turn, so clear any stuck flag.
      set({ view: res.state, error: null, botStuck: false, log: [...get().log, move] });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ error: msg });
    }
  }),

  runBot: () => serially(async () => {
    try {
      const res = await botMove();
      // A concede may have ended the game while this bot move was in flight.
      if (get().view?.phase === 'over') return;
      // If the bot couldn't act, flag it so the auto-driver stops instead of spinning.
      set({ view: res.state, error: null, botStuck: !res.acted, log: res.acted ? [...get().log, 'bot'] : get().log });
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      set({ error: msg });
    }
  }),
}));
