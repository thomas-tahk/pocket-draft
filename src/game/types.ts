// Mirrors server/view.go. The client renders these; it does not own them.
export type AttackView = { name: string; cost: string[]; damage: number };

export type CardView = {
  id: string;
  name: string;
  image?: string;
  stage: string;
  hp: number;
  type: string;
  isEx: boolean;
  evolvesFrom?: string;
  retreatCost: number;
  weakness?: string;
  attacks: AttackView[];
};

export type InPlayView = {
  card: CardView;
  damage: number;
  remainingHp: number;
  energy: Record<string, number>;
  totalEnergy: number;
  poisoned: boolean;
  burned: boolean;
};

export type PlayerView = {
  hand: CardView[]; // empty for the opponent: only your own hand is sent
  handCount: number;
  active: InPlayView | null;
  bench: (InPlayView | null)[]; // one entry per slot: left, behind, right; null is empty
  ready: boolean; // pressed Ready at setup; the opponent's picks stay hidden until both are
  points: number;
  deckCount: number;
  discardCount: number;
  energyZone: string; // this turn's energy; '' if none
  energyNext: string; // next turn's energy; '' before it is rolled
  energyTypes: string[];
  energyUsed: boolean;
  retreated: boolean;
  discardedEnergy: string[]; // oldest first
  legalMoves: Move[]; // everything you may do right now; empty on the opponent's side
  handBlocked: Record<string, string>; // your hand card id -> why it has nowhere to go
  energyBlocked: string; // why this turn's energy can't be attached; '' if it can
};

// One applied move (server actionView), for the opponent-turn playback (F14)
// and the history (F15). Target pos 0 is the Active, n is bench slot n-1;
// player -1 means it touched no Pokémon.
export type ActionView = {
  player: number;
  kind: string;
  card: CardView | null;
  target: { player: number; pos: number };
  result: string[];
};

export type PromptView = { player: number; kind: 'setup' | 'new_active' | 'retreat_energy' };

export type GameView = {
  players: [PlayerView, PlayerView];
  active: number;
  turn: number;
  phase: 'setup' | 'main' | 'over';
  pending: PromptView | null;
  winner: number; // -1 while ongoing, else 0 or 1
  outcome: '' | 'win' | 'loss' | 'tie'; // yours; '' until the game is over
  reason: '' | 'points' | 'no_pokemon' | 'concede' | 'turn_cap'; // why it ended (spec F19)
  narration: string[];
  actions: ActionView[]; // oldest first; hidden info already removed for you
};

// Mirrors server toEvent(): one shape per move type. `player` is always the
// actor (gameView.active, or pending.player during a prompt).
export type Move =
  | { type: 'SetupActive'; player: number; cardId: string }
  | { type: 'SetupBench'; player: number; cardId: string; slot: number }
  | { type: 'SetupReady'; player: number }
  | { type: 'PlayBasic'; player: number; cardId: string; slot: number }
  | { type: 'AttachEnergy'; player: number; target: number }
  | { type: 'Evolve'; player: number; handCardId: string; target: number }
  | { type: 'Retreat'; player: number; benchIndex: number }
  | { type: 'UseAttack'; player: number; index: number }
  | { type: 'EndTurn'; player: number }
  | { type: 'ChooseNewActive'; player: number; benchIndex: number }
  | { type: 'RetreatDiscard'; player: number; energy: string[] }
  | { type: 'Concede'; player: number };

export type MoveResp = { ok: boolean; error?: string; state: GameView };
export type BotResp = { acted: boolean; state: GameView };
