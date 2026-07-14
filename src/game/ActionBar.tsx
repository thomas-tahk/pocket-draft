import { useState } from 'react';
import type { GameView } from './types';
import { useGameStore } from './store';

// Client-side mirror of the server's canPay, for affordance styling only —
// the server remains the sole authority on whether an attack is legal.
function canAfford(energy: Record<string, number>, total: number, cost: string[]): boolean {
  if (total < cost.length) return false;
  const need: Record<string, number> = {};
  for (const sym of cost) if (sym !== 'Colorless') need[sym] = (need[sym] ?? 0) + 1;
  return Object.entries(need).every(([t, n]) => (energy[t] ?? 0) >= n);
}

export function ActionBar({ view }: { view: GameView }) {
  const { dispatch, runBot, restart, selection, clearSelection } = useGameStore();
  const [setupBench, setSetupBench] = useState<string[]>([]);
  const [setupActive, setSetupActive] = useState<string | null>(null);
  const [confirmConcede, setConfirmConcede] = useState(false);

  const actor = view.pending ? view.pending.player : view.active;
  const me = view.players[actor];
  const btn = { fontSize: 13, padding: '6px 12px', marginRight: 6, marginBottom: 6 } as const;

  if (view.phase === 'over') {
    return <div><button style={btn} onClick={() => restart()}>New game</button></div>;
  }

  // Setup: choose Active + up to 3 bench Basics from hand.
  if (view.pending?.kind === 'setup') {
    const activeId = setupActive ?? '';
    return (
      <div>
        <div style={{ fontSize: 12, marginBottom: 6 }}>
          Setup P{actor + 1}: pick an Active Basic (click a hand card, then “Set Active”), toggle up to 3 bench, then Place.
        </div>
        <button style={btn} onClick={() => {
          if (selection?.kind === 'hand') setSetupActive(selection.cardId);
        }}>Set Active</button>
        <button style={btn} onClick={() => {
          if (selection?.kind === 'hand') {
            const id = selection.cardId;
            setSetupBench((b) => (b.includes(id) ? b.filter((x) => x !== id) : b.length < 3 ? [...b, id] : b));
          }
        }}>Toggle bench ({setupBench.length}/3)</button>
        <button style={btn} disabled={!activeId} onClick={async () => {
          await dispatch({ type: 'SetupPlace', player: actor, activeCardId: activeId, benchCardIds: setupBench });
          setSetupBench([]);
          setSetupActive(null);
        }}>Place</button>
        <button style={btn} onClick={() => runBot()}>Bot place</button>
      </div>
    );
  }

  if (view.pending?.kind === 'new_active') {
    return (
      <div style={{ fontSize: 12 }}>
        P{actor + 1} was knocked out — click a benched Pokémon to promote it.
        <div>
          <button style={btn} onClick={() => {
            if (selection?.kind === 'bench') dispatch({ type: 'ChooseNewActive', player: actor, benchIndex: selection.index });
          }}>Promote selected bench</button>
          <button style={btn} onClick={() => runBot()}>Bot choose</button>
        </div>
      </div>
    );
  }

  // Main phase.
  const active = me.active;
  const attacks = active?.card.attacks ?? [];
  // Attach energy to whichever Pokémon is selected: a benched one if a bench
  // card is selected, otherwise the Active. (Engine target: 0 = Active, n = bench n-1.)
  const attachTarget = selection?.kind === 'bench' ? selection.index + 1 : 0;
  const attachLabel = selection?.kind === 'bench'
    ? `Attach energy → bench ${selection.index + 1}`
    : 'Attach energy → active';
  return (
    <div>
      {confirmConcede && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,.55)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50,
        }}>
          <div style={{ background: '#1c1f26', border: '1px solid #33363d', borderRadius: 10, padding: 20, maxWidth: 320 }}>
            <div style={{ fontSize: 14, marginBottom: 14 }}>Concede this game? Your opponent wins immediately.</div>
            <button style={{ ...btn, background: '#c0392b', color: '#fff' }} onClick={() => {
              dispatch({ type: 'Concede', player: actor });
              clearSelection();
              setConfirmConcede(false);
            }}>Concede</button>
            <button style={btn} onClick={() => setConfirmConcede(false)}>Cancel</button>
          </div>
        </div>
      )}
      <button style={btn} onClick={() => dispatch({ type: 'AttachEnergy', player: actor, target: attachTarget })}>{attachLabel}</button>
      {attacks.map((a, i) => {
        const ok = !!active && canAfford(active.energy, active.totalEnergy, a.cost);
        return (
          <button
            key={i}
            style={btn}
            disabled={!ok}
            title={ok ? '' : 'needs ' + a.cost.join(' ')}
            onClick={() => dispatch({ type: 'UseAttack', player: actor, index: i })}
          >
            Attack: {a.name} ({a.damage})
          </button>
        );
      })}
      <button style={btn} onClick={() => {
        if (selection?.kind === 'hand') dispatch({ type: 'PlayBasic', player: actor, cardId: selection.cardId });
      }}>Play basic (selected)</button>
      <button style={btn} onClick={() => {
        if (selection?.kind === 'hand') dispatch({ type: 'Evolve', player: actor, handCardId: selection.cardId, target: 0 });
      }}>Evolve active (selected)</button>
      <button style={btn} onClick={() => {
        if (selection?.kind === 'bench') dispatch({ type: 'Retreat', player: actor, benchIndex: selection.index });
      }}>Retreat → selected bench</button>
      <button style={btn} onClick={() => { dispatch({ type: 'EndTurn', player: actor }); clearSelection(); }}>End turn</button>
      <button style={btn} onClick={() => runBot()}>Bot move</button>
      <button style={btn} onClick={() => restart()}>New game</button>
      <button style={{ ...btn, marginLeft: 12, color: '#c0392b' }} onClick={() => setConfirmConcede(true)}>Concede</button>
    </div>
  );
}
