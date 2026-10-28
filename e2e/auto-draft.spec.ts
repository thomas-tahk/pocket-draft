import { test, expect, type Request } from '@playwright/test';

// Test shortcut: Auto-draft skips the draft, lands on review with a legal 20,
// and Play deck starts a battle with exactly that deck.

const isNewGame = (req: Request) => req.url().endsWith('/api/new') && req.method() === 'POST';

test('Auto-draft lands on review and plays its 20', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Auto-draft (test)' }).click();
  const play = page.getByRole('button', { name: 'Next battle' });
  await expect(play).toBeVisible();

  const [req] = await Promise.all([page.waitForRequest(isNewGame), play.click()]);
  const sent = (req.postDataJSON() as { you: string[] }).you;
  expect(sent).toHaveLength(20);
  await expect(page.getByTestId('setup-intro')).toBeVisible();
});
