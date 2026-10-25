import { useEffect, useState } from 'react';
import { useGameStore } from '../game/store';
import { useDraftStore } from '../stores/draftStore';
import type { InProgressBattle } from '../stores/run';
import { Board } from './Board';
import { ConfirmDialog, EndScreen, GameMenu, endTurnWarning, type Confirm } from './EndControls';
import { choiceFor } from './Choice';
import { SkipButton, flashAttrs, useOpponentTurn } from './Playback';
import { SetupIntro, useSetupIntro } from './Setup';
import './battle.css';

// The new battle screen (plan 1.19): renders the server's current game on the
// Pocket-style board, with End Turn, Concede and the end screen (1.27). Given a
// drafted deck (1.31) it starts a game with it; Rematch reuses that deck.
type Props = { deck?: string[]; onExit?: () => void; practice?: boolean; resume?: InProgressBattle };

export function BattleScreen({ deck, onExit, practice, resume }: Props) {
  const { view, status, error, log, init, startGame, resumeGame, idle, dispatch, runBot, botStuck, restart } = useGameStore();
  const setInProgress = useDraftStore((s) => s.setInProgress);
  const [confirm, setConfirm] = useState<Confirm | null>(null);

  useEffect(() => {
    if (resume) void resumeGame(resume);
    else if (deck) void startGame(deck);
    else void init();
  }, [deck, resume, init, startGame, resumeGame]);

  // A run's battle is saved as it is played, so a reload or a closed tab can pick it up again.
  useEffect(() => {
    const g = useGameStore.getState();
    if (!deck || practice || g.status !== 'ready' || g.view?.phase === 'over' || g.seed === null) return;
    // A bot step that lands while the page is hidden must not wipe the closed-tab stamp.
    const hidden = document.visibilityState === 'hidden';
    const stamp = useDraftStore.getState().run?.inProgress?.closedAt ?? new Date().toISOString();
    setInProgress({ seed: g.seed, yourDeck: deck, log: g.log, leftAt: null, closedAt: hidden ? stamp : null });
  }, [deck, practice, status, log, setInProgress]);

  // A run's game ends once: save its result and replayable record. The store
  // ignores a seed it already holds, so a re-render cannot record it twice.
  const recordBattle = useDraftStore((s) => s.recordBattle);
  const runStatus = useDraftStore((s) => s.run?.status);
  // On mount this runs after startGame's effect has cleared the store, but with
  // the render's stale values, so read the store's current game here.
  useEffect(() => {
    const { view: current, status: now } = useGameStore.getState();
    if (!deck || practice || now !== 'ready' || current?.phase !== 'over' || !current.outcome) return;
    const { seed, you, bot, actions } = current.record;
    recordBattle({ result: current.outcome, seed, yourDeck: you, botDeck: bot, actions, endedAt: new Date().toISOString() });
  }, [deck, practice, status, view, recordBattle]);

  // The bot (seat 1) plays whenever it is its move, one step at a time (F14).
  const { step, canSkip, skip, skipping } = useOpponentTurn(view, botStuck, runBot);
  const intro = useSetupIntro(view);

  if (!view && resume && error) {
    return <ResumeFailed onRetry={() => void resumeGame(resume)} onStartOver={() => { setInProgress(undefined); void startGame(deck ?? resume.yourDeck); }} />;
  }
  if (!view) {
    return <div className="bt-root"><div style={{ color: '#fff' }}>{error ? `Couldn't reach the game server: ${error}` : 'Loading game…'}</div></div>;
  }

  // Leaving a run's battle keeps it for 90 s on the run screen; a practice game just ends.
  const onLeave = async () => {
    await idle(); // a move still on its way to the server must be in the saved log
    const g = useGameStore.getState();
    if (!practice && deck && g.seed !== null) setInProgress({ seed: g.seed, yourDeck: deck, log: g.log, leftAt: new Date().toISOString(), closedAt: null });
    onExit?.();
  };

  const endTurn = () => void dispatch({ type: 'EndTurn', player: 0 });
  const onEndTurn = () => {
    const warning = endTurnWarning(view);
    if (warning) setConfirm({ text: warning, okLabel: 'End Turn', onOk: endTurn });
    else endTurn();
  };
  const onReady = () => void dispatch({ type: 'SetupReady', player: 0 });
  const onConcede = () =>
    setConfirm({ text: 'Concede this game? It counts as a loss.', okLabel: 'Concede', onOk: () => void dispatch({ type: 'Concede', player: 0 }) });

  return (
    // A long press zooms a card; never the browser's image menu (iOS callout in battle.css).
    <div className="bt-root" {...flashAttrs(step)} onContextMenu={(e) => e.preventDefault()}>
      <Board view={view} instant={skipping} onEndTurn={onEndTurn} onReady={onReady} choice={choiceFor(view, 0, (m) => void dispatch(m))}>
        {view.phase !== 'over' && <GameMenu onConcede={onConcede} onLeave={onExit && (() => void onLeave())} />}
        {canSkip && <SkipButton onSkip={skip} />}
        {intro.show && <SetupIntro view={view} onDone={intro.finish} />}
        {confirm && <ConfirmDialog confirm={confirm} onClose={() => setConfirm(null)} />}
        {view.phase === 'over' && <EndScreen view={view} canRematch={!deck || !runStatus || runStatus === 'active'} onRematch={() => void restart()} onBack={onExit ?? (() => window.location.assign('/'))} />}
      </Board>
    </div>
  );
}

function ResumeFailed({ onRetry, onStartOver }: { onRetry: () => void; onStartOver: () => void }) {
  return (
    <div className="bt-root">
      <div className="bt-dialog" role="alert" style={{ margin: 'auto' }}>
        <p>This battle could not be resumed.</p>
        <button className="bt-btn bt-btn-main" onClick={onRetry}>Try again</button>
        <button className="bt-btn" onClick={onStartOver}>Start a new battle</button>
      </div>
    </div>
  );
}
