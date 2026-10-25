import { test, expect, type CDPSession, type Locator, type Page } from '@playwright/test';
import { mkdirSync, writeFileSync } from 'node:fs';

// Issue 113: the draft, shop and deckbuild screens fit a phone (one centred column, a
// scannable 2-wide offer grid, the drafted list reachable, no off-screen panel, no hover
// zoom on tap) and stay usable on a 1440x900 desktop. Screenshots and measurements land
// in test-results/phone-layout/.

const OUT = 'test-results/phone-layout';
const TOTAL_PACKS = 20;
const GUTTER_MIN = 16;
const GUTTER_SKEW_MAX = 4;

type Rect = { x: number; y: number; width: number; height: number };

const rectOf = async (l: Locator): Promise<Rect> => {
  const b = await l.boundingBox();
  if (!b) throw new Error('element has no box');
  return b;
};

// Page-relative (not viewport-relative) top and bottom of a box.
const pageY = async (page: Page, l: Locator) => {
  const b = await rectOf(l);
  const scroll = await page.evaluate(() => window.scrollY);
  return { top: b.y + scroll, bottom: b.y + scroll + b.height };
};

async function noHorizontalScroll(page: Page) {
  const { scrollW, innerW } = await page.evaluate(() => ({
    scrollW: document.documentElement.scrollWidth,
    innerW: window.innerWidth,
  }));
  expect(scrollW).toBeLessThanOrEqual(innerW);
}

// Left and right space around the outermost cards of a grid.
async function gutters(page: Page, tiles: Locator) {
  const vw = page.viewportSize()!.width;
  const boxes = await Promise.all((await tiles.all()).map(rectOf));
  const left = Math.min(...boxes.map((b) => b.x));
  const right = vw - Math.max(...boxes.map((b) => b.x + b.width));
  return { left, right };
}

async function expectCentredColumn(page: Page, tiles: Locator) {
  const { left, right } = await gutters(page, tiles);
  expect(left).toBeGreaterThanOrEqual(GUTTER_MIN);
  expect(right).toBeGreaterThanOrEqual(GUTTER_MIN);
  expect(Math.abs(left - right)).toBeLessThanOrEqual(GUTTER_SKEW_MAX);
  return { left, right };
}

const hoverPreview = (page: Page) => page.locator('div[style*="position: fixed"]');
const zoom = (page: Page) => page.locator('.card-zoom');

const touch = (cdp: CDPSession, type: string, p?: { x: number; y: number }) =>
  cdp.send('Input.dispatchTouchEvent', { type, touchPoints: p ? [p] : [] });

async function centre(l: Locator) {
  const b = await rectOf(l);
  return { x: b.x + b.width / 2, y: b.y + b.height / 2 };
}

async function shot(page: Page, name: string) {
  mkdirSync(OUT, { recursive: true });
  await page.screenshot({ path: `${OUT}/${name}.png`, fullPage: true });
}

const offers = (page: Page) => page.locator('.offer-grid button');

// Card images load lazily; measure sizes only once they have.
const imagesLoaded = (page: Page) =>
  page.waitForFunction(() => [...document.images].every((i) => i.complete && i.naturalWidth > 0));

async function draftAll(page: Page) {
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  for (let i = 0; i < TOTAL_PACKS; i++) {
    await expect(page.getByText(`Pack ${i + 1} / ${TOTAL_PACKS}`)).toBeVisible();
    await expect(offers(page)).toHaveCount(5);
    await imagesLoaded(page);
    if (i === 0) await draftScreenChecks(page);
    await offers(page).first().click();
  }
}

// Runs on pack 1, before any pick, at whatever viewport the test set.
let measured: Record<string, unknown> = {};
async function draftScreenChecks(page: Page) {
  const vp = page.viewportSize()!;
  const phone = vp.width < 700;
  await noHorizontalScroll(page);
  const tiles = offers(page);
  const bottoms = await Promise.all((await tiles.all()).map((t) => pageY(page, t)));
  const lastBottom = Math.max(...bottoms.map((b) => b.bottom));
  const first = await rectOf(tiles.first());
  const pageHeight = await page.evaluate(() => document.documentElement.scrollHeight);
  const g = await gutters(page, tiles);
  measured.draft = { gutters: g, offerSize: [first.width, first.height], offersBottom: lastBottom, pageHeight };
  await shot(page, `draft-${vp.width}`);

  if (phone) {
    await expectCentredColumn(page, tiles);
    expect(lastBottom).toBeLessThanOrEqual(vp.height * 2);
    // 2-wide: the first two offers share a row.
    expect((await rectOf(tiles.nth(1))).y).toBeCloseTo(first.y, 0);
  } else {
    expect(lastBottom).toBeLessThanOrEqual(vp.height);
  }
}

