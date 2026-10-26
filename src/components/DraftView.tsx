import { useEffect, useState } from 'react';
import { AbandonDraft } from './AbandonDraft';
import { TypeTally } from '../battle/energy';
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
  // A new pack replaces the tiles; drop any preview or zoom left from the old ones.
  useEffect(() => setHover(null), [offer]);

  const tally = tallyTypes(picks);
  const sortedTally = (Object.entries(tally) as [EnergyType, number][]).sort(
    (a, b) => b[1] - a[1],
  );

  return (
    <div className="screen">
      <main className="screen-main">
        <header className="screen-header">
          <h1 className="screen-title">Pocket Draft</h1>
          <div className="screen-header-actions">
            <div className="muted">
              Pack {packIndex + 1} / {TOTAL_PACKS}
              {sortedTally.length > 0 && (
                <span style={{ marginLeft: 16 }}>
                  <TypeTally entries={sortedTally} />
                </span>
              )}
            </div>
            <AbandonDraft loses="all picks so far" onAbandon={onCancel} />
          </div>
        </header>

        <section style={{ marginTop: 16 }}>
          <OfferRow offer={offer} onPick={onPick} onHover={handleHover} />
        </section>
      </main>

      <aside className="screen-aside">
        <div className="muted-xs" style={{ marginBottom: 8 }}>
          Drafted ({picks.length})
        </div>
        <div className="tile-grid-sm">
          {picks.map((c, i) => (
            <CardTile key={`${c.id}-${i}`} card={c} size="sm" onHover={handleHover} />
          ))}
        </div>
      </aside>

      <CardPreview hover={hover} dock />
    </div>
  );
}
