import { test, expect, type Page } from '@playwright/test';

// #131: dark is an option, never forced. Auto follows the phone (prefers-color-scheme); the toggle on
// Home (Auto / Light / Dark) overrides it via data-theme on <html> and survives a reload.

const LIGHT_PAGE = '#0b63c9';
const DARK_PAGE = '#140b3d';

// The page colour token the blended look is drawn with right now.
const pageColour = (page: Page) =>
  page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--look-page').trim());
const themeAttr = (page: Page) => page.evaluate(() => document.documentElement.getAttribute('data-theme'));
const option = (page: Page, name: 'Auto' | 'Light' | 'Dark') =>
  page.getByRole('group', { name: 'Theme' }).getByRole('button', { name, exact: true });

async function home(page: Page) {
  await page.goto('/');
  await expect(page.getByRole('button', { name: 'Start draft' })).toBeVisible();
}

test('Auto follows the emulated colour scheme', async ({ page }) => {
  await page.emulateMedia({ colorScheme: 'light' });
  await home(page);
  await expect(option(page, 'Auto')).toHaveAttribute('aria-pressed', 'true');
  expect(await themeAttr(page)).toBeNull();
  expect(await pageColour(page)).toBe(LIGHT_PAGE);

  await page.emulateMedia({ colorScheme: 'dark' });
  await expect.poll(() => pageColour(page)).toBe(DARK_PAGE);
  // Dark body: near-black background, light text (the same check as pocket-look's dark test).
  const bodyBg = await page.evaluate(() => getComputedStyle(document.body).backgroundColor);
  expect(bodyBg).toBe('rgb(20, 11, 61)');
});

test('the toggle overrides the phone setting and survives a reload', async ({ page }) => {
  await page.emulateMedia({ colorScheme: 'dark' });
  await home(page);
  expect(await pageColour(page)).toBe(DARK_PAGE);

  // Light on a dark phone.
  await option(page, 'Light').click();
  await expect(option(page, 'Light')).toHaveAttribute('aria-pressed', 'true');
  await expect(option(page, 'Auto')).toHaveAttribute('aria-pressed', 'false');
  expect(await themeAttr(page)).toBe('light');
  expect(await pageColour(page)).toBe(LIGHT_PAGE);
  await page.reload();
  await expect(option(page, 'Light')).toHaveAttribute('aria-pressed', 'true');
  expect(await themeAttr(page)).toBe('light');
  expect(await pageColour(page)).toBe(LIGHT_PAGE);

  // Dark on a light phone, and it holds on other screens too.
  await page.emulateMedia({ colorScheme: 'light' });
  await option(page, 'Dark').click();
  expect(await pageColour(page)).toBe(DARK_PAGE);
  await page.reload();
  await expect(option(page, 'Dark')).toHaveAttribute('aria-pressed', 'true');
  expect(await pageColour(page)).toBe(DARK_PAGE);
  await page.getByRole('button', { name: /Past runs/ }).click();
  await expect(page.getByRole('heading', { name: 'Draft history' })).toBeVisible();
  expect(await pageColour(page)).toBe(DARK_PAGE);
  await page.getByRole('button', { name: 'Back' }).click();

  // Back to Auto: the override and its saved value are gone; the phone (light) decides again.
  await option(page, 'Auto').click();
  expect(await themeAttr(page)).toBeNull();
  expect(await pageColour(page)).toBe(LIGHT_PAGE);
  expect(await page.evaluate(() => localStorage.getItem('pocket-draft-theme'))).toBeNull();
  await page.reload();
  await expect(option(page, 'Auto')).toHaveAttribute('aria-pressed', 'true');
  expect(await pageColour(page)).toBe(LIGHT_PAGE);
});

test('a page whose storage is blocked still opens, on Auto, and the toggle still works', async ({ page }) => {
  await page.addInitScript(() => {
    const block = () => {
      throw new DOMException('blocked', 'SecurityError');
    };
    const real = { get: Storage.prototype.getItem, set: Storage.prototype.setItem };
    // Only the theme key is blocked, so the draft store (which has its own handling) is unaffected.
    Storage.prototype.getItem = function (k: string) { return k === 'pocket-draft-theme' ? block() : real.get.call(this, k); };
    Storage.prototype.setItem = function (k: string, v: string) { return k === 'pocket-draft-theme' ? block() : real.set.call(this, k, v); };
  });
  await page.emulateMedia({ colorScheme: 'light' });
  await home(page);
  await expect(option(page, 'Auto')).toHaveAttribute('aria-pressed', 'true');
  await option(page, 'Dark').click();
  expect(await pageColour(page)).toBe(DARK_PAGE);
});
