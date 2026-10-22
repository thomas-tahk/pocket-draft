import type { ActionView, GameView, InPlayView, PlayerView } from '../game/types';

// Board motion (M3-2, M3-3): when the board view changes, every card that
// changed place slides from where it was to where it is now (FLIP, with the
// Web Animations API); a card that just appeared comes from its origin (the
// deck, the opponent's hand); a played Trainer shows in the middle on its way
// to the discard; a knocked-out Pokémon fades toward the discard. The engine
// sends no per-card instance id, so a card's key is made here: player + card
// id + its count among that player's visible copies. Two copies of one card
// look the same, so which copy "moved" is never visible. Look and timing are
// placeholders for the owner (docs/questions/M3.md).

export const FLIP_MS = 220; // owner: each animation 400 ms at most, "doesn't have to be fancy"
const TRAINER_MS = 400;
const KO_MS = 300;
const FX = 'fx'; // Animation.id of every animation made here

// What a spot's key is, per zone of one player's side.
export type SideKeys = { hand: string[]; active: string; activeTool: string; bench: string[]; benchTool: string[] };

function counter(side: number) {
  const seen = new Map<string, number>();
  return (id: string) => {
    const n = seen.get(id) ?? 0;
    seen.set(id, n + 1);
    return `p${side}:${id}#${n}`;
  };
}

// Face-down cards (an opponent's hand, a hidden setup pick) have no identity; they are keyed by place.
export function sideKeys(side: number, p: PlayerView): SideKeys {
  const next = counter(side);
  const mon = (m: InPlayView | null, at: string) => (!m ? '' : m.faceDown ? `p${side}:down:${at}` : next(m.card.id));
  const tool = (m: InPlayView | null) => (m?.tool ? next(m.tool.id) : '');
  const hand = side === 0 ? p.hand.map((c) => next(c.id)) : Array.from({ length: p.handCount }, (_, i) => `p${side}:hand#${i}`);
  const active = mon(p.active, 'a');
  const activeTool = tool(p.active);
  const bench = p.bench.map((m, i) => mon(m, `b${i}`));
  return { hand, active, activeTool, bench, benchTool: p.bench.map(tool) };
}

export const viewKeys = (v: GameView): [SideKeys, SideKeys] => [sideKeys(0, v.players[0]), sideKeys(1, v.players[1])];

// The Stadium slot is shared, so its key has no owner.
export const stadiumKey = (v: GameView) => (v.stadium ? `st:${v.stadium.id}` : '');

export type Pt = { x: number; y: number };
export type Box = Pt & { w: number; h: number };
// Where a keyed element was: its centre and size, a copy of it to leave behind, and what it was.
export type Icon = Box & { type: string; clone: HTMLElement };
export type Snap = Box & { clone: HTMLElement; owner: 'my' | 'opp'; zone: 'hand' | 'mon' | 'other'; energy: Icon[] };

const ownerOf = (el: Element): 'my' | 'opp' => (el.closest('[data-testid^="opp-"]') ? 'opp' : 'my');
const zoneOf = (el: Element): Snap['zone'] => {
  if (el.closest('[data-testid$="-hand"]')) return 'hand';
  return el.closest('[data-testid$="-active"], [data-testid$="-bench"]') && !el.classList.contains('bt-tool') ? 'mon' : 'other';
};

export function boxOf(el: HTMLElement): Box {
  const r = el.getBoundingClientRect();
  return { x: r.left + r.width / 2, y: r.top + r.height / 2, w: el.offsetWidth, h: el.offsetHeight };
}

const STRIPPED = ['data-testid', 'data-fx', 'data-pick', 'data-zoom'];

// A copy with nothing a test or a tap could find it by.
function copyOf(el: HTMLElement): HTMLElement {
  const clone = el.cloneNode(true) as HTMLElement;
  for (const n of [clone, ...clone.querySelectorAll<HTMLElement>(STRIPPED.map((a) => `[${a}]`).join(','))]) {
    for (const a of STRIPPED) n.removeAttribute(a);
  }
  return clone;
}

export const energyIcons = (mon: HTMLElement) => [...mon.querySelectorAll<HTMLElement>('[data-energy]')];

