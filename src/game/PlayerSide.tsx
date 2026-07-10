import type { PlayerView, InPlayView } from './types';
import { GameCard } from './GameCard';
import { useGameStore } from './store';

function EnergyLine({ mon }: { mon: InPlayView }) {
  const parts = Object.entries(mon.energy).map(([t, n]) => `${t}×${n}`);
  return <div style={{ fontSize: 12, opacity: 0.8 }}>{parts.length ? parts.join(' ') : 'no energy'} · {mon.totalEnergy} total</div>;
}

function ActiveSlot({ mon, highlight }: { mon: InPlayView | null; highlight?: boolean }) {
  if (!mon) return <div style={{ fontSize: 13, opacity: 0.6 }}>— no active —</div>;
  return (
    <div>
      <GameCard card={mon.card} size="lg" highlight={highlight} />
      <div style={{ fontSize: 13 }}>HP {mon.remainingHp}/{mon.card.hp}{mon.poisoned ? ' ☠' : ''}{mon.burned ? ' 🔥' : ''}</div>
      <EnergyLine mon={mon} />
    </div>
  );
}

export function PlayerSide({ player, role }: { player: PlayerView; role: 'you' | 'opponent' }) {
  const { selection, select } = useGameStore();
  // Legal-target hint for the you-side Active: a retreat destination when a
  // bench mon is selected, or an evolve target when the selected hand card
  // evolves from the current Active.
  const selectedHandCard = selection?.kind === 'hand' ? player.hand.find((c) => c.id === selection.cardId) : undefined;
  const activeHighlight =
    role === 'you' &&
    !!player.active &&
    (selection?.kind === 'bench' || selectedHandCard?.evolvesFrom === player.active.card.name);
  return (
    <section style={{ padding: 12, border: '1px solid #33363d', borderRadius: 10 }}>
      <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 6 }}>
        {role === 'you' ? 'YOU' : 'OPPONENT'} · pts {player.points}/3 · deck {player.deckCount} · discard {player.discardCount}
        {role === 'opponent' ? ` · hand ${player.hand.length}` : ''}
      </div>
      <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
        <ActiveSlot mon={player.active} highlight={activeHighlight} />
        <div style={{ display: 'flex', gap: 8 }}>
          {player.bench.map((mon, i) => (
            <div key={i}>
              <GameCard
                card={mon.card}
                size="sm"
                selected={role === 'you' && selection?.kind === 'bench' && selection.index === i}
                onClick={role === 'you' ? () => select({ kind: 'bench', index: i }) : undefined}
              />
              <div style={{ fontSize: 11 }}>{mon.remainingHp}/{mon.card.hp} · {mon.totalEnergy}⚡</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
