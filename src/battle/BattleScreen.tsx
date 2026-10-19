import { useEffect, useState } from 'react';
import { useGameStore } from '../game/store';
import { Board } from './Board';
import { ConfirmDialog, EndScreen, GameMenu, endTurnWarning, type Confirm } from './EndControls';
import { choiceFor } from './Choice';
import './battle.css';

// The new battle screen (plan 1.19): renders the server's current game on the
// Pocket-style board, with End Turn, Concede and the end screen (1.27).
export function BattleScreen() {
  const { view, error, init, dispatch, runBot, botStuck, restart } = useGameStore();
  const [confirm, setConfirm] = useState<Confirm | null>(null);

  useEffect(() => {
    void init();
  }, [init]);

  // The bot (seat 1) plays whenever it is its move, one move per step, until
  // control comes back to you. Step-by-step playback (F14) is 1.26's.
  useEffect(() => {
    if (!view || view.phase === 'over' || botStuck) return;
    const actor = view.pending ? view.pending.player : view.active;
    if (actor !== 1) return;
    const t = setTimeout(() => void runBot(), 550);
    return () => clearTimeout(t);
  }, [view, botStuck, runBot]);

  if (!view) {
    return <div className="bt-root"><div style={{ color: '#fff' }}>{error ? `Couldn't reach the game server: ${error}` : 'Loading game…'}</div></div>;
  }

  const endTurn = () => void dispatch({ type: 'EndTurn', player: 0 });
  const onEndTurn = () => {
    const warning = endTurnWarning(view);
    if (warning) setConfirm({ text: warning, okLabel: 'End Turn', onOk: endTurn });
    else endTurn();
  };
  const onConcede = () =>
    setConfirm({ text: 'Concede this game? It counts as a loss.', okLabel: 'Concede', onOk: () => void dispatch({ type: 'Concede', player: 0 }) });

  return (
    <div className="bt-root">
      <Board view={view} onEndTurn={onEndTurn} choice={choiceFor(view, 0, (m) => void dispatch(m))}>
        {view.phase !== 'over' && <GameMenu onConcede={onConcede} />}
        {confirm && <ConfirmDialog confirm={confirm} onClose={() => setConfirm(null)} />}
        {view.phase === 'over' && <EndScreen view={view} onRematch={() => void restart()} onBack={() => window.location.assign('/')} />}
      </Board>
    </div>
  );
}
