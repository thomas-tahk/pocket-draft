import { useState } from 'react';
import type { Card, EnergyType } from '../types/card';
import { TOTAL_PACKS } from '../stores/draftStore';
import { tallyTypes } from '../draft/weights';
import { OfferRow } from './OfferRow';
import { CardTile, type HoverPayload } from './CardTile';
import { CardPreview, type HoverState } from './CardPreview';

type Props = {
  packIndex: number;
  offer: Card[];
  picks: Card[];
  onPick: (card: Card) => void;
  onCancel: () => void;
};

export function DraftView({ packIndex, offer, picks, onPick, onCancel }: Props) {
  const [hover, setHover] = useState<HoverState>(null);
  const handleHover = (p: HoverPayload | null) => setHover(p);

  const tally = tallyTypes(picks);
  const sortedTally = (Object.entries(tally) as [EnergyType, number][]).sort(
    (a, b) => b[1] - a[1],
  );

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <main style={{ flex: 1, padding: '0 24px 24px' }}>
        <header
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'baseline',
            padding: '16px 0',
            marginBottom: 8,
            position: 'sticky',
            top: 0,
            background: 'var(--surface)',
            zIndex: 5,
            borderBottom: '1px solid #8882',
          }}
        >
          <h1 style={{ fontSize: 20, margin: 0 }}>Pocket Draft</h1>
          <div style={{ display: 'flex', gap: 16, alignItems: 'baseline', flexWrap: 'wrap' }}>
            <div style={{ fontSize: 14, opacity: 0.75 }}>
              Pack {packIndex + 1} / {TOTAL_PACKS}
              {sortedTally.length > 0 && (
                <span style={{ marginLeft: 16 }}>
                  {sortedTally.map(([t, n]) => `${t} ×${n}`).join('  ')}
                </span>
              )}
            </div>
            <button
              onClick={() => {
                if (window.confirm('Cancel this draft? You will lose all picks so far.')) {
                  onCancel();
                }
              }}
              style={{ fontSize: 12 }}
            >
              Cancel draft
            </button>
          </div>
        </header>

        <section style={{ marginTop: 16 }}>
          <OfferRow offer={offer} onPick={onPick} onHover={handleHover} />
        </section>
      </main>

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
            gridTemplateColumns: 'repeat(3, 84px)',
            justifyContent: 'center',
            gap: 4,
            alignItems: 'start',
          }}
        >
          {picks.map((c, i) => (
            <CardTile key={`${c.id}-${i}`} card={c} size="sm" onHover={handleHover} />
          ))}
        </div>
      </aside>

      <CardPreview hover={hover} />
    </div>
  );
}
