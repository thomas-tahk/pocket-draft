import { useEffect, useMemo, useState } from 'react';
import { loadCardPool, type CardPool } from './data';
import { DraftView } from './components/DraftView';
import { ShopView } from './components/ShopView';
import { DeckbuildView } from './components/DeckbuildView';
import { ReviewView } from './components/ReviewView';
import { SharedDeckView } from './components/SharedDeckView';
import { HistoryView } from './components/HistoryView';
import { BattleScreen } from './battle/BattleScreen';
import { expandDeck } from './game/deck';
import { autoDeck } from './game/api';
import { derivePhase, HISTORY_CAP, TOTAL_PACKS, useDraftStore } from './stores/draftStore';
import { decodeDeck, type Decklist } from './share/deckCode';
import type { Card } from './types/card';

// A shared deck is opened read-only via ?deck=<code>. Parsed once at load;
// invalid or absent codes fall through to the normal draft app. See ADR 0003.
function readSharedDeck(): Decklist | null {
  const code = new URLSearchParams(window.location.search).get('deck');
  if (!code) return null;
  try {
    return decodeDeck(code);
  } catch {
    return null;
  }
}

export function App() {
  const [pool, setPool] = useState<CardPool | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [playing, setPlaying] = useState(false);
  const [sharedDeck, setSharedDeck] = useState<Decklist | null>(readSharedDeck);

  useEffect(() => {
    loadCardPool().then(setPool).catch((e) => setError(String(e)));
  }, []);

  const {
    pickIds,
    offerIds,
    shopPurchasedIds,
    shopFinalized,
    deck,
    deckFinalized,
    history,
    run,
    start,
    pick,
    purchase,
    unpurchase,
    finalizeShop,
    addToDeck,
    removeFromDeck,
    finalizeDeck,
    autoDraft,
    updateHistoryEntry,
    deleteHistoryEntry,
    reset,
  } = useDraftStore();

  const draftableById = useMemo(() => {
    if (!pool) return new Map<string, Card>();
    return new Map(pool.draftableCards.map((c) => [c.id, c]));
  }, [pool]);

  const fullById = useMemo(() => {
    if (!pool) return new Map<string, Card>();
    const m = new Map<string, Card>();
    for (const c of pool.fullCards) m.set(c.id, c);
    return m;
  }, [pool]);

  const picks = useMemo(
    () => pickIds.map((id) => draftableById.get(id)).filter((c): c is Card => Boolean(c)),
    [pickIds, draftableById],
  );

  const offer = useMemo(
    () => offerIds.map((id) => draftableById.get(id)).filter((c): c is Card => Boolean(c)),
    [offerIds, draftableById],
  );

  const purchases = useMemo(
    () => shopPurchasedIds.map((id) => fullById.get(id)).filter((c): c is Card => Boolean(c)),
    [shopPurchasedIds, fullById],
  );

  // The shop offers only cards the engine can play (owner 2026-10-19 (1)).
  const shopPool = useMemo(
    () => (pool ? pool.fullCards.filter((c) => pool.working.has(c.id)) : []),
    [pool],
  );

  // Stable across renders, so the battle screen starts its game once.
  const drafted = useMemo(() => expandDeck(deck), [deck]);

  const phase = derivePhase({ pickIds, offerIds, shopFinalized, deckFinalized });

  // The new battle screen (plan 1.19), on the server's current game.
  if (new URLSearchParams(window.location.search).has('battle')) return <BattleScreen />;

  // In-app deck play (1.31), launched from the review screen with the finalized deck.
  if (playing) return <BattleScreen deck={drafted} onExit={() => setPlaying(false)} />;

  if (error) return <main style={{ padding: 24 }}>Error: {error}</main>;
  if (!pool) return <main style={{ padding: 24 }}>Loading card data…</main>;

  if (sharedDeck) {
    return (
      <SharedDeckView
        pool={pool}
        deck={sharedDeck}
        onExit={() => {
          window.history.replaceState({}, '', window.location.pathname);
          setSharedDeck(null);
        }}
      />
    );
  }

  if (historyOpen) {
    return (
      <HistoryView
        history={history}
        pool={pool}
        onClose={() => setHistoryOpen(false)}
        onUpdate={updateHistoryEntry}
        onDelete={deleteHistoryEntry}
      />
    );
  }

  if (phase === 'unstarted') {
    return (
      <main style={{ padding: 24, maxWidth: 720, margin: '0 auto', textAlign: 'center' }}>
        <h1>Pocket Draft</h1>
        <p style={{ opacity: 0.75 }}>
          {pool.draftableCards.length} draftable cards (final-evolution Pokémon and Trainers).
        </p>
        <p style={{ opacity: 0.6, fontSize: 13 }}>
          {TOTAL_PACKS} rounds → shop → choose your 20-card deck.
        </p>
        <button style={{ fontSize: 16, padding: '8px 20px' }} onClick={() => start(pool.draftableCards)}>
          Start draft
        </button>{' '}
        <button
          style={{ fontSize: 16, padding: '8px 20px' }}
          onClick={() =>
            autoDeck()
              .then((d) => autoDraft(d.picks, d.deck))
              .catch((e) => setError(String(e)))
          }
        >
          Auto-draft (test)
        </button>
        {history.length > 0 && (
          <div style={{ marginTop: 24 }}>
            <button onClick={() => setHistoryOpen(true)} style={{ fontSize: 13 }}>
              View past drafts ({history.length} / {HISTORY_CAP})
            </button>
          </div>
        )}
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

  if (phase === 'shop') {
    return (
      <ShopView
        picks={picks}
        fullPool={shopPool}
        purchasedIds={shopPurchasedIds}
        onPurchase={purchase}
        onUnpurchase={unpurchase}
        onFinalize={finalizeShop}
        onCancel={reset}
      />
    );
  }

  if (phase === 'deckbuild') {
    return (
      <DeckbuildView
        picks={picks}
        purchases={purchases}
        fullPool={pool.fullCards}
        working={pool.working}
        deck={deck}
        onAdd={addToDeck}
        onRemove={removeFromDeck}
        onFinalize={finalizeDeck}
        onCancel={reset}
      />
    );
  }

  return (
    <ReviewView
      pool={pool}
      deck={deck}
      run={run}
      onPlay={() => setPlaying(true)}
      onNewDraft={() => start(pool.draftableCards)}
      onHome={reset}
    />
  );
}
