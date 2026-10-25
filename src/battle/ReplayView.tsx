import { useEffect, useState } from 'react';
import type { GameView } from '../game/types';
import type { RunBattle } from '../stores/run';
import { Board } from './Board';
import { captionFor } from './replayCaption';
import { buildReplayViews } from './replayViews';
import './battle.css';
import './replay.css';

const PLAY_MS = 700; // pace of Play; the owner's to tune

// Replay (#119): a finished battle stepped through on the same board, read-only.
// Wording is placeholder for the owner.
export function ReplayView({ battle, onExit }: { battle: RunBattle; onExit: () => void }) {
  const [views, setViews] = useState<GameView[] | null>(null);
  const [failed, setFailed] = useState(false);
  const [step, setStep] = useState(0);
  const [playing, setPlaying] = useState(false);

  useEffect(() => {
    const leave = new AbortController();
    buildReplayViews(battle, leave.signal).then(
      (v) => !leave.signal.aborted && setViews(v),
      () => !leave.signal.aborted && setFailed(true),
    );
    return () => leave.abort();
  }, [battle]);

  const last = (views?.length ?? 1) - 1;
  useEffect(() => {
    if (!playing) return;
    if (step >= last) { setPlaying(false); return; }
    const t = setTimeout(() => setStep((s) => s + 1), PLAY_MS);
    return () => clearTimeout(t);
  }, [playing, step, last]);

  if (failed) return <div className="bt-root"><div className="rp-msg" role="alert" data-testid="replay-error">This replay no longer matches its battle. <button className="rp-btn" data-testid="replay-exit" onClick={onExit}>Exit</button></div></div>;
  if (!views) return <div className="bt-root"><div className="rp-msg" role="status">Loading replay…</div></div>;

  const view = views[step];
  const caption = captionFor(view, step);
  const go = (to: number) => { setPlaying(false); setStep(Math.max(0, Math.min(last, to))); };
  const toggle = () => { if (step >= last) setStep(0); setPlaying(!playing); };
  return (
    <div className="bt-root rp-root" onContextMenu={(e) => e.preventDefault()}>
      <Board view={view} instant onEndTurn={() => {}}>
        <div className="rp-top">
          <span className="rp-count" data-testid="replay-step" aria-live="polite">Step {step} / {last}</span>
        </div>
        <div className="rp-bar" role="group" aria-label="Replay controls">
          <button className="rp-btn" data-testid="replay-exit" onClick={onExit}>Exit</button>
          <button className="rp-btn" data-testid="replay-back" disabled={step === 0} onClick={() => go(step - 1)} aria-label="Back one step">◀</button>
          <button className="rp-btn" data-testid="replay-play" onClick={toggle} aria-label={playing ? 'Pause' : 'Play'}>{playing ? '❚❚' : '▶'}</button>
          <button className="rp-btn" data-testid="replay-forward" disabled={step === last} onClick={() => go(step + 1)} aria-label="Forward one step">▶▶</button>
        </div>
        <p className="rp-caption" data-testid="replay-narration">{caption}</p>
      </Board>
    </div>
  );
}
