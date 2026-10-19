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
			return []engine.Event{setupStep(actor, pl)}
		case engine.PromptNewActive:
			var evs []engine.Event
			for bi, m := range pl.Bench {
				if m == nil {
					continue
				}
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

// setupStep is the bot's next setup move, one card at a time: the first Basic
// in hand to the Active, then Basics into the empty bench slots left to right,
// then Ready.
func setupStep(actor int, pl engine.Player) engine.Event {
	for _, c := range pl.Hand {
		if c.Stage != engine.Basic {
			continue
		}
		if pl.Active == nil {
			return engine.SetupActive{Player: actor, CardID: c.ID}
		}
		for slot, m := range pl.Bench {
			if m == nil {
				return engine.SetupBench{Player: actor, CardID: c.ID, Slot: slot}
			}
		}
	}
	return engine.SetupReady{Player: actor}
}

type botResp struct {
	Acted bool     `json:"acted"`
	State gameView `json:"state"`
}

// POST /api/bot — make one greedy move for the current actor.
func handleBot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, botStep())
}

// botStep takes the bot's single move. Split out of the handler so the
// WebAssembly build, which has no HTTP, can call the same code.
func botStep() botResp {
	mu.Lock()
	defer mu.Unlock()
	acted := botMove()
	return botResp{Acted: acted, State: toGameView(game, humanSeat)}
}
