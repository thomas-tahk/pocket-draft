import { useMemo, useState } from 'react';
import type { Card, EnergyType } from '../types/card';
import { DECK_SIZE } from '../stores/draftStore';
import {
  buildCollection,
  COPY_CAP_BY_NAME,
  deckCountByName,
  deckSize,
  type CollectionEntry,
  type CollectionSource,
} from '../draft/expansions';
import { tallyTypes } from '../draft/weights';
import { CardTile } from './CardTile';
import { CardPreview } from './CardPreview';

type Props = {
  picks: Card[];
  fullPool: Card[];
  universals: Card[];
  deck: Record<string, number>;
  onAdd: (cardId: string) => void;
  onRemove: (cardId: string) => void;
  onFinalize: () => void;
  onCancel: () => void;
};

const SOURCE_LABEL: Record<CollectionSource, string> = {
  drafted: 'Your draft',
  'under-evo': 'Free expansions (under-evolutions)',
  universal: 'Free universals',
};

const SOURCE_ORDER: CollectionSource[] = ['drafted', 'under-evo', 'universal'];

export function DeckbuildView({
  picks,
  fullPool,
  universals,
  deck,
  onAdd,
  onRemove,
  onFinalize,
  onCancel,
}: Props) {
  const [hovered, setHovered] = useState<Card | null>(null);

  const collection = useMemo(
    () => buildCollection(picks, fullPool, universals),
    [picks, fullPool, universals],
  );

  const idToCard = useMemo(() => {
    const m = new Map<string, Card>();
    for (const c of fullPool) m.set(c.id, c);
    for (const c of universals) m.set(c.id, c);
    return m;
  }, [fullPool, universals]);

  const grouped = useMemo(() => {
    const g: Record<CollectionSource, CollectionEntry[]> = {
      drafted: [],
      'under-evo': [],
      universal: [],
    };
    for (const e of collection) g[e.source].push(e);
    return g;
  }, [collection]);

  const size = deckSize(deck);
  const remaining = DECK_SIZE - size;

  const deckCards: { card: Card; count: number }[] = useMemo(() => {
    const out: { card: Card; count: number }[] = [];
    for (const [cardId, count] of Object.entries(deck)) {
      const c = idToCard.get(cardId);
      if (c) out.push({ card: c, count });
    }
    return out;
  }, [deck, idToCard]);

  const tally = useMemo(() => {
    const flat: Card[] = [];
    for (const { card, count } of deckCards) {
      for (let i = 0; i < count; i++) flat.push(card);
    }
    return tallyTypes(flat);
  }, [deckCards]);
  const sortedTally = (Object.entries(tally) as [EnergyType, number][]).sort(
    (a, b) => b[1] - a[1],
  );

  const canAdd = (entryCard: Card) => {
    if (size >= DECK_SIZE) return false;
    return deckCountByName(deck, idToCard, entryCard.name) < COPY_CAP_BY_NAME;
  };

  const renderEntry = (entry: CollectionEntry) => {
    const inDeck = deck[entry.card.id] ?? 0;
    const nameTotal = deckCountByName(deck, idToCard, entry.card.name);
    const addDisabled = !canAdd(entry.card);

    return (
      <div
        key={`${entry.source}-${entry.card.id}`}
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 4,
          padding: 6,
          borderRadius: 8,
          background: inDeck > 0 ? '#4c81' : 'transparent',
          outline: inDeck > 0 ? '2px solid #4c8' : 'none',
        }}
      >
        <CardTile card={entry.card} size="lg" onHover={setHovered} />
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 12 }}>
          <button
            onClick={() => onRemove(entry.card.id)}
            disabled={inDeck === 0}
            style={{ minWidth: 28 }}
          >
            −
          </button>
          <span style={{ opacity: 0.8 }}>
            {nameTotal} / {COPY_CAP_BY_NAME}
          </span>
          <button
            onClick={() => onAdd(entry.card.id)}
            disabled={addDisabled}
            style={{ minWidth: 28 }}
          >
            +
          </button>
        </div>
      </div>
    );
  };

  const renderSection = (source: CollectionSource) => {
    const entries = grouped[source];
    if (entries.length === 0) return null;
    return (
      <section key={source} style={{ marginTop: 16 }}>
        <h2 style={{ fontSize: 14, opacity: 0.75, margin: '0 0 8px' }}>
          {SOURCE_LABEL[source]} ({entries.length})
        </h2>
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(170px, 1fr))',
            gap: 8,
            alignItems: 'start',
          }}
        >
          {entries.map(renderEntry)}
        </div>
      </section>
    );
  };

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <main style={{ flex: 1, padding: '16px 24px', maxWidth: 1300 }}>
        <header
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'baseline',
            marginBottom: 8,
            gap: 12,
            flexWrap: 'wrap',
          }}
        >
          <h1 style={{ fontSize: 20, margin: 0 }}>Deckbuilding</h1>
          <div style={{ display: 'flex', gap: 12, alignItems: 'baseline' }}>
            <div style={{ fontSize: 14, opacity: 0.75 }}>
              Deck: {size} / {DECK_SIZE}
              {sortedTally.length > 0 && (
                <span style={{ marginLeft: 16 }}>
                  {sortedTally.map(([t, n]) => `${t} ×${n}`).join('  ')}
                </span>
              )}
            </div>
            <button
              onClick={onFinalize}
              disabled={size !== DECK_SIZE}
              title={size !== DECK_SIZE ? `Add ${remaining} more cards to finalize` : ''}
            >
              Finalize deck
            </button>
            <button
              onClick={() => {
                if (window.confirm('Cancel this draft? You will lose all picks and deck choices.')) {
                  onCancel();
                }
              }}
              style={{ fontSize: 12 }}
            >
              Cancel draft
            </button>
          </div>
        </header>
        <p style={{ fontSize: 13, opacity: 0.6, marginTop: 0 }}>
          Build a deck of exactly {DECK_SIZE} cards. Max 2 of any card by name. Drafted Pokémon
          unlock their under-evolutions for free; the 5 free universals are always available.
        </p>

        {SOURCE_ORDER.map(renderSection)}
      </main>

      <aside
        style={{
          width: 280,
          borderLeft: '1px solid #8884',
          padding: 16,
          height: '100vh',
          overflowY: 'auto',
          position: 'sticky',
          top: 0,
          boxSizing: 'border-box',
        }}
      >
        <div style={{ fontSize: 12, opacity: 0.6, marginBottom: 8 }}>
          Deck ({size} / {DECK_SIZE})
        </div>
        {deckCards.length === 0 ? (
          <p style={{ fontSize: 12, opacity: 0.5 }}>No cards yet.</p>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 4 }}>
            {deckCards.map(({ card, count }) =>
              Array.from({ length: count }).map((_, i) => (
                <CardTile
                  key={`${card.id}-${i}`}
                  card={card}
                  size="sm"
                  onPick={() => onRemove(card.id)}
                  onHover={setHovered}
                />
              )),
            )}
          </div>
        )}
      </aside>

      <CardPreview card={hovered} />
    </div>
  );
}
