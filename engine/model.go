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

// Player is one side's full state.
type Player struct {
	Deck        []Card // top of deck is index 0
	Hand        []Card
	Active      *InPlay
	Bench       []*InPlay // max 3
	Discard     []Card
	Points      int
	EnergyTypes []Energy
	EnergyZone  Energy // the energy available to attach this turn; NoEnergy if none
	EnergyUsed  bool   // attached energy already this turn
	Retreated   bool   // retreated already this turn
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
	resumeNone resumeAction = iota
	resumeConclude              // after the prompt, run end-of-turn + advance
	resumeAdvance               // after the prompt, just advance to the next turn
)

type PromptKind uint8

const (
	PromptSetup PromptKind = iota
	PromptNewActive
)

// Prompt is a pending decision the engine is blocked on: a specific player owes
// a specific choice before play can continue (ADR-0007, mode 3).
type Prompt struct {
	Player int
	Kind   PromptKind
}

// State is the full game state. It is always derivable by folding the event log
// from a seed (ADR-0005); nothing here is authoritative on its own.
type State struct {
	Players [2]Player
	Active  int // whose turn it is
	Turn    int // turn counter, starts at 1
	Phase   Phase
	Pending *Prompt
	Winner  int // -1 until the game ends

	resume      resumeAction
	firstPlayer int
	setup       [2]bool
}

// --- Events: the canonical log entries (ADR-0005). Player intents and prompt
// responses only — draws, energy generation, coin flips and knockouts are
// re-derived deterministically during the fold, never stored. ---

type Event interface{ event() }

// SetupPlace answers a setup prompt: place one Basic as Active and up to three
// more on the bench.
type SetupPlace struct {
	Player       int
	ActiveCardID string
	BenchCardIDs []string
}

// PlayBasic plays a Basic Pokémon from hand to the bench.
type PlayBasic struct {
	Player int
	CardID string
}

// AttachEnergy attaches the turn's Energy Zone energy to a Pokémon.
// Target 0 is the Active; Target n is bench index n-1.
type AttachEnergy struct {
	Player int
	Target int
}

// Evolve evolves an in-play Pokémon using an evolution card from hand.
type Evolve struct {
	Player     int
	HandCardID string
	Target     int // 0 active, n bench[n-1]
}

// Retreat swaps the Active with a benched Pokémon, paying the retreat cost in
// energy.
type Retreat struct {
	Player     int
	BenchIndex int
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
	BenchIndex int
}

// Concede forfeits the game on the conceding player's turn; the opponent wins.
type Concede struct {
	Player int
}

func (SetupPlace) event()      {}
func (PlayBasic) event()       {}
func (AttachEnergy) event()    {}
func (Evolve) event()          {}
func (Retreat) event()         {}
func (UseAttack) event()       {}
func (EndTurn) event()         {}
func (ChooseNewActive) event() {}
func (Concede) event()         {}
