import { useMemo, useState, type ReactNode } from 'react';
import { TypeTally } from '../battle/energy';
import type { Card, EnergyType } from '../types/card';
import type { CardPool } from '../data';
import { tallyTypes } from '../draft/weights';
import { runLine, type Run } from '../stores/run';
import { DeckShare } from './DeckShare';
import { CardTile, type HoverPayload } from './CardTile';
import { CardPreview, type HoverState } from './CardPreview';

type Props = {
  pool: CardPool;
  deck: Record<string, number>;
  run: Run | null;
  onPlay?: () => void;
  children?: ReactNode; // the run panel, shown above the deck
  onNewDraft: () => void;
  onHome: () => void;
};

export function ReviewView({ pool, deck, run, onPlay, children, onNewDraft, onHome }: Props) {
  const [hover, setHover] = useState<HoverState>(null);
  const handleHover = (p: HoverPayload | null) => setHover(p);

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
  const sortedTally = (Object.entries(tally) as [EnergyType, number][])
    .sort((a, b) => b[1] - a[1]);
  const total = deckCards.reduce((s, { count }) => s + count, 0);

  return (
    <main className="page">
      <header className="screen-header">
        <h1 className="screen-title">Deck finalized</h1>
        {run && (
          <span data-testid="run-record" className="record-chip">
            {runLine(run)} ({run.status})
          </span>
        )}
        <div className="review-actions">
          {onPlay && <button className="btn-primary" onClick={onPlay}>▶ Play deck</button>}
          <DeckShare deck={deck} pool={pool} />
          <button onClick={onNewDraft}>Start new draft</button>
          <button onClick={onHome}>
            Home
          </button>
        </div>
      </header>

      {children}

      <div className="muted-sm" style={{ marginBottom: 16 }}>
        {total} cards · <TypeTally entries={sortedTally} />
      </div>

      <h2 className="sr-only">Your deck</h2>
      <div className="deck-grid" role="group" aria-label="Your deck">
        {deckCards.flatMap(({ card, count }) =>
          Array.from({ length: count }).map((_, i) => (
            <CardTile key={`${card.id}-${i}`} card={card} size="lg" onHover={handleHover} />
          )),
        )}
      </div>

      <CardPreview hover={hover} />
    </main>
  );
}
