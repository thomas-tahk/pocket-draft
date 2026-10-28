import { useEffect, useRef, type FocusEvent, type KeyboardEvent, type PointerEvent, type ReactNode } from 'react';
import type { Card } from '../types/card';
import { cardFacts } from '../data/rarity';

// A mouse hover follows the cursor; a touch long-press (zoom) shows the card large and centred.
// `anchor` (keyboard focus only) is the focused element's box, so a phone-sized preview can keep clear of it.
export type HoverPayload =
  | { card: Card; position: { x: number; y: number }; anchor?: { top: number; bottom: number } }
  | { card: Card; zoom: true };

// Same hold and drag thresholds as the battle screen's long-press zoom (src/battle/useZoom.ts).
const HOLD_MS = 350;
const DRAG_START_PX = 8;
// Some touch browsers send a stray mouse pointer right after a tap (e.g. over the card that
// replaced the tapped one); mouse hover is ignored for this long after any touch.
const TOUCH_QUIET_MS = 1000;
// The click that follows a zoomed press arrives right after the release; swallow it for this long.
const CLICK_WINDOW_MS = 300;
let lastTouchAt = -Infinity;
// Whether the last input was the keyboard: focus that arrives from a key press (even one moved by the
// app after a pick) shows the card preview; focus from a mouse or touch press does not.
let keyboardMode = false;
if (typeof document !== 'undefined') {
  document.addEventListener('keydown', () => (keyboardMode = true), true);
  document.addEventListener('pointerdown', () => (keyboardMode = false), true);
}

type Props = {
  card: Card;
  onPick?: () => void;
  onHover?: (payload: HoverPayload | null) => void;
  size?: 'lg' | 'sm';
  // The accessible name of a pickable tile when the plain card facts are not enough
  // (e.g. a tile that removes the card from the deck). Defaults to name + rarity.
  label?: string;
  // A tile with no button of its own is a focus stop (so the keyboard can reach its preview);
  // only where no other control of the card is on screen.
  focusable?: boolean;
};

// The preview a keyboard user gets when a card's own control takes focus (as a mouse hover gives it).
// Escape hides it until focus moves to another card. Used by tiles and by FocusPreview cells.
export function useFocusPreview(card: Card, onHover?: (p: HoverPayload | null) => void) {
  const shown = useRef(false);
  const dismissed = useRef(false);
  const hoverRef = useRef(onHover);
  hoverRef.current = onHover;
  const hide = () => {
    if (shown.current) {
      shown.current = false;
      hoverRef.current?.(null);
    }
  };
  useEffect(() => hide, []); // eslint-disable-line react-hooks/exhaustive-deps
  return {
    shown,
    hide,
    onFocus: (e: FocusEvent<HTMLElement>) => {
      if (!keyboardMode || dismissed.current || !hoverRef.current) return;
      const r = e.currentTarget.getBoundingClientRect();
      shown.current = true;
      hoverRef.current({ card, position: { x: r.right - r.width / 3, y: r.top + r.height / 3 }, anchor: { top: r.top, bottom: r.bottom } });
    },
    onBlur: (e: FocusEvent<HTMLElement>) => {
      if (e.currentTarget.contains(e.relatedTarget as Node | null)) return;
      dismissed.current = false;
      hide();
    },
    onKeyDown: (e: KeyboardEvent) => {
      if (e.key === 'Escape' && shown.current) {
        e.stopPropagation();
        dismissed.current = true;
        hide();
      }
    },
  };
}

// A card's cell with its own buttons (shop and deckbuild +/-): focus anywhere inside previews the card.
export function FocusPreview({ card, onHover, className, children }: { card: Card; onHover: (p: HoverPayload | null) => void; className?: string; children: ReactNode }) {
  const f = useFocusPreview(card, onHover);
  return (
    <div className={className} onFocus={f.onFocus} onBlur={f.onBlur} onKeyDown={f.onKeyDown}>
      {children}
    </div>
  );
}

