import type { RefObject } from 'react';
import type { CardView } from '../game/types';
import { CardFace, EnergyIcon } from './pieces';
import { hintFor, type Holding } from './usePlay';

// Drawn over the board while choosing (F3): the dim, the dragged card under
// the finger, the bottom hint; and the brief illegal-reason flash (F17).
export function PlayLayer({ holding, flash, boardRef }: { holding: Holding | null; flash: string | null; boardRef: RefObject<HTMLDivElement | null> }) {
  return (
    <>
      {holding && <div className="bt-play-dim" data-testid="play-dim" />}
      {holding?.mode === 'drag' && <Ghost holding={holding} boardRef={boardRef} />}
      {holding && <div className="bt-hint" data-testid="play-hint">{hintFor(holding.held, holding.moves)}</div>}
      {flash && <div className="bt-reason" data-testid="play-reason" role="status">{flash}</div>}
    </>
  );
}

// A held card blown up, centred, over everything (F2). Size is a placeholder.
export function CardZoom({ card }: { card: CardView | null }) {
  if (!card) return null;
  return <div className="bt-zoom" data-testid="card-zoom"><CardFace card={card} width={320} /></div>;
}

function Ghost({ holding, boardRef }: { holding: Holding; boardRef: RefObject<HTMLDivElement | null> }) {
  const rect = boardRef.current?.getBoundingClientRect();
  if (!rect) return null;
  const { held, pos } = holding;
  return (
    <div className="bt-ghost" data-testid="play-ghost" style={{ left: pos.x - rect.left, top: pos.y - rect.top }}>
      {held.kind === 'hand' ? <CardFace card={held.card} width={80} /> : <EnergyIcon type={held.type} size={34} testId="ghost-energy" />}
    </div>
  );
}
