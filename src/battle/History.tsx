import { useState } from 'react';
import { createPortal } from 'react-dom';
import type { ActionView } from '../game/types';
import './history.css';

// History (spec F15): a small Poké Ball notebook icon on the board opens a
// scrollable list of the moves so far, newest first. There is no text log on
// the board itself. On a phone the list is a sheet over the board; on desktop
// it docks beside the board and shows more at once. Same entries in both.
// Actions arrive already stripped of hidden picks; render what is sent.

// Placeholder wording (owner's to change, see docs/questions/M1.md).
const KIND_LABEL: Record<string, string> = {
  setup: 'Placed', setup_return: 'Took back', setup_move: 'Moved', ready: 'Ready', play_basic: 'Played to Bench', attach_energy: 'Attached energy',
  evolve: 'Evolved', retreat: 'Retreated', retreat_pay: 'Paid retreat', attack: 'Attacked',
  end_turn: 'Ended turn', promote: 'New Active', concede: 'Conceded',
};

function Entry({ action }: { action: ActionView }) {
  const who = action.player === 0 ? 'You' : 'Opponent';
  return (
    <li className="hx-entry" data-testid="history-entry" data-kind={action.kind} data-player={action.player}>
      {action.card?.image ? <img className="hx-thumb" src={action.card.image} alt="" /> : <span className="hx-thumb hx-thumb-none" />}
      <div>
        <div className="hx-head">
          <span className={`hx-who hx-p${action.player}`}>{who}</span> {KIND_LABEL[action.kind] ?? action.kind}
          {action.card && <> · <b>{action.card.name}</b></>}
        </div>
        {action.result.map((line, i) => <div key={i} className="hx-line">{line}</div>)}
      </div>
    </li>
  );
}

function Panel({ actions, onClose }: { actions: ActionView[]; onClose: () => void }) {
  return (
    <section className="hx-panel" data-testid="history-panel" aria-label="History">
      <header className="hx-bar">
        History
        <button className="hx-close" data-testid="history-close" onClick={onClose} aria-label="Close history">×</button>
      </header>
      <ol className="hx-list" data-testid="history-list">
        {actions.length === 0 && <li className="hx-empty">No moves yet.</li>}
        {[...actions].reverse().map((a, i) => <Entry key={actions.length - i} action={a} />)}
      </ol>
    </section>
  );
}

export function History({ actions }: { actions: ActionView[] }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <button className="hx-icon" data-testid="history-button" aria-label="History" aria-expanded={open} onClick={() => setOpen(!open)}>
        <span className="hx-ball" />
      </button>
      {open && createPortal(<Panel actions={actions} onClose={() => setOpen(false)} />, document.body)}
    </>
  );
}
