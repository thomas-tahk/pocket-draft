import { useEffect, useMemo, useState } from 'react';
import { loadCardPool, type CardPool } from './data';
import { DraftView } from './components/DraftView';
import { DeckbuildView } from './components/DeckbuildView';
import { ReviewView } from './components/ReviewView';
import { derivePhase, TOTAL_PACKS, useDraftStore } from './stores/draftStore';
import type { Card } from './types/card';

export function App() {
  const [pool, setPool] = useState<CardPool | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadCardPool().then(setPool).catch((e) => setError(String(e)));
  }, []);

  const {
    pickIds,
    offerIds,
    deck,
    deckFinalized,
    start,
    pick,
    addToDeck,
    removeFromDeck,
    finalizeDeck,
    reset,
  } = useDraftStore();

  const draftableById = useMemo(() => {
    if (!pool) return new Map<string, Card>();
    return new Map(pool.draftableCards.map((c) => [c.id, c]));
  }, [pool]);

  // Picks may reference cards that come from the draftable pool only.
  const picks = useMemo(
    () => pickIds.map((id) => draftableById.get(id)).filter((c): c is Card => Boolean(c)),
    [pickIds, draftableById],
  );

  const offer = useMemo(
    () => offerIds.map((id) => draftableById.get(id)).filter((c): c is Card => Boolean(c)),
    [offerIds, draftableById],
  );

  const phase = derivePhase({ pickIds, offerIds, deckFinalized });

  if (error) return <main style={{ padding: 24 }}>Error: {error}</main>;
  if (!pool) return <main style={{ padding: 24 }}>Loading card data…</main>;

  if (phase === 'unstarted') {
    return (
      <main style={{ padding: 24, maxWidth: 720, margin: '0 auto', textAlign: 'center' }}>
        <h1>Pocket Draft</h1>
        <p style={{ opacity: 0.75 }}>
          {pool.draftableCards.length} draftable cards (final-evolution Pokémon and Trainers).
        </p>
        <p style={{ opacity: 0.6, fontSize: 13 }}>
          {TOTAL_PACKS} rounds. Pick one card from each pack to build your collection — then choose
          your 20-card deck.
        </p>
        <button style={{ fontSize: 16, padding: '8px 20px' }} onClick={() => start(pool.draftableCards)}>
          Start draft
        </button>
      </main>
    );
  }

  if (phase === 'drafting') {
    return (
      <DraftView
        packIndex={pickIds.length}
        offer={offer}
        picks={picks}
        onPick={(card) => pick(card.id, pool.draftableCards)}
        onCancel={reset}
      />
    );
  }

  if (phase === 'deckbuild') {
    return (
      <DeckbuildView
        picks={picks}
        fullPool={pool.fullCards}
        universals={pool.universals}
        deck={deck}
        onAdd={addToDeck}
        onRemove={removeFromDeck}
        onFinalize={finalizeDeck}
        onCancel={reset}
      />
    );
  }

  return <ReviewView pool={pool} deck={deck} onReset={reset} />;
}
