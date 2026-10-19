import type { CSSProperties } from 'react';
import type { CardView } from '../game/types';

// Small drawn pieces of the board. There are no icon assets in the repo, so
// energy, card backs and Poké Balls are CSS.

const ENERGY_COLOR: Record<string, string> = {
  Grass: '#3a9d3f', Fire: '#e2481f', Water: '#1f86de', Lightning: '#e8b800',
  Psychic: '#9a4fc6', Fighting: '#b8611f', Darkness: '#2c3646', Metal: '#7f8d9b',
  Dragon: '#a88a1c', Colorless: '#9da3ad',
};

// One letter per type, as the TCG prints them (Fire is R, Darkness is D).
const ENERGY_LETTER: Record<string, string> = {
  Grass: 'G', Fire: 'R', Water: 'W', Lightning: 'L', Psychic: 'P',
  Fighting: 'F', Darkness: 'D', Metal: 'M', Dragon: 'N', Colorless: 'C',
};

const unit = (n: number) => ({ '--w': n }) as CSSProperties;

export function EnergyIcon({ type, size = 14, spent = false, testId = 'energy-icon' }: { type: string; size?: number; spent?: boolean; testId?: string }) {
  const style = { '--s': size, background: ENERGY_COLOR[type] ?? '#b9bcc8' } as CSSProperties;
  return (
    <span className="bt-energy" style={style} data-testid={testId} data-spent={spent} title={spent ? 'used' : type} aria-label={spent ? 'energy used' : `${type} energy`}>
      {ENERGY_LETTER[type] ?? ''}
    </span>
  );
}

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
