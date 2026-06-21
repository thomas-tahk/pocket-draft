package main

import (
	"net/http"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// botMove makes ONE legal move for whoever the engine is currently waiting on,
// using a deliberately predictable greedy policy: get energy onto the Active,
// swing with it, otherwise pass. It is not meant to play well — it exists so the
// two-player flow can be exercised without a second human.
//
// It leans on the engine's "block the illegal" guarantee (ADR-0004/0007): a
// rejected Submit changes nothing, so we can try candidate moves in priority
// order and keep the first that succeeds. Caller must hold mu.
func botMove() bool {
	s := game.S
	if s.Phase == engine.PhaseOver {
		return false
	}
	actor := s.Active
	if s.Pending != nil {
		actor = s.Pending.Player
	}
	for _, ev := range candidates(actor) {
		if game.Submit(ev) == nil {
			return true
		}
	}
	return false
}

// candidates lists moves to try for the actor, best-effort first. EndTurn is the
// always-legal fallback in the main phase, so a turn can never stall.
func candidates(actor int) []engine.Event {
	s := game.S
	pl := s.Players[actor]

	if s.Pending != nil {
		switch s.Pending.Kind {
		case engine.PromptSetup:
			return []engine.Event{setupChoice(actor, pl)}
		case engine.PromptNewActive:
			var evs []engine.Event
			for bi := range pl.Bench {
				evs = append(evs, engine.ChooseNewActive{Player: actor, BenchIndex: bi})
			}
			return evs
		}
	}

	// Main phase: attach to the Active, then try each attack, then pass.
	evs := []engine.Event{engine.AttachEnergy{Player: actor, Target: 0}}
	if pl.Active != nil {
		for i := range pl.Active.Card.Attacks {
			evs = append(evs, engine.UseAttack{Player: actor, Index: i})
		}
	}
	return append(evs, engine.EndTurn{Player: actor})
}

// setupChoice puts the first Basic in hand into the Active spot and the next few
// Basics (up to the 3-slot bench) behind it.
func setupChoice(actor int, pl engine.Player) engine.SetupPlace {
	var active string
	var bench []string
	for _, c := range pl.Hand {
		if c.Stage != engine.Basic {
			continue
		}
		if active == "" {
			active = c.ID
		} else if len(bench) < 3 {
			bench = append(bench, c.ID)
		}
	}
	return engine.SetupPlace{Player: actor, ActiveCardID: active, BenchCardIDs: bench}
}

type botResp struct {
	Acted bool     `json:"acted"`
	State gameView `json:"state"`
}

// POST /api/bot — make one greedy move for the current actor.
func handleBot(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	acted := botMove()
	writeJSON(w, botResp{Acted: acted, State: toGameView(game)})
}
