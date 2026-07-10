import type { GameView, InPlayView } from './types';

export type ChangeSet = { hp: Set<string>; energy: Set<string> };

function slotKey(player: number, slot: string): string {
  return `${player}:${slot}`;
}

// Which on-board slots changed HP or total energy between two views. A null prev
// (fresh load / new game) reports no changes so nothing flashes on first render.
export function changedSlots(prev: GameView | null, next: GameView): ChangeSet {
  const hp = new Set<string>();
  const energy = new Set<string>();
  if (!prev) return { hp, energy };

  const compare = (p: number, slot: string, a: InPlayView | null, b: InPlayView | null) => {
    if (!a || !b) return;
    if (a.remainingHp !== b.remainingHp) hp.add(slotKey(p, slot));
    if (a.totalEnergy !== b.totalEnergy) energy.add(slotKey(p, slot));
  };

  for (let p = 0; p < 2; p++) {
    compare(p, 'active', prev.players[p].active, next.players[p].active);
    const pb = prev.players[p].bench;
    const nb = next.players[p].bench;
    for (let i = 0; i < nb.length; i++) compare(p, String(i), pb[i] ?? null, nb[i]);
  }
  return { hp, energy };
}
