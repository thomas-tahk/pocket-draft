import { test, expect, type Locator, type Page } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { draftTwenty, buildFromDraft } from './helpers/draft';

// #121: the owner's accessibility bar on every screen outside the battle board.
//  1. A whole run by keyboard only (Tab / Shift+Tab / Enter / Space): home, 20 picks, shop, deckbuild to 20,
//     Energy Zone, finalize, start a battle. The board is out of scope, so the battle is finished through the
//     game store; the run screen is then driven by keyboard through to the end-of-run summary.
//     Every control that takes focus shows a focus indicator.
//  2. Every button, link, input and textarea on the run, opening, draft, shop, deckbuild, end and history
//     screens is at least 44 x 44 px, at 390x844 and 1440x900.
//
// The card-image CDN is blocked in this sandbox, so card images are answered with a placeholder (as pocket-look.spec.ts).

const IMAGE_HOST = '**://limitlesstcg.nyc3.cdn.digitaloceanspaces.com/**';
const PLACEHOLDER = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 250 350"><rect width="250" height="350" rx="14" fill="#f4f1e6"/><rect x="24" y="48" width="202" height="150" rx="6" fill="#bfe3f2"/></svg>`;

async function stubImages(page: Page) {
  await page.route(IMAGE_HOST, (route) => route.fulfill({ status: 200, contentType: 'image/svg+xml', body: PLACEHOLDER }));
}

type Raw = { id: string; stage: string };
const stageById = new Map((JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as Raw[]).map((c) => [c.id, c.stage]));
const idOf = (title: string) => title.split(' · ').pop() ?? '';

// ---------- keyboard helpers ----------

type FocusFact = { tag: string; name: string; focusVisible: boolean; outline: string; shadow: string };
const seen: FocusFact[] = [];

// What the focused element looks like right now: a ring (non-none outline of real width) or a box-shadow.
const focusFact = (page: Page): Promise<FocusFact> =>
  page.evaluate(() => {
    const el = document.activeElement as HTMLElement;
    const cs = getComputedStyle(el);
    return {
      tag: el.tagName.toLowerCase(),
      name: el.getAttribute('aria-label') ?? el.textContent?.trim().slice(0, 40) ?? '',
      focusVisible: el.matches(':focus-visible'),
      outline: `${cs.outlineStyle} ${cs.outlineWidth}`,
      shadow: cs.boxShadow,
    };
  });

const hasIndicator = (f: FocusFact) => (f.outline !== 'none 0px' && !f.outline.startsWith('none') && !f.outline.endsWith(' 0px')) || f.shadow !== 'none';

// Checks the focused control shows focus, and remembers it for the summary assertion at the end.
async function checkFocus(page: Page) {
  const f = await focusFact(page);
  seen.push(f);
  expect(f.focusVisible, `${f.tag} "${f.name}" matches :focus-visible`).toBe(true);
  expect(hasIndicator(f), `${f.tag} "${f.name}" shows a focus indicator (outline ${f.outline}, shadow ${f.shadow})`).toBe(true);
}

const isFocused = (loc: Locator) => loc.evaluate((el) => el === document.activeElement).catch(() => false);

// Tab (or Shift+Tab) until `target` has focus; every stop on the way is checked for a focus indicator.
async function tabTo(page: Page, target: Locator, opts: { back?: boolean; max?: number } = {}) {
  const key = opts.back ? 'Shift+Tab' : 'Tab';
  for (let i = 0; i < (opts.max ?? 300); i++) {
    if (await isFocused(target)) {
      await checkFocus(page);
      return;
    }
    await page.keyboard.press(key);
    const tag = await page.evaluate(() => document.activeElement?.tagName);
    if (tag && tag !== 'BODY' && tag !== 'H1') await checkFocus(page);
  }
  throw new Error(`never reached ${target}`);
}

const press = async (page: Page, target: Locator, key: 'Enter' | 'Space', opts?: { back?: boolean }) => {
  await tabTo(page, target, opts);
  await page.keyboard.press(key === 'Space' ? ' ' : 'Enter');
};

const announcer = (page: Page) => page.getByTestId('announcer');
const headingFocused = (page: Page, name: string | RegExp) =>
  expect(page.getByRole('heading', { level: 1, name })).toBeFocused();

async function concedeByStore(page: Page) {
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('setup-intro')).toHaveCount(0, { timeout: 10_000 });
  await page.evaluate(async () => {
    const { useGameStore } = await import('/src/game/store.ts');
    await useGameStore.getState().dispatch({ type: 'Concede', player: 0 });
  });
  await expect(page.getByTestId('end-result')).toHaveText('Loss');
}