export function snapshot(root: HTMLElement): Map<string, Snap> {
  const out = new Map<string, Snap>();
  root.querySelectorAll<HTMLElement>('[data-fx]').forEach((el) => {
    const energy = energyIcons(el).map((i) => ({ ...boxOf(i), type: i.dataset.energy!, clone: copyOf(i) }));
    out.set(el.dataset.fx!, { ...boxOf(el), clone: copyOf(el), owner: ownerOf(el), zone: zoneOf(el), energy });
  });
  return out;
}

export const reducedMotion = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches;

export function run(el: HTMLElement, frames: Keyframe[], ms: number): Animation {
  const a = el.animate(frames, { duration: ms, easing: 'ease-out' });
  a.id = FX;
  return a;
}

// Slides el from where it was to where it is now. Uses the individual
// `translate` and `scale` properties so a card's own `transform` is left alone.
export function slide(el: HTMLElement, from: Box, to: Box, ms = FLIP_MS): Animation | null {
  const dx = from.x - to.x;
  const dy = from.y - to.y;
  const s = from.w && to.w ? from.w / to.w : 1;
  if (Math.abs(dx) + Math.abs(dy) < 2 && Math.abs(s - 1) < 0.02) return null;
  return run(el, [{ translate: `${dx}px ${dy}px`, scale: String(s), zIndex: 40 }, { translate: '0px 0px', scale: '1', zIndex: 40 }], ms);
}

// Animations of this module that are still running finish at once.
export function settle(root: HTMLElement) {
  root.getAnimations({ subtree: true }).filter((a) => a.id === FX).forEach((a) => a.finish());
}

export const zoneBox = (root: HTMLElement, id: string, inner?: string): Box | null => {
  const z = root.querySelector<HTMLElement>(`[data-testid="${id}"]`);
  const el = inner ? z?.querySelector<HTMLElement>(inner) : z;
  return el ? boxOf(el) : null;
};

// Where a card that has just appeared came from: the deck for a draw, the
// opponent's hand for what they play (their hand is face down and unkeyed).
function originOf(root: HTMLElement, el: HTMLElement): Box | null {
  if (el.closest('[data-testid="my-hand"]')) return zoneBox(root, 'my-deck', '.bt-back');
  if (el.closest('[data-testid="opp-hand"]')) return zoneBox(root, 'opp-deck', '.bt-back');
  if (el.closest('[data-testid^="opp-"]')) return zoneBox(root, 'opp-hand');
  return null;
}

export type Ghost = { node: HTMLElement; at: Box };

// A copy of a card that lives on the board only while it flies, then is removed.
export function launch(root: HTMLElement, g: Ghost, frames: Keyframe[], ms: number, z = 45) {
  const r = root.getBoundingClientRect();
  Object.assign(g.node.style, {
    position: 'absolute', left: `${g.at.x - r.left - g.at.w / 2}px`, top: `${g.at.y - r.top - g.at.h / 2}px`,
    margin: '0', pointerEvents: 'none', zIndex: String(z),
  });
  root.appendChild(g.node);
  const a = run(g.node, frames, ms);
  a.onfinish = () => g.node.remove();
  a.oncancel = () => g.node.remove();
}

export const frame = (from: Pt, to: Pt, scale: number, extra: Keyframe = {}): Keyframe => ({
  translate: `${to.x - from.x}px ${to.y - from.y}px`, scale: String(scale), ...extra,
});

// A knocked-out Pokémon shrinks and fades toward its discard pile.
function knockOut(root: HTMLElement, snap: Snap) {
  const pile = zoneBox(root, `${snap.owner}-discard`, '.bt-discard');
  if (!pile) return;
  launch(root, { node: snap.clone, at: snap }, [frame(snap, snap, 1, { opacity: 1 }), frame(snap, pile, 0.4, { opacity: 0 })], KO_MS);
}

// The Pokémon an evolution lands on stays under the new card until it arrives.
function stayUnder(root: HTMLElement, snap: Snap) {
  launch(root, { node: snap.clone, at: snap }, [{ opacity: 1 }, { opacity: 1 }], FLIP_MS, 30);
}

