import type { Card } from '../types/card';
import { CardTile } from './CardTile';

type Props = {
  picks: Card[];
  onHover: (card: Card | null) => void;
};

export function DeckSidebar({ picks, onHover }: Props) {
  return (
    <aside
      style={{
        width: 300,
        borderLeft: '1px solid #8884',
        padding: 16,
        height: '100vh',
        overflowY: 'auto',
        position: 'sticky',
        top: 0,
        boxSizing: 'border-box',
      }}
    >
      <div style={{ fontSize: 12, opacity: 0.6, marginBottom: 8 }}>
        Drafted ({picks.length})
      </div>
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(3, 1fr)',
          gap: 4,
          alignItems: 'start',
        }}
      >
        {picks.map((c, i) => (
          <CardTile key={`${c.id}-${i}`} card={c} size="sm" onHover={onHover} />
        ))}
      </div>
    </aside>
  );
}
