import type { Card } from '../types/card';
import { CardTile } from './CardTile';

type Props = {
  offer: Card[];
  onPick: (card: Card) => void;
};

export function OfferRow({ offer, onPick }: Props) {
  return (
    <div style={{ display: 'flex', justifyContent: 'center', gap: 8, flexWrap: 'wrap' }}>
      {offer.map((c) => (
        <CardTile key={c.id} card={c} onPick={() => onPick(c)} />
      ))}
    </div>
  );
}
