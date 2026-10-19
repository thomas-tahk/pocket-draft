import { useCallback, useEffect, useRef, useState } from 'react';
import type { CardView, GameView } from '../game/types';

// Hold to zoom (spec F2): a card pressed ~350 ms without moving opens large and
// centred until the press ends or the drag starts. Hold time is a placeholder
// the owner may change (questions file).

const HOLD_MS = 350;
export const DRAG_START_PX = 8;

type Press = { card: CardView; x: number; y: number };

// The card under the pointer: a hand card (data-pick) or a card in play (data-zoom).
export function zoomCardAt(el: Element | null, view: GameView): CardView | null {
  const [you, opp] = view.players;
  const pick = el?.closest('[data-pick^="hand:"]')?.getAttribute('data-pick');
  if (pick) return you.hand[Number(pick.slice(5))] ?? null;
  const id = el?.closest('[data-zoom]')?.getAttribute('data-zoom');
  if (!id) return null;
  const inPlay = [you, opp].flatMap((p) => [p.active, ...p.bench]);
  return inPlay.find((m) => m?.card.id === id)?.card ?? null;
}

export function useZoom() {
  const [press, setPress] = useState<Press | null>(null);
  const [zoom, setZoom] = useState<CardView | null>(null);
  // Whether the current press opened the zoom; read on release, after the zoom has closed.
  const opened = useRef(false);

  useEffect(() => {
    if (!press) return;
    const timer = window.setTimeout(() => { opened.current = true; setZoom(press.card); }, HOLD_MS);
    const end = () => { setPress(null); setZoom(null); };
    const move = (e: PointerEvent) => {
      if (Math.hypot(e.clientX - press.x, e.clientY - press.y) > DRAG_START_PX) end();
    };
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', end);
    window.addEventListener('pointercancel', end);
    return () => {
      window.clearTimeout(timer);
      window.removeEventListener('pointermove', move);
      window.removeEventListener('pointerup', end);
      window.removeEventListener('pointercancel', end);
    };
  }, [press]);

  // Called on every press; null when the press isn't on a card.
  const startPress = useCallback((p: Press | null) => { opened.current = false; setPress(p); }, []);
  return { zoom, opened, startPress };
}
