import { useEffect } from 'react';
import { useGameStore } from '../game/store';
import { Board } from './Board';
import './battle.css';

// The new battle screen (plan 1.19): renders the server's current game on the
// Pocket-style board. Layout and rendering only; input comes in later issues.
export function BattleScreen() {
  const { view, error, init } = useGameStore();

  useEffect(() => {
    void init();
  }, [init]);

  return (
    <div className="bt-root">
      {view ? <Board view={view} /> : <div style={{ color: '#fff' }}>{error ? `Couldn't reach the game server: ${error}` : 'Loading game…'}</div>}
    </div>
  );
}
