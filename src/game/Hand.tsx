import type { PlayerView } from './types';
import { GameCard } from './GameCard';
import { useGameStore } from './store';

export function Hand({ player }: { player: PlayerView }) {
  const { selection, select } = useGameStore();
  return (
    <div>
      <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 4 }}>Your hand</div>
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        {player.hand.map((c) => {
          const sel = selection?.kind === 'hand' && selection.cardId === c.id;
          const selSetup = selection?.kind === 'setupActive' && selection.cardId === c.id;
          return (
            <GameCard
              key={c.id}
              card={c}
              size="sm"
              selected={sel || selSetup}
              title={`${c.name} · ${c.stage} · ${c.id}`}
              onClick={() => select({ kind: 'hand', cardId: c.id })}
            />
          );
        })}
      </div>
    </div>
  );
}
