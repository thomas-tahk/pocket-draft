import { useEffect, useRef, useState } from 'react';
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
  const packKey = offer.map((c) => c.id).join();
  useEffect(() => setHover(null), [packKey]);
  // After a pick the tile that had focus is gone: keep the keyboard on the new pack.
  const packRef = useRef<HTMLElement>(null);
  const openingPack = useRef(packKey);
  useEffect(() => {
    if (openingPack.current === packKey) return; // the pack that opened the screen: the heading has focus
    // A frame later, so it lands after development's double effect run on the new tiles (which clears a preview).
    const frame = requestAnimationFrame(() => {
      packRef.current?.querySelector<HTMLElement>('.offer-grid button')?.focus({ preventScroll: true });
    });
    return () => cancelAnimationFrame(frame);
  }, [packKey]);

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
              <span role="status">Pack {packIndex + 1} / {TOTAL_PACKS}</span>
              {sortedTally.length > 0 && (
                <span style={{ marginLeft: 16 }}>
                  <TypeTally entries={sortedTally} />
                </span>
              )}
            </div>
            <AbandonDraft loses="all picks so far" onAbandon={onCancel} />
          </div>
        </header>

        <section ref={packRef} style={{ marginTop: 16 }}>
          <h2 className="sr-only">Choose a card</h2>
          <OfferRow packIndex={packIndex} offer={offer} onPick={onPick} onHover={handleHover} />
        </section>
      </main>

      <aside className="screen-aside" aria-labelledby="drafted-title">
        <h2 id="drafted-title" className="aside-title">
          Drafted ({picks.length})
        </h2>
        <div className="tile-grid-sm" role="group" aria-labelledby="drafted-title">
          {picks.map((c, i) => (
            <CardTile key={`${c.id}-${i}`} card={c} size="sm" onHover={handleHover} />
          ))}
        </div>
      </aside>

      <CardPreview hover={hover} dock />
    </div>
  );
}
