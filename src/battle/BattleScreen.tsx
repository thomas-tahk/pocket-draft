import { useEffect, useState } from 'react';
import { useGameStore } from '../game/store';
import { Board } from './Board';
import { ConfirmDialog, EndScreen, GameMenu, endTurnWarning, type Confirm } from './EndControls';
import { choiceFor } from './Choice';
import { SkipButton, flashAttrs, useOpponentTurn } from './Playback';
import { SetupIntro, useSetupIntro } from './Setup';
import './battle.css';

// The new battle screen (plan 1.19): renders the server's current game on the
// Pocket-style board, with End Turn, Concede and the end screen (1.27). Given a
// drafted deck (1.31) it starts a game with it; Rematch reuses that deck.
export function BattleScreen({ deck, onExit }: { deck?: string[]; onExit?: () => void }) {
  const { view, error, init, startGame, dispatch, runBot, botStuck, restart } = useGameStore();
  const [confirm, setConfirm] = useState<Confirm | null>(null);

  useEffect(() => {
    if (deck) void startGame(deck);
    else void init();
  }, [deck, init, startGame]);

  // The bot (seat 1) plays whenever it is its move, one step at a time (F14).
  const { step, canSkip, skip, skipping } = useOpponentTurn(view, botStuck, runBot);
  const intro = useSetupIntro(view);

  if (!view) {
    return <div className="bt-root"><div style={{ color: '#fff' }}>{error ? `Couldn't reach the game server: ${error}` : 'Loading game…'}</div></div>;
  }

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
    <div className="bt-root" {...flashAttrs(step)}>
      <Board view={view} instant={skipping} onEndTurn={onEndTurn} onReady={onReady} choice={choiceFor(view, 0, (m) => void dispatch(m))}>
        {view.phase !== 'over' && <GameMenu onConcede={onConcede} />}
        {canSkip && <SkipButton onSkip={skip} />}
        {intro.show && <SetupIntro view={view} onDone={intro.finish} />}
        {confirm && <ConfirmDialog confirm={confirm} onClose={() => setConfirm(null)} />}
        {view.phase === 'over' && <EndScreen view={view} onRematch={() => void restart()} onBack={onExit ?? (() => window.location.assign('/'))} />}
      </Board>
    </div>
  );
}
