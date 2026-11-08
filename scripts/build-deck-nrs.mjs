#!/usr/bin/env node
// Build public/data/deck-nrs.json: card id -> the number Pokémon TCG Pocket's
// deck-share QR uses for it ("deckBuilderNr").
//
// Source: cards.json from github.com/flibustier/pokemon-tcg-pocket-database
// (dist/cards.json). The number is in each card's image filename:
//   cPK_10_000010_... -> 1 (Pokémon #1); cTR_10_000080_... -> 1000008 (Trainer #8)
// Format notes: github.com/Nirostar/ptcgp-deck-qr (MIT).
//
// Usage: node scripts/build-deck-nrs.mjs <path to flibustier dist/cards.json>
// Re-run after `npm run scrape` brings in a new set; a unit test fails if any
// card in cards.json has no number.
import { readFile, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const TRAINER_OFFSET = 1_000_000;
const SET_ALIASES = { 'P-A': 'PROMO-A', 'P-B': 'PROMO-B', 'P-C': 'PROMO-C' };
const out = resolve(dirname(fileURLToPath(import.meta.url)), '..', 'public', 'data', 'deck-nrs.json');

function nrFromImage(image) {
  const m = /^c([A-Z]+)_\d+_(\d{6})_/.exec(image ?? '');
  if (!m || Number(m[2]) % 10 !== 0) return null;
  const nr = Number(m[2]) / 10;
  return m[1] === 'TR' ? TRAINER_OFFSET + nr : nr;
}

const src = process.argv[2];
if (!src) throw new Error('usage: build-deck-nrs.mjs <flibustier dist/cards.json>');
const theirs = JSON.parse(await readFile(src, 'utf8'));
const ours = JSON.parse(await readFile(resolve(dirname(out), 'cards.json'), 'utf8'));
const bySetNumber = new Map(theirs.map((c) => [`${c.set}/${c.number}`, c]));

const result = {};
const missing = [];
for (const c of ours) {
  const t = bySetNumber.get(`${SET_ALIASES[c.setId] ?? c.setId}/${c.number}`);
  const nr = t && nrFromImage(t.image);
  if (nr) result[c.id] = nr;
  else missing.push(c.id);
}
await writeFile(out, JSON.stringify(result) + '\n');
console.log(`${Object.keys(result).length} cards mapped, ${missing.length} missing${missing.length ? ': ' + missing.join(' ') : ''}`);
