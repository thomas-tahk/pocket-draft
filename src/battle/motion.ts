import { useLayoutEffect, useRef, type RefObject } from 'react';
import type { GameView, InPlayView, PlayerView } from '../game/types';

// Motion groundwork (M3-2): when the board view changes, every card that
// changed place slides from where it was to where it is now (FLIP, with the
// Web Animations API). The engine sends no per-card instance id, so a card's
// key is made here: player + card id + its count among that player's visible
// copies. Two copies of one card look the same, so which copy "moved" is
// never visible. Look and timing are placeholders for the owner (docs/questions/M3.md).

export const FLIP_MS = 220; // owner: each animation 400 ms at most, "doesn't have to be fancy"
const FX = 'fx'; // Animation.id of every animation made here

// What a spot's key is, per zone of one player's side.
export type SideKeys = { hand: string[]; active: string; activeTool: string; bench: string[]; benchTool: string[] };

function counter(side: number) {
  const seen = new Map<string, number>();
  return (id: string) => {
    const n = seen.get(id) ?? 0;
    seen.set(id, n + 1);
    return `p${side}:${id}#${n}`;
  };
}

// Face-down cards (an opponent's hand, a hidden setup pick) have no identity; they are keyed by place.
export function sideKeys(side: number, p: PlayerView): SideKeys {
  const next = counter(side);
  const mon = (m: InPlayView | null, at: string) => (!m ? '' : m.faceDown ? `p${side}:down:${at}` : next(m.card.id));
  const tool = (m: InPlayView | null) => (m?.tool ? next(m.tool.id) : '');
  const hand = side === 0 ? p.hand.map((c) => next(c.id)) : Array.from({ length: p.handCount }, (_, i) => `p${side}:hand#${i}`);
  const active = mon(p.active, 'a');
  const activeTool = tool(p.active);
  const bench = p.bench.map((m, i) => mon(m, `b${i}`));
  return { hand, active, activeTool, bench, benchTool: p.bench.map(tool) };
}

export const viewKeys = (v: GameView): [SideKeys, SideKeys] => [sideKeys(0, v.players[0]), sideKeys(1, v.players[1])];

type Snap = { x: number; y: number; w: number };

export function snapshot(root: HTMLElement): Map<string, Snap> {
  const out = new Map<string, Snap>();
  root.querySelectorAll<HTMLElement>('[data-fx]').forEach((el) => {
    const r = el.getBoundingClientRect();
    out.set(el.dataset.fx!, { x: r.left + r.width / 2, y: r.top + r.height / 2, w: el.offsetWidth });
  });
  return out;
}

export const reducedMotion = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches;

// Slides el from a snapshot's place to where it is now. Uses the individual
// `translate` and `scale` properties so a card's own `transform` is left alone.
export function slide(el: HTMLElement, from: Snap, to: Snap, ms = FLIP_MS): Animation | null {
  const dx = from.x - to.x;
  const dy = from.y - to.y;
  const s = from.w && to.w ? from.w / to.w : 1;
  if (Math.abs(dx) + Math.abs(dy) < 2 && Math.abs(s - 1) < 0.02) return null;
  const a = el.animate(
    [{ translate: `${dx}px ${dy}px`, scale: String(s), zIndex: 40 }, { translate: '0px 0px', scale: '1', zIndex: 40 }],
    { duration: ms, easing: 'ease-out' },
  );
  a.id = FX;
  return a;
}

// Animations of this module that are still running finish at once.
export function settle(root: HTMLElement) {
  root.getAnimations({ subtree: true }).filter((a) => a.id === FX).forEach((a) => a.finish());
}

export function useBoardMotion(root: RefObject<HTMLElement | null>, view: GameView, instant: boolean) {
  const before = useRef<Map<string, Snap> | null>(null);
  const skip = useRef(instant);
  skip.current = instant;
  // Skip finishes whatever is still moving.
  useLayoutEffect(() => {
    if (instant && root.current) settle(root.current);
  }, [root, instant]);
  useLayoutEffect(() => {
    const el = root.current;
    if (!el) return;
    settle(el);
    const after = snapshot(el);
    const was = before.current;
    before.current = after;
    if (!was || skip.current || reducedMotion()) return;
    el.querySelectorAll<HTMLElement>('[data-fx]').forEach((node) => {
      const key = node.dataset.fx!;
      const from = was.get(key);
      if (from) slide(node, from, after.get(key)!);
    });
  }, [root, view]);
}
