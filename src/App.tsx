import { useEffect, useMemo, useState } from 'react';
import { loadCardPool, type CardPool } from './data';
import { DraftView } from './components/DraftView';
import { ShopView } from './components/ShopView';
import { ReviewView } from './components/ReviewView';
import { derivePhase, TOTAL_PACKS, useDraftStore } from './stores/draftStore';
import type { Card } from './types/card';

export function App() {
  const [pool, setPool] = useState<CardPool | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadCardPool().then(setPool).catch((e) => setError(String(e)));
  }, []);

  const {
    pickIds,
    offerIds,
    packOffers,
    shopPurchasedIds,
    shopFinalized,
    start,
    pick,
    purchase,
    unpurchase,
    finalizeShop,
    reset,
  } = useDraftStore();

  const byId = useMemo(() => {
    if (!pool) return new Map<string, Card>();
    return new Map(pool.draftableCards.map((c) => [c.id, c]));
  }, [pool]);

  const picks = useMemo(
    () => pickIds.map((id) => byId.get(id)).filter((c): c is Card => Boolean(c)),
    [pickIds, byId],
  );
  const offer = useMemo(
    () => offerIds.map((id) => byId.get(id)).filter((c): c is Card => Boolean(c)),
    [offerIds, byId],
  );
  const purchased = useMemo(
    () => shopPurchasedIds.map((id) => byId.get(id)).filter((c): c is Card => Boolean(c)),
    [shopPurchasedIds, byId],
  );

  const phase = derivePhase({ pickIds, offerIds, shopFinalized });

  if (error) return <main style={{ padding: 24 }}>Error: {error}</main>;
  if (!pool) return <main style={{ padding: 24 }}>Loading card data…</main>;

  if (phase === 'unstarted') {
    return (
      <main style={{ padding: 24, maxWidth: 720, margin: '0 auto', textAlign: 'center' }}>
        <h1>Pocket Draft</h1>
        <p style={{ opacity: 0.75 }}>
          {pool.draftableCards.length} draftable cards across {pool.sets.length} sets.
        </p>
        <p style={{ opacity: 0.6, fontSize: 13 }}>
          {TOTAL_PACKS} rounds. Each round shows 5 cards — pick one. Then a quick shop, then your finished deck.
        </p>
        <button style={{ fontSize: 16, padding: '8px 20px' }} onClick={() => start(pool.draftableCards)}>
          Start draft
        </button>
      </main>
    );
  }

  if (phase === 'drafting') {
    return (
      <DraftView
        packIndex={pickIds.length}
        offer={offer}
        picks={picks}
        onPick={(card) => pick(card.id, pool.draftableCards)}
      />
    );
  }

  if (phase === 'shop') {
    return (
      <ShopView
        pool={pool.draftableCards}
        picks={picks}
        packOffers={packOffers}
        pickIds={pickIds}
        purchasedIds={shopPurchasedIds}
        onPurchase={purchase}
        onUnpurchase={unpurchase}
        onFinalize={finalizeShop}
      />
    );
  }

  return <ReviewView picks={picks} purchased={purchased} onReset={reset} />;
}
