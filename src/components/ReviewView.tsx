import type { Card, EnergyType } from '../types/card';
import { tallyTypes } from '../draft/weights';
import { CardTile } from './CardTile';

type Props = {
  picks: Card[];
  purchased: Card[];
  onReset: () => void;
};

function download(filename: string, content: string) {
  const blob = new Blob([content], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

export function ReviewView({ picks, purchased, onReset }: Props) {
  const all = [...picks, ...purchased];
  const tally = tallyTypes(all);
  const sortedTally = (Object.entries(tally) as [EnergyType, number][])
    .sort((a, b) => b[1] - a[1]);

  const exportPayload = {
    version: 'v0.2',
    timestamp: new Date().toISOString(),
    pickIds: picks.map((c) => c.id),
    purchasedIds: purchased.map((c) => c.id),
  };

  const Section = ({ title, cards }: { title: string; cards: Card[] }) =>
    cards.length === 0 ? null : (
      <section style={{ marginTop: 20 }}>
        <h2 style={{ fontSize: 14, opacity: 0.75, margin: '0 0 8px' }}>
          {title} ({cards.length})
        </h2>
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(5, 1fr)',
            gap: 8,
            alignItems: 'start',
          }}
        >
          {cards.map((c, i) => (
            <CardTile key={`${c.id}-${i}`} card={c} size="lg" />
          ))}
        </div>
      </section>
    );

  return (
    <main style={{ padding: '16px 24px', maxWidth: 1100, margin: '0 auto' }}>
      <header
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          marginBottom: 8,
        }}
      >
        <h1 style={{ fontSize: 20, margin: 0 }}>Draft complete</h1>
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            onClick={() =>
              download(`pocket-draft-${Date.now()}.json`, JSON.stringify(exportPayload, null, 2))
            }
          >
            Export JSON
          </button>
          <button onClick={onReset}>New draft</button>
        </div>
      </header>

      <div style={{ fontSize: 13, opacity: 0.75 }}>
        {all.length} cards · {sortedTally.map(([t, n]) => `${t} ×${n}`).join('  ')}
      </div>

      <Section title="Drafted" cards={picks} />
      <Section title="Purchased" cards={purchased} />
    </main>
  );
}
