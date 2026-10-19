import { useRef, type CSSProperties, type ReactNode } from 'react';
import type { CardView, GameView, PlayerView } from '../game/types';
import { ChoiceOverlay, type Choice } from './Choice';
import { History } from './History';
import { InPlayCard } from './InPlayCard';
import { CardBack, CardFace, EmptySlot, EnergyIcon, PointPips } from './pieces';
import { TurnBanner } from './TurnBanner';
import { usableCards } from './usable';
import { CardZoom, PlayLayer } from './PlayLayer';
import { PlayContext, usePlay, useHolding, type Spot } from './usePlay';

// The battle board, laid out from the owner's Pocket screenshot (spec L1–L4).
// Positions are in 390 × 844 grid units (see battle.css). Opponent on top,
// mirrored; you at the bottom.

type Place = { top?: number; left?: number; right?: number; bottom?: number; width?: number; height?: number };

function at(p: Place): CSSProperties {
  const style: Record<string, string> = {};
  for (const [k, v] of Object.entries(p)) style[k] = `calc(var(--u) * ${v})`;
  return style;
}

function Zone({ id, place, children, extra }: { id: string; place: Place; children: ReactNode; extra?: Record<string, string | number> }) {
  return <div className="bt-zone" data-testid={id} style={at(place)} {...extra}>{children}</div>;
}

const SLOT_NAMES = ['left', 'behind', 'right'] as const;
const BENCH_W = 74;
const ACTIVE_W = 108;
const PILE_W = 56;

// Marks a drop spot for press-and-drag play; it glows when the held card can go there (F3).
function useSpot(spot: Spot | null) {
  const holding = useHolding();
  if (!spot) return {};
  return { 'data-spot': spot, 'data-glow': holding?.moves.has(spot) ?? false };
}

type SlotProps = { i: number; mon: PlayerView['bench'][number]; mirrored: boolean; usable: boolean; choice?: Choice | null };

function SpotSlot({ i, mon, mirrored, usable, choice }: SlotProps) {
  const spot = useSpot(mirrored ? null : (`bench-${i}` as Spot));
  const p = pulse(usable);
  return (
    <div data-testid="bench-slot" data-slot={SLOT_NAMES[i]} data-pos={i + 1} {...p} className={['bt-spot', p.className].filter(Boolean).join(' ')} {...spot}>
      <Lit pos={i + 1} choice={choice}><InPlayCard mon={mon ?? null} width={BENCH_W} /></Lit>
    </div>
  );
}

function ActiveSpot({ mon }: { mon: PlayerView['active'] }) {
  return <div className="bt-spot" {...useSpot('active')}><InPlayCard mon={mon} width={ACTIVE_W} /></div>;
}

function HandCard({ card, index }: { card: CardView; index: number }) {
  const holding = useHolding();
  const lifted = holding?.held.kind === 'hand' && holding.held.index === index;
  return <div data-pick={`hand:${index}`} data-card-id={card.id} data-lifted={lifted}><CardFace card={card} width={80} /></div>;
}

// Bench slots in screen order. Yours read left, behind, right; the opponent
// faces you, so their left is on your right.
function Bench({ bench, mirrored, usable, choice }: { bench: PlayerView['bench']; mirrored: boolean; usable?: Set<number>; choice?: Choice | null }) {
  const order = mirrored ? [2, 1, 0] : [0, 1, 2];
  return (
    <div style={{ display: 'flex', gap: 'calc(var(--u) * 8)' }}>
      {order.map((i) => <SpotSlot key={i} i={i} mon={bench[i]} mirrored={mirrored} usable={!!usable?.has(i)} choice={choice} />)}
    </div>
  );
}

// A card on your side involved in the current choice is drawn above the dim
// layer; if tapping it answers the choice, it glows and takes the tap.
function Lit({ pos, choice, children }: { pos: number; choice?: Choice | null; children: ReactNode }) {
  const spot = choice?.lit.get(pos);
  if (!spot) return <>{children}</>;
  return <div className="bt-lit" data-lit="true" data-pickable={!!spot.pick} onClick={spot.pick}>{children}</div>;
}

function DeckPile({ count }: { count: number }) {
  return (
    <div className="bt-pile" aria-label={`deck, ${count} cards`}>
      {count > 0 ? <CardBack width={PILE_W} /> : <EmptySlot width={PILE_W} />}
      <span className="bt-count">{count}</span>
    </div>
  );
}

function DiscardPile({ count }: { count: number }) {
  return (
    <div className="bt-pile" aria-label={`discard pile, ${count} cards`}>
      <div className="bt-card bt-discard" style={{ '--w': PILE_W } as CSSProperties}>Discard</div>
      <span className="bt-count">{count}</span>
    </div>
  );
}

// L5: a usable card pulses (battle.css .bt-usable).
function pulse(usable = false) {
  return { className: usable ? 'bt-usable' : undefined, 'data-usable': String(usable) };
}

// Current energy big, next energy small. Once the current one is attached it
// leaves an empty grey circle (owner's L1 answer).
function EnergyZone({ player, scale }: { player: PlayerView; scale: number }) {
  const spent = player.energyUsed || !player.energyZone;
  return (
    <div style={{ position: 'relative' }}>
      <div className="bt-ez-ring" style={{ '--r': 50 * scale } as CSSProperties}>
        <EnergyIcon type={player.energyZone} size={34 * scale} spent={spent} testId="energy-current" />
      </div>
      <div style={{ position: 'absolute', left: `calc(var(--u) * ${46 * scale})`, top: `calc(var(--u) * ${30 * scale})` }}>
        <EnergyIcon type={player.energyNext} size={20 * scale} spent={!player.energyNext} testId="energy-next" />
      </div>
    </div>
  );
}

