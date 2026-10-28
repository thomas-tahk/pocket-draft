import { useMemo, useState } from 'react';
import { announce } from '../announce';
import { AbandonDraft } from './AbandonDraft';
import { EnergyIcon, TypeTally } from '../battle/energy';
import { defaultZoneTypes, MAX_ZONE_TYPES, ZONE_FALLBACK, ZONE_TYPES } from '../game/energy';
import type { Card, EnergyType } from '../types/card';
import { DECK_SIZE } from '../stores/draftStore';
import {
  buildCollection,
  COPY_CAP_BY_NAME,
  deckCountByName,
  deckSize,
  type CollectionEntry,
  type CollectionSource,
} from '../draft/expansions';
import { tallyTypes } from '../draft/weights';
import { CardTile, FocusPreview, type HoverPayload } from './CardTile';
import { CardPreview, type HoverState } from './CardPreview';

type Props = {
  picks: Card[];
  purchases: Card[];
  fullPool: Card[];
  working: ReadonlySet<string>; // ids the engine can play; only these are offered
  deck: Record<string, number>;
  energy: string[]; // the player's Energy Zone choice; [] until they touch it
  onEnergy: (types: string[]) => void;
  onAdd: (cardId: string) => void;
  onRemove: (cardId: string) => void;
  onFinalize: (energy: string[]) => void;
  onCancel: () => void;
};

const SOURCE_LABEL: Record<CollectionSource, string> = {
  drafted: 'Your draft',
  shop: 'Shop purchases',
  'under-evo': 'Free expansions (under-evolutions)',
};

const SOURCE_ORDER: CollectionSource[] = ['drafted', 'shop', 'under-evo'];

