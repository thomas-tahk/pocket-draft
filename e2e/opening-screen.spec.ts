import { test, expect } from '@playwright/test';

test('fresh app shows Draft, no Continue run; Practice disabled', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('button', { name: 'Start draft' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Continue run' })).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Practice' })).toBeDisabled();
  await page.screenshot({ path: 'test-results/opening-fresh.png' });
});

test('active run: reload shows Continue run, leads to deck view', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  await expect(page.getByRole('button', { name: '▶ Play deck' })).toBeVisible();
  await page.reload();
  const cont = page.getByRole('button', { name: 'Continue run' });
  await expect(cont).toBeVisible();
  await expect(page.getByRole('button', { name: 'Practice' })).toBeEnabled();
  await page.screenshot({ path: 'test-results/opening-active.png' });
  await cont.click();
  await expect(page.getByRole('button', { name: '▶ Play deck' })).toBeVisible();
});

test('Past runs opens history', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Past runs' }).click();
  await expect(page.getByRole('button', { name: /back|close/i }).first()).toBeVisible();
});
