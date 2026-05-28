import { useMemo, useState } from 'react';
import type { Card, EnergyType } from '../types/card';
import type { CardPool } from '../data';
import type { DraftHistoryEntry } from '../stores/draftStore';
import { tallyTypes } from '../draft/weights';
import { CardTile, type HoverPayload } from './CardTile';
import { CardPreview, type HoverState } from './CardPreview';

type Props = {
  history: DraftHistoryEntry[];
  pool: CardPool;
  onClose: () => void;
  onUpdate: (
    id: string,
    updates: Partial<Pick<DraftHistoryEntry, 'notes' | 'wins' | 'losses'>>,
  ) => void;
  onDelete: (id: string) => void;
};

export function HistoryView({ history, pool, onClose, onUpdate, onDelete }: Props) {
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [hover, setHover] = useState<HoverState>(null);
  const handleHover = (p: HoverPayload | null) => setHover(p);

  const idToCard = useMemo(() => {
    const m = new Map<string, Card>();
    for (const c of pool.fullCards) m.set(c.id, c);
    for (const c of pool.universals) m.set(c.id, c);
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

    return (
      <div
        key={entry.id}
        style={{
          border: '1px solid #8884',
          borderRadius: 8,
          padding: 12,
          marginBottom: 8,
        }}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            gap: 8,
            flexWrap: 'wrap',
          }}
        >
          <div style={{ fontSize: 13 }}>
            <div style={{ fontWeight: 600 }}>{formatDate(entry.completedAt)}</div>
            <div style={{ opacity: 0.75, fontSize: 12, marginTop: 2 }}>
              {total} cards · {sortedTally.map(([t, n]) => `${t} ×${n}`).join('  ')}
              {' '}·{' '}
              <span style={{ color: '#4c8' }}>{entry.wins}W</span>
              {' / '}
              <span style={{ color: '#c66' }}>{entry.losses}L</span>
            </div>
          </div>
          <div style={{ display: 'flex', gap: 6 }}>
            <button onClick={() => setExpandedId(expanded ? null : entry.id)}>
              {expanded ? 'Collapse' : 'View'}
            </button>
            <button
              onClick={() => {
                if (window.confirm('Delete this draft from history?')) onDelete(entry.id);
              }}
              style={{ fontSize: 11, opacity: 0.7 }}
            >
              Delete
            </button>
          </div>
        </div>

        {expanded && (
          <div style={{ marginTop: 12 }}>
            <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 8 }}>
              <label style={{ fontSize: 12 }}>
                Wins:{' '}
                <input
                  type="number"
                  min={0}
                  value={entry.wins}
                  onChange={(e) => onUpdate(entry.id, { wins: Math.max(0, Number(e.target.value)) })}
                  style={{ width: 50 }}
                />
              </label>
              <label style={{ fontSize: 12 }}>
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

            <textarea
              placeholder="Notes on how this deck fared…"
              value={entry.notes}
              onChange={(e) => onUpdate(entry.id, { notes: e.target.value })}
              rows={3}
              style={{
                width: '100%',
                boxSizing: 'border-box',
                fontSize: 12,
                fontFamily: 'inherit',
                marginBottom: 12,
              }}
            />

            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fill, minmax(96px, 1fr))',
                gap: 4,
              }}
            >
              {deckCards.flatMap(({ card, count }) =>
                Array.from({ length: count }).map((_, i) => (
                  <CardTile key={`${card.id}-${i}`} card={card} size="sm" onHover={handleHover} />
                )),
              )}
            </div>
          </div>
        )}
      </div>
    );
  };

  return (
    <main style={{ padding: '16px 24px', maxWidth: 900, margin: '0 auto' }}>
      <header
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          marginBottom: 16,
        }}
      >
        <h1 style={{ fontSize: 20, margin: 0 }}>Draft history</h1>
        <button onClick={onClose}>Back</button>
      </header>

      {history.length === 0 ? (
        <p style={{ opacity: 0.6, fontSize: 14 }}>No completed drafts yet.</p>
      ) : (
        history.map(renderEntry)
      )}

      <CardPreview hover={hover} />
    </main>
  );
}