export function DeckbuildView({
  picks,
  purchases,
  fullPool,
  working,
  deck,
  energy,
  onEnergy,
  onAdd,
  onRemove,
  onFinalize,
  onCancel,
}: Props) {
  const [hover, setHover] = useState<HoverState>(null);
  const handleHover = (p: HoverPayload | null) => setHover(p);

  const collection = useMemo(
    () =>
      buildCollection(picks, purchases, fullPool).filter((e) => working.has(e.card.id)),
    [picks, purchases, fullPool, working],
  );

  const idToCard = useMemo(() => {
    const m = new Map<string, Card>();
    for (const c of fullPool) m.set(c.id, c);
    return m;
  }, [fullPool]);

  const grouped = useMemo(() => {
    const g: Record<CollectionSource, CollectionEntry[]> = {
      drafted: [],
      shop: [],
      'under-evo': [],
    };
    for (const e of collection) g[e.source].push(e);
    return g;
  }, [collection]);

  const size = deckSize(deck);
  const remaining = DECK_SIZE - size;
  const hasBasic = Object.keys(deck).some((id) => idToCard.get(id)?.stage === 'Basic');
  const finalizeBlocker =
    size !== DECK_SIZE ? `Add ${remaining} more cards to finalize` : !hasBasic ? 'Add a Basic Pokémon to finalize' : '';

  const deckCards: { card: Card; count: number }[] = useMemo(() => {
    const out: { card: Card; count: number }[] = [];
    for (const [cardId, count] of Object.entries(deck)) {
      const c = idToCard.get(cardId);
      if (c) out.push({ card: c, count });
    }
    return out;
  }, [deck, idToCard]);

  const tally = useMemo(() => {
    const flat: Card[] = [];
    for (const { card, count } of deckCards) {
      for (let i = 0; i < count; i++) flat.push(card);
    }
    return tallyTypes(flat);
  }, [deckCards]);
  const sortedTally = (Object.entries(tally) as [EnergyType, number][]).sort(
    (a, b) => b[1] - a[1],
  );

  // Until the player picks, the zone defaults to the energy the deck's attacks cost
  // (engine DefaultEnergyTypes, mirrored by defaultZoneTypes).
  const defaultEnergy = useMemo(() => {
    const flat: Card[] = [];
    for (const { card, count } of deckCards) for (let i = 0; i < count; i++) flat.push(card);
    return defaultZoneTypes(flat);
  }, [deckCards]);
  const zone = energy.length > 0 ? energy : defaultEnergy;
  const toggleZone = (t: string) => {
    if (zone.includes(t)) {
      if (zone.length > 1) onEnergy(zone.filter((z) => z !== t));
    } else if (zone.length < MAX_ZONE_TYPES) {
      onEnergy([...zone, t]);
      if (zone.length + 1 === MAX_ZONE_TYPES) announce('Up to 3: deselect one to change.');
    }
  };

  const shopCountById = useMemo(() => {
    const m = new Map<string, number>();
    for (const p of purchases) m.set(p.id, (m.get(p.id) ?? 0) + 1);
    return m;
  }, [purchases]);

  // Shop entries cap their deck contribution at the per-card ticket count
  // (so 1 Erika ticket = max 1 Erika in deck, not 2). Other sources just
  // observe the global 2x-by-name cap.
  const capFor = (entry: CollectionEntry): number => {
    if (entry.source === 'shop') {
      return Math.min(COPY_CAP_BY_NAME, shopCountById.get(entry.card.id) ?? 0);
    }
    return COPY_CAP_BY_NAME;
  };

  const canAddEntry = (entry: CollectionEntry) => {
    if (size >= DECK_SIZE) return false;
    return deckCountByName(deck, idToCard, entry.card.name) < capFor(entry);
  };

  const renderEntry = (entry: CollectionEntry) => {
    const inDeck = deck[entry.card.id] ?? 0;
    const nameTotal = deckCountByName(deck, idToCard, entry.card.name);
    const cap = capFor(entry);
    const addDisabled = !canAddEntry(entry);

    return (
      <FocusPreview
        key={`${entry.source}-${entry.card.id}`}
        card={entry.card}
        onHover={handleHover}
        className={`shelf-cell${inDeck > 0 ? ' is-picked' : ''}`}
      >
        <CardTile card={entry.card} size="lg" onHover={handleHover} />
        <div className="step-row">
          <button
            onClick={() => onRemove(entry.card.id)}
            disabled={inDeck === 0}
            className="btn-step"
            aria-label={`− Remove ${entry.card.name}`}
          >
            −
          </button>
          <span className="muted-xs" aria-hidden="true">
            {nameTotal} / {cap}
          </span>
          <span className="sr-only">{nameTotal} of {cap} copies</span>
          <button
            onClick={() => onAdd(entry.card.id)}
            disabled={addDisabled}
            className="btn-step"
            aria-label={`+ Add ${entry.card.name}`}
          >
            +
          </button>
        </div>
      </FocusPreview>
    );
  };

  const renderSection = (source: CollectionSource) => {
    const entries = grouped[source];
    if (entries.length === 0) return null;
    return (
      <section key={source} style={{ marginTop: 16 }}>
        <h2 className="section-title">
          {SOURCE_LABEL[source]} ({entries.length})
        </h2>
        <div className="shelf-grid">
          {entries.map(renderEntry)}
        </div>
      </section>
    );
  };

  return (
    <div className="screen">
      <main className="screen-main wide">
        <header className="screen-header">
          <h1 className="screen-title">Deckbuilding</h1>
          <div className="screen-header-actions">
            <div className="muted">
              <span role="status">
                Deck: {size} / {DECK_SIZE}
                {finalizeBlocker && (
                  <small className="muted-sm" style={{ marginLeft: 8 }}>
                    {finalizeBlocker}
                  </small>
                )}
              </span>
              {sortedTally.length > 0 && (
                <span style={{ marginLeft: 16 }}>
                  <TypeTally entries={sortedTally} />
                </span>
              )}
            </div>
            <button
              className="btn-primary"
              onClick={() => onFinalize(zone)}
              disabled={finalizeBlocker !== ''}
              title={finalizeBlocker}
            >
              Finalize deck
            </button>
            <AbandonDraft loses="all picks and deck choices" onAbandon={onCancel} />
          </div>
        </header>
        <p className="muted-sm" style={{ marginTop: 0 }}>
          Build a deck of exactly {DECK_SIZE} cards. Max 2 of any card by name. Drafted Pokémon
          unlock their under-evolutions (all functional variants) for free. Trainers come only from
          your draft and the shop.
        </p>

        <div className="zone-picker" role="group" data-testid="energy-picker" aria-labelledby="zone-title">
          <h2 id="zone-title" className="section-title" style={{ color: 'var(--color-text)', margin: '0 0 8px' }}>
            Choose up to {MAX_ZONE_TYPES} energies to use{' '}
            <span data-testid="energy-count" className="muted-xs" aria-hidden="true">
              {zone.length}/{MAX_ZONE_TYPES}
            </span>
          </h2>
          {zone.length === 0 && (
            <p className="muted-xs" style={{ margin: '0 0 8px' }}>
              No typed energy needed yet: {ZONE_FALLBACK} only.
            </p>
          )}
          <div className="zone-picker-row" role="group" aria-label="Energy Zone types">
            {ZONE_TYPES.map((t) => {
              const on = zone.includes(t);
              const blocked = !on && zone.length >= MAX_ZONE_TYPES;
              return (
                <button
                  key={t}
                  type="button"
                  className="zone-chip"
                  aria-pressed={on}
                  disabled={blocked}
                  aria-label={t}
                  data-type={t}
                  onClick={() => toggleZone(t)}
                >
                  <EnergyIcon type={t} size={22} testId="zone-chip-icon" />
                  {t}
                </button>
              );
            })}
          </div>
        </div>

        {SOURCE_ORDER.map(renderSection)}
      </main>

      <aside className="screen-aside" aria-labelledby="deck-aside-title" style={{ ['--aside-w' as string]: '280px' }}>
        <h2 id="deck-aside-title" className="aside-title">
          Deck ({size} / {DECK_SIZE})
        </h2>
        {deckCards.length === 0 ? (
          <p className="muted-xs">No cards yet.</p>
        ) : (
          <div className="tile-grid-sm" role="group" aria-labelledby="deck-aside-title">
            {deckCards.map(({ card, count }) =>
              Array.from({ length: count }).map((_, i) => (
                <CardTile
                  key={`${card.id}-${i}`}
                  card={card}
                  size="sm"
                  label={`Remove ${card.name} from deck`}
                  onPick={() => onRemove(card.id)}
                  onHover={handleHover}
                />
              )),
            )}
          </div>
        )}
      </aside>

      <CardPreview hover={hover} />
    </div>
  );
}