test('a whole run by keyboard: draft, shop, deckbuild, Energy Zone, battle, summary', async ({ page }) => {
  test.setTimeout(300_000);
  await stubImages(page);
  await page.goto('/');

  // ---- home: focus starts on the heading; Tab reaches the primary action, Shift+Tab goes back ----
  await headingFocused(page, 'Pocket Draft');
  await page.keyboard.press('Tab');
  await expect(page.getByRole('button', { name: 'Start draft' })).toBeFocused();
  await checkFocus(page);
  await page.keyboard.press('Tab');
  await checkFocus(page);
  await page.keyboard.press('Shift+Tab');
  await expect(page.getByRole('button', { name: 'Start draft' })).toBeFocused();
  await page.keyboard.press('Enter');

  // ---- draft: 20 picks by keyboard ----
  await headingFocused(page, 'Pocket Draft');
  const offerBtns = page.locator('.offer-grid button');
  for (let pack = 1; pack <= 20; pack++) {
    await expect(page.getByText(`Pack ${pack} / 20`)).toBeVisible();
    await expect(offerBtns).toHaveCount(5);
    await expect(page.getByRole('group', { name: `Pack ${pack} of 20: choose a card` })).toBeVisible();
    if (pack > 1) {
      // After a pick, focus is on the first card of the new pack: no hunting for it.
      await expect(offerBtns.first()).toBeFocused();
    } else {
      // Pack 1: Abandon draft opens a confirm; Escape closes it and puts focus back on the button.
      await tabTo(page, page.getByRole('button', { name: 'Abandon draft' }), { back: false });
      await page.keyboard.press('Enter');
      const dialog = page.getByRole('alertdialog', { name: 'Abandon draft' });
      await expect(dialog.getByRole('button', { name: 'Cancel' })).toBeFocused();
      await checkFocus(page);
      await page.keyboard.press('Escape');
      await expect(dialog).toHaveCount(0);
      await expect(page.getByRole('button', { name: 'Abandon draft' })).toBeFocused();
      await checkFocus(page);
    }
    const titles = await offerBtns.evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''));
    const ids = titles.map(idOf);
    let want = ids.findIndex((i) => stageById.get(i) === 'Basic');
    if (want < 0) want = 0;
    // Every offer has a name + rarity (what is shown on the card), not an image-only button.
    const label = await offerBtns.nth(want).getAttribute('aria-label');
    expect(label).toMatch(/\S, \S/);
    await tabTo(page, offerBtns.nth(want), { back: false });
    // Focusing a card shows its preview, like hover does.
    await expect(page.getByTestId('card-preview'), `pack ${pack}`).toBeVisible();
    await page.keyboard.press(pack % 2 ? 'Enter' : ' ');
  }

  // ---- shop: buy one with Space, then continue with Enter ----
  await headingFocused(page, 'Shop');
  await expect(page.getByRole('status').filter({ hasText: 'Tickets: 4 / 4' })).toBeVisible();
  const add = page.locator('button[aria-label^="+ Add"]:not([disabled])');
  await press(page, add.first(), 'Space');
  await expect(page.getByRole('status').filter({ hasText: 'Tickets: 3 / 4' })).toBeVisible();
  await expect(add.first()).toBeFocused(); // still on the button: press again to buy more, or move on
  await press(page, page.getByRole('button', { name: 'Continue to deckbuild' }), 'Enter', { back: true });

  // ---- deckbuild: Space on every drafted card's + until the deck has 20 ----
  await headingFocused(page, 'Deckbuilding');
  await expect(page.getByRole('status').filter({ hasText: 'Deck: 0 / 20' })).toBeVisible();
  const deckStatus = page.getByRole('status').filter({ hasText: /^Deck: \d+ \/ 20/ });
  for (let guard = 0; guard < 400; guard++) {
    if (/Deck: 20 \/ 20/.test((await deckStatus.textContent()) ?? '')) break;
    await page.keyboard.press('Tab');
    const f = await focusFact(page);
    if (f.tag === 'body') continue;
    await checkFocus(page);
    if (/^\+ Add /.test(f.name) && (await page.evaluate(() => !(document.activeElement as HTMLButtonElement).disabled))) {
      await page.keyboard.press(guard % 2 ? 'Enter' : ' ');
    }
  }
  await expect(deckStatus).toContainText('Deck: 20 / 20');

  // ---- Energy Zone: toggle chips by keyboard; the cap is announced ----
  const picker = page.getByTestId('energy-picker');
  const chips = picker.getByRole('group', { name: 'Energy Zone types' }).getByRole('button');
  const pressedCount = () => picker.locator('button[aria-pressed="true"]').count();
  // The zone starts from the deck's own types (1 to 3). Make room if it is full, then fill it to the cap with Space.
  if ((await pressedCount()) === 3) await press(page, picker.locator('button[aria-pressed="true"]').first(), 'Space', { back: true });
  expect(await pressedCount()).toBeLessThan(3);
  for (let i = 0; i < 8 && (await pressedCount()) < 3; i++) {
    const chip = chips.nth(i);
    if ((await chip.getAttribute('aria-pressed')) === 'true') continue;
    await press(page, chip, 'Space', { back: true });
  }
  expect(await pressedCount()).toBe(3);
  await expect(picker).toContainText('Energy Zone (3 / 3)');
  await expect(announcer(page)).toContainText('Up to 3: deselect one to change.');
  // Unpress one (not the last): keyboard can undo.
  const on = picker.locator('button[aria-pressed="true"]').first();
  await press(page, on, 'Enter', { back: true });
  expect(await pressedCount()).toBe(2);
  await expect(picker).toContainText('Energy Zone (2 / 3)');

  // ---- finalize ----
  await press(page, page.getByRole('button', { name: 'Finalize deck' }), 'Enter', { back: true });
  await headingFocused(page, 'Deck finalized');

  // ---- run screen: the quit confirm, then the first battle ----
  await press(page, page.getByRole('button', { name: 'Quit run' }), 'Enter');
  const quit = page.getByRole('alertdialog', { name: 'Quit run' });
  await expect(quit.getByRole('button', { name: 'Keep playing' })).toBeFocused();
  await checkFocus(page);
  await page.keyboard.press('Escape');
  await expect(quit).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Quit run' })).toBeFocused();

  await press(page, page.getByRole('button', { name: 'Next battle' }), 'Enter', { back: true });
  await concedeByStore(page);

  // Back to the run screen from the end screen (the board's own button; focus() since the board is out of scope).
  await page.getByRole('button', { name: 'Back' }).focus();
  await page.keyboard.press('Enter');
  await headingFocused(page, 'Deck finalized');
  await expect(announcer(page)).toContainText('Battle 1: Loss. Run: 0–1');
  await expect(page.getByTestId('run-battles').locator('li')).toHaveCount(1);

  // Leave confirm: Home asks first; Escape keeps the run and focus returns to Home.
  await press(page, page.getByRole('button', { name: 'Home', exact: true }), 'Enter', { back: true });
  const leave = page.getByTestId('leave-confirm');
  await expect(leave.getByRole('button', { name: 'Keep playing' })).toBeFocused();
  await page.keyboard.press('Escape');
  await expect(leave).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Home', exact: true })).toBeFocused();

  // Second battle, second loss: the run ends on the summary.
  await press(page, page.getByRole('button', { name: 'Next battle' }), 'Enter');
  await concedeByStore(page);
  await page.getByRole('button', { name: 'Back' }).focus();
  await page.keyboard.press('Enter');

  await headingFocused(page, 'Run lost');
  await expect(announcer(page)).toContainText('Battle 2: Loss. Run: 0–2. Run lost');
  await expect(page.getByTestId('run-battles').locator('li')).toHaveCount(2);
  // Reach the summary's actions by keyboard.
  await tabTo(page, page.getByRole('button', { name: 'Replay' }).first());
  await tabTo(page, page.getByRole('button', { name: 'Draft again' }));
  await tabTo(page, page.getByRole('button', { name: 'Home', exact: true }));
  await tabTo(page, page.getByRole('button', { name: 'Copy code' }));

  // ---- the bar, over everything focused above ----
  expect(seen.length).toBeGreaterThan(100);
  const bad = seen.filter((f) => !f.focusVisible || !hasIndicator(f));
  expect(bad, JSON.stringify(bad.slice(0, 5))).toEqual([]);
});

