import { useMemo, useState } from 'react';
import { TypeTally } from '../battle/energy';
import type { Card, EnergyType } from '../types/card';
import type { CardPool } from '../data';
import type { Decklist } from '../share/deckCode';
import { tallyTypes } from '../draft/weights';
import { CardTile } from './CardTile';
import { CardPreview, type HoverState } from './CardPreview';

type Props = {
  pool: CardPool;
  deck: Decklist;
  onExit: () => void;
};

// Read-only display of a shared deck (opened via ?deck=… or a pasted code).
// This is display only — it never writes into the local draft. See ADR 0003.
export function SharedDeckView({ pool, deck, onExit }: Props) {
  const [hover, setHover] = useState<HoverState>(null);

  const idToCard = useMemo(() => {
    const m = new Map<string, Card>();
    for (const c of pool.fullCards) m.set(c.id, c);
    return m;
  }, [pool]);

  const deckCards: { card: Card; count: number }[] = useMemo(() => {
    const out: { card: Card; count: number }[] = [];
    for (const [cardId, count] of Object.entries(deck)) {
      const c = idToCard.get(cardId);
      if (c) out.push({ card: c, count });
    }
    return out.sort((a, b) => {
      const aT = a.card.cardType === 'Trainer' ? 1 : 0;
      const bT = b.card.cardType === 'Trainer' ? 1 : 0;
      if (aT !== bT) return aT - bT;
      return a.card.name.localeCompare(b.card.name);
    });
  }, [deck, idToCard]);

  const flatForTally = useMemo(() => {
    const out: Card[] = [];
    for (const { card, count } of deckCards) {
      for (let i = 0; i < count; i++) out.push(card);
    }
    return out;
  }, [deckCards]);

  const tally = tallyTypes(flatForTally);
  const sortedTally = (Object.entries(tally) as [EnergyType, number][]).sort((a, b) => b[1] - a[1]);
  const total = deckCards.reduce((s, { count }) => s + count, 0);
  const unknownCount =
    Object.values(deck).reduce((s, n) => s + n, 0) - total;

  return (
    <main className="page">
      <header className="screen-header">
        <h1 className="screen-title">Shared deck</h1>
        <button onClick={onExit}>Go to Pocket Draft</button>
      </header>

      {total === 0 ? (
        <p className="muted">
          This deck code didn’t match any known cards. It may be from a newer card set than this
          build has, or the code may be incomplete.
        </p>
      ) : (
        <>
          <div className="muted-sm" style={{ marginBottom: 16 }}>
            {total} cards · <TypeTally entries={sortedTally} />
            {unknownCount > 0 && ` · ${unknownCount} unrecognized card(s) hidden`}
          </div>

          <div className="deck-grid">
            {deckCards.flatMap(({ card, count }) =>
              Array.from({ length: count }).map((_, i) => (
                <CardTile key={`${card.id}-${i}`} card={card} size="lg" focusable onHover={setHover} />
              )),
            )}
          </div>
        </>
      )}

      <CardPreview hover={hover} />
    </main>
  );
}
