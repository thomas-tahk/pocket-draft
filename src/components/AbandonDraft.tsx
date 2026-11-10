import { useState } from 'react';
import { onEscape, useDialog } from './useDialog';

// The draft header's way out (#117): asks on screen, never window.confirm. Cancel keeps the draft.
export function AbandonDraft({ loses, onAbandon }: { loses: string; onAbandon: () => void }) {
  const [asking, setAsking] = useState(false);
  if (!asking) {
    return <button id="abandon-open" className="run-btn" onClick={() => setAsking(true)}>Abandon draft</button>;
  }
  return <AbandonConfirm loses={loses} onAbandon={onAbandon} onCancel={() => setAsking(false)} />;
}

function AbandonConfirm({ loses, onAbandon, onCancel }: { loses: string; onAbandon: () => void; onCancel: () => void }) {
  const ref = useDialog('abandon-open');
  return (
    <div ref={ref} role="alertdialog" aria-label="Abandon draft" aria-describedby="abandon-msg" className="run-confirm" data-testid="abandon-confirm" onKeyDown={onEscape(onCancel)}>
      <p id="abandon-msg">Abandon this draft? You will lose {loses}.</p>
      <button className="run-btn run-btn-danger" data-testid="abandon-ok" onClick={onAbandon}>Abandon draft</button>
      <button className="run-btn" data-testid="abandon-cancel" data-autofocus onClick={onCancel}>Cancel</button>
    </div>
  );
}
