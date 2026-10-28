import { test, expect, type Page, type Request } from '@playwright/test';
import { draftTwenty } from './helpers/draft';

// A non-default Energy Zone choice rides along with the run: Rematch and a replay rebuild
// send it to /api/new again. Three chosen types show the cap hint.

const isNewGame = (req: Request) => req.url().endsWith('/api/new') && req.method() === 'POST';
const chip = (page: Page, type: string) => page.getByTestId('energy-picker').getByRole('button', { name: type, exact: true });
const ALL = ['Grass', 'Fire', 'Water', 'Lightning', 'Psychic', 'Fighting', 'Darkness', 'Metal'];

async function choose(page: Page, want: string[]) {
  const pressed = async (t: string) => (await chip(page, t).getAttribute('aria-pressed')) === 'true';
  for (const t of ALL) {
    if (want.includes(t) || !(await pressed(t))) continue;
    const count = (await Promise.all(ALL.map(pressed))).filter(Boolean).length;
    if (count === 1) await chip(page, want.find((w) => w !== t)!).click();
    await chip(page, t).click();
  }
  for (const t of want) if (!(await pressed(t))) await chip(page, t).click();
}

const energyOf = async (req: Promise<Request>) => [...(((await req).postDataJSON() as { energy?: string[] }).energy ?? [])].sort();

async function concede(page: Page) {
  await expect(page.getByTestId('setup-intro')).toHaveCount(0, { timeout: 10_000 });
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();
  await expect(page.getByTestId('end-screen')).toBeVisible();
}

test('the chosen types reach the rematch and the replay rebuild; 3 chosen disables the rest', async ({ page }) => {
  test.setTimeout(150_000);
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  await draftTwenty(page);
  await page.getByRole('button', { name: 'Skip shop' }).click();
  const drafted = page.locator('section', { has: page.getByRole('heading', { name: /^Your draft/ }) });
  const plus = drafted.getByRole('button', { name: '+' });
  for (let i = 0; i < (await plus.count()); i++) {
    for (let c = 0; c < 2 && (await plus.nth(i).isEnabled()); c++) await plus.nth(i).click();
  }

  const want = ['Fire', 'Lightning', 'Metal'];
  await choose(page, want);
  const picker = page.getByTestId('energy-picker');
  await expect(picker.locator('button[aria-pressed="true"]')).toHaveCount(3);
  await expect(picker.getByTestId('energy-count')).toHaveText('3/3');
  // Owner 2026-10-28 (item 6): no visible cap line any more; the disabled chips say it.
  await expect(page.getByTestId('energy-cap-hint')).toHaveCount(0);
  await expect(picker.locator('button:disabled')).toHaveCount(5);
  await page.getByRole('button', { name: 'Finalize deck' }).click();

  const first = page.waitForRequest(isNewGame);
  await page.getByRole('button', { name: /Next battle/ }).click();
  expect(await energyOf(first)).toEqual(want);
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await concede(page);

  const end = page.getByTestId('end-screen');
  const rematch = page.waitForRequest(isNewGame);
  await end.getByRole('button', { name: 'Rematch' }).click();
  expect(await energyOf(rematch)).toEqual(want);
  await expect(end).toBeHidden();
  await concede(page);

  // The recorded battle replays with the same types: the rebuild's /api/new carries them.
  const replay = page.waitForRequest(isNewGame);
  await page.goto('/?replay=0');
  expect(await energyOf(replay)).toEqual(want);
});
