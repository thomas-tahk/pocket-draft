import { test, expect, type APIRequestContext } from '@playwright/test';
import { FIRE } from './decks';

// A card the opponent plays flies from their face-down hand. It must start at
// the size of a card in that hand, not stretched to the whole hand zone (it
// used to start ~2.7x its size and cover the board).

const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function toYourFirstTurn(api: APIRequestContext) {
  const start = (await (await api.post('/api/new', { data: { seed: 7, bot: BOT, you: FIRE } })).json()) as { players: { hand: { id: string; stage: string }[] }[] };
  const basic = start.players[0].hand.find((c) => c.stage === 'Basic')!;
  await api.post('/api/move', { data: { type: 'SetupActive', player: 0, cardId: basic.id } });
  await api.post('/api/move', { data: { type: 'SetupReady', player: 0 } });
  for (let i = 0; i < 6; i++) {
    const v = await (await api.post('/api/bot')).json();
    if ((v.state ?? v).phase !== 'setup') break;
  }
}

test("the opponent's played cards start at hand-card size", async ({ page, request }) => {
  await toYourFirstTurn(request);
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await page.addInitScript(() => {
    const w = window as unknown as { starts: { zone: string; scale: number }[] };
    w.starts = [];
    const original = Element.prototype.animate;
    Element.prototype.animate = function (this: Element, frames: Keyframe[], options?: number | KeyframeAnimationOptions) {
      const first = (frames as Keyframe[])[0] ?? {};
      const zone = this.closest('[data-testid^="opp-"]')?.getAttribute('data-testid') ?? (this.closest('.bt-board') ? 'board' : '');
      if (first.scale !== undefined && !this.closest('[data-testid$="-points"]')) w.starts.push({ zone, scale: Number(first.scale) });
      return original.call(this, frames, options);
    };
  });
  await page.goto('/?battle');
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 10_000 });
  await page.getByTestId('end-turn').getByRole('button').click();
  const confirm = page.getByTestId('confirm-ok');
  if (await confirm.isVisible().catch(() => false)) await confirm.click();
  await expect(page.getByTestId('end-turn').getByRole('button')).toBeEnabled({ timeout: 20_000 });

  const starts = await page.evaluate(() => (window as unknown as { starts: { zone: string; scale: number }[] }).starts);
  const oppPlays = starts.filter((s) => s.zone === 'opp-bench' || s.zone === 'opp-active');
  expect(oppPlays.length, 'the bot played onto its board').toBeGreaterThan(0);
  expect(Math.max(...starts.map((s) => s.scale)), JSON.stringify(starts)).toBeLessThanOrEqual(1.5);
});
