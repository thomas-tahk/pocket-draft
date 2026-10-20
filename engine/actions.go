package engine

// Action is the replay/history record of one applied move (spec F14, F15):
// who did what, with which card, to which spot, and what came of it. Actions
// are derived during the fold like Narration — never stored — so replaying
// the event log reproduces them exactly. Actions[i] records Events[i].
type Action struct {
	Player int
	Kind   ActionKind
	Card   Card       // the card played or acting; zero (ID "") when none
	Target Spot       // the spot it touched; NoSpot when none
	Result []string   // the narration lines this move produced
	Flips  []CoinFlip // every coin this move flipped, in order (spec F11)
}

// ActionKind names the move an Action records.
type ActionKind string

const (
	ActSetup        ActionKind = "setup" // a Basic placed at setup (Active or bench)
	ActReady        ActionKind = "ready"
	ActPlayBasic    ActionKind = "play_basic"
	ActAttachEnergy ActionKind = "attach_energy"
	ActEvolve       ActionKind = "evolve"
	ActRetreat      ActionKind = "retreat"
	ActRetreatPay   ActionKind = "retreat_pay"
	ActAbility      ActionKind = "ability"
	ActAttack       ActionKind = "attack"
	ActEndTurn      ActionKind = "end_turn"
	ActPromote      ActionKind = "promote"
	ActConcede      ActionKind = "concede"
)

// Spot is a place on the board: a player's Active (Pos 0) or bench slot
// Pos-1 — the same numbering as AttachEnergy.Target.
type Spot struct {
	Player int
	Pos    int
}

// NoSpot marks an Action that touched no Pokémon (end turn, concede).
var NoSpot = Spot{Player: -1, Pos: -1}

// describe builds the Action for ev from the state before it is applied.
// Lookups are guarded because ev has not been validated yet; an invalid ev is
// refused by validateAndApply and its Action is dropped.
func (g *Game) describe(ev Event) Action {
	switch e := ev.(type) {
	case SetupActive:
		return Action{e.Player, ActSetup, g.handCard(e.Player, e.CardID), Spot{e.Player, 0}, nil, nil}
	case SetupBench:
		return Action{e.Player, ActSetup, g.handCard(e.Player, e.CardID), Spot{e.Player, e.Slot + 1}, nil, nil}
	case SetupReady:
		return Action{e.Player, ActReady, Card{}, NoSpot, nil, nil}
	case PlayBasic:
		return Action{e.Player, ActPlayBasic, g.handCard(e.Player, e.CardID), Spot{e.Player, e.Slot + 1}, nil, nil}
	case AttachEnergy:
		return Action{e.Player, ActAttachEnergy, Card{}, Spot{e.Player, e.Target}, nil, nil}
	case Evolve:
		return Action{e.Player, ActEvolve, g.handCard(e.Player, e.HandCardID), Spot{e.Player, e.Target}, nil, nil}
	case Retreat:
		return Action{e.Player, ActRetreat, g.monCard(e.Player, 0), Spot{e.Player, e.BenchIndex + 1}, nil, nil}
	case RetreatDiscard:
		return Action{e.Player, ActRetreatPay, g.monCard(e.Player, 0), Spot{e.Player, 0}, nil, nil}
	case PlayTrainer:
		return Action{e.Player, ActPlayTrainer, g.handCard(e.Player, e.CardID), e.Target, nil, nil}
	case DiscardFossil:
		return Action{e.Player, ActDiscardFossil, g.monCard(e.Player, e.Pos), Spot{e.Player, e.Pos}, nil, nil}
	case UseAbility:
		return Action{e.Player, ActAbility, g.monCard(e.Player, e.Pos), Spot{e.Player, e.Pos}, nil, nil}
	case ChooseOption:
		return Action{e.Player, ActChoose, Card{}, NoSpot, nil, nil}
	case UseStadium:
		return Action{e.Player, ActUseStadium, g.stadiumCard(), StadiumSpot, nil, nil}
	case UseAttack:
		return Action{e.Player, ActAttack, g.monCard(e.Player, 0), Spot{1 - e.Player, 0}, nil, nil}
	case EndTurn:
		return Action{e.Player, ActEndTurn, Card{}, NoSpot, nil, nil}
	case ChooseNewActive:
		return Action{e.Player, ActPromote, g.monCard(e.Player, e.BenchIndex+1), Spot{e.Player, e.BenchIndex + 1}, nil, nil}
	case Concede:
		return Action{e.Player, ActConcede, Card{}, NoSpot, nil, nil}
	}
	return Action{Target: NoSpot}
}

func (g *Game) handCard(p int, id string) Card {
	if p != 0 && p != 1 {
		return Card{}
	}
	hand := g.S.Players[p].Hand
	if i := indexInHand(hand, id); i >= 0 {
		return hand[i]
	}
	return Card{}
}

func (g *Game) monCard(p, pos int) Card {
	if p != 0 && p != 1 {
		return Card{}
	}
	if mon := targetMon(&g.S.Players[p], pos); mon != nil {
		return mon.Card
	}
	return Card{}
}
