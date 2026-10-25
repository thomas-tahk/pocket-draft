import { useState } from 'react';
import type { GameView } from '../game/types';

// End Turn's confirm (F12), the Concede menu (F18) and the end screen (F19).
// All wording here is a placeholder for the owner (docs/questions/M1.md).

export type Confirm = { text: string; okLabel: string; onOk: () => void };

export function ConfirmDialog({ confirm, onClose }: { confirm: Confirm; onClose: () => void }) {
  return (
    <div className="bt-overlay bt-overlay-top" role="dialog" aria-modal="true" data-testid="confirm">
      <div className="bt-dialog">
        <p>{confirm.text}</p>
        <div className="bt-dialog-buttons">
          <button className="bt-btn" data-testid="confirm-cancel" onClick={onClose}>Cancel</button>
          <button className="bt-btn bt-btn-main" data-testid="confirm-ok" onClick={() => { onClose(); confirm.onOk(); }}>{confirm.okLabel}</button>
        </div>
      </div>
    </div>
  );
}

// What End Turn would throw away, or null if nothing (F12: one prompt names both).
export function endTurnWarning(view: GameView): string | null {
  const moves = view.players[0].legalMoves;
  const energy = moves.some((m) => m.type === 'AttachEnergy');
  const attack = moves.some((m) => m.type === 'UseAttack');
  if (!energy && !attack) return null;
  const what = energy && attack ? 'attach energy and attack' : energy ? 'attach energy' : 'attack';
  return `You can still ${what}. End your turn?`;
}

export function GameMenu({ onConcede }: { onConcede: () => void }) {
  const [open, setOpen] = useState(false);
  return (
    <div className="bt-menu">
      <button className="bt-menu-button" aria-label="Menu" data-testid="menu" onClick={() => setOpen(!open)}>☰</button>
      {open && (
        <button className="bt-btn bt-menu-item" data-testid="menu-concede" onClick={() => { setOpen(false); onConcede(); }}>Concede</button>
      )}
    </div>
  );
}

const RESULT = { win: 'Win', loss: 'Loss', tie: 'Tie', '': '' } as const;
const REASON = { points: '3 points', no_pokemon: 'No Pokémon left', concede: 'Concede', turn_cap: 'Turn cap', '': '' } as const;

export function EndScreen({ view, canRematch, onRematch, onBack }: { view: GameView; canRematch: boolean; onRematch: () => void; onBack: () => void }) {
  const [you, opp] = view.players;
  return (
    <div className="bt-overlay" data-testid="end-screen">
      <div className="bt-dialog bt-end">
        <h1 data-testid="end-result">{RESULT[view.outcome]}</h1>
        <p data-testid="end-reason">{REASON[view.reason]}</p>
        <p data-testid="end-score">You {you.points} – {opp.points} Opponent</p>
        <div className="bt-dialog-buttons">
          <button className="bt-btn" onClick={onBack}>Back</button>
          {canRematch && <button className="bt-btn bt-btn-main" onClick={onRematch}>Rematch</button>}
        </div>
      </div>
    </div>
  );
}
