import type { CardView } from './types';

type Props = {
  card: CardView;
  size?: 'lg' | 'sm';
  selected?: boolean;
  highlight?: boolean; // legal target for the current selection
  disabled?: boolean;
  onClick?: () => void;
  title?: string;
};

export function GameCard({ card, size = 'lg', selected, highlight, disabled, onClick, title }: Props) {
  const width = size === 'lg' ? 140 : 84;
  const Tag = onClick ? 'button' : 'div';
  const ring = selected ? '#2b5cff' : highlight ? '#ffce54' : 'transparent';

  return (
    <Tag
      onClick={onClick}
      disabled={disabled}
      title={title ?? `${card.name} · ${card.id}`}
      style={{
        all: 'unset',
        cursor: onClick && !disabled ? 'pointer' : 'default',
        display: 'block',
        width,
        opacity: disabled ? 0.45 : 1,
        outline: `3px solid ${ring}`,
        outlineOffset: 2,
        borderRadius: 8,
      }}
    >
      {card.image ? (
        <img src={card.image} alt={card.name} loading="lazy" style={{ width: '100%', borderRadius: 8, display: 'block' }} />
      ) : (
        <div style={{ width: '100%', aspectRatio: '5 / 7', borderRadius: 8, background: '#33363d', color: '#e6e6e8', display: 'grid', placeItems: 'center', fontSize: 12, textAlign: 'center', padding: 4 }}>
          {card.name}
        </div>
      )}
    </Tag>
  );
}
