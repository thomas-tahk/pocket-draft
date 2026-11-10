// Mirrors server/view.go. The client renders these; it does not own them.
// effect is the printed text; printed is the damage as printed ("40+")
export type AttackView = { name: string; cost: string[]; damage: number; effect?: string; printed?: string };

export type AbilityView = { name: string; effect: string }; // effect is the printed text

export type CardView = {
  id: string;
  name: string;
  image?: string;
  kind: 'pokemon' | 'item' | 'supporter' | 'tool' | 'stadium';
  stage: string; // '' on a Trainer
  hp: number;
  type: string;
  isEx: boolean;
  evolvesFrom?: string;
  retreatCost: number;
  weakness?: string;
  attacks: AttackView[];
  ability?: AbilityView;
};

export type InPlayView = {
  card: CardView;
  damage: number;
  remainingHp: number;
  energy: Record<string, number>;
  totalEnergy: number;
  poisoned: boolean;
  burned: boolean;
  asleep: boolean;
  paralyzed: boolean;
  confused: boolean;
  tool?: CardView; // the attached Pokémon Tool
  faceDown?: boolean; // an opponent's setup pick: the card is hidden until both are Ready
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
  // every coin the move flipped, in order (spec F11); waits: hold for the flipper's tap
  flips: { flipper: number; heads: boolean; waits: boolean }[];
  damage?: number; // an attack's damage after Weakness
};

// title and labels word a card's choice ('choose'); only the chooser gets them.
export type PromptView = {
  player: number;
  kind: 'setup' | 'new_active' | 'retreat_energy' | 'choose';
  title?: string;
  labels?: string[];
};

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
  openingFlip: { flipper: number; heads: boolean }; // the coin flip for first (spec S4)
  firstPlayer: number; // 0 or 1: who goes first
  stadium: CardView | null; // the Stadium in play
  record: GameRecord; // replayable record; empty until the game is over
};

// Mirrors server recordView. Filled only once the game is over: POST /api/new with
// {seed, you, bot}, then POST /api/move for each action in order, replays it.
export type GameRecord = { seed: number; you: string[]; bot: string[]; actions: Move[] };

// Mirrors server toEvent(): one shape per move type. `player` is always the
// actor (gameView.active, or pending.player during a prompt).
export type Move =
  | { type: 'SetupActive'; player: number; cardId: string }
  | { type: 'SetupBench'; player: number; cardId: string; slot: number }
  | { type: 'SetupReady'; player: number }
  | { type: 'SetupReturn'; player: number; target: number } // spot 0 Active, n Bench slot n-1
  | { type: 'SetupMove'; player: number; index: number; target: number } // from index to target
  | { type: 'PlayBasic'; player: number; cardId: string; slot: number }
  | { type: 'AttachEnergy'; player: number; target: number }
  | { type: 'Evolve'; player: number; handCardId: string; target: number }
  | { type: 'Retreat'; player: number; benchIndex: number }
  | { type: 'UseAttack'; player: number; index: number }
  | { type: 'EndTurn'; player: number }
  | { type: 'ChooseNewActive'; player: number; benchIndex: number }
  | { type: 'RetreatDiscard'; player: number; energy: string[] }
  | { type: 'ChooseOption'; player: number; index: number }
  | { type: 'Concede'; player: number }
  // spot: where the Trainer goes (pos 0 Active, n Bench slot n-1; player -1
  // pos -2 is the Stadium slot); absent for a Trainer with no target
  | { type: 'PlayTrainer'; player: number; cardId: string; spot?: { player: number; pos: number } }
  | { type: 'DiscardFossil'; player: number; target: number }
  // target: the Pokémon using it (0 Active, n Bench slot n-1); spot: what it is aimed at
  | { type: 'UseAbility'; player: number; target: number; spot?: { player: number; pos: number } };

export type MoveResp = { ok: boolean; error?: string; state: GameView };
export type BotResp = { acted: boolean; state: GameView };
