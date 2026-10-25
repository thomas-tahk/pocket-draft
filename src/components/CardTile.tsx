import { useRef, type PointerEvent } from 'react';
import type { Card } from '../types/card';

export type HoverPayload = {
  card: Card;
  position: { x: number; y: number };
  // Set by a touch long-press: show the card large and centred until the press ends.
  zoom?: boolean;
};

// Same hold and drag thresholds as the battle screen's long-press zoom (src/battle/useZoom.ts).
const HOLD_MS = 350;
const DRAG_START_PX = 8;
// Some touch browsers send a stray mouse pointer right after a tap (e.g. over the card that
// replaced the tapped one); mouse hover is ignored for this long after any touch.
const TOUCH_QUIET_MS = 1000;
let lastTouchAt = -Infinity;

type Props = {
  card: Card;
  onPick?: () => void;
  onHover?: (payload: HoverPayload | null) => void;
  size?: 'lg' | 'sm';
};

export function CardTile({ card, onPick, onHover, size = 'lg' }: Props) {
  const Tag = onPick ? 'button' : 'div';
  const holdTimer = useRef<number | undefined>(undefined);
  const pressStart = useRef({ x: 0, y: 0 });
  // Whether the current press opened the zoom; swallows the click that follows the release.
  const zoomed = useRef(false);

  // Hover zoom is for a mouse; a tap on touch must not open it.
  const trackHover = (e: PointerEvent) => {
    if (e.pointerType !== 'mouse' || Date.now() - lastTouchAt < TOUCH_QUIET_MS) return;
    onHover?.({ card, position: { x: e.clientX, y: e.clientY } });
  };

  const endPress = () => {
    lastTouchAt = Date.now();
    window.clearTimeout(holdTimer.current);
    if (zoomed.current) onHover?.(null);
  };

  const startPress = (e: PointerEvent) => {
    if (e.pointerType === 'mouse') return;
    lastTouchAt = Date.now();
    zoomed.current = false;
    pressStart.current = { x: e.clientX, y: e.clientY };
    holdTimer.current = window.setTimeout(() => {
      zoomed.current = true;
      onHover?.({ card, position: pressStart.current, zoom: true });
    }, HOLD_MS);
  };

  const movePress = (e: PointerEvent) => {
    if (e.pointerType === 'mouse') return trackHover(e);
    const { x, y } = pressStart.current;
    if (Math.hypot(e.clientX - x, e.clientY - y) <= DRAG_START_PX) return;
    window.clearTimeout(holdTimer.current);
    if (zoomed.current) onHover?.(null);
    zoomed.current = false;
  };

  const handlers = onHover
    ? {
        onPointerEnter: trackHover,
        onPointerMove: movePress,
        onPointerLeave: (e: PointerEvent) => (e.pointerType === 'mouse' ? onHover(null) : endPress()),
        onPointerDown: startPress,
        onPointerUp: endPress,
        onPointerCancel: endPress,
        onContextMenu: (e: { preventDefault: () => void }) => e.preventDefault(),
      }
    : {};

  const click = () => {
    if (zoomed.current) {
      zoomed.current = false;
      return;
    }
    onPick?.();
  };

  return (
    <Tag
      className={`tile-${size}`}
      onClick={click}
      {...handlers}
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
        alt={card.name}
        loading="lazy"
        draggable={false}
        style={{ width: '100%', borderRadius: 8, display: 'block' }}
      />
      {size === 'lg' && (
        <div style={{ fontSize: 12, marginTop: 4, opacity: 0.75 }}>
          {card.name} · {card.draftRarity}
          {card.isMegaEx ? ' M-EX' : card.isEx ? ' ex' : ''}
        </div>
      )}
    </Tag>
  );
}
