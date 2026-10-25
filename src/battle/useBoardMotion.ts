import { useLayoutEffect, useRef, type RefObject } from 'react';
import type { GameView } from '../game/types';
import { moveCards, popNewPips, reducedMotion, removedCards, settle, snapshot, type Snap } from './motion';
import { resultsLand } from './results';

// Runs every board animation when the view changes (motion.ts: where cards
// go; results.ts: what the moves did). Skip (instant) and reduced motion run none.
export function useBoardMotion(root: RefObject<HTMLElement | null>, view: GameView, instant: boolean) {
  const last = useRef<{ snaps: Map<string, Snap>; view: GameView } | null>(null);
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
    const prev = last.current;
    const after = snapshot(el);
    last.current = { snaps: after, view };
    const newGame = !prev || view.actions.length < prev.view.actions.length;
    if (newGame || prev.view === view || skip.current || reducedMotion()) return;
    const change = { root: el, was: prev.snaps, after, before: prev.view, now: view };
    moveCards(change);
    removedCards(change);
    resultsLand(change);
    popNewPips(el, prev.view, view);
  }, [root, view]);
}