// ---------- 44 x 44 targets ----------

const TARGETS = 'button, a[href], [role="button"], input, textarea, select, summary';

// Every visible control that is too small, as "label WxH".
async function tooSmall(page: Page, screen: string): Promise<string[]> {
  const small = await page.evaluate((sel) => {
    const out: string[] = [];
    for (const el of document.querySelectorAll<HTMLElement>(sel)) {
      const cs = getComputedStyle(el);
      if (cs.display === 'none' || cs.visibility === 'hidden') continue;
      const r = el.getBoundingClientRect();
      if (r.width === 0 && r.height === 0) continue;
      if (r.width < 43.99 || r.height < 43.99) {
        out.push(`${el.tagName.toLowerCase()} "${(el.getAttribute('aria-label') ?? el.textContent ?? '').trim().slice(0, 30)}" ${r.width.toFixed(1)}x${r.height.toFixed(1)}`);
      }
    }
    return out;
  }, TARGETS);
  return small.map((s) => `${screen}: ${s}`);
}

for (const vp of [
  { width: 390, height: 844 },
  { width: 1440, height: 900 },
]) {
  test(`every button, link and input is at least 44x44 at ${vp.width}x${vp.height}`, async ({ page }) => {
    test.setTimeout(300_000);
    await page.setViewportSize(vp);
    await stubImages(page);
    const small: string[] = [];

    await page.goto('/');
    await expect(page.getByRole('button', { name: 'Start draft' })).toBeVisible();
    small.push(...(await tooSmall(page, 'opening')));

    await page.getByRole('button', { name: 'Start draft' }).click();
    await expect(page.locator('.offer-grid button')).toHaveCount(5);
    await page.getByRole('button', { name: 'Abandon draft' }).click();
    small.push(...(await tooSmall(page, 'draft (abandon confirm open)')));
    await page.getByRole('button', { name: 'Cancel' }).click();
    await draftTwenty(page);
    await expect(page.getByRole('heading', { name: 'Shop' })).toBeVisible();
    await page.getByRole('button', { name: /^\+ Add / }).first().click();
    small.push(...(await tooSmall(page, 'shop')));
    await page.getByRole('button', { name: 'Continue to deckbuild' }).click();

    await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
    small.push(...(await tooSmall(page, 'deckbuild (empty deck)')));
    await buildFromDraft(page);
    await expect(page.getByRole('heading', { name: 'Deck finalized' })).toBeVisible();
    small.push(...(await tooSmall(page, 'run')));
    await page.getByRole('button', { name: 'Quit run' }).click();
    small.push(...(await tooSmall(page, 'run (quit confirm open)')));
    await page.getByRole('button', { name: 'Keep playing' }).click();

    for (let i = 0; i < 2; i++) {
      await page.getByRole('button', { name: 'Next battle' }).click();
      await concedeByStore(page);
      // (the battle's own end screen is the board's, out of scope: its Back/Rematch/History are 34-43 px)
      await page.getByRole('button', { name: 'Back' }).click();
    }
    await expect(page.getByRole('heading', { name: 'Run lost' })).toBeVisible();
    small.push(...(await tooSmall(page, 'end of run')));

    await page.getByRole('button', { name: 'Home', exact: true }).click();
    await page.getByRole('button', { name: /Past runs/ }).click();
    await expect(page.getByRole('heading', { name: 'Draft history' })).toBeVisible();
    small.push(...(await tooSmall(page, 'history')));
    await page.getByRole('button', { name: 'View' }).first().click();
    small.push(...(await tooSmall(page, 'history (entry open)')));

    expect(small, small.join('\n')).toEqual([]);
  });
}

