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

// A new pack ignores a touch tap for this long after it appears: a double-tap's second tap (or a slow
// finger lifting) would otherwise land on the new pack's tile at the same spot. Keyboard and mouse are exempt.
export const PACK_SETTLE_MS = 350;

export function DraftView({ packIndex, offer, picks, onPick, onCancel }: Props) {
  const [hover, setHover] = useState<HoverState>(null);
  // When this pack appeared, and the last kind of input ('touch' | 'pen' | 'mouse' | 'key').
  const shownAt = useRef(-Infinity); // the pack that opens the screen is not guarded, only the ones a pick brings
  const lastInput = useRef('mouse');
  const handlePick = (card: Card) => {
    const touch = lastInput.current === 'touch' || lastInput.current === 'pen';
    if (touch && performance.now() - shownAt.current < PACK_SETTLE_MS) return;
    onPick(card);
  };
  const handleHover = (p: HoverPayload | null) => setHover(p);
  // A new pack replaces the tiles; drop any preview or zoom left from the old ones.
  const packKey = offer.map((c) => c.id).join();
  const firstPack = useRef(packKey);
  useEffect(() => {
    if (firstPack.current !== packKey) shownAt.current = performance.now();
    setHover(null);
  }, [packKey]);
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
              Pack {packIndex + 1} / {TOTAL_PACKS}
              {sortedTally.length > 0 && (
                <span style={{ marginLeft: 16 }}>
                  <TypeTally entries={sortedTally} />
                </span>
              )}
            </div>
            <AbandonDraft loses="all picks so far" onAbandon={onCancel} />
          </div>
          {/* The same count as "Pack n / 20", drawn as segments; filled up to the pack on screen. */}
          <div className="pack-progress" aria-hidden="true">
            {Array.from({ length: TOTAL_PACKS }, (_, i) => (
              <i key={i} className={i <= packIndex ? 'on' : undefined} />
            ))}
          </div>
        </header>

        <section
          ref={packRef}
          style={{ marginTop: 16 }}
          onPointerDownCapture={(e) => (lastInput.current = e.pointerType)}
          onKeyDownCapture={() => (lastInput.current = 'key')}
        >
          <h2 className="sr-only">Choose a card</h2>
          <OfferRow packIndex={packIndex} offer={offer} onPick={handlePick} onHover={handleHover} />
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
