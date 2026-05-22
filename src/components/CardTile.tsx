import type { Card } from '../types/card';

type Props = {
  card: Card;
  onPick?: () => void;
  onHover?: (card: Card | null) => void;
  size?: 'lg' | 'sm';
};

export function CardTile({ card, onPick, onHover, size = 'lg' }: Props) {
  const width = size === 'lg' ? 180 : 84;
  const Tag = onPick ? 'button' : 'div';

  const handlers = onHover
    ? {
        onMouseEnter: () => onHover(card),
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
        display: 'inline-block',
        width,
        textAlign: 'center',
        margin: 2,
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
