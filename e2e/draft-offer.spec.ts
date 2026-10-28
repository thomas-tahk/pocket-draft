import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { idOf } from './helpers/draft';
import { generateOffer } from '../src/draft/offer';
import { indexByTier } from '../src/draft/tiers';
import {
  buildSynergy,
  setSupportTypes,
  weightOf,
  SYNERGY_MULT,
  TYPE_LEAN_CAP,
  TYPE_SUPPORT_COEF,
  TYPE_WEIGHT_COEF,
} from '../src/draft/weights';
import type { Card } from '../src/types/card';

// The TS offer generator (what the SPA's manual draft uses), over 20 seeded
// drafts of a synthetic pool. Mirrors the Go tests in server/draft_offer_test.go.

const TYPES = ['Fire', 'Water', 'Grass', 'Lightning'] as const;
const RARITIES = ['◊◊◊◊', '◊◊◊', '◊◊', '◊'] as const;

const pool: Card[] = TYPES.flatMap((type) =>
  RARITIES.flatMap((draftRarity) =>
    [0, 1, 2, 3, 4, 5].map(
      (i) => ({ id: `${type}-${draftRarity}-${i}`, name: `${type}-${draftRarity}-${i}`, cardType: type, draftRarity, isEx: draftRarity === '◊◊◊◊' }) as Card,
    ),
  ),
);
const byTier = indexByTier(pool);

