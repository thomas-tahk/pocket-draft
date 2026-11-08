import { useEffect, useMemo, useState } from 'react';
import type { CardPool } from '../data';
import { announce } from '../announce';
import { defaultZoneTypes } from '../game/energy';
import { deckList, deckQrPayload, type DeckNrs } from '../share/deckQr';

let nrsLoad: Promise<DeckNrs> | null = null;
function loadDeckNrs(): Promise<DeckNrs> {
  nrsLoad ??= fetch(`${import.meta.env.BASE_URL}data/deck-nrs.json`).then((r) => {
    if (!r.ok) throw new Error(`deck-nrs.json: ${r.status}`);
    return r.json() as Promise<DeckNrs>;
  });
  nrsLoad.catch(() => { nrsLoad = null; });
  return nrsLoad;
}

// Deck QR (scan it with Pocket's Scan Code) and Copy list (Limitless text format).
// `energy` is the deck's Energy Zone choice; absent, the default the battle would use.
export function DeckShare({ deck, pool, energy }: { deck: Record<string, number>; pool: CardPool; energy?: string[] }) {
  const [copied, setCopied] = useState(false);
  const [showQr, setShowQr] = useState(false);
  const [qr, setQr] = useState<{ src: string } | 'unavailable' | null>(null);

  const cards = useMemo(() => new Map(pool.fullCards.map((c) => [c.id, c])), [pool]);
  const zone = useMemo(() => {
    if (energy && energy.length > 0) return energy;
    const flat = Object.entries(deck).flatMap(([id, n]) => (cards.has(id) ? Array(n).fill(cards.get(id)!) : []));
    return defaultZoneTypes(flat);
  }, [energy, deck, cards]);

  useEffect(() => {
    if (!showQr) return;
    let live = true;
    Promise.all([loadDeckNrs(), import('qrcode')])
      .then(async ([nrs, QRCode]) => {
        const payload = deckQrPayload(deck, nrs, zone);
        if (!payload) return live && setQr('unavailable');
        const src = await QRCode.toDataURL(payload, { errorCorrectionLevel: 'M', margin: 2, width: 280 });
        if (live) setQr({ src });
      })
      .catch(() => live && setQr('unavailable'));
    return () => { live = false; };
  }, [showQr, deck, zone]);

  const copyList = () => {
    navigator.clipboard.writeText(deckList(deck, cards, zone)).then(() => {
      setCopied(true);
      announce('Copied!');
      setTimeout(() => setCopied(false), 1500);
    });
  };

  return (
    <>
      <button aria-expanded={showQr} onClick={() => setShowQr((v) => !v)}>{showQr ? 'Hide QR' : 'Deck QR'}</button>
      <button onClick={copyList}>{copied ? 'Copied!' : 'Copy list'}</button>
      {showQr && (
        <div className="deck-qr" data-testid="deck-qr" style={{ flexBasis: '100%' }}>
          {qr === null && <p className="muted-sm">Making QR…</p>}
          {qr === 'unavailable' && <p className="muted-sm">QR unavailable for this deck. Use Copy list.</p>}
          {qr && qr !== 'unavailable' && (
            <>
              <img src={qr.src} width={280} height={280} alt="Deck QR code" style={{ background: '#fff' }} />
              <p className="muted-sm">In Pocket: Decks → Scan Code.</p>
            </>
          )}
        </div>
      )}
    </>
  );
}
