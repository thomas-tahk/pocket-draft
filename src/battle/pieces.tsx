import type { CSSProperties } from 'react';
import type { CardView } from '../game/types';

export { EnergyIcon } from './energy';

// Small drawn pieces of the board. There are no icon assets in the repo, so
// energy, card backs and Poké Balls are CSS.

const unit = (n: number) => ({ '--w': n }) as CSSProperties;

export function CardBack({ width }: { width: number }) {
  return <div className="bt-back" style={unit(width)} data-testid="card-back" aria-label="face-down card" />;
}

export function CardFace({ card, width }: { card: CardView; width: number }) {
  return (
    <div className="bt-card" style={unit(width)} title={card.name}>
      {card.image ? <img src={card.image} alt={card.name} draggable={false} /> : <div className="bt-card-name">{card.name}</div>}
    </div>
  );
}

export function EmptySlot({ width }: { width: number }) {
  return <div className="bt-card bt-slot-empty" style={unit(width)} aria-label="empty slot" />;
}

export function PointPips({ points }: { points: number }) {
  return (
    <span className="bt-pips" aria-label={`${points} of 3 points`}>
      {[0, 1, 2].map((i) => <span key={i} className="bt-pip" data-testid="pip" data-filled={i < points} />)}
    </span>
  );
}
