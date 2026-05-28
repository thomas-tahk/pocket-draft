import type { MouseEvent } from 'react';
import type { Card } from '../types/card';

export type HoverPayload = {
  card: Card;
  position: { x: number; y: number };
};

type Props = {
  card: Card;
  onPick?: () => void;
  onHover?: (payload: HoverPayload | null) => void;
  size?: 'lg' | 'sm';
};

export function CardTile({ card, onPick, onHover, size = 'lg' }: Props) {
  const width = size === 'lg' ? 180 : 84;
  const Tag = onPick ? 'button' : 'div';

  const trackHover = (e: MouseEvent) => {
    onHover?.({ card, position: { x: e.clientX, y: e.clientY } });
  };

  const handlers = onHover
    ? {
        onMouseEnter: trackHover,
        onMouseMove: trackHover,
        onMouseLeave: () => onHover(null),
      }
    : {};

  return (
    <Tag
      onClick={onPick}
      {...handlers}
      title={`${card.name} · ${card.draftRarity}${card.isMegaEx ? ' Mega-EX' : card.isEx ? ' EX' : ''} · ${card.id}`}
      style={{
        all: 'unset',
        cursor: onPick ? 'pointer' : 'default',
        display: 'block',
        width,
        textAlign: 'center',
      }}
    >
      <img
        src={card.imageThumb}
        alt={card.name}
        loading="lazy"
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
