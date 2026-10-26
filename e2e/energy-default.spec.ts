import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';

// The deckbuild picker's default (src/game/energy.ts) must equal the engine's
// DefaultEnergyTypes (engine/energy.go). Same real deck through both: the app's function
// on the card data, and the Go server's answer to /api/new with no energy chosen.

type Raw = { id: string; cardType: string };
const cards = JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as Raw[];
const byId = new Map(cards.map((c) => [c.id, c]));

const twice = (ids: string[]) => ids.flatMap((id) => [id, id]);
// Druddigon costs Fire + Water + Colorless; its Dragon type must not hide that.
const working = new Set((JSON.parse(readFileSync('public/data/draft-pool.json', 'utf8')) as { working: string[] }).working);
const psychic = new Map<string, string>(); // name -> id of a playable regular Psychic Basic
for (const c of cards as (Raw & { name: string; stage: string; exKind: string })[]) {
  if (c.cardType === 'Psychic' && c.stage === 'Basic' && c.exKind === 'regular' && working.has(c.id) && !psychic.has(c.name)) psychic.set(c.name, c.id);
}
const DECK = twice(['A1a-056', ...[...psychic.values()].slice(0, 9)]);

test('default Energy Zone types match the engine: Druddigon with Psychic Pokémon', async ({ page, request }) => {
  await page.goto('/');
  const mine = await page.evaluate(async (deck) => {
    const { defaultZoneTypes } = await import('/src/game/energy.ts');
    const all = (await (await fetch('/data/cards.json')).json()) as { id: string }[];
    const index = new Map(all.map((c) => [c.id, c]));
    return defaultZoneTypes(deck.map((id) => index.get(id)) as never);
  }, DECK);

  const res = await request.post('/api/new', { data: { you: DECK, seed: 1 } });
  const view = await res.json();
  expect(view.error, 'server refused the deck').toBeUndefined();
  expect(mine).toEqual(view.players[0].energyTypes);
  // Not what the Pokémon types alone would give: Druddigon is a Dragon, yet Fire and Water count.
  expect(DECK.some((id) => byId.get(id)!.cardType === 'Dragon')).toBe(true);
  expect(mine).toContain('Fire');
  expect(mine).toContain('Water');
});
