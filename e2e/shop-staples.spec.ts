import { test, expect } from '@playwright/test';
import { draftTwenty } from './helpers/draft';

// #137/#138: the shop shows Rare Candy and sells Trainers only (no Pokémon shelf).
// Card images are blocked in this sandbox, so the CDN is answered with a placeholder.
const IMAGE_HOST = '**://limitlesstcg.nyc3.cdn.digitaloceanspaces.com/**';
const PLACEHOLDER = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 250 350"><rect width="250" height="350" rx="14" fill="#f4f1e6"/></svg>`;
const OUT = 'docs/chief/look/m4-fixes';

for (const [w, h] of [[390, 844], [1440, 900]]) {
  test(`shop shows Rare Candy and no Pokémon shelf at ${w}x${h}`, async ({ page }) => {
    await page.setViewportSize({ width: w, height: h });
    await page.route(IMAGE_HOST, (route) => route.fulfill({ status: 200, contentType: 'image/svg+xml', body: PLACEHOLDER }));
    await page.goto('/');
    await page.getByRole('button', { name: 'Start draft' }).click();
    await draftTwenty(page);
    await expect(page.getByRole('heading', { name: 'Shop' })).toBeVisible();

    await expect(page.getByRole('button', { name: '+ Add Rare Candy' })).toBeVisible();
    await expect(page.getByRole('heading', { name: /Generic staple Trainers/ })).toBeVisible();
    await expect(page.getByRole('heading', { name: /EX \/ Mega-EX/ })).toHaveCount(0);
    await expect(page.locator('main .shelf-cell')).not.toHaveCount(0);
    // Every shelf card is a Trainer: none carries an ex marker.
    await expect(page.locator('main .shelf-cell [title*=" EX "], main .shelf-cell [title*="Mega-EX"]')).toHaveCount(0);
    if (process.env.SHOP_SHOT) await page.screenshot({ path: `${OUT}/shop-${w}.png`, fullPage: true });
  });
}
