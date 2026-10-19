package engine

// Mode is who the game is against. The zero value is VsBot, so every game is
// against the bot unless a caller says otherwise.
type Mode uint8

const (
	VsBot Mode = iota
	VsPlayer
)

// turnCap is the last turn of the game (spec M3/M4). Every player's turn counts
// as one turn; when this turn ends, the game is a tie regardless of points.
func (m Mode) turnCap() int {
	if m == VsPlayer {
		return 30
	}
	return 50
}

// EndReason is why the game ended (spec F19). NoReason until it does.
type EndReason uint8

const (
	NoReason EndReason = iota
	ReasonPoints
	ReasonNoPokemon
	ReasonConcede
	ReasonTurnCap
)

func (r EndReason) String() string {
	switch r {
	case ReasonPoints:
		return "3 points"
	case ReasonNoPokemon:
		return "no Pokémon left"
	case ReasonConcede:
		return "concede"
	case ReasonTurnCap:
		return "turn cap"
	default:
		return "none"
	}
}

// Outcome is the game's result from one player's point of view.
type Outcome uint8

const (
	Undecided Outcome = iota
	Win
	Loss
	Tie
)

func (o Outcome) String() string {
	switch o {
	case Win:
		return "Win"
	case Loss:
		return "Loss"
	case Tie:
		return "Tie"
	default:
		return "Undecided"
	}
}

// Outcome reports the result for player p. A finished game with no winner is a tie.
func (s *State) Outcome(p int) Outcome {
	switch {
	case s.Phase != PhaseOver:
		return Undecided
	case s.Winner < 0:
		return Tie
	case s.Winner == p:
		return Win
	default:
		return Loss
	}
}

// tie ends the game with no winner.
func (g *Game) tie(reason EndReason) {
	g.S.Phase = PhaseOver
	g.S.Winner = -1
	g.S.Reason = reason
	g.S.Pending = nil
	g.narrate("Game over — tie (%s), %d-%d.", reason, g.S.Players[0].Points, g.S.Players[1].Points)
}
