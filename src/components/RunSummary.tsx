import type { CardPool } from '../data';
import { runLine, runScore, STATUS_LABEL, type Run } from '../stores/run';
import { BattleList } from './BattleList';
import { DeckShare } from './DeckShare';

type Props = {
  pool: CardPool;
  run: Run;
  onDraftAgain: () => void;
  onHome: () => void;
  onReplay: (index: number) => void;
};

// The ending of a run: result, every battle in order, and the way on.
export function RunSummary({ pool, run, onDraftAgain, onHome, onReplay }: Props) {
  const { wins, losses } = runScore(run);
  return (
    <main className="run-summary" aria-label="Run summary">
      <h1 data-testid="run-status" className="run-status">{STATUS_LABEL[run.status]}</h1>
      <p data-testid="run-record" className="run-record">{runLine(run)} ({run.status})</p>
      <p className="run-tally">{wins} {wins === 1 ? 'win' : 'wins'}, {losses} {losses === 1 ? 'loss' : 'losses'}</p>
      <h2 className="run-sub">Battles</h2>
      {run.battles.length === 0 ? (
        <p className="run-empty">No battles were played.</p>
      ) : (
        <BattleList run={run} onReplay={onReplay} />
      )}
      <div className="run-summary-actions">
        <button className="run-btn run-btn-primary" onClick={onDraftAgain}>Draft again</button>
        <button className="run-btn" onClick={onHome}>Home</button>
      </div>
      <div className="run-summary-actions run-summary-share">
        <DeckShare deck={run.deck} pool={pool} />
      </div>
    </main>
  );
}
