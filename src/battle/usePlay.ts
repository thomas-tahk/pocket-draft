import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';
import type { PointerEvent as ReactPointerEvent, RefObject } from 'react';
import { useGameStore } from '../game/store';
import type { CardView, GameView, Move } from '../game/types';
import { DRAG_START_PX, useZoom, zoomCardAt } from './useZoom';

// Press-and-drag play (spec F1, F3–F5, F17). What glows and what a drop does
// come only from the server's legalMoves; the client knows no rules. Mouse and
// touch share one pointer-event path. A press that never moves becomes a tap:
// the card stays held until the next tap plays it or cancels (tap-then-tap).
// A press held long enough to zoom (F2) and then released lets the card go.

// A spot a card can be dropped on: your Active or Bench slot 0–2, the
// opponent's ('opp-' prefix), the Stadium slot, or 'board' for a Trainer with
// no target, dropped anywhere on the board but the hands (F7).
type Pos = 'active' | 'bench-0' | 'bench-1' | 'bench-2';
export type Spot = Pos | `opp-${Pos}` | 'stadium' | 'board';
type Held = { kind: 'hand'; index: number; card: CardView } | { kind: 'energy'; type: string };
type Point = { x: number; y: number };
export type Holding = { held: Held; moves: Map<Spot, Move>; mode: 'press' | 'drag' | 'tap'; start: Point; pos: Point };

const FLASH_MS = 1800;

// In-play target numbers: 0 is the Active, n is Bench slot n-1.
const spotOfTarget = (t: number): Pos => (t === 0 ? 'active' : (`bench-${t - 1}` as Pos));

// Where a PlayTrainer move drops (F6, F7). Player 0 is you (humanSeat).
function trainerSpot(spot?: { player: number; pos: number }): Spot {
  if (!spot) return 'board';
  if (spot.player === -1) return 'stadium';
  return spot.player === 0 ? spotOfTarget(spot.pos) : `opp-${spotOfTarget(spot.pos)}`;
}

function movesFor(held: Held, legal: Move[]): Map<Spot, Move> {
  const out = new Map<Spot, Move>();
  for (const m of legal) {
    if (held.kind === 'energy' && m.type === 'AttachEnergy') out.set(spotOfTarget(m.target), m);
    if (held.kind !== 'hand') continue;
    if (m.type === 'PlayBasic' && m.cardId === held.card.id) out.set(`bench-${m.slot}` as Spot, m);
    if (m.type === 'SetupActive' && m.cardId === held.card.id) out.set('active', m);
    if (m.type === 'SetupBench' && m.cardId === held.card.id) out.set(`bench-${m.slot}` as Spot, m);
    if (m.type === 'Evolve' && m.handCardId === held.card.id) out.set(spotOfTarget(m.target), m);
    if (m.type === 'PlayTrainer' && m.cardId === held.card.id) out.set(trainerSpot(m.spot), m);
  }
  return out;
}

// The bottom line while choosing. Wording is the owner's to change (questions file).
export function hintFor(held: Held, moves: Map<Spot, Move>): string {
  if (held.kind === 'energy') return 'Choose a Pokémon to attach energy to';
  if (moves.get('active')?.type === 'SetupActive') return `Place ${held.card.name} as your Active`;
  if (moves.has('board')) return `Drop ${held.card.name} anywhere on the board`;
  if (moves.has('stadium')) return `Drop ${held.card.name} on the Stadium spot`;
  if (held.card.kind !== 'pokemon') return `Choose a Pokémon for ${held.card.name}`;
  if (held.card.stage === 'Basic') return `Choose a Bench spot for ${held.card.name}`;
  return `Choose a Pokémon to evolve into ${held.card.name}`;
}

function heldAt(el: Element | null, view: GameView): Held | null {
  const pick = el?.closest('[data-pick]')?.getAttribute('data-pick');
  const you = view.players[0];
  if (pick === 'energy') return { kind: 'energy', type: you.energyZone };
  if (!pick?.startsWith('hand:')) return null;
  const index = Number(pick.slice(5));
  return you.hand[index] ? { kind: 'hand', index, card: you.hand[index] } : null;
}

const spotAt = (el: Element | null) => (el?.closest('[data-spot]')?.getAttribute('data-spot') ?? null) as Spot | null;

