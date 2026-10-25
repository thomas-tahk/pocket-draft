import type { Card } from '../types/card';
import { CardTile, type HoverPayload } from './CardTile';

type Props = {
  offer: Card[];
  onPick: (card: Card) => void;
  onHover: (payload: HoverPayload | null) => void;
};

export function OfferRow({ offer, onPick, onHover }: Props) {
  return (
    <div className="offer-grid">
      {offer.map((c) => (
        <CardTile key={c.id} card={c} onPick={() => onPick(c)} onHover={onHover} />
      ))}
    </div>
  );
}
