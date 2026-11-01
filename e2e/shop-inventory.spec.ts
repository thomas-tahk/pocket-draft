import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { enrich } from '../src/data/rarity';
import { buildShopInventory } from '../src/shop/inventory';
import type { Card, RawCard } from '../src/types/card';

// Pure unit tests of the shop shelves (no browser, no server).
const pool: Card[] = (JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as RawCard[])
  .map(enrich)
  .filter((c): c is Card => c !== null);
const byName = (n: string) => pool.find((c) => c.name === n)!;
const pokemon = pool.filter((c) => c.cardType !== 'Trainer');

test.describe('shop shelves', () => {
  test('Rare Candy is a staple for any draft', () => {
    for (const picks of [[], pokemon.slice(0, 20), pokemon.filter((c) => c.isEx).slice(0, 5)]) {
      const names = buildShopInventory(picks, pool).staples.map((c) => c.name);
      expect(names).toContain('Rare Candy');
    }
  });

  test('Rare Candy is left out once drafted', () => {
    const names = buildShopInventory([byName('Rare Candy')], pool).staples.map((c) => c.name);
    expect(names).not.toContain('Rare Candy');
  });
});
