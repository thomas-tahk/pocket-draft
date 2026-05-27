import { useMemo } from 'react';
import type { Card, EnergyType } from '../types/card';
import type { CardPool } from '../data';
import { tallyTypes } from '../draft/weights';
import { CardTile } from './CardTile';

type Props = {
  pool: CardPool;
  deck: Record<string, number>;
  onReset: () => void;
};

function download(filename: string, content: string) {
  const blob = new Blob([content], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

export function ReviewView({ pool, deck, onReset }: Props) {
  const idToCard = useMemo(() => {
    const m = new Map<string, Card>();
    for (const c of pool.fullCards) m.set(c.id, c);
    for (const c of pool.universals) m.set(c.id, c);
    return m;
  }, [pool]);

  const deckCards: { card: Card; count: number }[] = useMemo(() => {
    const out: { card: Card; count: number }[] = [];
    for (const [cardId, count] of Object.entries(deck)) {
      const c = idToCard.get(cardId);
      if (c) out.push({ card: c, count });
    }
    return out.sort((a, b) => {
      // Pokemon first, then Trainers, then by name.
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

  const exportPayload = {
    version: 'v0.3',
    timestamp: new Date().toISOString(),
    deck: deckCards.map(({ card, count }) => ({ id: card.id, name: card.name, count })),
  };

  return (
    <main style={{ padding: '16px 24px', maxWidth: 1100, margin: '0 auto' }}>
      <header
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          marginBottom: 8,
        }}
      >
        <h1 style={{ fontSize: 20, margin: 0 }}>Deck finalized</h1>
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            onClick={() =>
              download(`pocket-draft-${Date.now()}.json`, JSON.stringify(exportPayload, null, 2))
            }
          >
            Export JSON
          </button>
          <button onClick={onReset}>New draft</button>
        </div>
      </header>

      <div style={{ fontSize: 13, opacity: 0.75, marginBottom: 16 }}>
        {total} cards · {sortedTally.map(([t, n]) => `${t} ×${n}`).join('  ')}
      </div>

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(5, 1fr)',
          gap: 8,
          alignItems: 'start',
        }}
      >
        {deckCards.flatMap(({ card, count }) =>
          Array.from({ length: count }).map((_, i) => (
            <CardTile key={`${card.id}-${i}`} card={card} size="lg" />
          )),
        )}
      </div>
    </main>
  );
}
