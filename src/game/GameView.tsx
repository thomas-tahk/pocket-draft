import { useEffect } from 'react';
import { useGameStore } from './store';
import { PlayerSide } from './PlayerSide';
import { GameCard } from './GameCard';
import { Log } from './Log';

export function GameView() {
  const { view, status, error, init } = useGameStore();

  useEffect(() => {
    void init();
  }, [init]);

  if (status !== 'ready' || !view) return <main style={{ padding: 24 }}>Loading game…</main>;

  const [you, opp] = view.players;
  const turnLabel = view.phase === 'over'
    ? `Game over — winner: P${view.winner + 1}`
    : `Turn ${view.turn} · phase ${view.phase} · active P${view.active + 1}` + (view.pending ? ` · waiting on P${view.pending.player + 1} (${view.pending.kind})` : '');

  return (
    <main style={{ padding: 16, maxWidth: 900, margin: '0 auto', display: 'grid', gap: 12 }}>
      <PlayerSide player={opp} role="opponent" />
      <div style={{ textAlign: 'center', fontSize: 13, fontWeight: 600 }}>{turnLabel}</div>
      {error && <div style={{ color: '#ff6b6b', fontSize: 13, textAlign: 'center' }}>{error}</div>}
      <PlayerSide player={you} role="you" />
      <div>
        <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 4 }}>Your hand</div>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          {you.hand.map((c) => <GameCard key={c.id} card={c} size="sm" />)}
        </div>
      </div>
      <Log lines={view.narration} />
    </main>
  );
}
