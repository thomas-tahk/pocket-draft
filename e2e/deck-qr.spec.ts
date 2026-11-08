import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { decodeDeckQr, deckList, encodeDeckQr, deckQrPayload, TRAINER_OFFSET } from '../src/share/deckQr';
import type { Card } from '../src/types/card';

const read = <T>(f: string) => JSON.parse(readFileSync(`public/data/${f}`, 'utf8')) as T;
const nrs = read<Record<string, number>>('deck-nrs.json');
const cards = read<Card[]>('cards.json');

// A real Limitless "Open as Image" QR (Expansion 1, cards 1-20, Grass), scanned in-game
// and published with ptcgp-deckcode's tests. Order in the payload is irrelevant.
const REAL_A1_GRASS = 'ABQAAFAAAFoAAGQAADIAADwAAEYAAAoAABQAAB4AACgAAKAAAKoAALQAAL4AAMgAAG4AAHgAAIIAAIwAAJYBAQ==';

test('encoder reproduces a real scanned code for A1 #1-20 + Grass', () => {
  const a1 = Array.from({ length: 20 }, (_, i) => nrs[`A1-${String(i + 1).padStart(3, '0')}`]);
  expect(a1).toEqual(Array.from({ length: 20 }, (_, i) => i + 1));
  const ours = decodeDeckQr(encodeDeckQr(a1, ['Grass']));
  const real = decodeDeckQr(REAL_A1_GRASS);
  expect(ours.pokemon.sort((a, b) => a - b)).toEqual(real.pokemon.sort((a, b) => a - b));
  expect(ours).toMatchObject({ trainers: [], energy: [1] });
  expect(encodeDeckQr(a1, ['Grass'])).toHaveLength(REAL_A1_GRASS.length);
});

test('segments: Trainers first, Pokémon second, then Energy ids in order', () => {
  const bytes = Array.from(atob(encodeDeckQr([TRAINER_OFFSET + 8, TRAINER_OFFSET + 8, 35], ['Fire', 'Water'])), (c) => c.charCodeAt(0));
  const trainer = [0x98, 0x96, 0xd0]; // (1000008 x 10) = 10000080, 3 bytes big-endian
  const pokemon = [0x00, 0x01, 0x5e]; // 35 x 10 = 350
  expect(bytes).toEqual([2, ...trainer, ...trainer, 1, ...pokemon, 2, 2, 3]);
});

test('card numbers: known cards from several sets keep their identity', () => {
  expect(nrs['A1-001']).toBe(1); // Bulbasaur, Pokémon #1
  expect(nrs['A1-216']).toBe(TRAINER_OFFSET + 8); // Helix Fossil
  const sameNr = (a: string, b: string) => expect(nrs[a]).toBe(nrs[b]); // reprints/alt-arts share a number
  sameNr('A1-004', 'A1-251'); // Venusaur ex and its Special Art
  // One name <-> one number, Trainer numbers only on Trainers: catches a bad per-set mapping.
  const byNr = new Map<number, Set<string>>();
  for (const c of cards) {
    expect(nrs[c.id], c.id).toBeGreaterThan(0);
    expect(nrs[c.id] >= TRAINER_OFFSET, c.id).toBe(c.cardType === 'Trainer');
    byNr.set(nrs[c.id], (byNr.get(nrs[c.id]) ?? new Set()).add(c.name.replace(/’/g, "'")));
  }
  expect([...byNr].filter(([, names]) => names.size > 1)).toEqual([]);
});

test('a deck with an unmapped card gives no QR rather than a wrong one', () => {
  expect(deckQrPayload({ 'A1-001': 2, 'ZZ-999': 1 }, nrs, ['Grass'])).toBeNull();
  expect(deckQrPayload({ 'A1-001': 2 }, nrs, ['Grass'])).not.toBeNull();
});

test('Copy list uses the Limitless text format', () => {
  const map = new Map(cards.map((c) => [c.id, c]));
  const f = cards.find((c) => c.name === 'Frigibax' && c.setId === 'B2a')!;
  const text = deckList({ [f.id]: 2, 'A1-216': 1 }, map, ['Water']);
  expect(text).toBe(`2 Frigibax B2a ${f.number}\n1 Helix Fossil A1 216\n\nEnergy: Water\n`);
});
