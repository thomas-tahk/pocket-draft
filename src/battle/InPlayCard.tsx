import type { InPlayView } from '../game/types';
import { CardFace, EmptySlot, EnergyIcon } from './pieces';

// A Pokémon in play (L2): real card image, HP number and bar at the top right,
// attached energy as type icons along the bottom, status icons.

function hpColor(fraction: number): string {
  if (fraction > 0.5) return '#2fbf4a';
  if (fraction > 0.25) return '#f0a020';
  return '#e33a2a';
}

function HpBadge({ mon }: { mon: InPlayView }) {
  const fraction = Math.max(0, mon.remainingHp) / mon.card.hp;
  return (
    <div className="bt-hp">
      <span className="bt-hp-num" data-testid="hp">{Math.max(0, mon.remainingHp)}</span>
      <span className="bt-hp-bar"><span style={{ width: `${fraction * 100}%`, background: hpColor(fraction) }} /></span>
    </div>
  );
}

function AttachedEnergy({ energy, size }: { energy: Record<string, number>; size: number }) {
  const icons = Object.entries(energy).flatMap(([type, n]) => Array.from({ length: n }, () => type));
  return (
    <div className="bt-attached">
      {icons.map((type, i) => <EnergyIcon key={i} type={type} size={size} />)}
    </div>
  );
}

function StatusIcons({ mon }: { mon: InPlayView }) {
  if (!mon.poisoned && !mon.burned) return null;
  return (
    <div className="bt-status">
      {mon.poisoned && <span style={{ background: '#8a3cc4' }} title="Poisoned">PSN</span>}
      {mon.burned && <span style={{ background: '#e2481f' }} title="Burned">BRN</span>}
    </div>
  );
}

export function InPlayCard({ mon, width }: { mon: InPlayView | null; width: number }) {
  if (!mon) return <EmptySlot width={width} />;
  return (
    <div style={{ position: 'relative' }} data-zoom={mon.card.id}>
      <CardFace card={mon.card} width={width} />
      <HpBadge mon={mon} />
      <StatusIcons mon={mon} />
      <AttachedEnergy energy={mon.energy} size={width > 90 ? 18 : 14} />
    </div>
  );
}
