import { useEffect, useMemo, useState } from 'react';
import { loadCardPool, type CardPool } from './data';
import { DraftView } from './components/DraftView';
import { ShopView } from './components/ShopView';
import { DeckbuildView } from './components/DeckbuildView';
import { RunView } from './components/RunView';
import { SharedDeckView } from './components/SharedDeckView';
import { HomeView } from './components/HomeView';
import { HistoryView } from './components/HistoryView';
import { BattleScreen } from './battle/BattleScreen';
import { ReplayView } from './battle/ReplayView';
import { expandDeck } from './game/deck';
import { autoDeck } from './game/api';
import { derivePhase, useDraftStore } from './stores/draftStore';
import { runScore } from './stores/run';
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

// ?replay=<i> opens run.battles[i] as a replay (handy for sharing and tests). An index
// with no battle behind it is dropped (see the effect in App).
function readReplayIndex(): number | null {
  const raw = new URLSearchParams(window.location.search).get('replay');
  const n = Number(raw);
  return raw !== null && Number.isInteger(n) && n >= 0 ? n : null;
}

export function App() {
  const [pool, setPool] = useState<CardPool | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [homeOpen, setHomeOpen] = useState(true);
  const [playing, setPlaying] = useState(false);
  // Index into run.battles of the battle being replayed: the run screen's Replay button
  // sets it, and ?replay=<i> opens it directly.
  const [replaying, setReplaying] = useState<number | null>(readReplayIndex);
  const [practicing, setPracticing] = useState(false);
  const [sharedDeck, setSharedDeck] = useState<Decklist | null>(readSharedDeck);

  const closeReplay = () => {
    window.history.replaceState({}, '', window.location.pathname);
    setReplaying(null);
  };

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
    quitRun,
  } = useDraftStore();

  // A ?replay index with no battle behind it goes to the run screen, param dropped.
  const replayMissing = replaying !== null && !run?.battles[replaying];
  useEffect(() => {
    if (!replayMissing) return;
    closeReplay();
    setHomeOpen(false);
  }, [replayMissing]);

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
  if (playing) {
    return (
      <BattleScreen
        deck={drafted}
        practice={practicing}
        onExit={() => {
          setPlaying(false);
          setPracticing(false);
        }}
      />
    );
  }

  const replayed = replaying !== null ? run?.battles[replaying] : undefined;
  if (replayed) return <ReplayView battle={replayed} onExit={closeReplay} />;

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
        run={run}
        pool={pool}
        onClose={() => setHistoryOpen(false)}
        onUpdate={updateHistoryEntry}
        onDelete={deleteHistoryEntry}
      />
    );
  }

  if (phase === 'unstarted' || (homeOpen && phase === 'review')) {
    return (
      <HomeView
        cardCount={pool.draftableCards.length}
        hasActiveRun={run?.status === 'active'}
        hasFinishedRun={run !== null && run.status !== 'active'}
        runScore={run ? runScore(run) : null}
        canPractice={deckFinalized}
        historyCount={history.length}
        onDraft={() => {
          setHomeOpen(false);
          start(pool.draftableCards);
        }}
        onContinue={() => setHomeOpen(false)}
        onPractice={() => {
          setHomeOpen(false);
          setPracticing(true);
          setPlaying(true);
        }}
        onPastRuns={() => setHistoryOpen(true)}
        onAutoDraft={() => {
          setHomeOpen(false);
          autoDeck()
            .then((d) => autoDraft(d.picks, d.deck))
            .catch((e) => setError(String(e)));
        }}
      />
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
    <RunView
      pool={pool}
      deck={deck}
      run={run}
      onNextBattle={() => setPlaying(true)}
      onReplay={setReplaying}
      onQuit={quitRun}
      onNewDraft={() => start(pool.draftableCards)}
      onHome={reset}
    />
  );
}