// mulberry32: small seeded rng.
function seeded(seed: number): () => number {
  let a = seed;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function fireShare(picks: Card[]): number {
  let onType = 0;
  let total = 0;
  for (let seed = 1; seed <= 20; seed++) {
    const rng = seeded(seed);
    for (let i = 0; i < 10; i++) {
      for (const c of generateOffer(byTier, { isPack1: false, picks, rng })) {
        total++;
        if (c.cardType === 'Fire') onType++;
      }
    }
  }
  return onType / total;
}

test('pack 1 always offers an ex over 20 seeded drafts', () => {
  for (let seed = 1; seed <= 20; seed++) {
    const offer = generateOffer(byTier, { isPack1: true, picks: [], rng: seeded(seed) });
    expect(offer.some((c) => c.isEx), `seed ${seed}`).toBe(true);
  }
});

test('later offers favour the types already picked', () => {
  const fire = pool.filter((c) => c.cardType === 'Fire').slice(0, 8);
  const weighted = fireShare(fire);
  const baseline = fireShare([]);
  expect(weighted).toBeGreaterThan(baseline * 1.5);
});

// Synergy (#109). Mirrors server/draft_synergy_test.go.
const mon = (name: string, cardType: string) => ({ id: name + cardType, name, cardType, attacks: [] }) as unknown as Card;
const trainer = (name: string, trainerText: string) =>
  ({ id: name + trainerText, name, cardType: 'Trainer', trainerText, attacks: [] }) as unknown as Card;

test('a Trainer naming a picked Pokemon gets the bonus, one that does not does not', () => {
  const syn = buildSynergy([mon('Onix', 'Fighting')]);
  expect(weightOf(trainer('Brock', 'Attach an Energy to Golem or Onix.'), {}, syn)).toBe(SYNERGY_MULT);
  expect(weightOf(trainer('Misty', 'Attach an Energy to your Pokémon.'), {}, syn)).toBe(1);
  expect(weightOf(trainer('Other', 'Attach to Onixia.'), {}, syn)).toBe(1);
  const ex = buildSynergy([mon('Mewtwo ex', 'Psychic')]);
  expect(weightOf(trainer('X', 'Heal your Mewtwo.'), {}, ex)).toBe(SYNERGY_MULT);
});

test('text naming a picked type leans a Trainer or boosts a Pokémon; costs and Colorless do not', () => {
  const picks = [mon('Squirtle', 'Water')];
  const syn = buildSynergy(picks);
  const tally = { Water: 1 };
  for (const text of ['Choose 1 of your [W] Pokémon.', 'Attach a {W} Energy.', 'Heal your Water Pokémon.', 'Discard a Water Energy.'])
    expect(weightOf(trainer('T', text), tally, syn), text).toBe(1 + TYPE_SUPPORT_COEF);
  expect(weightOf(trainer('Blaine', 'Choose 1 of your [R] Pokémon.'), tally, syn)).toBe(1);
  expect(weightOf(trainer('C', 'Attach a [C] Energy.'), tally, syn)).toBe(1);
  const psyduck = { ...mon('Psyduck', 'Psychic'), attacks: [{ cost: '[W]', name: '[W] Splash', damage: '', effect: '' }] } as unknown as Card;
  expect(weightOf(psyduck, tally, syn)).toBe(1);
  const withEffect = { ...psyduck, id: 'psyduck-effect', attacks: [{ cost: 'W', name: 'X', damage: '', effect: 'Take a [W] Energy.' }] } as unknown as Card;
  expect(weightOf(withEffect, tally, syn)).toBe(SYNERGY_MULT);
});

test('a card naming its own type never counts, as pick or candidate', () => {
  const withText = (name: string, cardType: string, effect: string) =>
    ({ ...mon(name, cardType), id: name, attacks: [{ cost: 'W', name: 'X', damage: '', effect }] }) as unknown as Card;
  // 3b: a Water pick whose text says [W] does not boost other Water Pokémon.
  const waterPick = buildSynergy([withText('Vaporeon', 'Water', 'Take a [W] Energy.')]);
  expect(weightOf(mon('Squirtle', 'Water'), {}, waterPick)).toBe(1);
  const firePick = buildSynergy([withText('Flareon', 'Fire', 'Take a [W] Energy.')]);
  expect(weightOf(mon('Squirtle', 'Water'), {}, firePick)).toBe(SYNERGY_MULT);
  // 3a: a Water candidate whose own text says [W] is not boosted by Water picks.
  const candidate = withText('Kyogre', 'Water', 'Take a [W] Energy.');
  expect(weightOf(candidate, {}, buildSynergy([mon('Squirtle', 'Water')]))).toBe(1);
});

test('a [N] Trainer gets the bonus after a Dragon pick', () => {
  const syn = buildSynergy([mon('Dratini', 'Dragon')]);
  const tally = { Dragon: 1 };
  expect(weightOf(trainer('Lance', 'Choose 1 of your [N] Pokémon.'), tally, syn)).toBe(1 + TYPE_SUPPORT_COEF);
  expect(weightOf(trainer('Blaine', 'Choose 1 of your [R] Pokémon.'), tally, syn)).toBe(1);
});

test('a pick whose text names a type boosts candidates of that type, once', () => {
  const syn = buildSynergy([trainer('Misty', 'Choose 1 of your [W] Pokémon.')]);
  expect(weightOf(mon('Squirtle', 'Water'), {}, syn)).toBe(SYNERGY_MULT);
  expect(weightOf(mon('Charmander', 'Fire'), {}, syn)).toBe(1);
  expect(weightOf(trainer('Other', 'Draw a card.'), {}, syn)).toBe(1);
  const colorless = buildSynergy([trainer('Poké Ball', 'Put a [C] Pokémon in hand.')]);
  expect(weightOf(mon('Meowth', 'Colorless'), {}, colorless)).toBe(1);
  const two = buildSynergy([mon('Squirtle', 'Water')]);
  expect(weightOf(trainer('X', 'Heal your Squirtle and your [W] Pokémon.'), {}, two)).toBe(SYNERGY_MULT);
});

// A pick is secured for the whole draft (#124). Mirrors
// TestBotDraftNeverReoffersAPickedName in server/draft_no_repeat_test.go.
test('generateOffer never offers a picked name, other printings included', () => {
  const picks = pool.slice(0, 20).map((c, i) => ({ ...c, id: `reprint-${i}` }) as Card);
  const names = new Set(picks.map((c) => c.name));
  for (let seed = 1; seed <= 20; seed++) {
    for (const c of generateOffer(byTier, { isPack1: false, picks, rng: seeded(seed) })) {
      expect(names.has(c.name), `seed ${seed}: ${c.name}`).toBe(false);
    }
  }
});

test('a full 20-pick draft in the UI never shows an already-picked name', async ({ page }) => {
  const raw = JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as { id: string; name: string }[];
  const nameById = new Map(raw.map((c) => [c.id, c.name]));
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  const pickedNames = new Set<string>();
  for (let pack = 1; pack <= 20; pack++) {
    await expect(page.getByText(`Pack ${pack} / 20`)).toBeVisible();
    const offer = page.locator('main button[title]');
    await expect(offer).toHaveCount(5);
    const ids = (await offer.evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''))).map(idOf);
    const names = ids.map((id) => nameById.get(id)!);
    for (const name of names) expect(pickedNames.has(name), `pack ${pack} offered ${name}`).toBe(false);
    // Always take the first card: nothing here steers around repeats.
    pickedNames.add(names[0]);
    await offer.nth(0).click();
  }
});

