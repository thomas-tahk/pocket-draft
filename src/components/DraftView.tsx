import { useState } from 'react';
import type { Card, EnergyType } from '../types/card';
import { TOTAL_PACKS } from '../stores/draftStore';
import { tallyTypes } from '../draft/weights';
import { OfferRow } from './OfferRow';
import { DeckSidebar } from './DeckSidebar';
import { CardPreview } from './CardPreview';

type Props = {
  packIndex: number; // 0..19
  offer: Card[];
  picks: Card[];
  onPick: (card: Card) => void;
};

export function DraftView({ packIndex, offer, picks, onPick }: Props) {
  const [hovered, setHovered] = useState<Card | null>(null);
  const tally = tallyTypes(picks);
  const sortedTally = (Object.entries(tally) as [EnergyType, number][])
    .sort((a, b) => b[1] - a[1]);

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <main style={{ flex: 1, padding: '16px 24px' }}>
        <header
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'baseline',
            marginBottom: 16,
          }}
        >
          <h1 style={{ fontSize: 20, margin: 0 }}>Pocket Draft</h1>
          <div style={{ fontSize: 14, opacity: 0.75 }}>
            Pack {packIndex + 1} / {TOTAL_PACKS}
            {sortedTally.length > 0 && (
              <span style={{ marginLeft: 16 }}>
                {sortedTally.map(([t, n]) => `${t} ×${n}`).join('  ')}
              </span>
            )}
          </div>
        </header>

        <section>
          <OfferRow offer={offer} onPick={onPick} />
        </section>
      </main>

      <DeckSidebar picks={picks} onHover={setHovered} />
      <CardPreview card={hovered} />
    </div>
  );
}
