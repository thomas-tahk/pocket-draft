import { RESULT_LABEL, type Run } from '../stores/run';

// Every battle of a run in order, each with its Replay button.
export function BattleList({ run, onReplay }: { run: Run; onReplay: (index: number) => void }) {
  return (
    <ol className="run-battles" data-testid="run-battles">
      {run.battles.map((b, i) => (
        <li key={b.seed} className="run-battle">
          <span id={`battle-${i}`}>Battle {i + 1}: <strong>{RESULT_LABEL[b.result]}</strong></span>
          <button className="run-btn" aria-describedby={`battle-${i}`} onClick={() => onReplay(i)}>Replay</button>
        </li>
      ))}
    </ol>
  );
}