// The spot a drop on el lands on: a marked spot, else 'board' when the held
// card plays anywhere and el is on the board outside both hands (F7).
function dropAt(el: Element | null, moves: Map<Spot, Move>): Move | undefined {
  const spot = spotAt(el);
  if (spot && moves.has(spot)) return moves.get(spot);
  const onBoard = el?.closest('.bt-board') && !el.closest('[data-testid="my-hand"], [data-testid="opp-hand"]');
  return onBoard ? moves.get('board') : undefined;
}

const sameHeld = (a: Held, b: Held) => a.kind === b.kind && (a.kind === 'energy' || a.index === (b as typeof a).index);

function blockedReason(held: Held, view: GameView): string {
  const you = view.players[0];
  const reason = held.kind === 'energy' ? you.energyBlocked : you.handBlocked?.[held.card.id];
  return reason || 'That card has nowhere to go right now';
}

export function usePlay(view: GameView) {
  const dispatch = useGameStore((s) => s.dispatch);
  const [holding, setHolding] = useState<Holding | null>(null);
  const [flash, setFlash] = useState<string | null>(null);
  const flashTimer = useRef<number | undefined>(undefined);
  const busy = useRef(false);
  const { zoom, opened, startPress } = useZoom();

  const showReason = useCallback((reason: string) => {
    window.clearTimeout(flashTimer.current);
    setFlash(reason);
    flashTimer.current = window.setTimeout(() => setFlash(null), FLASH_MS);
  }, []);

  const play = useCallback(async (move: Move) => {
    setHolding(null);
    busy.current = true;
    await dispatch(move);
    busy.current = false;
    const err = useGameStore.getState().error;
    if (err) showReason(err);
  }, [dispatch, showReason]);

  // Board-level pointerdown: plays or cancels a tap-held card, or picks one up.
  const onPointerDown = (e: ReactPointerEvent) => {
    if (busy.current) return;
    const target = e.target as Element;
    const held = heldAt(target, view);
    const zoomCard = zoomCardAt(target, view);
    if (holding?.mode === 'tap') {
      const move = dropAt(target, holding.moves);
      if (move) return void play(move);
      setHolding(null);
    }
    startPress(zoomCard && { card: zoomCard, x: e.clientX, y: e.clientY });
    if (holding?.mode === 'tap' && (!held || sameHeld(held, holding.held))) return;
    if (!held) return;
    const moves = movesFor(held, view.players[0].legalMoves ?? []);
    if (moves.size === 0) return showReason(blockedReason(held, view));
    const p = { x: e.clientX, y: e.clientY };
    setHolding({ held, moves, mode: 'press', start: p, pos: p });
  };

  usePointerTracking({ holding, setHolding, play, zoomed: opened });
  return { holding, flash, zoom, zoomOpened: opened, onPointerDown };
}

type Tracking = { holding: Holding | null; setHolding: (h: Holding | null) => void; play: (m: Move) => void; zoomed: RefObject<boolean> };

// While a card is pressed or dragged, follow the pointer anywhere on the page.
function usePointerTracking({ holding, setHolding, play, zoomed }: Tracking) {
  const active = holding && holding.mode !== 'tap';
  useEffect(() => {
    if (!active || !holding) return;
    const move = (e: PointerEvent) => {
      const pos = { x: e.clientX, y: e.clientY };
      const moved = Math.hypot(pos.x - holding.start.x, pos.y - holding.start.y) > DRAG_START_PX;
      setHolding({ ...holding, pos, mode: holding.mode === 'drag' || moved ? 'drag' : 'press' });
    };
    const up = (e: PointerEvent) => {
      if (holding.mode === 'press') return setHolding(zoomed.current ? null : { ...holding, mode: 'tap' });
      const m = dropAt(document.elementFromPoint(e.clientX, e.clientY), holding.moves);
      if (m) play(m);
      else setHolding(null); // dropped anywhere else: cancel (F3)
    };
    const cancel = () => setHolding(null);
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', up);
    window.addEventListener('pointercancel', cancel);
    return () => {
      window.removeEventListener('pointermove', move);
      window.removeEventListener('pointerup', up);
      window.removeEventListener('pointercancel', cancel);
    };
  }, [active, holding, setHolding, play, zoomed]);
}

// Lets board pieces ask whether they glow or are the held card.
export const PlayContext = createContext<Holding | null>(null);
export const useHolding = () => useContext(PlayContext);
