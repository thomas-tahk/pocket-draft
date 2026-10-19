package engine

// InPlay is a Pokémon on the board: its current card (which changes on
// evolution), accumulated damage, attached energy, and status conditions.
type InPlay struct {
	Card       Card
	Damage     int
	Energy     map[Energy]int
	Poisoned   bool
	Burned     bool
	PlacedTurn int // the Turn number it entered play / last evolved
}

func (p *InPlay) remainingHP() int { return p.Card.HP - p.Damage }

func (p *InPlay) totalEnergy() int {
	t := 0
	for _, n := range p.Energy {
		t += n
	}
	return t
}

// The Bench is three fixed slots around the Active (spec T7). Slot numbers run
// left to right on screen, and the slot a Pokémon was placed in is kept.
const (
	BenchLeft = iota
	BenchBehind
	BenchRight
	BenchSlots
)

// hasPokemon reports whether the player has any Pokémon left in play.
func (p *Player) hasPokemon() bool { return p.Active != nil || p.BenchCount() > 0 }

// BenchCount is how many Pokémon are on the bench.
func (p *Player) BenchCount() int {
	n := 0
	for _, m := range p.Bench {
		if m != nil {
			n++
		}
	}
	return n
}

// Player is one side's full state.
type Player struct {
	Deck        []Card // top of deck is index 0
	Hand        []Card
	Active      *InPlay
	Bench       [BenchSlots]*InPlay // indexed by slot; nil is an empty slot
	Discard     []Card
	Points      int
	EnergyTypes []Energy
	EnergyZone  Energy // the energy available to attach this turn; NoEnergy if none
	EnergyNext  Energy // next turn's energy, shown small in the zone (spec T4)
	EnergyUsed  bool   // attached energy already this turn
	Retreated   bool   // retreated already this turn

	// DiscardedEnergy is every energy this player has discarded, oldest first,
	// for cards that count discarded energy (spec T6).
	DiscardedEnergy []Energy
}

type Phase uint8

const (
	PhaseSetup Phase = iota
	PhaseMain
	PhaseOver
)

// resumeAction records what the engine still owes once a pending prompt is
// answered — see ADR-0007. It must live in State so replay reconstructs it
// across the pause a prompt creates.
type resumeAction uint8

const (
	resumeNone     resumeAction = iota
	resumeConclude              // after the prompt, run end-of-turn + advance
	resumeAdvance               // after the prompt, just advance to the next turn
)

type PromptKind uint8

const (
	PromptSetup PromptKind = iota
	PromptNewActive
)

// Prompt is a pending decision the engine is blocked on: a specific player owes
// a specific choice before play can continue (ADR-0007, mode 3). Every choice
// is raised through Game.ask, so a prompt with one option never reaches the
// player (spec F8).
type Prompt struct {
	Player  int        // the player who chooses
	Kind    PromptKind // what is being chosen
	Options []Event    // every answer the chooser may give; the cards involved
}

// State is the full game state. It is always derivable by folding the event log
// from a seed (ADR-0005); nothing here is authoritative on its own.
type State struct {
	Players [2]Player
	Active  int // whose turn it is
	Turn    int // turn counter, starts at 1
	Phase   Phase
	Pending *Prompt
	Winner  int       // -1 until the game ends, and on a tie
	Reason  EndReason // why the game ended; NoReason until it does
	Mode    Mode      // who the game is against; set before the first event

	// OpeningFlip is the coin flip that decided FirstPlayer (spec S4), kept so
	// the screen can show who flipped and how it landed.
	OpeningFlip CoinFlip
	FirstPlayer int

	resume resumeAction
	setup  [2]bool // which players have pressed Ready at setup
}

// Ready reports whether player p has placed their setup and pressed Ready
// (spec S5). Until both are Ready, each side's picks are hidden from the other.
func (s *State) Ready(p int) bool { return s.setup[p] }

// CoinFlip records one flip: which player flipped and whether it was heads.
type CoinFlip struct {
	Flipper int
	Heads   bool
}

// --- Events: the canonical log entries (ADR-0005). Player intents and prompt
// responses only — draws, energy generation, coin flips and knockouts are
// re-derived deterministically during the fold, never stored. ---

type Event interface{ event() }

// SetupActive places a Basic from hand as the Active during setup (spec S5).
type SetupActive struct {
	Player int
	CardID string
}

// SetupBench places a Basic from hand into an empty bench slot during setup.
type SetupBench struct {
	Player int
	CardID string
	Slot   int // BenchLeft, BenchBehind or BenchRight
}

// SetupReady ends a player's setup. It needs an Active; once both players are
// Ready the opponent's picks are revealed and turn 1 starts (spec S5).
type SetupReady struct {
	Player int
}

// PlayBasic plays a Basic Pokémon from hand into an empty bench slot.
type PlayBasic struct {
	Player int
	CardID string
	Slot   int // BenchLeft, BenchBehind or BenchRight
}

// AttachEnergy attaches the turn's Energy Zone energy to a Pokémon.
// Target 0 is the Active; Target n is bench slot n-1.
type AttachEnergy struct {
	Player int
	Target int
}

// Evolve evolves an in-play Pokémon using an evolution card from hand.
type Evolve struct {
	Player     int
	HandCardID string
	Target     int // 0 active, n bench slot n-1
}

// Retreat swaps the Active with a benched Pokémon, paying the retreat cost in
// energy.
type Retreat struct {
	Player     int
	BenchIndex int // the bench slot to swap with
}

// UseAttack uses one of the Active Pokémon's attacks. Attacking ends the turn.
type UseAttack struct {
	Player int
	Index  int
}

// EndTurn passes the turn without attacking.
type EndTurn struct {
	Player int
}

// ChooseNewActive answers a "promote a new Active" prompt after a knockout.
type ChooseNewActive struct {
	Player     int
	BenchIndex int // the bench slot to promote from
}

// Concede forfeits the game at any time, from either player; the opponent wins.
type Concede struct {
	Player int
}

func (SetupActive) event()     {}
func (SetupBench) event()      {}
func (SetupReady) event()      {}
func (PlayBasic) event()       {}
func (AttachEnergy) event()    {}
func (Evolve) event()          {}
func (Retreat) event()         {}
func (UseAttack) event()       {}
func (EndTurn) event()         {}
func (ChooseNewActive) event() {}
func (Concede) event()         {}
