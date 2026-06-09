import { useMemo, useState } from 'react';
import type { Card, EnergyType } from '../types/card';
import type { CardPool } from '../data';
import { tallyTypes } from '../draft/weights';
import { encodeDeck } from '../share/deckCode';
import { CardTile, type HoverPayload } from './CardTile';
import { CardPreview, type HoverState } from './CardPreview';

type Props = {
  pool: CardPool;
  deck: Record<string, number>;
  onNewDraft: () => void;
  onHome: () => void;
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

export function ReviewView({ pool, deck, onNewDraft, onHome }: Props) {
  const [hover, setHover] = useState<HoverState>(null);
  const [copied, setCopied] = useState<string | null>(null);
  const handleHover = (p: HoverPayload | null) => setHover(p);

  const deckCode = useMemo(() => encodeDeck(deck), [deck]);
  const shareLink = `${location.origin}${location.pathname}?deck=${encodeURIComponent(deckCode)}`;

  const copy = (text: string, label: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(label);
      setTimeout(() => setCopied((c) => (c === label ? null : c)), 1500);
    });
  };

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
    version: 'v0.4',
    timestamp: new Date().toISOString(),
    deck: deckCards.map(({ card, count }) => ({ id: card.id, name: card.name, count })),
  };

  return (
    <main style={{ padding: '0 24px 24px', maxWidth: 1100, margin: '0 auto' }}>
      <header
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          padding: '16px 0',
          marginBottom: 8,
          position: 'sticky',
          top: 0,
          background: 'var(--surface)',
          zIndex: 5,
          borderBottom: '1px solid #8882',
          gap: 12,
          flexWrap: 'wrap',
        }}
      >
        <h1 style={{ fontSize: 20, margin: 0 }}>Deck finalized</h1>
        <div style={{ display: 'flex', gap: 8 }}>
          <button onClick={() => copy(deckCode, 'code')}>
            {copied === 'code' ? 'Copied!' : 'Copy code'}
          </button>
          <button onClick={() => copy(shareLink, 'link')}>
            {copied === 'link' ? 'Copied!' : 'Copy link'}
          </button>
          <button
            onClick={() =>
              download(`pocket-draft-${Date.now()}.json`, JSON.stringify(exportPayload, null, 2))
            }
          >
            Export JSON
          </button>
          <button onClick={onNewDraft}>Start new draft</button>
          <button onClick={onHome} style={{ fontSize: 12 }}>
            Home
          </button>
        </div>
      </header>

      <div style={{ fontSize: 13, opacity: 0.75, marginBottom: 16 }}>
        {total} cards · {sortedTally.map(([t, n]) => `${t} ×${n}`).join('  ')}
      </div>

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, 180px)',
          justifyContent: 'start',
          gap: 12,
          alignItems: 'start',
        }}
      >
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
