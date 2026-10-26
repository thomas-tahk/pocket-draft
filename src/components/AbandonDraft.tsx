import { useState } from 'react';

// The draft header's way out (#117): asks on screen, never window.confirm. Cancel keeps the draft.
export function AbandonDraft({ loses, onAbandon }: { loses: string; onAbandon: () => void }) {
  const [asking, setAsking] = useState(false);
  if (!asking) {
    return <button className="run-btn" onClick={() => setAsking(true)}>Abandon draft</button>;
  }
  return (
    <div role="alertdialog" aria-label="Abandon draft" className="run-confirm" data-testid="abandon-confirm">
      <p>Abandon this draft? You will lose {loses}.</p>
      <button className="run-btn run-btn-danger" data-testid="abandon-ok" onClick={onAbandon}>Abandon draft</button>
      <button className="run-btn" data-testid="abandon-cancel" onClick={() => setAsking(false)}>Cancel</button>
    </div>
  );
}
