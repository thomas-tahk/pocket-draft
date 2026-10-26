import { useMemo, useState } from 'react';
import { AbandonDraft } from './AbandonDraft';
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
      return <p className="muted-sm">No candidates.</p>;
    }
    return (
      <div className="shelf-grid">
        {cards.map((c) => {
          const count = countById.get(c.id) ?? 0;
          const atPerCardCap = count >= COPY_CAP_BY_NAME;
          const noTickets = ticketsLeft === 0;
          const addDisabled = atPerCardCap || noTickets;
          return (
            <div
              key={c.id}
              className={`shelf-cell${count > 0 ? ' is-picked' : ''}${addDisabled && count === 0 ? ' is-dim' : ''}`}
            >
              <CardTile card={c} size="lg" onHover={handleHover} />
              <div className="step-row">
                <button
                  onClick={() => onUnpurchase(c.id)}
                  disabled={count === 0}
                  className="btn-step"
                  aria-label={`− Remove ${c.name}`}
                >
                  −
                </button>
                <span className="muted-xs" style={{ minWidth: 36, textAlign: 'center' }}>
                  {count > 0 ? `×${count}` : '1 ticket'}
                </span>
                <button
                  onClick={() => onPurchase(c.id)}
                  disabled={addDisabled}
                  className="btn-step"
                  aria-label={`+ Add ${c.name}`}
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
    <div className="screen">
      <main className="screen-main wide">
        <header className="screen-header">
          <h1 className="screen-title">Shop</h1>
          <div className="screen-header-actions">
            <div className="muted">
              <span role="status">Tickets: {ticketsLeft} / {SHOP_TICKETS}</span>
            </div>
            <button className="btn-primary" onClick={onFinalize}>
              {purchasedIds.length === 0 ? 'Skip shop' : 'Continue to deckbuild'}
            </button>
            <AbandonDraft loses="all picks and purchases" onAbandon={onCancel} />
          </div>
        </header>
        <p className="muted-sm" style={{ marginTop: 0 }}>
          Each ticket buys one deck copy of a card. Spend up to {COPY_CAP_BY_NAME} tickets on the
          same card. Tickets don't carry over. Drafted-by-name cards aren't shown — you already
          have access to them.
        </p>

        <section style={{ marginTop: 16 }}>
          <h2 className="section-title">
            Generic staple Trainers ({inventory.staples.length})
          </h2>
          {renderShelf(inventory.staples)}
        </section>

        <section style={{ marginTop: 24 }}>
          <h2 className="section-title">
            Same-type EX / Mega-EX ({inventory.alts.length})
          </h2>
          {renderShelf(inventory.alts)}
        </section>
      </main>

      <aside className="screen-aside" aria-label="Your cards" style={{ ['--aside-w' as string]: '260px' }}>
        <h2 id="shop-drafted" className="aside-title">
          Drafted ({picks.length})
        </h2>
        <div className="tile-grid-sm" role="group" aria-labelledby="shop-drafted" style={{ marginBottom: 16 }}>
          {picks.map((c, i) => (
            <CardTile key={`d-${c.id}-${i}`} card={c} size="sm" onHover={handleHover} />
          ))}
        </div>

        {purchased.length > 0 && (
          <>
            <h2 id="shop-purchased" className="aside-title">
              Purchased ({purchased.length})
            </h2>
            <div className="tile-grid-sm" role="group" aria-labelledby="shop-purchased">
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
