import type { Move } from '../game/types';

// L5: a card is usable exactly when some legal move uses it. Hand cards are
// named by cardId (play, setup) or handCardId (evolve); the Active by an attack
// or a retreat; a Bench Pokémon by being chosen as the new Active.
export type Usable = { hand: Set<string>; active: boolean; bench: Set<number> };

export function usableCards(moves: Move[]): Usable {
  const out: Usable = { hand: new Set(), active: false, bench: new Set() };
  for (const m of moves) {
    if ('cardId' in m && m.cardId) out.hand.add(m.cardId);
    if (m.type === 'Evolve') out.hand.add(m.handCardId);
    if (m.type === 'UseAttack' || m.type === 'Retreat') out.active = true;
    if (m.type === 'UseAbility') { if (m.target === 0) out.active = true; else out.bench.add(m.target - 1); }
    if (m.type === 'ChooseNewActive') out.bench.add(m.benchIndex);
  }
  return out;
}
