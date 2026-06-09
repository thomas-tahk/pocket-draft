// Deck code: a compact, copy-pasteable encoding of a 20-card decklist.
//
// Format:  PKD1.<base64url>
// where the decoded payload is  "<id>*<count>;<id>*<count>;..."  (the "*count"
// is omitted when the count is 1). Card ids (e.g. "A1-001") are stable across
// card-data refreshes, and the leading "PKD1" version tag lets the format
// evolve later without breaking codes already shared.
//
// This module is deliberately free of any card-data dependency so the future
// constructed/simulator mode can reuse it. Validating that ids actually exist
// is the caller's job (the viewer drops ids it doesn't recognise).
// See docs/adr/0003-deck-code-format.md.

export type Decklist = Record<string, number>; // cardId -> count (1..2)

const VERSION = 'PKD1';
const MAX_COPIES = 2;

function toBase64Url(ascii: string): string {
  return btoa(ascii).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function fromBase64Url(s: string): string {
  return atob(s.replace(/-/g, '+').replace(/_/g, '/'));
}

export function encodeDeck(deck: Decklist): string {
  const payload = Object.entries(deck)
    .filter(([, n]) => n > 0)
    .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0)) // stable, order-independent codes
    .map(([id, n]) => (n === 1 ? id : `${id}*${n}`))
    .join(';');
  return `${VERSION}.${toBase64Url(payload)}`;
}

export function decodeDeck(code: string): Decklist {
  const trimmed = code.trim();
  const dot = trimmed.indexOf('.');
  if (dot === -1) throw new Error('Not a deck code');

  const version = trimmed.slice(0, dot);
  if (version !== VERSION) throw new Error(`Unsupported deck code version: ${version}`);

  const payload = fromBase64Url(trimmed.slice(dot + 1));
  const deck: Decklist = {};
  for (const part of payload.split(';')) {
    if (!part) continue;
    const star = part.indexOf('*');
    const id = star === -1 ? part : part.slice(0, star);
    const parsed = star === -1 ? 1 : parseInt(part.slice(star + 1), 10);
    const count = Math.max(1, Math.min(MAX_COPIES, Number.isFinite(parsed) ? parsed : 1));
    if (id) deck[id] = count;
  }
  return deck;
}
