import { useEffect, useState } from 'react';
import type { Card } from '../types/card';

export type HoverState = {
  card: Card;
  position: { x: number; y: number };
} | null;

type Props = {
  hover: HoverState;
};

const PREVIEW_W = 260;
const PREVIEW_H = 360;
const OFFSET = 20;

function placeNearCursor(x: number, y: number, winW: number, winH: number) {
  let left = x + OFFSET;
  let top = y + OFFSET;
  if (left + PREVIEW_W > winW) left = x - PREVIEW_W - OFFSET;
  if (left < 8) left = 8;
  if (top + PREVIEW_H > winH) top = y - PREVIEW_H - OFFSET;
  if (top < 8) top = 8;
  return { left, top };
}

export function CardPreview({ hover }: Props) {
  const [vp, setVp] = useState({ w: 1280, h: 800 });

  useEffect(() => {
    const update = () => setVp({ w: window.innerWidth, h: window.innerHeight });
    update();
    window.addEventListener('resize', update);
    return () => window.removeEventListener('resize', update);
  }, []);

  if (!hover) return null;
  const { card, position } = hover;
  const src = card.imageFull || card.imageThumb;
  const { left, top } = placeNearCursor(position.x, position.y, vp.w, vp.h);
  return (
    <div
      style={{
        position: 'fixed',
        left,
        top,
        width: PREVIEW_W,
        pointerEvents: 'none',
        zIndex: 100,
        boxShadow: '0 8px 24px rgba(0,0,0,0.45)',
        borderRadius: 12,
        overflow: 'hidden',
        background: '#0008',
      }}
    >
      <img src={src} alt={card.name} style={{ width: '100%', display: 'block' }} />
    </div>
  );
}
