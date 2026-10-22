import { useEffect, useRef, useState } from 'react';
import type { ActionView, GameView } from '../game/types';
import { CardFace } from './pieces';

// The bot's turn, played back one action at a time (spec F14): the bot moves
// one step every STEP_MS; each step shows its card and result and lights the
// spot it touched. Skip runs the rest of the bot's turn with no wait.

const STEP_MS = 550; // F14 "about 0.5 s each"; the exact pace is the owner's
const SLOT_NAMES = ['left', 'behind', 'right'];
const SILENT_KINDS = ['setup', 'setup_return', 'setup_move', 'ready', 'end_turn', 'concede']; // nothing to play back

export type Step = { index: number; action: ActionView };

const actorOf = (v: GameView) => (v.pending ? v.pending.player : v.active);

// Drives the bot (seat 1) whenever it is its move and tracks the step on show.
export function useOpponentTurn(view: GameView | null, botStuck: boolean, runBot: () => Promise<void>) {
  const [skipping, setSkipping] = useState(false);
  const [step, setStep] = useState<Step | null>(null);
  const seen = useRef<number | null>(null);
  // Setup is simultaneous: the bot places while you do (owner 2026-10-19 (8)).
  const botSetting = !!view && view.phase === 'setup' && !view.players[1].ready;
  const botToMove = !!view && view.phase !== 'over' && !botStuck && (botSetting || actorOf(view) === 1);

  useEffect(() => {
    if (!botToMove) return;
    const t = setTimeout(() => void runBot(), skipping ? 0 : STEP_MS);
    return () => clearTimeout(t);
  }, [view, botToMove, skipping, runBot]);

  // Each new bot action becomes the step on show. A skip lasts until the bot
  // hands control back, its last action included.
  useEffect(() => {
    if (!view) return;
    if (!botToMove) setSkipping(false);
    const before = seen.current ?? view.actions.length;
    seen.current = view.actions.length;
    if (view.actions.length <= before) return;
    const last = view.actions.length - 1;
    const action = view.actions[last];
    if (action.player !== 1) return;
    const shows = !skipping && view.phase === 'main' && !SILENT_KINDS.includes(action.kind);
    setStep(shows ? { index: last, action } : null);
  }, [view, botToMove, skipping]);

  // The last step stays up for its half second once the turn is handed back.
  useEffect(() => {
    if (!step || botToMove) return;
    const t = setTimeout(() => setStep(null), STEP_MS);
    return () => clearTimeout(t);
  }, [step, botToMove]);

  const skip = () => {
    setSkipping(true);
    setStep(null);
  };
  const canSkip = botToMove && view.phase === 'main' && !skipping;
  return { step, canSkip, skip };
}

// Data attributes for the board root that light the step's target (battle.css).
export function flashAttrs(step: Step | null): Record<string, string> {
  const t = step?.action.target;
  if (!t || t.player < 0) return {};
  const side = t.player === 0 ? 'my' : 'opp';
  if (t.pos === 0) return { 'data-flash-zone': `${side}-active` };
  return { 'data-flash-zone': `${side}-bench`, 'data-flash-slot': SLOT_NAMES[t.pos - 1] };
}

export function PlaybackStep({ step }: { step: Step }) {
  const lines = step.action.result.filter((l) => !l.startsWith('---'));
  return (
    <div className="bt-playback" data-testid="playback" data-step={step.index} data-kind={step.action.kind}>
      {step.action.card && <CardFace card={step.action.card} width={72} />}
      {lines.map((l, i) => <p key={i}>{l}</p>)}
    </div>
  );
}

export function SkipButton({ onSkip }: { onSkip: () => void }) {
  return <button className="bt-skip" onClick={onSkip}>Skip</button>;
}
