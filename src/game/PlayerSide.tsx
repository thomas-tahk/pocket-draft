import type { PlayerView, InPlayView } from './types';
import { GameCard } from './GameCard';

function EnergyLine({ mon }: { mon: InPlayView }) {
  const parts = Object.entries(mon.energy).map(([t, n]) => `${t}×${n}`);
  return <div style={{ fontSize: 12, opacity: 0.8 }}>{parts.length ? parts.join(' ') : 'no energy'} · {mon.totalEnergy} total</div>;
}

function ActiveSlot({ mon }: { mon: InPlayView | null }) {
  if (!mon) return <div style={{ fontSize: 13, opacity: 0.6 }}>— no active —</div>;
  return (
    <div>
      <GameCard card={mon.card} size="lg" />
      <div style={{ fontSize: 13 }}>HP {mon.remainingHp}/{mon.card.hp}{mon.poisoned ? ' ☠' : ''}{mon.burned ? ' 🔥' : ''}</div>
      <EnergyLine mon={mon} />
    </div>
  );
}

export function PlayerSide({ player, role }: { player: PlayerView; role: 'you' | 'opponent' }) {
  return (
    <section style={{ padding: 12, border: '1px solid #33363d', borderRadius: 10 }}>
      <div style={{ fontSize: 12, opacity: 0.7, marginBottom: 6 }}>
        {role === 'you' ? 'YOU' : 'OPPONENT'} · pts {player.points}/3 · deck {player.deckCount} · discard {player.discardCount}
        {role === 'opponent' ? ` · hand ${player.hand.length}` : ''}
      </div>
      <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
        <ActiveSlot mon={player.active} />
        <div style={{ display: 'flex', gap: 8 }}>
          {player.bench.map((mon, i) => (
            <div key={i}>
              <GameCard card={mon.card} size="sm" />
              <div style={{ fontSize: 11 }}>{mon.remainingHp}/{mon.card.hp} · {mon.totalEnergy}⚡</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
