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
  hand: CardView[];
  active: InPlayView | null;
  bench: InPlayView[];
  points: number;
  deckCount: number;
  discardCount: number;
  energyZone: string;
  energyUsed: boolean;
  retreated: boolean;
};

export type PromptView = { player: number; kind: 'setup' | 'new_active' };

export type GameView = {
  players: [PlayerView, PlayerView];
  active: number;
  turn: number;
  phase: 'setup' | 'main' | 'over';
  pending: PromptView | null;
  winner: number; // -1 while ongoing, else 0 or 1
  narration: string[];
};

// Mirrors server toEvent(): one shape per move type. `player` is always the
// actor (gameView.active, or pending.player during a prompt).
export type Move =
  | { type: 'SetupPlace'; player: number; activeCardId: string; benchCardIds: string[] }
  | { type: 'PlayBasic'; player: number; cardId: string }
  | { type: 'AttachEnergy'; player: number; target: number }
  | { type: 'Evolve'; player: number; handCardId: string; target: number }
  | { type: 'Retreat'; player: number; benchIndex: number }
  | { type: 'UseAttack'; player: number; index: number }
  | { type: 'EndTurn'; player: number }
  | { type: 'ChooseNewActive'; player: number; benchIndex: number }
  | { type: 'Concede'; player: number };

export type MoveResp = { ok: boolean; error?: string; state: GameView };
export type BotResp = { acted: boolean; state: GameView };
