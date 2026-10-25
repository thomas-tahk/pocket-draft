import { useState } from 'react';
import type { CardPool } from '../data';
import { RUN_LOSSES_TO_LOSE, RUN_WINS_TO_WIN, runLine, runScore, type Run } from '../stores/run';
import { ReviewView } from './ReviewView';
import { RunSummary } from './RunSummary';

type Props = {
  pool: CardPool;
  deck: Record<string, number>;
  run: Run | null;
  onNextBattle: () => void;
  onReplay?: (index: number) => void; // absent until the replay viewer lands
  onQuit: () => void;
  onNewDraft: () => void;
  onHome: () => void;
};

const RESULT = { win: 'Win', loss: 'Loss', tie: 'Tie' } as const;
const STATUS = { active: 'In progress', won: 'Run won', lost: 'Run lost', quit: 'Run quit' } as const;

// Round for a win, square for a loss; filled = earned. Shape and the word differ, never colour alone.
function Marks({ kind, count, total }: { kind: 'win' | 'loss'; count: number; total: number }) {
  const label = kind === 'win' ? 'Wins' : 'Losses';
  return (
    <div className="run-marks" data-testid={kind === 'win' ? 'run-wins' : 'run-losses'}>
      <span className="run-marks-label">{label} {count} of {total}</span>
      <span className="run-marks-row" aria-hidden="true">
        {Array.from({ length: total }, (_, i) => (
          <span key={i} className={`run-mark run-mark-${kind}${i < count ? ' filled' : ''}`}>
            {i < count ? (kind === 'win' ? 'W' : 'L') : ''}
          </span>
        ))}
      </span>
    </div>
  );
}

export function RunView({ pool, deck, run, onNextBattle, onReplay, onQuit, onNewDraft, onHome }: Props) {
  const [confirmingQuit, setConfirmingQuit] = useState(false);
  const [leaving, setLeaving] = useState<(() => void) | null>(null);
  if (!run) return <ReviewView pool={pool} deck={deck} run={run} onNewDraft={onNewDraft} onHome={onHome} />;
  if (run.status !== 'active') return <RunSummary run={run} onDraftAgain={onNewDraft} onHome={onHome} onReplay={onReplay} />;
  const { wins, losses, ties } = runScore(run);
  const active = run.status === 'active';
  // Leaving mid-run quits it first, so no button ends a run silently.
  const guard = (leave: () => void) => () => {
    if (active) setLeaving(() => leave);
    else leave();
  };

  return (
    <ReviewView pool={pool} deck={deck} run={null} onNewDraft={guard(onNewDraft)} onHome={guard(onHome)}>
      <section className="run-panel" aria-label="Run">
        <h2 className="run-title">Your run</h2>
        <p data-testid="run-record" className="run-record">{runLine(run)} ({run.status})</p>
        <p data-testid="run-status" className="run-status">{STATUS[run.status]}</p>
        <div className="run-marks-wrap">
          <Marks kind="win" count={wins} total={RUN_WINS_TO_WIN} />
          <Marks kind="loss" count={losses} total={RUN_LOSSES_TO_LOSE} />
        </div>
        {ties > 0 && <p className="run-ties">{ties} {ties === 1 ? 'tie' : 'ties'} (count toward neither)</p>}

        {active && (
          <div className="run-next">
            <button className="run-btn run-btn-primary" onClick={onNextBattle}>Next battle</button>
            <p className="run-opponent"><span aria-hidden="true">🤖</span> Opponent: practice bot</p>
          </div>
        )}

        <h3 className="run-sub">Battles</h3>
        {run.battles.length === 0 ? (
          <p className="run-empty">No battles yet.</p>
        ) : (
          <ol className="run-battles" data-testid="run-battles">
            {run.battles.map((b, i) => (
              <li key={b.seed} className="run-battle">
                <span>Battle {i + 1}: <strong>{RESULT[b.result]}</strong></span>
                <button
                  className="run-btn"
                  disabled={!onReplay}
                  title={onReplay ? undefined : 'Replay coming'}
                  onClick={() => onReplay?.(i)}
                >
                  Replay
                </button>
              </li>
            ))}
          </ol>
        )}

        {active && !confirmingQuit && (
          <button className="run-btn" onClick={() => setConfirmingQuit(true)}>Quit run</button>
        )}
        {active && confirmingQuit && (
          <div role="alertdialog" aria-label="Quit run" className="run-confirm">
            <p>Quit this run? It ends at {wins}–{losses}.</p>
            <button className="run-btn run-btn-danger" data-testid="quit-confirm" onClick={onQuit}>Yes, quit run</button>
            <button className="run-btn" onClick={() => setConfirmingQuit(false)}>Keep playing</button>
          </div>
        )}
        {active && leaving && (
          <div role="alertdialog" aria-label="Leave run" className="run-confirm" data-testid="leave-confirm">
            <p>This quits your run at {wins}–{losses}.</p>
            <button className="run-btn run-btn-danger" data-testid="leave-ok" onClick={() => { onQuit(); leaving(); }}>Quit run and leave</button>
            <button className="run-btn" data-testid="leave-cancel" onClick={() => setLeaving(null)}>Keep playing</button>
          </div>
        )}
      </section>
    </ReviewView>
  );
}
