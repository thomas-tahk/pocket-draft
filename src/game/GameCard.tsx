import { useState } from 'react';
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
  const [hovered, setHovered] = useState(false);
  const width = size === 'lg' ? 140 : 84;
  const Tag = onClick ? 'button' : 'div';
  const ring = selected ? '#2b5cff' : highlight ? '#ffce54' : 'transparent';
  const hoverable = !!onClick && !disabled;

  return (
    <Tag
      onClick={onClick}
      disabled={disabled}
      title={title ?? `${card.name} · ${card.id}`}
      onMouseEnter={hoverable ? () => setHovered(true) : undefined}
      onMouseLeave={hoverable ? () => setHovered(false) : undefined}
      style={{
        all: 'unset',
        cursor: onClick && !disabled ? 'pointer' : 'default',
        display: 'block',
        width,
        opacity: disabled ? 0.45 : 1,
        outline: `3px solid ${ring}`,
        outlineOffset: 2,
        borderRadius: 8,
        boxShadow: hoverable && hovered ? '0 4px 12px rgba(0, 0, 0, 0.35)' : '0 0 0 rgba(0, 0, 0, 0)',
        transition: 'box-shadow 120ms ease-out',
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
