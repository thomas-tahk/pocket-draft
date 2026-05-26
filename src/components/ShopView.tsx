import { useMemo, useState } from 'react';
import type { Card } from '../types/card';
import { SHOP_TICKETS } from '../stores/draftStore';
import { STAPLE_SHELF_IDS } from '../shop/staples';
import { buildUndraftedShelf } from '../shop/synergy';
import { CardTile } from './CardTile';
import { CardPreview } from './CardPreview';

type Props = {
  pool: Card[];
  picks: Card[];
  packOffers: string[][];
  pickIds: string[];
  purchasedIds: string[];
  onPurchase: (cardId: string) => void;
  onUnpurchase: (cardId: string) => void;
  onFinalize: () => void;
  onCancel: () => void;
};

export function ShopView({
  pool,
  picks,
  packOffers,
  pickIds,
  purchasedIds,
  onPurchase,
  onUnpurchase,
  onFinalize,
  onCancel,
}: Props) {
  const [hovered, setHovered] = useState<Card | null>(null);

  const byId = useMemo(() => new Map(pool.map((c) => [c.id, c])), [pool]);

  const staples = useMemo(
    () =>
      STAPLE_SHELF_IDS.map((id) => byId.get(id)).filter((c): c is Card => Boolean(c)),
    [byId],
  );

  const undrafted = useMemo(
    () => buildUndraftedShelf(packOffers, pickIds, byId),
    [packOffers, pickIds, byId],
  );

  const ticketsLeft = SHOP_TICKETS - purchasedIds.length;
  const purchased = purchasedIds.map((id) => byId.get(id)).filter((c): c is Card => Boolean(c));

  const handleClick = (card: Card) => {
    if (purchasedIds.includes(card.id)) {
      onUnpurchase(card.id);
    } else if (ticketsLeft > 0) {
      onPurchase(card.id);
    }
  };

  const renderShelf = (cards: Card[]) => (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(190px, 1fr))',
        gap: 8,
        alignItems: 'start',
      }}
    >
      {cards.map((c) => {
        const isPurchased = purchasedIds.includes(c.id);
        const disabled = !isPurchased && ticketsLeft === 0;
        return (
          <div
            key={c.id}
            style={{
              opacity: disabled ? 0.4 : 1,
              outline: isPurchased ? '2px solid #4c8' : 'none',
              outlineOffset: 2,
              borderRadius: 10,
            }}
          >
            <CardTile
              card={c}
              size="lg"
              onPick={disabled ? undefined : () => handleClick(c)}
              onHover={setHovered}
            />
            <div style={{ textAlign: 'center', fontSize: 11, opacity: 0.65, marginTop: 2 }}>
              {isPurchased ? '✓ purchased' : '1 ticket'}
            </div>
          </div>
        );
      })}
    </div>
  );

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <main style={{ flex: 1, padding: '16px 24px', maxWidth: 1200 }}>
        <header
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'baseline',
            marginBottom: 16,
          }}
        >
          <h1 style={{ fontSize: 20, margin: 0 }}>Shop</h1>
          <div style={{ display: 'flex', gap: 12, alignItems: 'baseline' }}>
            <div style={{ fontSize: 14, opacity: 0.75 }}>
              Tickets: {ticketsLeft} / {SHOP_TICKETS}
            </div>
            <button onClick={onFinalize}>
              {purchasedIds.length === 0 ? 'Skip shop' : 'Finish shop'}
            </button>
            <button
              onClick={() => {
                if (window.confirm('Cancel this draft? You will lose all picks and purchases.')) {
                  onCancel();
                }
              }}
              style={{ fontSize: 12 }}
            >
              Cancel draft
            </button>
          </div>
        </header>

        <p style={{ fontSize: 13, opacity: 0.65, marginTop: 0 }}>
          Spend up to {SHOP_TICKETS} tickets. Click to buy; click again to unbuy. Tickets don't carry over.
        </p>

        <h2 style={{ fontSize: 15, marginTop: 24, marginBottom: 8 }}>
          Staples ({staples.length})
        </h2>
        {renderShelf(staples)}

        <h2 style={{ fontSize: 15, marginTop: 24, marginBottom: 8 }}>
          From undrafted ({undrafted.length})
        </h2>
        {undrafted.length === 0 ? (
          <p style={{ fontSize: 13, opacity: 0.5 }}>No candidates found.</p>
        ) : (
          renderShelf(undrafted)
        )}
      </main>

      <aside
        style={{
          width: 260,
          borderLeft: '1px solid #8884',
          padding: 16,
          height: '100vh',
          overflowY: 'auto',
          position: 'sticky',
          top: 0,
          boxSizing: 'border-box',
        }}
      >
        <div style={{ fontSize: 12, opacity: 0.6, marginBottom: 6 }}>
          Drafted ({picks.length})
        </div>
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(3, 1fr)',
            gap: 4,
            alignItems: 'start',
            marginBottom: 16,
          }}
        >
          {picks.map((c, i) => (
            <CardTile key={`d-${c.id}-${i}`} card={c} size="sm" onHover={setHovered} />
          ))}
        </div>

        {purchased.length > 0 && (
          <>
            <div style={{ fontSize: 12, opacity: 0.6, marginBottom: 6 }}>
              Purchased ({purchased.length})
            </div>
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(3, 1fr)',
                gap: 4,
                alignItems: 'start',
              }}
            >
              {purchased.map((c) => (
                <CardTile key={`p-${c.id}`} card={c} size="sm" onHover={setHovered} />
              ))}
            </div>
          </>
        )}
      </aside>

      <CardPreview card={hovered} />
    </div>
  );
}
