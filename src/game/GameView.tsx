import { useEffect } from 'react';
import { useGameStore } from './store';
import { PlayerSide } from './PlayerSide';
import { Hand } from './Hand';
import { ActionBar } from './ActionBar';
import { Log } from './Log';
import { changedSlots } from './diff';

export function GameView({ deck, onExit }: { deck?: string[]; onExit?: () => void }) {
  const { view, prev, status, error, init, startGame } = useGameStore();

  useEffect(() => {
    if (deck) void startGame(deck);
    else void init();
  }, [deck, init, startGame]);

  if (status !== 'ready' || !view) {
    return (
      <main style={{ padding: 24 }}>
        {error ? (
          <div>
            <div style={{ color: '#ff6b6b' }}>Couldn't reach the game server: {error}</div>
            <button style={{ marginTop: 8 }} onClick={() => (deck ? void startGame(deck) : void init())}>
              Retry
            </button>
          </div>
        ) : (
          'Loading game…'
        )}
        {onExit && (
          <div>
            <button style={{ marginTop: 12, fontSize: 12 }} onClick={onExit}>← Back to deck</button>
          </div>
        )}
      </main>
    );
  }

  const [you, opp] = view.players;
  const changes = changedSlots(prev, view);
  const turnLabel = view.phase === 'over'
    ? `Game over — winner: P${view.winner + 1}`
    : `Turn ${view.turn} · phase ${view.phase} · active P${view.active + 1}` + (view.pending ? ` · waiting on P${view.pending.player + 1} (${view.pending.kind})` : '');

  return (
    <main style={{ padding: 16, maxWidth: 900, margin: '0 auto', display: 'grid', gap: 12 }}>
      {onExit && (
        <button onClick={onExit} style={{ justifySelf: 'start', fontSize: 12 }}>← Back to deck</button>
      )}
      <PlayerSide player={opp} role="opponent" playerIndex={1} changes={changes} />
      <div style={{ textAlign: 'center', fontSize: 13, fontWeight: 600 }}>{turnLabel}</div>
      {error && <div style={{ color: '#ff6b6b', fontSize: 13, textAlign: 'center' }}>{error}</div>}
      <PlayerSide player={you} role="you" playerIndex={0} changes={changes} />
      <Hand player={you} />
      <ActionBar view={view} />
      <Log lines={view.narration} />
    </main>
  );
}
