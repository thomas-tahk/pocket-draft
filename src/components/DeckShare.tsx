import { useMemo, useState } from 'react';
import type { CardPool } from '../data';
import { encodeDeck } from '../share/deckCode';

function download(filename: string, content: string) {
  const blob = new Blob([content], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

// Copy code, Copy link and Export JSON for a deck.
export function DeckShare({ deck, pool }: { deck: Record<string, number>; pool: CardPool }) {
  const [copied, setCopied] = useState<string | null>(null);
  const deckCode = useMemo(() => encodeDeck(deck), [deck]);
  const shareLink = `${location.origin}${location.pathname}?deck=${encodeURIComponent(deckCode)}`;

  const copy = (text: string, label: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(label);
      setTimeout(() => setCopied((c) => (c === label ? null : c)), 1500);
    });
  };

  const exportJson = () => {
    const names = new Map(pool.fullCards.map((c) => [c.id, c.name]));
    const cards = Object.entries(deck)
      .filter(([id]) => names.has(id))
      .map(([id, count]) => ({ id, name: names.get(id), count }));
    const payload = { version: 'v0.4', timestamp: new Date().toISOString(), deck: cards };
    download(`pocket-draft-${Date.now()}.json`, JSON.stringify(payload, null, 2));
  };

  return (
    <>
      <button onClick={() => copy(deckCode, 'code')}>{copied === 'code' ? 'Copied!' : 'Copy code'}</button>
      <button onClick={() => copy(shareLink, 'link')}>{copied === 'link' ? 'Copied!' : 'Copy link'}</button>
      <button onClick={exportJson}>Export JSON</button>
    </>
  );
}
