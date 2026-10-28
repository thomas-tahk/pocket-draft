import type { Card } from '../types/card';
import { TOTAL_PACKS } from '../stores/draftStore';
import { CardTile, type HoverPayload } from './CardTile';

type Props = {
  packIndex: number;
  offer: Card[];
  onPick: (card: Card) => void;
  onHover: (payload: HoverPayload | null) => void;
};

export function OfferRow({ packIndex, offer, onPick, onHover }: Props) {
  return (
    <div className="offer-grid" role="group" aria-label={`Pack ${packIndex + 1} of ${TOTAL_PACKS}: choose a card`}>
      {offer.map((c) => (
        <CardTile key={c.id} card={c} onPick={() => onPick(c)} onHover={onHover} />
      ))}
    </div>
  );
}
