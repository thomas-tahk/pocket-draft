import { test, expect, type APIRequestContext } from '@playwright/test';
import { animationsSoFar, logAnimations, settled } from './fx-log';

// The board follows cards by a key made from their card id and copy number
// (src/battle/motion.ts). A card in play must keep its key when another copy
// of it enters or leaves the hand, or the wrong card moves: the Active would
// fly in from the hand when its twin is benched.

const DECK = ['A1-074', 'A1-075', 'A1a-020', 'A2-043', 'A1-062', 'A2b-015', 'B1-048', 'A1-058', 'A4-049', 'A2a-015'].flatMap((id) => [id, id]);
const TWIN = 'A4-049'; // seed 9 opens with two of it

async function activeWithTwinInHand(api: APIRequestContext) {
  await api.post('/api/new', { data: { seed: 9, you: DECK, bot: DECK } });
  await api.post('/api/move', { data: { type: 'SetupActive', player: 0, cardId: TWIN } });
}

test('benching a copy of your Active moves only the benched card', async ({ page, request }) => {
  await activeWithTwinInHand(request);
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await logAnimations(page);
  await page.goto('/?battle');
  await expect(page.getByTestId('setup-intro')).toHaveCount(0, { timeout: 8000 });
  await settled(page);
  const before = (await animationsSoFar(page)).length;

  const hand = (await page.getByTestId('my-hand').locator(`[data-card-id="${TWIN}"]`).first().boundingBox())!;
  const slot = (await page.getByTestId('my-bench').locator('[data-slot="left"]').boundingBox())!;
  await page.mouse.move(hand.x + 10, hand.y + hand.height / 2);
  await page.mouse.down();
  await page.mouse.move(slot.x + slot.width / 2, slot.y + slot.height / 2, { steps: 8 });
  await page.mouse.up();
  await expect(page.getByTestId('my-bench').locator('[data-slot="left"] img')).toHaveCount(1);

  const log = (await animationsSoFar(page)).slice(before);
  expect(log.some((a) => a.zone === 'my-bench' && a.key !== ''), 'the benched card slides in').toBe(true);
  expect(log.filter((a) => a.zone === 'my-active'), 'the Active stays put').toEqual([]);
});
