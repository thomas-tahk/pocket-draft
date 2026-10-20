package main

import (
	"net/http"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// botMove makes ONE move for whoever the engine is currently waiting on. The
// move always comes from the engine's LegalMoves (botChoice), so Submit never
// refuses it. It is not meant to play well — it exists so the two-player flow
// can be exercised without a second human. Caller must hold mu.
func botMove() bool {
	ev, ok := botChoice(game)
	return ok && game.Submit(ev) == nil
}

// botChoice picks the bot's move from LegalMoves, never Concede. It asks the
// player the engine is waiting on first, then the other, so any prompt — of a
// kind the bot has never heard of included — gets answered. False only when
// the game is over or nobody has a move.
func botChoice(g *engine.Game) (engine.Event, bool) {
	first := g.S.Active
	if g.S.Pending != nil {
		first = g.S.Pending.Player
	}
	if g.S.Phase == engine.PhaseSetup {
		first = 1 - humanSeat // the bot places while the human still is
	}
	for _, p := range []int{first, 1 - first} {
		if ev, ok := bestMove(g.LegalMoves(p)); ok {
			return ev, true
		}
	}
	return nil, false
}

// bestMove is the highest-ranked move, the earliest listed on a tie.
func bestMove(moves []engine.Event) (engine.Event, bool) {
	var best engine.Event
	bestRank := 0
	for _, ev := range moves {
		r, ok := moveRank(ev)
		if ok && (best == nil || r > bestRank) {
			best, bestRank = ev, r
		}
	}
	return best, best != nil
}

// moveRank is the bot's greedy policy: at setup place the Active, fill the
// bench, then press Ready; in play get energy onto the Active, swing with it,
// otherwise pass. Main-phase moves it has no policy
// for rank below EndTurn, so it skips them; a prompt's answers are the only
// moves listed, so they are chosen whatever their kind. Concede is never ok.
func moveRank(ev engine.Event) (int, bool) {
	switch e := ev.(type) {
	case engine.Concede:
		return 0, false
	case engine.SetupActive:
		return 3, true
	case engine.SetupBench:
		return 2, true
	case engine.SetupReady:
		return 1, true
	case engine.SetupReturn, engine.SetupMove:
		return 0, false
	case engine.AttachEnergy:
		if e.Target == 0 {
			return 3, true
		}
	case engine.UseAttack:
		return 2, true
	case engine.EndTurn:
		return 1, true
	}
	return 0, true
}

type botResp struct {
	Acted bool     `json:"acted"`
	State gameView `json:"state"`
}

// POST /api/bot — make one bot move for the current actor.
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
