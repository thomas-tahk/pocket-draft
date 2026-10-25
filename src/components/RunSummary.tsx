import { runLine, runScore, type Run } from '../stores/run';

type Props = {
  run: Run;
  onDraftAgain: () => void;
  onHome: () => void;
  onReplay: (index: number) => void;
};

const RESULT = { win: 'Win', loss: 'Loss', tie: 'Tie' } as const;
const HEADLINE = { active: 'In progress', won: 'Run won', lost: 'Run lost', quit: 'Run quit' } as const;

// The ending of a run: result, every battle in order, and the way on.
export function RunSummary({ run, onDraftAgain, onHome, onReplay }: Props) {
  const { wins, losses } = runScore(run);
  return (
    <main className="run-summary" aria-label="Run summary">
      <h1 data-testid="run-status" className="run-status">{HEADLINE[run.status]}</h1>
      <p data-testid="run-record" className="run-record">{runLine(run)} ({run.status})</p>
      <p className="run-ties">{wins} {wins === 1 ? 'win' : 'wins'}, {losses} {losses === 1 ? 'loss' : 'losses'}</p>
      <h2 className="run-sub">Battles</h2>
      {run.battles.length === 0 ? (
        <p className="run-empty">No battles were played.</p>
      ) : (
        <ol className="run-battles" data-testid="run-battles">
          {run.battles.map((b, i) => (
            <li key={b.seed} className="run-battle">
              <span>Battle {i + 1}: <strong>{RESULT[b.result]}</strong></span>
              <button className="run-btn" onClick={() => onReplay(i)}>Replay</button>
            </li>
          ))}
        </ol>
      )}
      <div className="run-summary-actions">
        <button className="run-btn run-btn-primary" onClick={onDraftAgain}>Draft again</button>
        <button className="run-btn" onClick={onHome}>Home</button>
      </div>
    </main>
  );
}
