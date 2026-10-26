import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { announce } from './announce';
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
import { RESULT_LABEL, STATUS_LABEL, leaveMsLeft, runLine, runScore, type InProgressBattle } from './stores/run';
import { useCloseStamp, useLeaveExpiry } from './battle/useLeaveExpiry';
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

// A battle still being played when the page loaded (a reload) goes straight back to it,
// unless it was closed for 90 s or more: that one is a loss, recorded on load.
function readInterrupted(): InProgressBattle | null {
  const saved = useDraftStore.getState().run?.inProgress;
  if (!saved || saved.leftAt) return null;
  return (leaveMsLeft(saved, Date.now()) ?? 1) > 0 ? saved : null;
}

// A screen change remounts this wrapper, which fades in (.screen-fade; none with reduced motion).
// On a new screen, focus moves to its heading (tabindex -1, no ring) so keyboard and screen-reader
// users start at the top instead of on a control that no longer exists. The battle board keeps its own.
function ScreenRoot({ name, children }: { name: string; children: ReactNode }) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (name === 'battle' || name === 'replay') return;
    const h = ref.current?.querySelector<HTMLElement>('h1');
    if (!h) return;
    h.tabIndex = -1;
    h.focus({ preventScroll: true });
  }, [name]);
  return <div ref={ref} className="screen-fade" data-screen={name}>{children}</div>;
}
const screen = (name: string, el: ReactNode) => <ScreenRoot key={name} name={name}>{el}</ScreenRoot>;

export function App() {
  const [pool, setPool] = useState<CardPool | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [historyOpen, setHistoryOpen] = useState(false);
  // A run with a battle in progress opens on that run, not Home.
  const [homeOpen, setHomeOpen] = useState(() => !useDraftStore.getState().run?.inProgress);
  const [resuming, setResuming] = useState<InProgressBattle | null>(readInterrupted);
  const [playing, setPlaying] = useState(() => resuming !== null);
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
    energy,
    history,
    run,
    start,
    pick,
    purchase,
    unpurchase,
    finalizeShop,
    addToDeck,
    removeFromDeck,
    setEnergy,
    finalizeDeck,
    autoDraft,
    updateHistoryEntry,
    deleteHistoryEntry,
    reset,
    quitRun,
    setInProgress,
  } = useDraftStore();
  // Past 90 s the battle on screen is lost: back to the run screen.
  const onExpired = useCallback(() => {
    setPlaying(false);
    setResuming(null);
  }, []);
  useLeaveExpiry(onExpired);
  useCloseStamp();

  // A battle result and a run ending are announced; the screen they appear on has no navigation of its own.
  const battleCount = run?.battles.length ?? 0;
  const lastBattle = run?.battles[battleCount - 1];
  const runStatus = run?.status;
  const seenBattles = useRef(battleCount);
  useEffect(() => {
    if (battleCount > seenBattles.current && run && lastBattle) {
      const ended = run.status !== 'active' ? `. ${STATUS_LABEL[run.status]}` : '';
      announce(`Battle ${battleCount}: ${RESULT_LABEL[lastBattle.result]}. ${runLine(run)}${ended}`);
    }
    seenBattles.current = battleCount;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [battleCount, runStatus]);

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
  if (new URLSearchParams(window.location.search).has('battle')) return screen('battle', <BattleScreen />);

  // In-app deck play (1.31), launched from the review screen with the finalized deck.
  if (playing) {
    return screen('battle', (
      <BattleScreen
        deck={drafted}
        energy={run?.energy}
        practice={practicing}
        resume={resuming ?? undefined}
        onExit={() => {
          setPlaying(false);
          setPracticing(false);
          setResuming(null);
        }}
      />
    ));
  }

  const replayed = replaying !== null ? run?.battles[replaying] : undefined;
  if (replayed) return screen('replay', <ReplayView battle={replayed} onExit={closeReplay} />);

  if (error) return screen('error', <main role="alert" style={{ padding: 24 }}>Error: {error}</main>);
  if (!pool) return <main role="status" style={{ padding: 24 }}>Loading card data…</main>;

  if (sharedDeck) {
    return screen('shared', (
      <SharedDeckView
        pool={pool}
        deck={sharedDeck}
        onExit={() => {
          window.history.replaceState({}, '', window.location.pathname);
          setSharedDeck(null);
        }}
      />
    ));
  }

  if (historyOpen) {
    return screen('history', (
      <HistoryView
        history={history}
        run={run}
        pool={pool}
        onClose={() => setHistoryOpen(false)}
        onUpdate={updateHistoryEntry}
        onDelete={deleteHistoryEntry}
      />
    ));
  }

  if (phase === 'unstarted' || (homeOpen && phase === 'review')) {
    return screen('home', (
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
    ));
  }

  if (phase === 'drafting') {
    return screen('draft', (
      <DraftView
        packIndex={pickIds.length}
        offer={offer}
        picks={picks}
        onPick={(card) => pick(card.id, pool.draftableCards)}
        onCancel={reset}
      />
    ));
  }

  if (phase === 'shop') {
    return screen('shop', (
      <ShopView
        picks={picks}
        fullPool={shopPool}
        purchasedIds={shopPurchasedIds}
        onPurchase={purchase}
        onUnpurchase={unpurchase}
        onFinalize={finalizeShop}
        onCancel={reset}
      />
    ));
  }

  if (phase === 'deckbuild') {
    return screen('deckbuild', (
      <DeckbuildView
        picks={picks}
        purchases={purchases}
        fullPool={pool.fullCards}
        working={pool.working}
        deck={deck}
        energy={energy ?? []}
        onEnergy={setEnergy}
        onAdd={addToDeck}
        onRemove={removeFromDeck}
        onFinalize={finalizeDeck}
        onCancel={reset}
      />
    ));
  }

  return screen(run ? (run.status === 'active' ? 'run' : 'summary') : 'review', (
    <RunView
      pool={pool}
      deck={deck}
      run={run}
      onNextBattle={() => setPlaying(true)}
      onReturn={() => {
        if (!run?.inProgress) return;
        setInProgress({ ...run.inProgress, leftAt: null });
        setResuming(run.inProgress);
        setPlaying(true);
      }}
      onReplay={setReplaying}
      onQuit={quitRun}
      onNewDraft={() => start(pool.draftableCards)}
      onHome={reset}
    />
  ));
}
