import type { GameView, InPlayView } from '../game/types';
import { boxOf, energyIcons, launch, run, slide, viewKeys, zoneBox, type Change, type Pt } from './motion';

// What a move did, shown where it happened (M3-4): Energy flies from the
// Energy Zone to the Pokémon it goes to, an attacker nudges toward its target,
// a number pops where HP changed, a Special Condition flashes, discarded Energy
// fades off. Everything is found by diffing the old and new view per Pokémon
// key (motion.ts). The HP bar drains by a CSS transition (battle.css).
// Look and timing are placeholders for the owner (docs/questions/M3.md).

const POP_MS = 400;
const NUDGE_MS = 200;
const FLASH_MS = 250;
const FADE_MS = 220;
const CONDITIONS = ['poisoned', 'burned', 'asleep', 'paralyzed', 'confused'] as const;

function monsByKey(v: GameView): Map<string, InPlayView> {
  const keys = viewKeys(v);
  const out = new Map<string, InPlayView>();
  v.players.forEach((p, i) => {
    if (p.active && keys[i].active && !p.active.faceDown) out.set(keys[i].active, p.active);
    p.bench.forEach((m, j) => { if (m && keys[i].bench[j] && !m.faceDown) out.set(keys[i].bench[j], m); });
  });
  return out;
}

// A number rises from a point on the board: red for damage, green for healing.
export function showNumber(root: HTMLElement, at: Pt, text: string, kind: 'damage' | 'heal') {
  const u = root.getBoundingClientRect().width / 390;
  const node = document.createElement('div');
  node.className = 'bt-pop';
  node.dataset.kind = kind;
  node.textContent = text;
  const box = { x: at.x, y: at.y - 12 * u, w: 64 * u, h: 28 * u };
  Object.assign(node.style, { width: `${box.w}px`, height: `${box.h}px` });
  launch(root, { node, at: box }, [
    { translate: '0 0', scale: '0.5', opacity: 0 },
    { scale: '1.2', opacity: 1, offset: 0.25 },
    { translate: `0 ${-14 * u}px`, scale: '1', opacity: 1, offset: 0.7 },
    { translate: `0 ${-22 * u}px`, opacity: 0 },
  ], POP_MS);
}

function hpChange(root: HTMLElement, el: HTMLElement, was: InPlayView, now: InPlayView) {
  const delta = now.remainingHp - was.remainingHp;
  if (delta === 0) return;
  showNumber(root, boxOf(el), delta < 0 ? `-${-delta}` : `+${delta}`, delta < 0 ? 'damage' : 'heal');
}

function energyChange(change: Change, key: string, el: HTMLElement, was: InPlayView, now: InPlayView) {
  const { root, was: snaps, after } = change;
  for (const type of new Set([...Object.keys(was.energy), ...Object.keys(now.energy)])) {
    const diff = (now.energy[type] ?? 0) - (was.energy[type] ?? 0);
    if (diff > 0) flyEnergyIn(root, after.get(key)?.owner ?? 'my', energyIcons(el).filter((i) => i.dataset.energy === type).slice(-diff));
    if (diff < 0) fadeEnergyOut(root, snaps.get(key)?.energy.filter((i) => i.type === type).slice(diff) ?? []);
  }
}

// Attached Energy comes from the Energy Zone.
function flyEnergyIn(root: HTMLElement, side: 'my' | 'opp', icons: HTMLElement[]) {
  const from = zoneBox(root, `${side}-energy`, '[data-testid="energy-current"]');
  if (from) icons.forEach((icon) => slide(icon, from, boxOf(icon)));
}

// Discarded Energy fades where it sat.
function fadeEnergyOut(root: HTMLElement, icons: { clone: HTMLElement; x: number; y: number; w: number; h: number }[]) {
  for (const icon of icons) launch(root, { node: icon.clone, at: icon }, [{ opacity: 1, scale: '1' }, { opacity: 0, scale: '0.5' }], FADE_MS);
}

function conditionFlash(el: HTMLElement, was: InPlayView, now: InPlayView) {
  if (CONDITIONS.some((c) => now[c] && !was[c])) run(el, [{ filter: 'none' }, { filter: 'brightness(1.9) saturate(1.6)', offset: 0.4 }, { filter: 'none' }], FLASH_MS);
}

// The attacker leans toward the Pokémon it hit and back.
function nudge(root: HTMLElement, player: number) {
  const [mine, theirs] = player === 0 ? ['my', 'opp'] : ['opp', 'my'];
  const attacker = root.querySelector<HTMLElement>(`[data-testid="${mine}-active"] [data-fx]`);
  const target = zoneBox(root, `${theirs}-active`);
  if (!attacker || !target) return;
  const from = boxOf(attacker);
  const len = Math.hypot(target.x - from.x, target.y - from.y) || 1;
  const reach = (root.getBoundingClientRect().width / 390) * 16;
  const move = `${((target.x - from.x) / len) * reach}px ${((target.y - from.y) / len) * reach}px`;
  run(attacker, [{ translate: '0 0' }, { translate: move, offset: 0.4 }, { translate: '0 0' }], NUDGE_MS);
}

export function resultsLand(change: Change) {
  const { root, was, after, before, now } = change;
  const olds = monsByKey(before);
  const els = new Map([...root.querySelectorAll<HTMLElement>('[data-fx]')].map((e) => [e.dataset.fx!, e]));
  for (const [key, mon] of monsByKey(now)) {
    const old = olds.get(key);
    const el = els.get(key);
    if (!old || !el) continue;
    hpChange(root, el, old, mon);
    energyChange(change, key, el, old, mon);
    conditionFlash(el, old, mon);
  }
  const attacks = now.actions.slice(before.actions.length).filter((a) => a.kind === 'attack');
  for (const a of attacks) nudge(root, a.player);
  for (const [key, snap] of was) {
    const p = snap.owner === 'my' ? 0 : 1;
    const knockedOut = !after.has(key) && snap.zone === 'mon' && now.players[p].discardCount > before.players[p].discardCount;
    const old = olds.get(key);
    if (!knockedOut || !old || old.remainingHp <= 0) continue;
    // A knocked-out Pokémon shows the hit it took, not just the HP it had left.
    const hit = key === viewKeys(before)[p].active ? attacks.find((a) => a.player !== p)?.damage : undefined;
    showNumber(root, snap, `-${hit ?? old.remainingHp}`, 'damage');
  }
}
