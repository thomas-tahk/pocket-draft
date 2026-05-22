import type { Card, EnergyType } from '../types/card';
import { tallyTypes } from '../draft/weights';
import { CardTile } from './CardTile';

type Props = {
  picks: Card[];
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

export function ReviewView({ picks, onReset }: Props) {
  const tally = tallyTypes(picks);
  const sortedTally = (Object.entries(tally) as [EnergyType, number][])
    .sort((a, b) => b[1] - a[1]);

  const exportPayload = {
    version: 'v0.1',
    timestamp: new Date().toISOString(),
    pickIds: picks.map((c) => c.id),
  };

  return (
    <main style={{ padding: '16px 24px', maxWidth: 1100, margin: '0 auto' }}>
      <header
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          marginBottom: 16,
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

      <div style={{ fontSize: 13, opacity: 0.75, marginBottom: 16 }}>
        20 cards · {sortedTally.map(([t, n]) => `${t} ×${n}`).join('  ')}
      </div>

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(5, 1fr)',
          gap: 8,
          alignItems: 'start',
        }}
      >
        {picks.map((c, i) => (
          <CardTile key={`${c.id}-${i}`} card={c} size="lg" />
        ))}
      </div>
    </main>
  );
}
