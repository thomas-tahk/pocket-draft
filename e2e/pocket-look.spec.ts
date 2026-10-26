import { test, expect, type Locator, type Page } from '@playwright/test';
import { buildFromDraft, draftTwenty } from './helpers/draft';

// #120: one Pocket-style look across every screen. This spec
//  1. visits each screen at 390x844 and 1440x900 and saves test-results/look/<screen>-<w>.png,
//  2. checks that a screen change has no transition with reduced motion (and a short one without),
//  3. checks that hovering a draft offer never covers the other offers.
//
// The card-image CDN is blocked in this sandbox, so every card image is answered with a local
// placeholder card (page.route): the screenshots show the layout, not broken images.

const OUT = 'test-results/look';
const IMAGE_HOST = '**://limitlesstcg.nyc3.cdn.digitaloceanspaces.com/**';
const KEY = 'pocket-draft-v0.4';

// A neutral 5:7 card face: rounded frame, art window, name bar. Not a real card.
const PLACEHOLDER = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 250 350">
<rect width="250" height="350" rx="14" fill="#f4f1e6"/>
<rect x="10" y="10" width="230" height="330" rx="10" fill="#fff8d8" stroke="#c9b458" stroke-width="4"/>
<rect x="24" y="48" width="202" height="150" rx="6" fill="#bfe3f2"/>
<circle cx="125" cy="123" r="42" fill="#7fc4e0"/>
<rect x="24" y="22" width="120" height="14" rx="7" fill="#8a8466"/>
<rect x="24" y="214" width="150" height="10" rx="5" fill="#8a8466"/>
<rect x="24" y="236" width="202" height="8" rx="4" fill="#bdb58f"/>
<rect x="24" y="254" width="180" height="8" rx="4" fill="#bdb58f"/>
<rect x="24" y="272" width="190" height="8" rx="4" fill="#bdb58f"/>
</svg>`;

async function stubImages(page: Page) {
  await page.route(IMAGE_HOST, (route) => route.fulfill({ status: 200, contentType: 'image/svg+xml', body: PLACEHOLDER }));
}

const offers = (page: Page) => page.locator('.offer-grid button');

// Images in view have loaded (lazy ones off screen are not waited for).
const settled = (page: Page) =>
  page.waitForFunction(() =>
    [...document.images].every((i) => {
      const r = i.getBoundingClientRect();
      const onScreen = r.bottom > 0 && r.top < innerHeight && r.right > 0 && r.left < innerWidth;
      return !onScreen || (i.complete && i.naturalWidth > 0);
    }),
  );

async function shot(page: Page, name: string) {
  await settled(page);
  // The screen fade is 160 ms at most; with reduced motion it is already done.
  await page.waitForTimeout(250);
  await page.screenshot({ path: `${OUT}/${name}-${page.viewportSize()!.width}.png` });
}

const state = (page: Page) => page.evaluate((k) => localStorage.getItem(k), KEY);

for (const vp of [
  { width: 390, height: 844 },
  { width: 1440, height: 900 },
]) {
  test.describe(`screens at ${vp.width}x${vp.height}`, () => {
    test.use({ viewport: vp, isMobile: vp.width < 700, hasTouch: vp.width < 700 });

    test('every screen has a screenshot', async ({ page }) => {
      test.setTimeout(240_000);
      await stubImages(page);

      // home
      await page.goto('/');
      await expect(page.getByRole('button', { name: 'Start draft' })).toBeVisible();
      await shot(page, 'home');

      // draft
      await page.getByRole('button', { name: 'Start draft' }).click();
      await expect(offers(page)).toHaveCount(5);
      await shot(page, 'draft');
      await draftTwenty(page);

      // shop
      await expect(page.getByRole('heading', { name: 'Shop' })).toBeVisible();
      await shot(page, 'shop');
      await page.getByRole('button', { name: 'Skip shop' }).click();

      // deckbuild, with the Energy Zone picker on screen
      await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
      await page.getByTestId('energy-picker').scrollIntoViewIfNeeded();
      await page.evaluate(() => window.scrollTo(0, 0));
      await expect(page.getByTestId('energy-picker')).toBeInViewport();
      await shot(page, 'deckbuild');
      await buildFromDraft(page);

      // run (fresh), then review: the same deck screen with no run behind it
      await expect(page.getByTestId('run-record')).toHaveText('Run: 0–0 (active)');
      await shot(page, 'run');
      const saved = (await state(page))!;
      const noRun = JSON.parse(saved);
      noRun.state.run = null;
      await page.evaluate(([k, v]) => localStorage.setItem(k, v), [KEY, JSON.stringify(noRun)]);
      await page.goto('/?replay=99'); // a replay with no battle behind it opens the deck screen
      await expect(page.getByRole('heading', { name: 'Deck finalized' })).toBeVisible();
      await expect(page.getByTestId('run-record')).toHaveCount(0);
      await shot(page, 'review');
      await page.evaluate(([k, v]) => localStorage.setItem(k, v), [KEY, saved]);
      await page.goto('/');
      await page.getByRole('button', { name: 'Continue run' }).click();
      await expect(page.getByTestId('run-record')).toHaveText('Run: 0–0 (active)');

      // battle, its end screen, and the run with a battle in the list
      await page.getByRole('button', { name: 'Next battle' }).click();
      await expect(page.getByTestId('battle-board')).toBeVisible();
      await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 5000 });
      await shot(page, 'battle');
      await page.getByTestId('menu').click();
      await page.getByTestId('menu-concede').click();
      await page.getByTestId('confirm-ok').click();
      await expect(page.getByTestId('end-result')).toHaveText('Loss');
      await shot(page, 'battle-end');
      await page.getByRole('button', { name: 'Back' }).click();
      await expect(page.getByTestId('run-battles').locator('li')).toHaveCount(1);
      await shot(page, 'run-after-battle');

      // replay of that battle
      await page.getByRole('button', { name: 'Replay' }).click();
      await expect(page.getByTestId('replay-exit')).toBeVisible();
      await shot(page, 'replay');
      await page.getByTestId('replay-exit').click();

      // end-of-run summary
      await page.getByRole('button', { name: 'Quit run' }).click();
      await page.getByTestId('quit-confirm').click();
      await expect(page.getByRole('main', { name: 'Run summary' })).toBeVisible();
      await shot(page, 'summary');

      // history, collapsed and with the entry open
      await page.getByRole('button', { name: 'Home' }).click();
      await page.getByRole('button', { name: /Past runs/ }).click();
      await expect(page.getByRole('heading', { name: 'Draft history' })).toBeVisible();
      await shot(page, 'history');
      await page.getByRole('button', { name: 'View' }).first().click();
      await shot(page, 'history-open');
    });
  });
}

test.describe('screen change motion', () => {
  const toDraft = async (page: Page) => {
    await stubImages(page);
    await page.goto('/');
    await page.getByRole('button', { name: 'Start draft' }).click();
    await expect(offers(page)).toHaveCount(5);
    return page.locator('[data-screen="draft"]');
  };
  const durations = (l: Locator) =>
    l.evaluate((el) => {
      const s = getComputedStyle(el);
      return { animation: s.animationDuration, transition: s.transitionDuration, name: s.animationName };
    });

  test('with reduced motion a screen change has no transition', async ({ page }) => {
    await page.emulateMedia({ reducedMotion: 'reduce' });
    const screen = await toDraft(page);
    const d = await durations(screen);
    expect(d.animation).toBe('0s');
    expect(d.transition).toBe('0s');
    expect(d.name).toBe('none');
    expect(await screen.evaluate((el) => el.getAnimations().length)).toBe(0);
  });

  test('without reduced motion the change is a fade of at most 200 ms, opacity only', async ({ page }) => {
    await page.emulateMedia({ reducedMotion: 'no-preference' });
    await stubImages(page);
    await page.goto('/');
    // Watch the incoming screen for its animation as it mounts.
    await page.evaluate(() => {
      (window as unknown as { seen: unknown[] }).seen = [];
      new MutationObserver(() => {
        for (const el of document.querySelectorAll('[data-screen="draft"]')) {
          for (const a of el.getAnimations()) {
            const t = a.effect!.getTiming();
            const props = new Set(a.effect!.getKeyframes().flatMap((k) => Object.keys(k)));
            (window as unknown as { seen: unknown[] }).seen.push({ ms: t.duration, props: [...props] });
          }
        }
      }).observe(document.body, { childList: true, subtree: true });
    });
    await page.getByRole('button', { name: 'Start draft' }).click();
    await expect(offers(page)).toHaveCount(5);
    const seen = (await page.evaluate(() => (window as unknown as { seen: { ms: number; props: string[] }[] }).seen)) ?? [];
    expect(seen.length).toBeGreaterThan(0);
    for (const s of seen) {
      expect(s.ms).toBeGreaterThan(0);
      expect(s.ms).toBeLessThanOrEqual(200);
      expect(s.props.filter((p) => !['offset', 'easing', 'composite', 'computedOffset', 'opacity'].includes(p))).toEqual([]);
    }
  });
});

test.describe('hover on draft offers', () => {
  test.use({ viewport: { width: 1440, height: 900 }, isMobile: false, hasTouch: false });

  test('the preview never covers the other offers', async ({ page }) => {
    await stubImages(page);
    await page.goto('/');
    await page.getByRole('button', { name: 'Start draft' }).click();
    await expect(offers(page)).toHaveCount(5);
    await settled(page);
    const boxes = await Promise.all((await offers(page).all()).map((o) => o.boundingBox()));
    const preview = page.getByTestId('card-preview');
    for (let i = 0; i < 5; i++) {
      await offers(page).nth(i).hover();
      await expect(preview).toHaveCount(1);
      await settled(page);
      const p = (await preview.boundingBox())!;
      boxes.forEach((b, j) => {
        const overlap = p.x < b!.x + b!.width && p.x + p.width > b!.x && p.y < b!.y + b!.height && p.y + p.height > b!.y;
        expect(overlap, `hovering offer ${i} covers offer ${j}`).toBe(false);
      });
      if (i === 4) await page.screenshot({ path: `${OUT}/draft-hover-1440.png` });
    }
    await page.mouse.move(5, 5);
    await expect(preview).toHaveCount(0);
  });
});
