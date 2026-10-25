import { test, expect } from '@playwright/test';

// Holding a card on a phone zooms it; the browser's image menu must not open.
test('a long press on a card does not open the browser menu', async ({ page, request }) => {
  await request.post('/api/new', { data: { seed: 1, preset: 'fire' } });
  await page.goto('/?battle');
  const card = page.locator('.bt-root img').first();
  await expect(card).toBeVisible();

  const blocked = await card.evaluate((img) => !img.dispatchEvent(new MouseEvent('contextmenu', { bubbles: true, cancelable: true })));

  expect(blocked).toBe(true);
});
