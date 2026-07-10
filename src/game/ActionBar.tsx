import { useState } from 'react';
import type { GameView } from './types';
import { useGameStore } from './store';

export function ActionBar({ view }: { view: GameView }) {
  const { dispatch, runBot, restart, selection, clearSelection } = useGameStore();
  const [setupBench, setSetupBench] = useState<string[]>([]);
  const [setupActive, setSetupActive] = useState<string | null>(null);

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
  const attacks = me.active?.card.attacks ?? [];
  return (
    <div>
      <button style={btn} onClick={() => dispatch({ type: 'AttachEnergy', player: actor, target: 0 })}>Attach energy</button>
      {attacks.map((a, i) => (
        <button key={i} style={btn} onClick={() => dispatch({ type: 'UseAttack', player: actor, index: i })}>
          Attack: {a.name} ({a.damage})
        </button>
      ))}
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
    </div>
  );
}