function Points({ name, points }: { name: string; points: number }) {
  return <div className="bt-points">{name} <PointPips points={points} /></div>;
}

// A row of overlapping cards inside a fixed width, fanned slightly.
function Fan({ count, width, cardW, render }: { count: number; width: number; cardW: number; render: (i: number) => ReactNode }) {
  const step = count > 1 ? Math.min(cardW + 4, (width - cardW) / (count - 1)) : 0;
  const start = (width - (cardW + step * (count - 1))) / 2;
  return (
    <>
      {Array.from({ length: count }, (_, i) => {
        const tilt = (i - (count - 1) / 2) * 3;
        return (
          <div key={i} className="bt-hand-card" style={{ left: `calc(var(--u) * ${start + step * i})`, top: 0, transform: `rotate(${tilt}deg)` }}>
            {render(i)}
          </div>
        );
      })}
    </>
  );
}

function OpponentSide({ p }: { p: PlayerView }) {
  return (
    <>
      <Zone id="opp-points" place={{ top: 6, right: 0 }}><Points name="Opponent" points={p.points} /></Zone>
      <Zone id="opp-hand" place={{ top: 40, left: 105, width: 180, height: 72 }} extra={{ 'data-count': p.handCount }}>
        <Fan count={p.handCount} width={150} cardW={50} render={() => <CardBack width={50} />} />
        <span className="bt-count" style={{ left: 'auto', right: 0, bottom: 0, transform: 'none' }}>{p.handCount}</span>
      </Zone>
      <Zone id="opp-energy" place={{ top: 80, left: 14 }}><EnergyZone player={p} scale={0.8} /></Zone>
      <Zone id="opp-bench" place={{ top: 124, left: 76 }}><Bench bench={p.bench} mirrored /></Zone>
      <Zone id="opp-discard" place={{ top: 168, left: 6 }}><DiscardPile count={p.discardCount} /></Zone>
      <Zone id="opp-deck" place={{ top: 262, left: 6 }}><DeckPile count={p.deckCount} /></Zone>
      <Zone id="opp-active" place={{ top: 240, left: 141 }}><InPlayCard mon={p.active} width={ACTIVE_W} /></Zone>
    </>
  );
}

function YourSide({ p, choice }: { p: PlayerView; choice: Choice | null }) {
  const usable = usableCards(p.legalMoves);
  return (
    <>
      <Zone id="my-active" place={{ top: 440, left: 141 }}>
        <div data-testid="active-card" {...pulse(usable.active)}><Lit pos={0} choice={choice}><ActiveSpot mon={p.active} /></Lit></div>
      </Zone>
      <Zone id="my-deck" place={{ top: 446, right: 6 }}><DeckPile count={p.deckCount} /></Zone>
      <Zone id="my-discard" place={{ top: 536, right: 6 }}><DiscardPile count={p.discardCount} /></Zone>
      <Zone id="my-bench" place={{ top: 604, left: 76 }}><Bench bench={p.bench} mirrored={false} usable={usable.bench} choice={choice} /></Zone>
      <Zone id="my-hand" place={{ top: 722, left: 36, width: 274, height: 112 }}>
        <Fan count={p.hand.length} width={274} cardW={80} render={(i) => (
          <div data-testid="hand-card" {...pulse(usable.hand.has(p.hand[i].id))}><HandCard card={p.hand[i] as CardView} index={i} /></div>
        )} />
      </Zone>
      <Zone id="my-energy" place={{ top: 712, left: 318 }} extra={{ 'data-pick': 'energy' }}><EnergyZone player={p} scale={1} /></Zone>
      <Zone id="my-points" place={{ bottom: 0, left: 0 }}><Points name="You" points={p.points} /></Zone>
    </>
  );
}

export function Board({ view, onEndTurn, choice = null, children }: { view: GameView; onEndTurn: () => void; choice?: Choice | null; children?: ReactNode }) {
  const [you, opp] = view.players;
  const yourTurn = view.phase === 'main' && view.active === 0 && !view.pending;
  const boardRef = useRef<HTMLDivElement>(null);
  const { holding, flash, zoom, onPointerDown } = usePlay(view);
  return (
    <div className="bt-board" data-testid="battle-board" ref={boardRef} onPointerDown={onPointerDown}>
      <PlayContext.Provider value={holding}>
      <OpponentSide p={opp} />
      {/* No Stadium is tracked by the engine yet; the slot marks where it goes. */}
      <Zone id="stadium" place={{ top: 432, left: 6, width: 66, height: 92 }}>{null}</Zone>
      <Zone id="end-turn" place={{ top: 400, right: 10 }}>
        <button className="bt-endturn" disabled={!yourTurn} onClick={onEndTurn}>End Turn</button>
      </Zone>
      <YourSide p={you} choice={choice} />
      </PlayContext.Provider>
      <History actions={view.actions} />
      <PlayLayer holding={holding} flash={flash} boardRef={boardRef} />
      <CardZoom card={zoom} />
      <TurnBanner view={view} />
      {choice && <ChoiceOverlay choice={choice} />}
      {children}
    </div>
  );
}