test('Trainers naming the picked type are offered measurably more often', () => {
  const mistys = [0, 1, 2].map(
    (i) => ({ ...trainer('Misty', 'Choose 1 of your [W] Pokémon.'), id: `m${i}`, draftRarity: '◊', isEx: false }) as unknown as Card,
  );
  const tiers = indexByTier([...pool, ...mistys]);
  const count = (picks: Card[]) => {
    let n = 0;
    for (let seed = 1; seed <= 20; seed++) {
      const rng = seeded(seed);
      for (let i = 0; i < 100; i++)
        for (const c of generateOffer(tiers, { isPack1: false, picks, rng })) if (c.name === 'Misty') n++;
    }
    return n;
  };
  const withWater = count([mon('W', 'Water')]);
  const withFire = count([mon('F', 'Fire')]);
  expect(withFire).toBeGreaterThan(0);
  expect(withWater).toBeGreaterThan(withFire * 1.5);
});

// Type lean (#125). Mirrors server/draft_type_lean_test.go.
test('type support Trainers lean with the picks, by text, by Go’s answer or by the table; opposing ones do not', () => {
  // Marlon's text names Pokémon, not a type: only Go (which looks the Pokémon up
  // in cards.json) can say Water / Psychic, through draft-pool.json.
  const marlon = trainer('Marlon', 'Heal 70 damage from 1 of your Carracosta or Jellicent.');
  setSupportTypes(null);
  expect(weightOf(marlon, { Water: 3 }, buildSynergy([]))).toBe(1);
  setSupportTypes({ Marlon: ['Psychic', 'Water'] });
  try {
    expect(weightOf(marlon, { Water: 3 }, buildSynergy([]))).toBe(1 + TYPE_SUPPORT_COEF * 3);
    expect(weightOf(marlon, { Fire: 3 }, buildSynergy([]))).toBe(1);
  } finally {
    setSupportTypes(null);
  }
  // The table is the fallback for text naming neither a type nor a Pokémon.
  const drayden = trainer('Drayden', 'During this turn, 1 of your opponent’s Pokémon is chosen 1 more time for the Draco Meteor attack.');
  expect(weightOf(drayden, { Dragon: 3 }, buildSynergy([]))).toBe(1 + TYPE_SUPPORT_COEF * 3);
  const bottle = trainer('Squirt Bottle', "Discard a [R] Energy from your opponent's Active Pokémon.");
  expect(weightOf(bottle, { Fire: 3 }, buildSynergy([]))).toBe(1);
  const misty = trainer('Misty', 'Choose 1 of your [W] Pokémon.');
  expect(weightOf(misty, { Water: 3 }, buildSynergy([]))).toBe(1 + TYPE_SUPPORT_COEF * 3);
  expect(weightOf(mon('Squirtle', 'Water'), { Water: 3 }, buildSynergy([]))).toBe(1 + TYPE_WEIGHT_COEF * 3);
});

// Mirrors server/draft_type_lean_test.go's TestTypeLeanIsCapped.
test('the type lean keeps its strength at 2-3 picks and stops growing at the cap', () => {
  const squirtle = mon('Squirtle', 'Water');
  const misty = trainer('Misty', 'Choose 1 of your [W] Pokémon.');
  for (const c of [squirtle, misty]) {
    const w = (n: number) => weightOf(c, { Water: n }, buildSynergy([]));
    expect(w(2)).toBe(5);
    expect(w(3)).toBe(7);
    expect(w(TYPE_LEAN_CAP)).toBe(1 + 2 * TYPE_LEAN_CAP);
    expect(w(19)).toBe(w(TYPE_LEAN_CAP));
  }
  expect(TYPE_LEAN_CAP).toBe(4);
});

test('3 Water picks lean Water Pokémon but do not force them', () => {
  const water = pool.filter((c) => c.cardType === 'Water').slice(0, 3);
  const share = (picks: Card[], type: string) => {
    let n = 0;
    let total = 0;
    for (let seed = 1; seed <= 20; seed++) {
      const rng = seeded(seed);
      for (let i = 0; i < 50; i++)
        for (const c of generateOffer(byTier, { isPack1: false, picks, rng })) {
          total++;
          if (c.cardType === type) n++;
        }
    }
    return n / total;
  };
  expect(share(water, 'Water')).toBeGreaterThan(share([], 'Water') * 1.5);
  expect(share(water, 'Fire')).toBeGreaterThan(0);
});

// Niche odds (#126). Mirrors server/draft_niche_test.go.
const mark = (c: Card, id: string) => ({ ...c, id, draftRarity: '◊◊', isEx: false }) as unknown as Card;
const withEvolvesFrom = (c: Card, evolvesFrom: string) => ({ ...c, evolvesFrom }) as unknown as Card;

