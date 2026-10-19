import { useState } from 'react';
import { useGameStore } from '../game/store';
import type { GameView, InPlayView, Move } from '../game/types';
import type { Choice } from './Choice';
import { CardFace, EnergyIcon } from './pieces';

// Tap your Active to open its card (spec F9): its attacks, the ones no legal
// move names greyed, and Retreat. Tapping an attack uses it. Retreat closes the
// card and lights the Bench Pokémon you can switch in; tapping one retreats.
// What is usable comes only from legalMoves. Look and wording are the owner's
// (docs/questions/M1.md).

type Mode = 'closed' | 'card' | 'retreat';

export type Sheet = { mon: InPlayView; moves: Move[]; pick: (m: Move) => void; retreat: () => void; close: () => void };

// The Bench lights up like a choice; Cancel backs out without retreating.
function retreatChoice(moves: Move[], pick: (m: Move) => void, cancel: () => void): Choice {
  const lit = new Map<number, { pick: () => void }>();
  for (const m of moves) if (m.type === 'Retreat') lit.set(m.benchIndex + 1, { pick: () => pick(m) });
  return { title: 'Choose a Pokémon to switch in', lit, buttons: [{ key: 'cancel', label: 'Cancel', pick: cancel }] };
}

export function useActiveCard(view: GameView) {
  const dispatch = useGameStore((s) => s.dispatch);
  const [mode, setMode] = useState<Mode>('closed');
  const you = view.players[0];
  const yourTurn = view.phase === 'main' && view.active === 0 && !view.pending && !!you.active;
  const shown: Mode = yourTurn ? mode : 'closed';
  const close = () => setMode('closed');
  const pick = (m: Move) => {
    close();
    void dispatch(m);
  };
  const open = () => { if (yourTurn) setMode('card'); };
  const sheet: Sheet | null = shown === 'card' && you.active
    ? { mon: you.active, moves: you.legalMoves, pick, retreat: () => setMode('retreat'), close }
    : null;
  const retreat = shown === 'retreat' ? retreatChoice(you.legalMoves, pick, close) : null;
  return { open, sheet, retreat };
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

export function ActiveSheet({ sheet }: { sheet: Sheet }) {
  const { mon, moves, pick } = sheet;
  const canRetreat = moves.some((m) => m.type === 'Retreat');
  const attackMove = (i: number) => moves.find((m) => m.type === 'UseAttack' && m.index === i);
  return (
    <div className="bt-sheet-backdrop" onClick={sheet.close}>
      <div className="bt-sheet" data-testid="active-sheet" role="dialog" aria-label={mon.card.name} onClick={(e) => e.stopPropagation()}>
        <CardFace card={mon.card} width={200} />
        {mon.card.attacks.map((a, i) => <AttackRow key={i} {...a} move={attackMove(i)} pick={pick} />)}
        <button className="bt-sheet-row bt-sheet-retreat" data-testid="sheet-retreat" data-usable={canRetreat} disabled={!canRetreat} onClick={sheet.retreat}>
          <span className="bt-sheet-cost">{Array.from({ length: mon.card.retreatCost }, (_, i) => <EnergyIcon key={i} type="Colorless" size={18} />)}</span>
          <span className="bt-sheet-name">Retreat</span>
        </button>
      </div>
    </div>
  );
}
