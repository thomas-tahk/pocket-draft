import { useMemo, useRef, useState } from 'react';
import { announce } from '../announce';
import { TypeTally } from '../battle/energy';
import type { Card, EnergyType } from '../types/card';
import type { CardPool } from '../data';
import type { DraftHistoryEntry } from '../stores/draftStore';
import { tallyTypes } from '../draft/weights';
import { runScore, STATUS_LABEL, type Run } from '../stores/run';
import { CardTile, type HoverPayload } from './CardTile';
import { CardPreview, type HoverState } from './CardPreview';

type Props = {
  history: DraftHistoryEntry[];
  run: Run | null; // the live run: its entry shows it before it is archived
  pool: CardPool;
  onClose: () => void;
  onUpdate: (
    id: string,
    updates: Partial<Pick<DraftHistoryEntry, 'notes' | 'wins' | 'losses'>>,
  ) => void;
  onDelete: (id: string) => void;
};

export function HistoryView({ history, run, pool, onClose, onUpdate, onDelete }: Props) {
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [hover, setHover] = useState<HoverState>(null);
  const handleHover = (p: HoverPayload | null) => setHover(p);
  const titleRef = useRef<HTMLHeadingElement>(null);

  const idToCard = useMemo(() => {
    const m = new Map<string, Card>();
    for (const c of pool.fullCards) m.set(c.id, c);
    return m;
  }, [pool]);

  const formatDate = (iso: string) => {
    const d = new Date(iso);
    return d.toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
  };

  const renderEntry = (entry: DraftHistoryEntry) => {
    const deckCards = Object.entries(entry.deck)
      .map(([id, count]) => ({ card: idToCard.get(id), count }))
      .filter((x): x is { card: Card; count: number } => Boolean(x.card));

    const flat: Card[] = [];
    for (const { card, count } of deckCards) {
      for (let i = 0; i < count; i++) flat.push(card);
    }
    const tally = tallyTypes(flat);
    const sortedTally = (Object.entries(tally) as [EnergyType, number][])
      .sort((a, b) => b[1] - a[1])
      .slice(0, 4);
    const total = deckCards.reduce((s, { count }) => s + count, 0);

    const expanded = expandedId === entry.id;
    // A played run fills its own record; older entries keep the typed one.
    const played = entry.run ?? (run?.historyId === entry.id ? run : undefined);
    const record = played ? runScore(played) : { wins: entry.wins, losses: entry.losses };

    return (
      <div key={entry.id} className="panel history-entry">
        <div className="history-entry-head">
          <div style={{ fontSize: 'var(--fs-sm)' }}>
            <h2 id={`entry-${entry.id}`} style={{ margin: 0, fontSize: 'inherit', fontWeight: 'var(--fw-bold)' }}>{formatDate(entry.completedAt)}</h2>
            <div className="muted-xs" style={{ marginTop: 2 }}>
              {total} cards · <TypeTally entries={sortedTally} />
              {' '}·{' '}
              <span className="count-w">{record.wins}W</span>
              {' / '}
              <span className="count-l">{record.losses}L</span>
              {played && <> · <span data-testid="history-run-status">{STATUS_LABEL[played.status]}</span></>}
            </div>
          </div>
          <div className="history-actions">
            <button
              aria-expanded={expanded}
              aria-describedby={`entry-${entry.id}`}
              onClick={() => setExpandedId(expanded ? null : entry.id)}
            >
              {expanded ? 'Collapse' : 'View'}
            </button>
            <button
              onClick={() => {
                if (!window.confirm('Delete this draft from history?')) return;
                onDelete(entry.id);
                // The row (and the button that had focus) is gone: say so and keep focus on the page.
                announce('Draft deleted from history.');
                titleRef.current?.focus();
              }}
              aria-describedby={`entry-${entry.id}`}
              className="btn-quiet"
            >
              Delete
            </button>
          </div>
        </div>

        {expanded && (
          <div style={{ marginTop: 12 }}>
            {played ? (
              <p className="muted-xs" style={{ margin: '0 0 8px' }}>
                Record {record.wins}–{record.losses}, from the run.
              </p>
            ) : (
              <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 8 }}>
                <label className="muted-xs">
                  Wins:{' '}
                  <input
                    type="number"
                    min={0}
                    value={entry.wins}
                    onChange={(e) => onUpdate(entry.id, { wins: Math.max(0, Number(e.target.value)) })}
                    style={{ width: 50 }}
                  />
                </label>
                <label className="muted-xs">
                  Losses:{' '}
                  <input
                    type="number"
                    min={0}
                    value={entry.losses}
                    onChange={(e) =>
                      onUpdate(entry.id, { losses: Math.max(0, Number(e.target.value)) })
                    }
                    style={{ width: 50 }}
                  />
                </label>
              </div>
            )}

            <textarea
              placeholder="Notes on how this deck fared…"
              aria-label="Notes on how this deck fared"
              value={entry.notes}
              onChange={(e) => onUpdate(entry.id, { notes: e.target.value })}
              rows={3}
              className="history-notes"
            />

            <div className="history-grid">
              {deckCards.flatMap(({ card, count }) =>
                Array.from({ length: count }).map((_, i) => (
                  <CardTile key={`${card.id}-${i}`} card={card} size="sm" focusable onHover={handleHover} />
                )),
              )}
            </div>
          </div>
        )}
      </div>
    );
  };

  return (
    <main className="page page-narrow">
      <header className="screen-header screen-header-static">
        <h1 className="screen-title" ref={titleRef} tabIndex={-1}>Draft history</h1>
        <button onClick={onClose}>Back</button>
      </header>

      {history.length === 0 ? (
        <p className="muted">No completed drafts yet.</p>
      ) : (
        history.map(renderEntry)
      )}

      <CardPreview hover={hover} />
    </main>
  );
}
