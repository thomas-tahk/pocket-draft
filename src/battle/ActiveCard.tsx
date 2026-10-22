import { useState } from 'react';
import { useGameStore } from '../game/store';
import type { GameView, InPlayView, Move } from '../game/types';
import { litKey, type Choice } from './Choice';
import { CardFace, EnergyIcon } from './pieces';
import { EnergyText } from './energy';

// Tap your Active to open its card (spec F9): its attacks, the ones no legal
// move names greyed, its Ability, and Retreat. Tapping an attack uses it.
// Retreat closes the card and lights the Bench Pokémon you can switch in;
// tapping one retreats. Tap a Benched Pokémon with a usable Ability to open its
// card and use it (F10). An Ability that can be aimed at several Pokémon lights
// them the way Retreat does. What is usable comes only from legalMoves. Look
// and wording are the owner's (docs/questions/M1.md).

type Mode = 'closed' | 'card' | 'retreat' | 'aim';

// pos: the open card's Pokémon, 0 the Active, n Bench slot n-1.
export type Sheet = { mon: InPlayView; pos: number; moves: Move[]; pick: (m: Move) => void; retreat: () => void; close: () => void };

// The Bench lights up like a choice; Cancel backs out without retreating.
function retreatChoice(moves: Move[], pick: (m: Move) => void, cancel: () => void): Choice {
  const lit = new Map<string, { pick: () => void }>();
  for (const m of moves) if (m.type === 'Retreat') lit.set(litKey(0, m.benchIndex + 1), { pick: () => pick(m) });
  return { title: 'Choose a Pokémon to switch in', lit, buttons: [{ key: 'cancel', label: 'Cancel', pick: cancel }] };
}

// An Ability with several targets lights each Pokémon, yours or the opponent's, it can be aimed at.
function aimChoice(moves: Move[], pick: (m: Move) => void, cancel: () => void): Choice {
  const lit = new Map<string, { pick: () => void }>();
  for (const m of moves) if (m.type === 'UseAbility' && m.spot) lit.set(litKey(m.spot.player, m.spot.pos), { pick: () => pick(m) });
  return { title: 'Choose a Pokémon', lit, buttons: [{ key: 'cancel', label: 'Cancel', pick: cancel }] };
}

const abilityMoves = (moves: Move[], pos: number) => moves.filter((m) => m.type === 'UseAbility' && m.target === pos);

export function useActiveCard(view: GameView) {
  const dispatch = useGameStore((s) => s.dispatch);
  const [mode, setMode] = useState<Mode>('closed');
  const [pos, setPos] = useState(0);
  const you = view.players[0];
  const yourTurn = view.phase === 'main' && view.active === 0 && !view.pending && !!you.active;
  const shown: Mode = yourTurn ? mode : 'closed';
  const mon = pos === 0 ? you.active : you.bench[pos - 1];
  const close = () => setMode('closed');
  const pick = (m: Move) => {
    close();
    void dispatch(m);
  };
  const openAt = (at: number) => { setPos(at); setMode('card'); };
  const open = () => { if (yourTurn) openAt(0); };
  // A Benched Pokémon opens only when it has an Ability to use.
  const openBench = (slot: number) => { if (yourTurn && abilityMoves(you.legalMoves, slot + 1).length) openAt(slot + 1); };
  // One way to use it goes straight through; several ask which Pokémon.
  const useAbility = (m: Move) => (abilityMoves(you.legalMoves, pos).length > 1 ? setMode('aim') : pick(m));
  const sheet: Sheet | null = shown === 'card' && mon
    ? { mon, pos, moves: you.legalMoves, pick: (m) => (m.type === 'UseAbility' ? useAbility(m) : pick(m)), retreat: () => setMode('retreat'), close }
    : null;
  const choice = shown === 'retreat' ? retreatChoice(you.legalMoves, pick, close)
    : shown === 'aim' ? aimChoice(abilityMoves(you.legalMoves, pos), pick, close) : null;
  return { open, openBench, sheet, retreat: choice };
}

function AttackRow({ name, cost, damage, move, pick }: { name: string; cost: string[]; damage: number; move?: Move; pick: (m: Move) => void }) {
  return (
    <button className="bt-sheet-row" data-testid="sheet-attack" data-usable={!!move} disabled={!move} onClick={() => move && pick(move)}>
      <span className="bt-sheet-cost">{cost.map((t, i) => <EnergyIcon key={i} type={t} size={18} />)}</span>
      <span className="bt-sheet-name">{name}</span>
      <span className="bt-sheet-damage">{damage || ''}</span>
    </button>
  );
}

function AbilityRow({ mon, move, pick }: { mon: InPlayView; move?: Move; pick: (m: Move) => void }) {
  const ability = mon.card.ability;
  if (!ability) return null;
  return (
    <button className="bt-sheet-row bt-sheet-ability" data-testid="sheet-ability" data-usable={!!move} disabled={!move} onClick={() => move && pick(move)}>
      <span className="bt-sheet-name">{ability.name}</span>
      <span className="bt-sheet-effect"><EnergyText text={ability.effect} size={13} /></span>
    </button>
  );
}

export function ActiveSheet({ sheet }: { sheet: Sheet }) {
  const { mon, pos, moves, pick } = sheet;
  const isActive = pos === 0;
  const canRetreat = moves.some((m) => m.type === 'Retreat');
  const attackMove = (i: number) => moves.find((m) => m.type === 'UseAttack' && m.index === i);
  return (
    <div className="bt-sheet-backdrop" onClick={sheet.close}>
      <div className="bt-sheet" data-testid="active-sheet" role="dialog" aria-label={mon.card.name} onClick={(e) => e.stopPropagation()}>
        <CardFace card={mon.card} width={200} />
        {isActive && mon.card.attacks.map((a, i) => <AttackRow key={i} {...a} move={attackMove(i)} pick={pick} />)}
        <AbilityRow mon={mon} move={abilityMoves(moves, pos)[0]} pick={pick} />
        {isActive && (
          <button className="bt-sheet-row bt-sheet-retreat" data-testid="sheet-retreat" data-usable={canRetreat} disabled={!canRetreat} onClick={sheet.retreat}>
            <span className="bt-sheet-cost">{Array.from({ length: mon.card.retreatCost }, (_, i) => <EnergyIcon key={i} type="Colorless" size={18} />)}</span>
            <span className="bt-sheet-name">Retreat</span>
          </button>
        )}
      </div>
    </div>
  );
}
