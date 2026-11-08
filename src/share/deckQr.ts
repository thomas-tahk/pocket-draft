// The deck-share QR Pokémon TCG Pocket scans (Scan Code), and the Limitless text list.
//
// Payload = base64 of [Trainer segment][Pokémon segment][Energy block].
//   Segment: 1-byte count N, then N x 3-byte big-endian (card number x 10), one entry per copy.
//   Energy:  1-byte count, then 1 byte per Energy Zone type.
// Format reverse-engineered by github.com/Nirostar/ptcgp-deck-qr (MIT); our own encoder.
// A card's number comes from public/data/deck-nrs.json (scripts/build-deck-nrs.mjs).

import type { Card } from '../types/card';

export const TRAINER_OFFSET = 1_000_000;
export type DeckNrs = Record<string, number>; // card id -> deck-share number

const ENERGY_ID: Record<string, number> = {
  Grass: 1, Fire: 2, Water: 3, Lightning: 4, Psychic: 5, Fighting: 6, Darkness: 7, Metal: 8,
};

function segment(nrs: number[]): number[] {
  const bytes = [nrs.length];
  for (const nr of [...nrs].sort((a, b) => a - b)) {
    const v = nr * 10;
    bytes.push((v >> 16) & 0xff, (v >> 8) & 0xff, v & 0xff);
  }
  return bytes;
}

// card numbers (one per copy) + Energy Zone type names -> the base64 the game scans.
export function encodeDeckQr(cardNrs: number[], energy: string[]): string {
  const energyIds = energy.map((e) => ENERGY_ID[e]).filter((id): id is number => id !== undefined);
  const bytes = [
    ...segment(cardNrs.filter((n) => n >= TRAINER_OFFSET)),
    ...segment(cardNrs.filter((n) => n < TRAINER_OFFSET)),
    energyIds.length,
    ...energyIds,
  ];
  return btoa(String.fromCharCode(...bytes));
}

// Inverse of encodeDeckQr (tests, and a sanity check on what we show).
export function decodeDeckQr(code: string): { trainers: number[]; pokemon: number[]; energy: number[] } {
  const raw = Array.from(atob(code), (c) => c.charCodeAt(0));
  let i = 0;
  const read = () => {
    const n = raw[i++];
    const out: number[] = [];
    for (let k = 0; k < n; k++, i += 3) out.push(((raw[i] << 16) | (raw[i + 1] << 8) | raw[i + 2]) / 10);
    return out;
  };
  const trainers = read();
  const pokemon = read();
  const count = raw[i++] ?? 0;
  return { trainers, pokemon, energy: raw.slice(i, i + count) };
}

// The QR payload for a deck, or null if any card has no known number (never a wrong deck).
export function deckQrPayload(deck: Record<string, number>, nrs: DeckNrs, energy: string[]): string | null {
  const cardNrs: number[] = [];
  for (const [id, count] of Object.entries(deck)) {
    const nr = nrs[id];
    if (!nr) return null;
    for (let k = 0; k < count; k++) cardNrs.push(nr);
  }
  return encodeDeckQr(cardNrs, energy);
}

// Limitless text list: "2 Frigibax B2a 34" lines, a blank line, then "Energy: Water".
export function deckList(deck: Record<string, number>, cards: Map<string, Card>, energy: string[]): string {
  const lines = Object.entries(deck)
    .filter(([id, n]) => n > 0 && cards.has(id))
    .map(([id, n]) => {
      const c = cards.get(id)!;
      return { trainer: c.cardType === 'Trainer', line: `${n} ${c.name} ${c.setId} ${c.number}` };
    })
    .sort((a, b) => Number(a.trainer) - Number(b.trainer) || (a.line < b.line ? -1 : 1))
    .map((l) => l.line);
  return `${lines.join('\n')}\n\nEnergy: ${energy.join(', ')}\n`;
}
