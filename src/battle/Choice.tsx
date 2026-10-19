import type { ReactNode } from 'react';
import type { GameView, Move } from '../game/types';
import { EnergyIcon } from './pieces';

// The choice overlay (spec F8, F13). While a prompt waits on you, the answers
// are exactly your legal moves (plus Concede). Each answer lights the card it
// names; answers that name no card become buttons on the prompt panel. Setup
// prompts are the setup screen's, not this overlay's.

// A lit spot on your side: 0 is your Active, n is Bench slot n-1. `pick` is set
// when tapping the card answers the prompt; unset when the card is only involved.
export type Spot = { pick?: () => void };

export type Choice = {
  title: string;
  lit: Map<number, Spot>;
  buttons: { key: string; label: ReactNode; pick: () => void }[];
};

// Placeholder wording for prompts F13 doesn't word (see docs/questions/M1.md).
const TITLES: Record<string, string> = {
  new_active: 'Choose your new Active',
  retreat_energy: 'Choose energy to discard',
};

function energyLabel(energy: string[]): ReactNode {
  return energy.map((type, i) => <EnergyIcon key={i} type={type} size={18} />);
}

function addAnswer(choice: Choice, move: Move, pick: () => void) {
  const key = `${choice.buttons.length}`;
  switch (move.type) {
    case 'ChooseNewActive':
      choice.lit.set(move.benchIndex + 1, { pick });
      return;
    case 'RetreatDiscard': // the energy comes off your Active
      choice.lit.set(0, {});
      choice.buttons.push({ key, label: energyLabel(move.energy), pick });
      return;
    default: // a prompt kind this overlay doesn't know yet still gets answered
      choice.buttons.push({ key, label: move.type, pick });
  }
}

export function choiceFor(view: GameView, viewer: number, send: (m: Move) => void): Choice | null {
  const prompt = view.pending;
  if (!prompt || prompt.player !== viewer || prompt.kind === 'setup') return null;
  const choice: Choice = { title: TITLES[prompt.kind] ?? 'Choose', lit: new Map(), buttons: [] };
  for (const move of view.players[viewer].legalMoves) {
    if (move.type !== 'Concede') addAnswer(choice, move, () => send(move));
  }
  return choice;
}

// The dim layer and prompt panel. Lit cards sit above the dim layer (see
// .bt-lit in battle.css); everything else is under it and can't be tapped.
export function ChoiceOverlay({ choice }: { choice: Choice }) {
  return (
    <>
      <div className="bt-dim" data-testid="choice-dim" />
      <div className="bt-prompt">
        <div data-testid="choice-prompt">{choice.title}</div>
        {choice.buttons.length > 0 && (
          <div className="bt-prompt-buttons">
            {choice.buttons.map((b) => (
              <button key={b.key} className="bt-prompt-button" data-testid="choice-button" onClick={b.pick}>{b.label}</button>
            ))}
          </div>
        )}
      </div>
    </>
  );
}