test('a Pokemon named in a picked Trainer gets the bonus, once; Snorlax is not Snorlax ex', () => {
  const nemona = trainer('Nemona', 'During this turn, attacks used by your Pawmot do +80 damage.');
  const syn = buildSynergy([nemona]);
  expect(weightOf(mon('Pawmot', 'Lightning'), {}, syn)).toBe(SYNERGY_MULT);
  expect(weightOf(mon('Pawmotte', 'Lightning'), {}, syn)).toBe(1);
  const snorlaxEx = buildSynergy([trainer('T', 'Heal your Snorlax.')]);
  expect(weightOf(mon('Snorlax ex', 'Colorless'), {}, snorlaxEx)).toBe(SYNERGY_MULT);
  const onlyEx = buildSynergy([trainer('T', 'Heal your Snorlax ex.')]);
  expect(weightOf(mon('Snorlax', 'Colorless'), {}, onlyEx)).toBe(1);
  // Named by the text and in the table together: still one multiplication.
  const sada = buildSynergy([trainer('Professor Sada', 'Heal your Roaring Moon.')]);
  expect(weightOf(mon('Roaring Moon', 'Darkness'), {}, sada)).toBe(SYNERGY_MULT);
});

test('table groups raise Trainer and Pokemon both ways, cross-kind only', () => {
  const moon = buildSynergy([mon('Roaring Moon', 'Darkness')]);
  expect(weightOf(trainer('Professor Sada', 'x'), {}, moon)).toBe(SYNERGY_MULT);
  expect(weightOf(trainer('Ancient Booster Energy Capsule', 'x'), {}, moon)).toBe(SYNERGY_MULT);
  expect(weightOf(trainer('Professor Turo', 'x'), {}, moon)).toBe(1);
  expect(weightOf(mon('Great Tusk', 'Fighting'), {}, moon)).toBe(1);
  const sada = buildSynergy([trainer('Professor Sada', 'x')]);
  expect(weightOf(mon('Roaring Moon', 'Darkness'), {}, sada)).toBe(SYNERGY_MULT);
  expect(weightOf(mon('Aerodactyl ex', 'Fighting'), {}, buildSynergy([trainer('Old Amber', 'x')]))).toBe(SYNERGY_MULT);
  const bag = buildSynergy([trainer('Eevee Bag', 'Heal each Pokémon that evolves from Eevee.')]);
  expect(weightOf(withEvolvesFrom(mon('Vaporeon', 'Water'), 'Eevee'), {}, bag)).toBe(SYNERGY_MULT);
  expect(weightOf(mon('Vaporeon', 'Water'), {}, bag)).toBe(1); // no evolvesFrom, no match
  const vap = buildSynergy([withEvolvesFrom(mon('Vaporeon', 'Water'), 'Eevee')]);
  expect(weightOf(trainer('Eevee Bag', 'x'), {}, vap)).toBe(SYNERGY_MULT);
});

test('picking a Trainer that names a Pokemon, or the Pokemon, raises the other in the offers', () => {
  const pawmots = [0, 1, 2].map((i) => mark(mon('Pawmot', 'Lightning'), `pawmot${i}`));
  const nemonas = [0, 1, 2].map((i) =>
    mark(trainer('Nemona', 'During this turn, attacks used by your Pawmot do +80 damage.'), `nemona${i}`),
  );
  const tiers = indexByTier([...pool, ...pawmots, ...nemonas]);
  const count = (picks: Card[], name: string) => {
    let n = 0;
    for (let seed = 1; seed <= 20; seed++) {
      const rng = seeded(seed);
      for (let i = 0; i < 100; i++)
        for (const c of generateOffer(tiers, { isPack1: false, picks, rng })) if (c.name === name) n++;
    }
    return n;
  };
  const control = mark(trainer('Poké Ball', 'Put 1 random Basic Pokémon from your deck into your hand.'), 'ctl');
  const lightning = pool.find((c) => c.cardType === 'Lightning')!;
  const afterTrainer = count([nemonas[0]], 'Pawmot');
  const base = count([control], 'Pawmot');
  expect(base).toBeGreaterThan(0);
  expect(afterTrainer).toBeGreaterThan(base * 2);
  const afterPokemon = count([pawmots[0]], 'Nemona');
  const baseTrainer = count([lightning], 'Nemona');
  expect(baseTrainer).toBeGreaterThan(0);
  expect(afterPokemon).toBeGreaterThan(baseTrainer * 2);
});

