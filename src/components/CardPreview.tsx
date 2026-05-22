import type { Card } from '../types/card';

type Props = {
  card: Card | null;
};

// Fixed floating preview shown when a small tile is hovered. Anchored to the
// top-right of the viewport so it never overlaps the offer row or the cursor.
export function CardPreview({ card }: Props) {
  if (!card) return null;
  const src = card.imageFull || card.imageThumb;
  return (
    <div
      style={{
        position: 'fixed',
        top: 80,
        right: 24,
        width: 280,
        pointerEvents: 'none',
        zIndex: 50,
        boxShadow: '0 8px 24px rgba(0,0,0,0.35)',
        borderRadius: 12,
        overflow: 'hidden',
        background: '#0008',
      }}
    >
      <img
        src={src}
        alt={card.name}
        style={{ width: '100%', display: 'block' }}
      />
    </div>
  );
}
