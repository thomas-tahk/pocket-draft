import { HISTORY_CAP } from '../stores/draftStore';

type Props = {
  cardCount: number;
  hasActiveRun: boolean;
  canPractice: boolean;
  historyCount: number;
  onDraft: () => void;
  onContinue: () => void;
  onPractice: () => void;
  onPastRuns: () => void;
  onAutoDraft: () => void;
};

export function HomeView(p: Props) {
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
      {p.hasActiveRun && (
        <button className="home-link" onClick={p.onDraft}>Start a new draft instead</button>
      )}
      <button className="home-link" onClick={p.onAutoDraft}>Auto-draft (test)</button>
    </main>
  );
}
