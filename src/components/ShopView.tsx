import { useMemo, useState } from 'react';
import type { Card } from '../types/card';
import { SHOP_TICKETS } from '../stores/draftStore';
import { COPY_CAP_BY_NAME } from '../draft/expansions';
import { buildShopInventory } from '../shop/inventory';
import { CardTile, type HoverPayload } from './CardTile';
import { CardPreview, type HoverState } from './CardPreview';

type Props = {
  picks: Card[];
  fullPool: Card[];
  purchasedIds: string[];
  onPurchase: (cardId: string) => void;
  onUnpurchase: (cardId: string) => void;
  onFinalize: () => void;
  onCancel: () => void;
};

export function ShopView({
  picks,
  fullPool,
  purchasedIds,
  onPurchase,
  onUnpurchase,
  onFinalize,
  onCancel,
}: Props) {
  const [hover, setHover] = useState<HoverState>(null);
  const handleHover = (p: HoverPayload | null) => setHover(p);

  const inventory = useMemo(() => buildShopInventory(picks, fullPool), [picks, fullPool]);
  const byId = useMemo(() => new Map(fullPool.map((c) => [c.id, c])), [fullPool]);
  const purchased = purchasedIds.map((id) => byId.get(id)).filter((c): c is Card => Boolean(c));
  const ticketsLeft = SHOP_TICKETS - purchasedIds.length;

  const countById = useMemo(() => {
    const m = new Map<string, number>();
    for (const id of purchasedIds) m.set(id, (m.get(id) ?? 0) + 1);
    return m;
  }, [purchasedIds]);

  const renderShelf = (cards: Card[]) => {
    if (cards.length === 0) {
      return <p style={{ fontSize: 13, opacity: 0.5 }}>No candidates.</p>;
    }
    return (
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, 180px)',
          justifyContent: 'start',
          gap: 12,
          alignItems: 'start',
        }}
      >
        {cards.map((c) => {
          const count = countById.get(c.id) ?? 0;
          const atPerCardCap = count >= COPY_CAP_BY_NAME;
          const noTickets = ticketsLeft === 0;
          const addDisabled = atPerCardCap || noTickets;
          return (
            <div
              key={c.id}
              style={{
                opacity: addDisabled && count === 0 ? 0.4 : 1,
                outline: count > 0 ? '2px solid #4c8' : 'none',
                outlineOffset: 2,
                borderRadius: 10,
                padding: 2,
                width: 180,
                boxSizing: 'border-box',
              }}
            >
              <CardTile card={c} size="lg" onHover={handleHover} />
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 6,
                  fontSize: 12,
                  marginTop: 4,
                }}
              >
                <button
                  onClick={() => onUnpurchase(c.id)}
                  disabled={count === 0}
                  style={{ minWidth: 28 }}
                  aria-label="Remove one"
                >
                  −
                </button>
                <span style={{ opacity: 0.8, minWidth: 36, textAlign: 'center' }}>
                  {count > 0 ? `×${count}` : '1 ticket'}
                </span>
                <button
                  onClick={() => onPurchase(c.id)}
                  disabled={addDisabled}
                  style={{ minWidth: 28 }}
                  aria-label="Add one"
                  title={
                    atPerCardCap
                      ? `Max ${COPY_CAP_BY_NAME} of any card`
                      : noTickets
                        ? 'No tickets left'
                        : ''
                  }
                >
                  +
                </button>
              </div>
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <main style={{ flex: 1, padding: '0 24px 24px', maxWidth: 1300 }}>
        <header
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'baseline',
            padding: '16px 0',
            marginBottom: 8,
            gap: 12,
            flexWrap: 'wrap',
            position: 'sticky',
            top: 0,
            background: 'var(--surface)',
            zIndex: 20,
            borderBottom: '1px solid #8882',
          }}
        >
          <h1 style={{ fontSize: 20, margin: 0 }}>Shop</h1>
          <div style={{ display: 'flex', gap: 12, alignItems: 'baseline' }}>
            <div style={{ fontSize: 14, opacity: 0.75 }}>
              Tickets: {ticketsLeft} / {SHOP_TICKETS}
            </div>
            <button onClick={onFinalize}>
              {purchasedIds.length === 0 ? 'Skip shop' : 'Continue to deckbuild'}
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
        <p style={{ fontSize: 13, opacity: 0.6, marginTop: 0 }}>
          Each ticket buys one deck copy of a card. Spend up to {COPY_CAP_BY_NAME} tickets on the
          same card. Tickets don't carry over. Drafted-by-name cards aren't shown — you already
          have access to them.
        </p>

        <section style={{ marginTop: 16 }}>
          <h2 style={{ fontSize: 14, opacity: 0.75, margin: '0 0 8px' }}>
            Generic staple Trainers ({inventory.staples.length})
          </h2>
          {renderShelf(inventory.staples)}
        </section>

        <section style={{ marginTop: 24 }}>
          <h2 style={{ fontSize: 14, opacity: 0.75, margin: '0 0 8px' }}>
            Same-type EX / Mega-EX ({inventory.alts.length})
          </h2>
          {renderShelf(inventory.alts)}
        </section>
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
          background: 'var(--surface)',
          zIndex: 10,
        }}
      >
        <div style={{ fontSize: 12, opacity: 0.6, marginBottom: 6 }}>
          Drafted ({picks.length})
        </div>
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(3, 84px)',
            justifyContent: 'center',
            gap: 4,
            alignItems: 'start',
            marginBottom: 16,
          }}
        >
          {picks.map((c, i) => (
            <CardTile key={`d-${c.id}-${i}`} card={c} size="sm" onHover={handleHover} />
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
                gridTemplateColumns: 'repeat(3, 84px)',
                justifyContent: 'center',
                gap: 4,
                alignItems: 'start',
              }}
            >
              {purchased.map((c, i) => (
                <CardTile key={`p-${c.id}-${i}`} card={c} size="sm" onHover={handleHover} />
              ))}
            </div>
          </>
        )}
      </aside>

      <CardPreview hover={hover} />
    </div>
  );
}
