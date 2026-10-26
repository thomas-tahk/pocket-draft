import { useEffect, useState } from 'react';
import type { HoverPayload } from './CardTile';

export type HoverState = HoverPayload | null;

type Props = {
  hover: HoverState;
  // Draft offers: park the preview over the side panel instead of beside the cursor, so it never covers the offers.
  dock?: boolean;
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

export function CardPreview({ hover, dock = false }: Props) {
  const [vp, setVp] = useState({ w: 1280, h: 800 });

  useEffect(() => {
    const update = () => setVp({ w: window.innerWidth, h: window.innerHeight });
    update();
    window.addEventListener('resize', update);
    return () => window.removeEventListener('resize', update);
  }, []);

  if (!hover) return null;
  const { card } = hover;
  const src = card.imageFull || card.imageThumb;
  if ('zoom' in hover) {
    return (
      <div className="card-zoom" data-testid="card-zoom">
        <img src={src} alt={card.name} />
      </div>
    );
  }
  // The side panel is beside the offers above 700 px; below that the cursor-follow preview is used.
  if (dock && vp.w > 700) {
    return (
      <div className="card-preview card-preview-dock" data-testid="card-preview">
        <img src={src} alt={card.name} />
      </div>
    );
  }
  const { left, top } = placeNearCursor(hover.position.x, hover.position.y, vp.w, vp.h);
  return (
    <div className="card-preview" data-testid="card-preview" style={{ left, top }}>
      <img src={src} alt={card.name} />
    </div>
  );
}
