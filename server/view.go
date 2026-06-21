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
	Hand         []cardView   `json:"hand"`
	Active       *inPlayView  `json:"active"`
	Bench        []inPlayView `json:"bench"`
	Points       int          `json:"points"`
	DeckCount    int          `json:"deckCount"`
	DiscardCount int          `json:"discardCount"`
	EnergyZone   string       `json:"energyZone"`
	EnergyUsed   bool         `json:"energyUsed"`
	Retreated    bool         `json:"retreated"`
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

func toPlayerView(p engine.Player) playerView {
	hand := make([]cardView, len(p.Hand))
	for i, c := range p.Hand {
		hand[i] = toCardView(c)
	}
	bench := make([]inPlayView, len(p.Bench))
	for i, m := range p.Bench {
		bench[i] = *toInPlayView(m)
	}
	return playerView{
		Hand:         hand,
		Active:       toInPlayView(p.Active),
		Bench:        bench,
		Points:       p.Points,
		DeckCount:    len(p.Deck),
		DiscardCount: len(p.Discard),
		EnergyZone:   energyLabel(p.EnergyZone),
		EnergyUsed:   p.EnergyUsed,
		Retreated:    p.Retreated,
	}
}

func toGameView(g *engine.Game) gameView {
	s := g.S
	var pending *promptView
	if s.Pending != nil {
		pending = &promptView{Player: s.Pending.Player, Kind: promptKindName(s.Pending.Kind)}
	}
	return gameView{
		Players:   [2]playerView{toPlayerView(s.Players[0]), toPlayerView(s.Players[1])},
		Active:    s.Active,
		Turn:      s.Turn,
		Phase:     phaseName(s.Phase),
		Pending:   pending,
		Winner:    s.Winner,
		Narration: g.Narration,
	}
}
