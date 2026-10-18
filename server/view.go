package main

// This file is the "view" layer: it translates the engine's internal game state
// into plain, JSON-friendly structs the browser can render. The engine never
// learns about JSON or presentation — that separation is the whole point
// (ADR-0004: the client renders state, it does not own rules).

import "github.com/thomas-tahk/pocket-draft/engine"

type attackView struct {
	Name   string   `json:"name"`
	Cost   []string `json:"cost"`
	Damage int      `json:"damage"`
}

type cardView struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Image       string       `json:"image,omitempty"`
	Stage       string       `json:"stage"`
	HP          int          `json:"hp"`
	Type        string       `json:"type"`
	IsEX        bool         `json:"isEx"`
	EvolvesFrom string       `json:"evolvesFrom,omitempty"`
	RetreatCost int          `json:"retreatCost"`
	Weakness    string       `json:"weakness,omitempty"`
	Attacks     []attackView `json:"attacks"`
}

type inPlayView struct {
	Card        cardView       `json:"card"`
	Damage      int            `json:"damage"`
	RemainingHP int            `json:"remainingHp"`
	Energy      map[string]int `json:"energy"`
	TotalEnergy int            `json:"totalEnergy"`
	Poisoned    bool           `json:"poisoned"`
	Burned      bool           `json:"burned"`
}

type playerView struct {
	Hand         []cardView    `json:"hand"`
	HandCount    int           `json:"handCount"`
	Active       *inPlayView   `json:"active"`
	Bench        []*inPlayView `json:"bench"` // one entry per slot (left, behind, right); null is empty
	Points       int           `json:"points"`
	DeckCount    int           `json:"deckCount"`
	DiscardCount int           `json:"discardCount"`
	EnergyZone   string        `json:"energyZone"`
	EnergyTypes  []string      `json:"energyTypes"` // the deck's Energy Zone types (spec S2)
	EnergyUsed   bool          `json:"energyUsed"`
	Retreated    bool          `json:"retreated"`
	// LegalMoves is everything the viewer may do right now, each in the exact
	// shape POST /api/move accepts. Only filled on the viewer's own side.
	LegalMoves []moveReq `json:"legalMoves"`
}

type promptView struct {
	Player int    `json:"player"`
	Kind   string `json:"kind"`
}

type gameView struct {
	Players   [2]playerView `json:"players"`
	Active    int           `json:"active"`
	Turn      int           `json:"turn"`
	Phase     string        `json:"phase"`
	Pending   *promptView   `json:"pending"`
	Winner    int           `json:"winner"`
	Narration []string      `json:"narration"`
	Actions   []actionView  `json:"actions"`
}

type spotView struct {
	Player int `json:"player"`
	Pos    int `json:"pos"` // 0 is the Active, n is bench slot n-1; -1 is none
}

// actionView is one applied move for the opponent-turn playback and the
// history list (spec F14, F15).
type actionView struct {
	Player int       `json:"player"`
	Kind   string    `json:"kind"`
	Card   *cardView `json:"card"` // null when the move used no card, or it is hidden
	Target spotView  `json:"target"`
	Result []string  `json:"result"`
}

func stageName(s engine.Stage) string {
	switch s {
	case engine.Basic:
		return "Basic"
	case engine.Stage1:
		return "Stage 1"
	case engine.Stage2:
		return "Stage 2"
	}
	return "?"
}

func phaseName(p engine.Phase) string {
	switch p {
	case engine.PhaseSetup:
		return "setup"
	case engine.PhaseMain:
		return "main"
	case engine.PhaseOver:
		return "over"
	}
	return "?"
}

func promptKindName(k engine.PromptKind) string {
	if k == engine.PromptNewActive {
		return "new_active"
	}
	return "setup"
}

func energyLabel(e engine.Energy) string {
	if e == engine.NoEnergy {
		return ""
	}
	return e.String()
}

func energyLabels(es []engine.Energy) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.String()
	}
	return out
}

