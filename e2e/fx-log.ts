import type { Page } from '@playwright/test';

// Records every Web Animations call the board makes (src/battle/motion.ts), so
// a test can ask "did the card that moved get an animation?" without racing
// the 200 ms it lasts. Call before page.goto.

export type Fx = {
  zone: string; // the board zone the animated element is in ('' for a flying copy outside any)
  key: string; // its data-fx key ('' for a copy or a pip)
  copy: boolean; // a copy that lives on the board only while it flies
  props: string[]; // the CSS properties animated
  fades: boolean; // ends transparent
  stops: number; // keyframes
  energy: string; // the Energy type, for an attached Energy icon
  text: string; // a number that pops
};

const ZONES = ['my-hand', 'opp-hand', 'my-active', 'opp-active', 'my-bench', 'opp-bench', 'my-points', 'opp-points', 'stadium', 'my-energy', 'opp-energy'];

export async function logAnimations(page: Page) {
  await page.addInitScript((zones) => {
    const w = window as unknown as { fxLog: unknown[]; fxTransitions: string[] };
    w.fxLog = [];
    w.fxTransitions = [];
    new MutationObserver((muts) => muts.forEach((m) => (m.target as Element).getAnimations().forEach((a) => a instanceof CSSTransition && w.fxTransitions.push(`${a.transitionProperty}@${(m.target as Element).parentElement?.className}`)))).observe(document, { attributes: true, subtree: true, attributeFilter: ['style'] });
    const original = Element.prototype.animate;
    Element.prototype.animate = function (this: Element, frames: Keyframe[], options?: number | KeyframeAnimationOptions) {
      const list = frames as Keyframe[];
      const zone = zones.find((z) => this.closest(`[data-testid="${z}"]`)) ?? '';
      w.fxLog.push({
        zone,
        key: (this as HTMLElement).dataset?.fx ?? '',
        copy: !!this.closest('.bt-board') && (this as HTMLElement).style.position === 'absolute' && !zone,
        props: Object.keys(list[0] ?? {}).filter((k) => k !== 'offset'),
        fades: Number(list[list.length - 1]?.opacity ?? 1) < 0.5,
        stops: list.length,
        energy: (this as HTMLElement).dataset?.energy ?? '',
        text: (this as HTMLElement).classList?.contains('bt-pop') ? this.textContent : '',
      });
      return original.call(this, frames, options);
    };
  }, ZONES);
}

export const animationsSoFar = (page: Page) => page.evaluate(() => (window as unknown as { fxLog: Fx[] }).fxLog);

// Waits until nothing from the board is still moving.
export const settled = (page: Page) =>
  page.evaluate(() => Promise.all(document.getAnimations().filter((a) => a.id === 'fx').map((a) => a.finished.catch(() => {}))));

export const transitionsSoFar = (page: Page) => page.evaluate(() => (window as unknown as { fxTransitions: string[] }).fxTransitions);
