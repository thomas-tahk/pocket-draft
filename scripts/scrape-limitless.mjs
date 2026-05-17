#!/usr/bin/env node
// Scrape Pokemon TCG Pocket card data from pocket.limitlesstcg.com.
// Output: public/data/cards.json + public/data/sets.json (served directly by Vite).
// Politeness: 250ms delay between requests, custom UA, sequential.
// Idempotent: re-runs fully overwrite output.
//
// Known limitations (deferred to v0.2 when pack composition matters):
// - Multi-pack shared cards (e.g. A1 Eevee in Charizard/Mewtwo/Pikachu packs)
//   have pack="" because Limitless only lists them on the set index, not on
//   individual detail pages.
// - Single-pack mini-sets (A1a, A2a, B1a, A3a/b, A4a, B2a/b, B3, etc.) have
//   sets.packs=[] because Limitless never writes "X pack" when the set has
//   exactly one pack.

import { load as cheerioLoad } from 'cheerio';
import { mkdir, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const BASE = 'https://pocket.limitlesstcg.com';
const UA = 'Mozilla/5.0 (compatible; pocket-draft-scraper/0.1; +https://github.com/thomas-tahk/pocket-draft)';
const DELAY_MS = 250;
const __dirname = dirname(fileURLToPath(import.meta.url));
const OUT_DIR = resolve(__dirname, '..', 'public', 'data');

const ENERGY_TYPES = new Set([
  'Grass','Fire','Water','Lightning','Psychic','Fighting','Darkness','Metal','Dragon','Colorless',
]);

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function fetchHtml(url, attempt = 0) {
  const res = await fetch(url, { headers: { 'User-Agent': UA, 'Accept-Language': 'en-US,en;q=0.9' } });
  if (!res.ok) {
    if (res.status >= 500 && attempt < 2) {
      await sleep(1000 * (attempt + 1));
      return fetchHtml(url, attempt + 1);
    }
    throw new Error(`${url} → ${res.status}`);
  }
  return res.text();
}

function detectExKind(name) {
  if (name.startsWith('Mega ') && name.endsWith(' ex')) return 'mega-ex';
  if (name.endsWith(' ex')) return 'ex';
  return 'regular';
}

function parseTitleLine(titleText) {
  // Format examples:
  //   "Bulbasaur  - Grass  - 70 HP"
  //   "Mega Pinsir ex - Grass - 170 HP"
  //   "Beast Wall" (Trainer; no extra)
  // We've already grabbed name separately, so titleText here is the *raw* of `.card-text-title`
  // minus the anchor. Just split on dashes and look for an energy type and an "X HP".
  const parts = titleText
    .replace(/\s+/g, ' ')
    .split('-')
    .map((s) => s.trim())
    .filter(Boolean);
  let cardType = 'Trainer';
  let hp = null;
  for (const part of parts) {
    const m = part.match(/^(\d+)\s*HP$/i);
    if (m) {
      hp = Number(m[1]);
      continue;
    }
    if (ENERGY_TYPES.has(part)) {
      cardType = part;
    }
  }
  return { cardType, hp };
}

function parseTypeLine(typeText) {
  // Pokemon example:
  //   "Pokémon - Basic"
  //   "Pokémon - Stage 2 - Evolves from Charmeleon"
  // Trainer example:
  //   "Trainer - Item"
  //   "Trainer - Supporter"
  const clean = typeText.replace(/\s+/g, ' ').trim();
  const parts = clean.split(/\s+-\s+/).map((s) => s.trim()).filter(Boolean);
  const head = parts[0] ?? '';
  let stage = null;
  let trainerKind = null;
  let evolvesFrom = null;
  if (head === 'Trainer') {
    if (parts[1] === 'Item' || parts[1] === 'Supporter') trainerKind = parts[1];
  } else {
    // Pokémon
    if (parts[1] === 'Basic' || parts[1] === 'Stage 1' || parts[1] === 'Stage 2') {
      stage = parts[1];
    }
    const ef = parts.find((p) => p.startsWith('Evolves from'));
    if (ef) evolvesFrom = ef.replace(/^Evolves from\s+/, '').trim();
  }
  return { stage, trainerKind, evolvesFrom };
}

function parseAttack($, el) {
  const $el = $(el);
  const infoP = $el.find('.card-text-attack-info').first();
  const cost = infoP.find('.ptcg-symbol').first().text().trim();
  // Strip the cost span to get "Name Damage"
  const infoClone = infoP.clone();
  infoClone.find('.ptcg-symbol').remove();
  const rest = infoClone.text().replace(/\s+/g, ' ').trim();
  // rest may be "Vine Whip 40" or "Critical Scissors 80+" or "Psychic" (no damage)
  // Damage = trailing token if numeric (optionally with + / x)
  const m = rest.match(/^(.*?)\s+(\d+[+x]?)$/);
  let name = rest;
  let damage = '';
  if (m) {
    name = m[1].trim();
    damage = m[2];
  }
  const effect = $el.find('.card-text-attack-effect').first().text().replace(/\s+/g, ' ').trim();
  return { cost, name, damage, effect };
}

function parseAbility($, el) {
  const $el = $(el);
  const infoText = $el.find('.card-text-ability-info').first().text();
  const name = infoText.replace(/Ability:\s*/i, '').replace(/\s+/g, ' ').trim();
  const effect = $el.find('.card-text-ability-effect').first().text().replace(/\s+/g, ' ').trim();
  return { name, effect };
}

function parseWrr(text) {
  // "Weakness: Water Retreat: 2"
  const w = text.match(/Weakness:\s*([A-Za-z]+)/);
  const r = text.match(/Retreat:\s*(\d+)/);
  let weakness = null;
  if (w) {
    const t = w[1];
    if (ENERGY_TYPES.has(t)) weakness = t;
    else if (t === 'none' || t === 'None' || t === '-' || t === '—') weakness = null;
  }
  const retreat = r ? Number(r[1]) : null;
  return { weakness, retreat };
}

const RARITY_SYMBOLS = new Set(['◊','◊◊','◊◊◊','◊◊◊◊','☆','☆☆','☆☆☆','♕']);

function parseRarity(rarityChunk, setId) {
  // Pack-set example:   "#36 · ◊◊◊◊ · Mewtwo pack"
  // Promo from pack:    "#14 · Promo pack"
  // Promo from shop:    "#1 · Shop"
  // Mid-dot is U+00B7.
  const text = rarityChunk.replace(/\s+/g, ' ').trim();
  const parts = text.split(/\s*·\s*/).map((s) => s.trim()).filter(Boolean);
  let rarity = null;
  let pack = '';
  for (const part of parts) {
    if (/^#\d+$/.test(part)) continue;
    if (RARITY_SYMBOLS.has(part)) {
      rarity = part;
      continue;
    }
    if (part === 'Promo') {
      rarity = 'Promo';
      continue;
    }
    // "Crown Rare" is the chase-tier rarity (rendered as text rather than a
    // glyph on the detail page). Map to the existing crown symbol — these are
    // cosmetic reprints of base-art cards and will be filtered out of the draft pool.
    if (part === 'Crown Rare') {
      rarity = '♕';
      continue;
    }
    // Otherwise treat as pack hint. Strip a trailing " pack" if present.
    pack = part.replace(/\s*pack$/i, '').trim();
  }
  // Promo sets don't print a rarity symbol; mark them explicitly.
  if (rarity === null && setId.startsWith('P-')) rarity = 'Promo';
  return { rarity, pack };
}

async function fetchSetList() {
  const html = await fetchHtml(`${BASE}/cards`);
  const $ = cheerioLoad(html);
  // First-column anchors in the set table:
  //   <a href="/cards/A1"><span class="set-icon">…</span> Genetic Apex <span class="code annotation">A1</span></a>
  // The visible name is the text between the icon span and the code-annotation span.
  const seen = new Set();
  const sets = [];
  $('a').each((_, a) => {
    const href = $(a).attr('href') ?? '';
    const m = href.match(/^\/cards\/([A-Za-z0-9-]+)\/?$/);
    if (!m) return;
    const id = m[1];
    if (seen.has(id)) return;
    const $a = $(a);
    // Skip non-name anchors (date cell, card-count cell).
    if ($a.find('.set-icon').length === 0) return;
    const clone = $a.clone();
    clone.find('.set-icon, .code.annotation, .code, .annotation').remove();
    const name = clone.text().replace(/\s+/g, ' ').trim() || id;
    seen.add(id);
    sets.push({ id, name });
  });
  return sets;
}

async function fetchSetCards(setId) {
  const html = await fetchHtml(`${BASE}/cards/${setId}`);
  const $ = cheerioLoad(html);
  const items = [];
  const seen = new Set();
  $('a').each((_, a) => {
    const href = $(a).attr('href') ?? '';
    const m = href.match(new RegExp(`^/cards/${setId}/(\\d+)/?$`));
    if (!m) return;
    const num = Number(m[1]);
    if (seen.has(num)) return;
    seen.add(num);
    items.push(num);
  });
  items.sort((a, b) => a - b);
  return items;
}

async function fetchCardDetail(setId, num) {
  const url = `${BASE}/cards/${setId}/${num}`;
  const html = await fetchHtml(url);
  const $ = cheerioLoad(html);

  const titleP = $('.card-text-title').first();
  const name = titleP.find('.card-text-name a').text().trim();
  const titleRest = titleP.clone();
  titleRest.find('.card-text-name').remove();
  const titleRestText = titleRest.text();
  const { cardType, hp } = parseTitleLine(titleRestText);

  const typeText = $('.card-text-type').first().text();
  const { stage, trainerKind, evolvesFrom } = parseTypeLine(typeText);

  const exKind = detectExKind(name);

  const attacks = $('.card-text-attack')
    .map((_, el) => parseAttack($, el))
    .get();

  const abilityEl = $('.card-text-ability').first();
  const ability = abilityEl.length ? parseAbility($, abilityEl) : null;

  const wrrText = $('.card-text-wrr').not('.card-text-mega-rule').first().text();
  const { weakness, retreat } = parseWrr(wrrText);

  const artist = $('.card-text-artist a').first().text().trim() || null;
  const flavor = $('.card-text-flavor').first().text().replace(/\s+/g, ' ').trim() || null;

  // Print + rarity + pack
  const currentDetails = $('.card-prints-current .prints-current-details').first();
  // Skip the .text-lg (set name); the *other* span has the rarity+pack chunk.
  let rarityRaw = '';
  currentDetails.find('span').each((_, el) => {
    const $el = $(el);
    if ($el.hasClass('text-lg')) return;
    rarityRaw += ' ' + $el.text();
  });
  const { rarity, pack } = parseRarity(rarityRaw, setId);

  // Images
  const imgEl = $('.card-image img.card').first();
  const imageThumb = imgEl.attr('src') ?? '';
  const imageFull = imgEl.attr('data-src') ?? '';

  const id = `${setId}-${String(num).padStart(3, '0')}`;

  return {
    id,
    setId,
    number: num,
    name,
    rarity,
    exKind,
    cardType,
    trainerKind,
    hp,
    stage,
    evolvesFrom,
    attacks,
    ability,
    weakness,
    retreat,
    pack,
    imageThumb,
    imageFull,
    artist,
    flavor,
  };
}

async function main() {
  const targetSets = process.argv.slice(2); // optional whitelist
  await mkdir(OUT_DIR, { recursive: true });

  console.log('▸ Fetching set index…');
  const sets = await fetchSetList();
  const filtered = targetSets.length ? sets.filter((s) => targetSets.includes(s.id)) : sets;
  console.log(`  ${filtered.length} sets to scrape: ${filtered.map((s) => s.id).join(', ')}`);

  const allCards = [];
  const setSummaries = [];

  for (const set of filtered) {
    process.stdout.write(`▸ ${set.id} (${set.name}) – fetching card list… `);
    await sleep(DELAY_MS);
    const nums = await fetchSetCards(set.id);
    console.log(`${nums.length} cards`);

    const packs = new Set();
    for (const num of nums) {
      await sleep(DELAY_MS);
      try {
        const card = await fetchCardDetail(set.id, num);
        if (card.pack) packs.add(card.pack);
        allCards.push(card);
        process.stdout.write('.');
      } catch (err) {
        process.stdout.write('!');
        console.error(`\n  failed ${set.id}/${num}: ${err.message}`);
      }
    }
    process.stdout.write('\n');
    setSummaries.push({
      id: set.id,
      name: set.name,
      packs: [...packs].sort(),
      cardCount: nums.length,
    });
  }

  const cardsPath = resolve(OUT_DIR, 'cards.json');
  const setsPath = resolve(OUT_DIR, 'sets.json');
  await writeFile(cardsPath, JSON.stringify(allCards, null, 2));
  await writeFile(setsPath, JSON.stringify(setSummaries, null, 2));
  console.log(`\n✓ Wrote ${allCards.length} cards → ${cardsPath}`);
  console.log(`✓ Wrote ${setSummaries.length} sets → ${setsPath}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