func toCardView(c engine.Card) cardView {
	atks := make([]attackView, len(c.Attacks))
	for i, a := range c.Attacks {
		cost := make([]string, len(a.Cost))
		for j, e := range a.Cost {
			cost[j] = e.String()
		}
		atks[i] = attackView{Name: a.Name, Cost: cost, Damage: a.Damage}
	}
	return cardView{
		ID:          c.ID,
		Name:        c.Name,
		Image:       cardImages[c.ID],
		Stage:       stageName(c.Stage),
		HP:          c.HP,
		Type:        c.Type.String(),
		IsEX:        c.IsEX,
		EvolvesFrom: c.EvolvesFrom,
		RetreatCost: c.RetreatCost,
		Weakness:    energyLabel(c.Weakness),
		Attacks:     atks,
	}
}

func toInPlayView(p *engine.InPlay) *inPlayView {
	if p == nil {
		return nil
	}
	energy := map[string]int{}
	total := 0
	for e, n := range p.Energy {
		if n > 0 {
			energy[e.String()] = n
			total += n
		}
	}
	return &inPlayView{
		Card:        toCardView(p.Card),
		Damage:      p.Damage,
		RemainingHP: p.Card.HP - p.Damage,
		Energy:      energy,
		TotalEnergy: total,
		Poisoned:    p.Poisoned,
		Burned:      p.Burned,
	}
}

// toPlayerView renders one player's side. Only the viewer's own hand is shown;
// for anyone else the hand is empty and only its size is reported. The deck is
// only ever a count, so its order never leaves the server.
func toPlayerView(p engine.Player, isViewer bool) playerView {
	hand := []cardView{}
	if isViewer {
		for _, c := range p.Hand {
			hand = append(hand, toCardView(c))
		}
	}
	bench := make([]*inPlayView, len(p.Bench))
	for i, m := range p.Bench {
		bench[i] = toInPlayView(m)
	}
	return playerView{
		Hand:         hand,
		HandCount:    len(p.Hand),
		Active:       toInPlayView(p.Active),
		Bench:        bench,
		Points:       p.Points,
		DeckCount:    len(p.Deck),
		DiscardCount: len(p.Discard),
		EnergyZone:   energyLabel(p.EnergyZone),
		EnergyTypes:  energyLabels(p.EnergyTypes),
		EnergyUsed:   p.EnergyUsed,
		Retreated:    p.Retreated,
	}
}

// humanSeat is the player the browser plays; the bot is the other seat.
const humanSeat = 0

// toGameView is the game as seen from viewer's seat. It is the one place hidden
// information is removed: anything the viewer may not see is stripped here.
func toGameView(g *engine.Game, viewer int) gameView {
	s := g.S
	var pending *promptView
	if s.Pending != nil {
		pending = &promptView{Player: s.Pending.Player, Kind: promptKindName(s.Pending.Kind)}
	}
	players := [2]playerView{toPlayerView(s.Players[0], viewer == 0), toPlayerView(s.Players[1], viewer == 1)}
	players[viewer].LegalMoves = toMoveReqs(g.LegalMoves(viewer))
	players[1-viewer].LegalMoves = []moveReq{}
	return gameView{
		Players:   players,
		Active:    s.Active,
		Turn:      s.Turn,
		Phase:     phaseName(s.Phase),
		Pending:   pending,
		Winner:    s.Winner,
		Narration: g.Narration,
		Actions:   toActionViews(g, viewer),
	}
}

func toMoveReqs(evs []engine.Event) []moveReq {
	out := make([]moveReq, len(evs))
	for i, ev := range evs {
		out[i] = fromEvent(ev)
	}
	return out
}

// toActionViews renders the action log for viewer. Setup is placed in secret,
// so an opponent's setup pick shows only that they placed until the game has
// left setup (Turn is 0 until then).
func toActionViews(g *engine.Game, viewer int) []actionView {
	out := make([]actionView, 0, len(g.Actions))
	for _, a := range g.Actions {
		v := actionView{Player: a.Player, Kind: string(a.Kind), Target: spotView{a.Target.Player, a.Target.Pos}, Result: a.Result}
		if a.Card.ID != "" {
			c := toCardView(a.Card)
			v.Card = &c
		}
		if a.Kind == engine.ActSetup && a.Player != viewer && g.S.Turn == 0 {
			v.Card, v.Result = nil, []string{}
		}
		out = append(out, v)
	}
	return out
}