test('Alolan Marowak is not Marowak, Eevee Bag names no Eevee ex, and " ex" reads the same both ways', () => {
  const kiawe = trainer('Kiawe', 'Choose 1 of your Alolan Marowak or Turtonator. Attach 2 [R] Energy.');
  const afterKiawe = buildSynergy([kiawe]);
  expect(weightOf(mon('Marowak', 'Fighting'), {}, afterKiawe)).toBe(1);
  expect(weightOf(mon('Alolan Marowak', 'Fire'), {}, afterKiawe)).toBe(SYNERGY_MULT);
  expect(weightOf(kiawe, {}, buildSynergy([mon('Marowak', 'Fighting')]))).toBe(weightOf(kiawe, {}, buildSynergy([])));
  expect(weightOf(kiawe, {}, buildSynergy([mon('Alolan Marowak', 'Fire')]))).toBe(SYNERGY_MULT);

  const bag = buildSynergy([trainer('Eevee Bag', 'Heal each Pokémon that evolves from Eevee.')]);
  expect(weightOf(mon('Eevee ex', 'Colorless'), {}, bag)).toBe(1);

  const snorlax = mon('Snorlax', 'Colorless');
  const snorlaxEx = mon('Snorlax ex', 'Colorless');
  const onlyEx = trainer('A', 'Heal your Snorlax ex.');
  const plain = trainer('B', 'Heal your Snorlax.');
  const w = (c: Card, ...picks: Card[]) => weightOf(c, {}, buildSynergy(picks));
  expect(w(snorlax, onlyEx)).toBe(1);
  expect(w(snorlaxEx, onlyEx)).toBe(SYNERGY_MULT);
  expect(w(snorlax, plain)).toBe(SYNERGY_MULT);
  expect(w(snorlaxEx, plain)).toBe(SYNERGY_MULT);
  expect(w(onlyEx, snorlax)).toBe(1);
  expect(w(onlyEx, snorlaxEx)).toBe(SYNERGY_MULT);
  expect(w(plain, snorlax)).toBe(SYNERGY_MULT);
  expect(w(plain, snorlaxEx)).toBe(SYNERGY_MULT);

  // Raging Bolt is Ancient (the engine's list), not Future.
  const turo = buildSynergy([trainer('Professor Turo', 'x')]);
  const sada = buildSynergy([trainer('Professor Sada', 'x')]);
  expect(weightOf(mon('Raging Bolt', 'Dragon'), {}, sada)).toBe(SYNERGY_MULT);
  expect(weightOf(mon('Raging Bolt', 'Dragon'), {}, turo)).toBe(1);
});

// Owner 2026-10-28, decision 3. Mirrors server/draft_niche_test.go.
test('Ancient / Future only when the card says so: not the B2a Koraidon or Miraidon', () => {
  const turo = buildSynergy([trainer('Professor Turo', 'x')]);
  const sada = buildSynergy([trainer('Professor Sada', 'x')]);
  expect(weightOf(mon('Koraidon', 'Fighting'), {}, sada)).toBe(1);
  expect(weightOf(mon('Miraidon', 'Lightning'), {}, turo)).toBe(1);
  expect(weightOf(mon('Koraidon ex', 'Fighting'), {}, sada)).toBe(SYNERGY_MULT);
  expect(weightOf(mon('Miraidon ex', 'Lightning'), {}, turo)).toBe(SYNERGY_MULT);
  expect(weightOf(mon('Flutter Mane ex', 'Psychic'), {}, sada)).toBe(SYNERGY_MULT);
  expect(weightOf(trainer('Professor Sada', 'x'), {}, buildSynergy([mon('Koraidon', 'Fighting')]))).toBe(1);
});

// Owner 2026-10-28, decision 4. Mirrors server/draft_eevee_test.go.
test('Eevee ex and the Eeveelutions raise each other, both ways, never stacked', () => {
  const eevee = mon('Eevee ex', 'Colorless');
  const vaporeon = withEvolvesFrom(mon('Vaporeon', 'Water'), 'Eevee');
  const bag = trainer('Eevee Bag', 'Heal each Pokémon that evolves from Eevee.');
  const w = (c: Card, ...picks: Card[]) => weightOf(c, {}, buildSynergy(picks));
  expect(w(eevee, vaporeon)).toBe(3);
  expect(w(vaporeon, eevee)).toBe(3);
  expect(w(mon('Vaporeon', 'Water'), eevee)).toBe(1); // no evolvesFrom
  expect(w(mon('Snorlax', 'Colorless'), eevee)).toBe(1);
  expect(w(vaporeon, eevee, bag)).toBe(SYNERGY_MULT); // one multiplication
});
