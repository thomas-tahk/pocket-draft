import { useCallback, useEffect, useState } from 'react';
import type { GameView } from '../game/types';

// The setup screen's opening (spec S4, F11): the player picked to flip is
// named, the coin lands, then "You go first / second". Then the board is
// yours to place on, with Ready. The server already decided the flip; this
// only shows it. Wording and timing are placeholders (questions file).

const OPPONENT_FLIP_MS = 700; // the bot's flip runs on its own (F11)
const LANDED_MS = 900;
const ORDER_MS = 1400;

type Stage = 'flip' | 'landed' | 'order';

// Shown once per game, while setup is open and you haven't placed anything.
export function useSetupIntro(view: GameView | null) {
  const [done, setDone] = useState(false);
  const phase = view?.phase;
  useEffect(() => {
    if (phase && phase !== 'setup') setDone(false);
  }, [phase]);
  const you = view?.players[0];
  const fresh = !!you && !you.ready && !you.active && you.bench.every((b) => !b);
  const finish = useCallback(() => setDone(true), []);
  return { show: phase === 'setup' && fresh && !done, finish };
}

export function SetupIntro({ view, onDone }: { view: GameView; onDone: () => void }) {
  const youFlip = view.openingFlip.flipper === 0;
  const [stage, setStage] = useState<Stage>('flip');

  useEffect(() => {
    const next: Partial<Record<Stage, [Stage | 'done', number]>> = {
      landed: ['order', LANDED_MS],
      order: ['done', ORDER_MS],
      ...(youFlip ? {} : { flip: ['landed', OPPONENT_FLIP_MS] }),
    };
    const step = next[stage];
    if (!step) return;
    const t = setTimeout(() => (step[0] === 'done' ? onDone() : setStage(step[0] as Stage)), step[1]);
    return () => clearTimeout(t);
  }, [stage, youFlip, onDone]);

  const face = view.openingFlip.heads ? 'heads' : 'tails';
  const tap = youFlip && stage === 'flip' ? () => setStage('landed') : undefined;
  return (
    <div className="bt-overlay bt-setup-intro" data-testid="setup-intro" data-stage={stage}>
      <p>{youFlip ? 'You flip the coin' : 'Opponent flips the coin'}</p>
      <button className="bt-coin" data-testid="setup-coin" data-face={stage === 'flip' ? '' : face} disabled={!tap} onClick={tap}>
        {stage === 'flip' ? (tap ? 'Tap' : '') : face.toUpperCase()}
      </button>
      {stage === 'order' && <p data-testid="setup-order">{view.firstPlayer === 0 ? 'You go first' : 'You go second'}</p>}
    </div>
  );
}

// Ready (spec S5): lit once a legal SetupReady exists, i.e. your Active is placed.
export function ReadyButton({ view, onReady }: { view: GameView; onReady: () => void }) {
  const canReady = view.players[0].legalMoves.some((m) => m.type === 'SetupReady');
  return <button className="bt-endturn" data-testid="setup-ready" disabled={!canReady} onClick={onReady}>Ready</button>;
}