// ---------- names without glyphs; keyboard-focus previews; the leave countdown ----------

// Every accessible name that carries a rarity is in words: no diamond, star or crown glyph, no other symbol.
async function expectNamesInWords(page: Page, screen: string, minTiles: number) {
  const { bad, tiles } = await page.evaluate(() => {
    const names = [...document.querySelectorAll<HTMLElement>('[aria-label]')].map((e) => e.getAttribute('aria-label') ?? '');
    const alts = [...document.images].map((i) => i.alt);
    const glyph = /[^\p{L}\p{N}\s,.'’:+−\-–()/]/u;
    return {
      bad: [...names, ...alts].filter((n) => glyph.test(n)),
      tiles: names.filter((n) => /, \d (diamond|star)s?(,|$)|, crown/.test(n)).length,
    };
  });
  expect(bad, `${screen}: names with symbols`).toEqual([]);
  expect(tiles, `${screen}: tiles named with rarity words`).toBeGreaterThanOrEqual(minTiles);
}

const previewBox = (page: Page) => page.getByTestId('card-preview').boundingBox();
const overlaps = (a: { x: number; y: number; width: number; height: number }, b: typeof a) =>
  a.x < b.x + b.width - 0.5 && b.x < a.x + a.width - 0.5 && a.y < b.y + b.height - 0.5 && b.y < a.y + a.height - 0.5;

// Keyboard focus on a card's control shows that card's preview; on a phone it never covers the card.
async function focusPreviewChecks(page: Page, add: Locator, phone: boolean) {
  const preview = page.getByTestId('card-preview');
  const cell = (n: number) => add.nth(n).locator('xpath=ancestor::div[contains(@class,"shelf-cell")]');
  for (const where of phone ? (['top', 'bottom'] as const) : (['top'] as const)) {
    const c = cell(0);
    await c.evaluate((el, y) => window.scrollBy(0, el.getBoundingClientRect().top - y), where === 'top' ? 60 : 460);
    await page.keyboard.press('Shift'); // the keyboard is the input in use
    await add.nth(0).focus();
    await expect(preview, `${where}: preview on focus`).toBeVisible();
    const tile = (await c.locator('.tile-lg').boundingBox())!;
    const pv = (await previewBox(page))!;
    if (phone) {
      expect(overlaps(pv, tile), `${where}: preview ${JSON.stringify(pv)} covers tile ${JSON.stringify(tile)}`).toBe(false);
      expect(pv.y).toBeGreaterThanOrEqual(0);
      expect(pv.y + pv.height).toBeLessThanOrEqual(page.viewportSize()!.height);
    }
    // Escape hides it and it stays hidden while focus stays on this card's buttons.
    await page.keyboard.press('Escape');
    await expect(preview).toHaveCount(0);
    // Another card's control: its preview.
    await add.nth(1).focus();
    await expect(preview).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(preview).toHaveCount(0);
  }
}

for (const vp of [
  { width: 390, height: 844 },
  { width: 1440, height: 900 },
]) {
  test(`card names in words and previews on keyboard focus at ${vp.width}x${vp.height}`, async ({ page }) => {
    test.setTimeout(300_000);
    await page.setViewportSize(vp);
    await stubImages(page);
    await page.goto('/');
    await page.getByRole('button', { name: 'Start draft' }).click();
    await expect(page.locator('.offer-grid button')).toHaveCount(5);
    await expectNamesInWords(page, 'draft offers', 5);
    await draftTwenty(page);
    await expect(page.getByRole('heading', { name: 'Shop' })).toBeVisible();
    await expectNamesInWords(page, 'shop', 1);
    await focusPreviewChecks(page, page.locator('button[aria-label^="+ Add"]'), vp.width <= 700);
    await page.getByRole('button', { name: 'Skip shop' }).click();
    await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
    await focusPreviewChecks(page, page.locator('button[aria-label^="+ Add"]'), vp.width <= 700);
    await buildFromDraft(page);
    await expect(page.getByRole('heading', { name: 'Deck finalized' })).toBeVisible();
    await expectNamesInWords(page, 'run / review', 20);

    // The review's cards are focus stops with a preview.
    await page.keyboard.press('Shift');
    await page.locator('.deck-grid [role="group"]').first().focus();
    await expect(page.getByTestId('card-preview')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.getByTestId('card-preview')).toHaveCount(0);

    // Deck-side tiles (deckbuild) and history: end the run, open Past runs.
    for (let i = 0; i < 2; i++) {
      await page.getByRole('button', { name: 'Next battle' }).click();
      await concedeByStore(page);
      await page.getByRole('button', { name: 'Back' }).click();
    }
    await expect(page.getByRole('heading', { name: 'Run lost' })).toBeVisible();
    await page.getByRole('button', { name: 'Home', exact: true }).click();
    await page.getByRole('button', { name: /Past runs/ }).click();
    await page.getByRole('button', { name: 'View' }).first().click();
    await expectNamesInWords(page, 'history', 20);
    await page.keyboard.press('Shift');
    await page.locator('.history-grid [role="group"]').first().focus();
    await expect(page.getByTestId('card-preview')).toBeVisible();
  });
}

test('leave countdown: when time runs out under the Return button, focus goes to the heading, not the page', async ({ page }) => {
  test.setTimeout(120_000);
  await stubImages(page);
  await page.clock.install();
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await expect(page.getByTestId('run-record')).toBeVisible();
  await page.getByRole('button', { name: 'Next battle' }).click();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-leave').click();
  await expect(page.getByTestId('leave-countdown')).toBeVisible();
  const ret = page.getByRole('button', { name: 'Return to battle' });
  await tabTo(page, ret);
  await expect(ret).toBeFocused();
  await page.clock.fastForward(91_000);
  await expect(page.getByTestId('leave-countdown')).toHaveCount(0);
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–1 (active)');
  await expect(page.getByRole('heading', { level: 1 })).toBeFocused();
});
