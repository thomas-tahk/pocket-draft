import { useState } from 'react';
import { HISTORY_CAP } from '../stores/draftStore';
import { onEscape, useDialog } from './useDialog';

type Props = {
  cardCount: number;
  hasActiveRun: boolean;
  hasFinishedRun: boolean;
  runScore: { wins: number; losses: number } | null;
  canPractice: boolean;
  historyCount: number;
  onDraft: () => void;
  onContinue: () => void;
  onPractice: () => void;
  onPastRuns: () => void;
  onAutoDraft: () => void;
};

export function HomeView(p: Props) {
  const [confirming, setConfirming] = useState(false);
  return (
    <main className="home">
      <h1>Pocket Draft</h1>
      <p className="home-note">{p.cardCount} draftable cards. Draft, shop, build a 20-card deck, then play a run.</p>
      {p.hasActiveRun ? (
        <button className="home-primary" onClick={p.onContinue}>Continue run</button>
      ) : (
        <button className="home-primary" onClick={p.onDraft}>Start draft</button>
      )}
      <div className="home-secondary">
        {p.hasFinishedRun && <button onClick={p.onContinue}>View last run</button>}
        <button onClick={p.onPractice} disabled={!p.canPractice} aria-describedby="practice-reason">
          Practice
        </button>
        {!p.canPractice && (
          <span id="practice-reason" className="home-note">Finish a draft first to practice with your deck.</span>
        )}
        <button onClick={p.onPastRuns}>
          Past runs{p.historyCount > 0 ? ` (${p.historyCount} / ${HISTORY_CAP})` : ''}
        </button>
      </div>
      {p.hasActiveRun && !confirming && (
        <button id="home-new-draft" className="home-link" onClick={() => setConfirming(true)}>Start a new draft instead</button>
      )}
      {p.hasActiveRun && confirming && p.runScore && (
        <EndRunConfirm score={p.runScore} onDraft={p.onDraft} onKeep={() => setConfirming(false)} />
      )}
      <button className="home-link" onClick={p.onAutoDraft}>Auto-draft (test)</button>
    </main>
  );
}

function EndRunConfirm({ score, onDraft, onKeep }: { score: { wins: number; losses: number }; onDraft: () => void; onKeep: () => void }) {
  const ref = useDialog('home-new-draft');
  return (
    <div ref={ref} role="alertdialog" aria-label="End current run" aria-describedby="end-run-msg" className="home-secondary" onKeyDown={onEscape(onKeep)}>
      <p id="end-run-msg" className="home-note">This ends your current run ({score.wins}–{score.losses}).</p>
      <button onClick={onDraft}>End run and draft</button>
      <button data-autofocus onClick={onKeep}>Keep run</button>
    </div>
  );
}