function cardGhost(root: HTMLElement, action: ActionView): Ghost | null {
  if (!action.card) return null;
  const r = root.getBoundingClientRect();
  const w = (r.width / 390) * 96;
  const node = document.createElement('div');
  node.className = 'bt-card';
  node.style.setProperty('--w', '96');
  if (action.card.image) {
    const img = document.createElement('img');
    img.src = action.card.image;
    node.appendChild(img);
  } else node.textContent = action.card.name;
  return { node, at: { x: r.left + r.width / 2, y: r.top + r.height / 2, w, h: (w * 7) / 5 } };
}

// A Trainer shows in the middle for a moment, then goes to the discard pile.
function showTrainer(root: HTMLElement, action: ActionView, from: Box | null) {
  const ghost = cardGhost(root, action);
  const pile = zoneBox(root, `${action.player === 0 ? 'my' : 'opp'}-discard`, '.bt-discard');
  const start = from ?? zoneBox(root, action.player === 0 ? 'my-hand' : 'opp-hand');
  if (!ghost || !pile || !start) return;
  const c = ghost.at;
  launch(root, ghost, [
    frame(c, start, start.w / c.w, { opacity: 0.6, offset: 0 }),
    frame(c, c, 1, { opacity: 1, offset: 0.3 }),
    frame(c, c, 1, { opacity: 1, offset: 0.65 }),
    frame(c, pile, pile.w / c.w, { opacity: 0.8, offset: 1 }),
  ], TRAINER_MS);
}

export function popNewPips(root: HTMLElement, before: GameView, now: GameView) {
  (['my', 'opp'] as const).forEach((side, p) => {
    const pips = root.querySelectorAll<HTMLElement>(`[data-testid="${side}-points"] [data-testid="pip"]`);
    for (let i = before.players[p].points; i < now.players[p].points; i++) {
      if (pips[i]) run(pips[i], [{ scale: '1.8' }, { scale: '1' }], FLIP_MS);
    }
  });
}

export type Change = { root: HTMLElement; was: Map<string, Snap>; after: Map<string, Snap>; before: GameView; now: GameView };

// A Stadium comes from the hand of whoever played it.
function stadiumOrigin(root: HTMLElement, { was, after, now, before }: Change, key: string): Box | null {
  const player = now.actions.slice(before.actions.length).find((a) => a.card?.kind === 'stadium')?.player ?? 0;
  const owner = player === 0 ? 'my' : 'opp';
  const id = key.slice(3);
  const gone = [...was].find(([k, s]) => !after.has(k) && s.zone === 'hand' && s.owner === owner && k.includes(`:${id}#`));
  return gone?.[1] ?? zoneBox(root, `${owner}-hand`);
}

// Every keyed element that is still there slides; one that is new comes from its origin.
export function moveCards(change: Change) {
  const { root, was, before } = change;
  root.querySelectorAll<HTMLElement>('[data-fx]').forEach((node) => {
    const key = node.dataset.fx!;
    // An opponent's setup picks are revealed (a new key) when both are Ready: that is not a play.
    const reveal = before.phase === 'setup' && !key.includes(':down:') && node.closest('[data-testid^="opp-"]');
    const from = was.get(key) ?? (key.startsWith('st:') ? stadiumOrigin(root, change, key) : reveal ? null : originOf(root, node));
    if (from) slide(node, from, boxOf(node));
  });
}

// Cards that left the board: a played Trainer, a knocked-out Pokémon, the Pokémon an evolution covers.
export function removedCards({ root, was, after, before, now }: Change) {
  const fresh = now.actions.slice(before.actions.length);
  const evolved = fresh.some((a) => a.kind === 'evolve');
  const played = fresh.filter((a) => a.kind === 'play_trainer' && (a.card?.kind === 'item' || a.card?.kind === 'supporter'));
  for (const a of played) {
    const owner = a.player === 0 ? 'my' : 'opp';
    const gone = [...was].find(([k, s]) => !after.has(k) && s.zone === 'hand' && s.owner === owner && k.includes(`:${a.card!.id}#`));
    showTrainer(root, a, a.player === 0 ? (gone?.[1] ?? null) : null);
  }
  for (const [key, snap] of was) {
    if (after.has(key) || snap.zone !== 'mon') continue;
    const p = snap.owner === 'my' ? 0 : 1;
    if (now.players[p].discardCount > before.players[p].discardCount) knockOut(root, snap);
    else if (evolved) stayUnder(root, snap);
  }
}
