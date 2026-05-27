import { useMemo, useState } from 'react';
import type { Card } from '../types/card';
import { SHOP_TICKETS } from '../stores/draftStore';
import { buildShopInventory } from '../shop/inventory';
import { CardTile } from './CardTile';
import { CardPreview } from './CardPreview';

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
  const [hovered, setHovered] = useState<Card | null>(null);

  const inventory = useMemo(() => buildShopInventory(picks, fullPool), [picks, fullPool]);

  const byId = useMemo(() => new Map(fullPool.map((c) => [c.id, c])), [fullPool]);

  const purchased = purchasedIds
    .map((id) => byId.get(id))
    .filter((c): c is Card => Boolean(c));

  const ticketsLeft = SHOP_TICKETS - purchasedIds.length;

  const handleClick = (card: Card) => {
    if (purchasedIds.includes(card.id)) {
      onUnpurchase(card.id);
    } else if (ticketsLeft > 0) {
      onPurchase(card.id);
    }
  };

  const renderShelf = (cards: Card[]) => {
    if (cards.length === 0) {
      return <p style={{ fontSize: 13, opacity: 0.5 }}>No candidates.</p>;
    }
    return (
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
                padding: 4,
              }}
            >
              <CardTile
                card={c}
                size="lg"
                onPick={disabled ? undefined : () => handleClick(c)}
                onHover={setHovered}
              />
              <div style={{ textAlign: 'center', fontSize: 11, opacity: 0.65, marginTop: 2 }}>
                {isPurchased ? '✓ in collection' : '1 ticket'}
              </div>
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <main style={{ flex: 1, padding: '16px 24px', maxWidth: 1300 }}>
        <header
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'baseline',
            marginBottom: 8,
            gap: 12,
            flexWrap: 'wrap',
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
          1 ticket adds a card to your collection at the 2× copy cap. Tickets don't carry over.
          Drafted-by-name cards aren't shown — you already have access to them.
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