export function CardTile({ card, onPick, onHover, size = 'lg', label, focusable }: Props) {
  const Tag = onPick ? 'button' : 'div';
  const holdTimer = useRef<number | undefined>(undefined);
  const clickTimer = useRef<number | undefined>(undefined);
  const pressStart = useRef({ x: 0, y: 0 });
  // The press opened the zoom: its release must not pick the card, even after a drag.
  const swallowClick = useRef(false);
  // This tile currently owns the parent's hover/zoom state.
  const showing = useRef(false);
  const hoverRef = useRef(onHover);
  hoverRef.current = onHover;

  const show = (payload: HoverPayload | null) => {
    showing.current = payload !== null;
    hoverRef.current?.(payload);
  };

  // Unmounting (e.g. the pack changed after a pick) must not leave a preview or timer behind.
  useEffect(
    () => () => {
      window.clearTimeout(holdTimer.current);
      window.clearTimeout(clickTimer.current);
      if (showing.current) hoverRef.current?.(null);
    },
    [],
  );

  // Hover zoom is for a mouse; a tap on touch must not open it.
  const trackHover = (e: PointerEvent) => {
    if (e.pointerType !== 'mouse' || Date.now() - lastTouchAt < TOUCH_QUIET_MS) return;
    show({ card, position: { x: e.clientX, y: e.clientY } });
  };

  const endPress = () => {
    lastTouchAt = Date.now();
    window.clearTimeout(holdTimer.current);
    if (showing.current) show(null);
    // A release is followed by a click; after the window, a later click is a fresh one.
    window.clearTimeout(clickTimer.current);
    clickTimer.current = window.setTimeout(() => (swallowClick.current = false), CLICK_WINDOW_MS);
  };

  // A cancelled press (scroll, system gesture) is never followed by a click.
  const cancelPress = () => {
    endPress();
    swallowClick.current = false;
  };

  const startPress = (e: PointerEvent) => {
    window.clearTimeout(clickTimer.current);
    swallowClick.current = false;
    if (e.pointerType === 'mouse') return;
    lastTouchAt = Date.now();
    pressStart.current = { x: e.clientX, y: e.clientY };
    holdTimer.current = window.setTimeout(() => {
      swallowClick.current = true;
      show({ card, zoom: true });
    }, HOLD_MS);
  };

  const movePress = (e: PointerEvent) => {
    if (e.pointerType === 'mouse') return trackHover(e);
    const { x, y } = pressStart.current;
    if (Math.hypot(e.clientX - x, e.clientY - y) <= DRAG_START_PX) return;
    window.clearTimeout(holdTimer.current);
    if (showing.current) show(null);
  };

  // The tile's own preview state is shared with the hover's (`showing`), so a focus preview is cleaned up with it.
  const focus = useFocusPreview(card, (p) => (p ? show(p) : showing.current && show(null)));
  const facts = cardFacts(card);
  const handlers = onHover
    ? {
        onFocus: focus.onFocus,
        onBlur: focus.onBlur,
        onKeyDown: focus.onKeyDown,
        onPointerEnter: trackHover,
        onPointerMove: movePress,
        onPointerLeave: (e: PointerEvent) => (e.pointerType === 'mouse' ? show(null) : endPress()),
        onPointerDown: startPress,
        onPointerUp: endPress,
        onPointerCancel: cancelPress,
        onContextMenu: (e: { preventDefault: () => void }) => e.preventDefault(),
      }
    : {};

  const click = () => {
    if (swallowClick.current) {
      swallowClick.current = false;
      return;
    }
    onPick?.();
  };

  return (
    <Tag
      className={`tile-${size}`}
      data-type={card.cardType}
      onClick={click}
      {...handlers}
      aria-label={onPick ? (label ?? facts) : facts}
      role={onPick ? undefined : 'group'}
      tabIndex={!onPick && focusable ? 0 : undefined}
      title={`${card.name} · ${card.draftRarity}${card.isMegaEx ? ' Mega-EX' : card.isEx ? ' EX' : ''} · ${card.id}`}
      style={{
        all: 'unset',
        cursor: onPick ? 'pointer' : 'default',
        display: 'block',
        width: 'var(--tile-w)',
        textAlign: 'center',
        userSelect: 'none',
        WebkitTouchCallout: 'none',
      }}
    >
      <img
        src={card.imageThumb}
        alt={onPick ? card.name : ''}
        loading="lazy"
        draggable={false}
        className="card-img"
      />
      {size === 'lg' && (
        <div className="card-cap" aria-hidden="true">
          {card.name} · {card.draftRarity}
          {card.isMegaEx ? ' M-EX' : card.isEx ? ' ex' : ''}
        </div>
      )}
    </Tag>
  );
}