async function draftedListReachable(page: Page) {
  const vw = page.viewportSize()!.width;
  const heading = page.getByText(/^Drafted \(\d+\)$/);
  await heading.scrollIntoViewIfNeeded();
  const box = await rectOf(heading);
  expect(box.x).toBeGreaterThanOrEqual(0);
  expect(box.x + box.width).toBeLessThanOrEqual(vw);
  await noHorizontalScroll(page);
}

async function shopChecks(page: Page) {
  const vp = page.viewportSize()!;
  await expect(page.getByRole('heading', { name: 'Shop' })).toBeVisible();
  await noHorizontalScroll(page);
  const tiles = page.locator('.shelf-grid .shelf-cell');
  await expect(tiles.first()).toBeVisible();
  if (vp.width < 700) measured.shop = await expectCentredColumn(page, tiles);
  await shot(page, `shop-${vp.width}`);
  await draftedListReachable(page);
}

async function deckbuildChecks(page: Page) {
  const vp = page.viewportSize()!;
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
  await noHorizontalScroll(page);
  const tiles = page.locator('.shelf-grid .shelf-cell');
  await expect(tiles.first()).toBeVisible();
  if (vp.width < 700) measured.deckbuild = await expectCentredColumn(page, tiles);
  await shot(page, `deckbuild-${vp.width}`);
  const heading = page.getByText(/^Deck \(\d+ \/ 20\)$/);
  await heading.scrollIntoViewIfNeeded();
  const box = await rectOf(heading);
  expect(box.x + box.width).toBeLessThanOrEqual(vp.width);
  await noHorizontalScroll(page);
}

function record(name: string) {
  mkdirSync(OUT, { recursive: true });
  writeFileSync(`${OUT}/${name}.json`, JSON.stringify(measured, null, 2));
}

test.describe('phone 390x844', () => {
  test('draft, shop and deckbuild fit; tap does not zoom, long-press does', async ({ page }) => {
    measured = {};
    await page.goto('/');
    // Touch from the start: a mouse click here would leave a real mouse hover over the offers.
    await page.getByRole('button', { name: 'Start draft' }).tap();
    await expect(offers(page)).toHaveCount(5);
    await imagesLoaded(page);
    await draftScreenChecks(page);

    // A tap picks (does not open the hover zoom).
    await offers(page).first().tap();
    await expect(page.getByText(`Pack 2 / ${TOTAL_PACKS}`)).toBeVisible();
    await expect(hoverPreview(page)).toHaveCount(0);
    await expect(zoom(page)).toHaveCount(0);
    await draftedListReachable(page);

    // A long press zooms while held, and releasing it does not pick.
    await page.evaluate(() => window.scrollTo(0, 0));
    const cdp = await page.context().newCDPSession(page);
    await touch(cdp, 'touchStart', await centre(offers(page).first()));
    await expect(zoom(page)).toBeVisible();
    await touch(cdp, 'touchEnd');
    await expect(zoom(page)).toHaveCount(0);
    await expect(page.getByText(`Pack 2 / ${TOTAL_PACKS}`)).toBeVisible();

    for (let i = 1; i < TOTAL_PACKS; i++) {
      await offers(page).first().click();
      if (i < TOTAL_PACKS - 1) await expect(page.getByText(`Pack ${i + 2} / ${TOTAL_PACKS}`)).toBeVisible();
    }

    await shopChecks(page);
    // A tap on a shop card opens nothing.
    await page.locator('.shelf-cell .tile-lg').first().tap();
    await expect(hoverPreview(page)).toHaveCount(0);
    await expect(zoom(page)).toHaveCount(0);

    await page.getByRole('button', { name: /Skip shop|Continue to deckbuild/ }).click();
    await deckbuildChecks(page);
    record('measurements-390');
    console.log('MEASURED@390', JSON.stringify(measured));
  });
});

test.describe('desktop 1440x900', () => {
  test.use({ viewport: { width: 1440, height: 900 }, isMobile: false, hasTouch: false });

  test('draft, shop and deckbuild stay usable', async ({ page }) => {
    measured = {};
    await draftAll(page);
    // Side panel stays beside the offers, on screen.
    await expect(page.getByText(/^Drafted \(\d+\)$/)).toBeVisible();
    await shopChecks(page);
    await page.getByRole('button', { name: /Skip shop|Continue to deckbuild/ }).click();
    await deckbuildChecks(page);
    record('measurements-1440');
    console.log('MEASURED@1440', JSON.stringify(measured));
  });

  test('mouse hover still previews a card', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('button', { name: 'Start draft' }).click();
    await offers(page).first().hover();
    await expect(hoverPreview(page)).toHaveCount(1);
  });
});
