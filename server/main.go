// Command server runs the pocket-draft simulator as a single local game over
// plain HTTP. One game lives in memory; the browser asks for its state and posts
// moves. This is the hotseat slice — no rooms, no networking between machines
// yet (that arrives with multiplayer). See ADR-0001/0004/0007.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// One game, guarded by a mutex so overlapping HTTP requests can't corrupt it.
// Player 1's curated stand-in deck is built once at startup and reused; player
// 2's (the bot's) deck is drafted fresh for every new game so it varies by seed
// (issue #9).
var (
	mu       sync.Mutex
	game     *engine.Game
	fireDeck []engine.Card
)

func newGame(seed uint64) error {
	bot, err := draftBotDeck(seed)
	if err != nil {
		return err
	}
	game = engine.NewGame(seed, fireDeck, bot)
	return nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// GET /api/state — the current game as the browser sees it.
func handleState(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, toGameView(game))
}

// newReq is the optional JSON body for POST /api/new. `you` is your deck as a
// list of card IDs; when present you play it (player 0) against the drafted bot
// deck (player 1). An empty body falls back to the demo matchup (player 1's
// curated stand-in vs the bot).
type newReq struct {
	You  []string `json:"you"`
	Seed *uint64  `json:"seed"`
}

// POST /api/new — start a fresh game. With a {you:[cardId...], seed} body you play
// your deck vs a freshly drafted bot deck; with no body it's the demo matchup. A
// fixed seed makes a game reproducible (same seed -> same shuffles and the same
// bot deck). ?seed=N still works.
func handleNew(w http.ResponseWriter, r *http.Request) {
	var req newReq
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req) // empty/invalid body -> no deck
	}
	if req.Seed == nil {
		if s := r.URL.Query().Get("seed"); s != "" {
			if n, err := strconv.ParseUint(s, 10, 64); err == nil {
				req.Seed = &n
			}
		}
	}
	view, err := startGame(req)
	if err != nil {
		w.WriteHeader(err.status)
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, view)
}

// startGame builds both decks and replaces the in-memory game. It is the whole
// body of POST /api/new minus the HTTP, so the WebAssembly build can reuse it.
func startGame(req newReq) (gameView, *startErr) {
	seed := uint64(1)
	if req.Seed != nil {
		seed = *req.Seed
	}

	bot, err := draftBotDeck(seed)
	if err != nil {
		return gameView{}, &startErr{err, http.StatusInternalServerError}
	}

	p0, p1 := fireDeck, bot
	if len(req.You) > 0 {
		d, err := deckFromIDs(req.You)
		if err != nil {
			return gameView{}, &startErr{err, http.StatusBadRequest}
		}
		p0 = d
	}

	mu.Lock()
	defer mu.Unlock()
	game = engine.NewGame(seed, p0, p1)
	return toGameView(game), nil
}

// startErr carries the status code a failed start should report over HTTP.
type startErr struct {
	error
	status int
}

// moveReq is the superset of fields any move might carry; only the ones relevant
// to the given type are read.
type moveReq struct {
	Type         string   `json:"type"`
	Player       int      `json:"player"`
	CardID       string   `json:"cardId"`
	ActiveCardID string   `json:"activeCardId"`
	BenchCardIDs []string `json:"benchCardIds"`
	HandCardID   string   `json:"handCardId"`
	Target       int      `json:"target"`
	BenchIndex   int      `json:"benchIndex"`
	Index        int      `json:"index"`
}

type moveResp struct {
	OK    bool     `json:"ok"`
	Error string   `json:"error,omitempty"`
	State gameView `json:"state"`
}

// POST /api/move — validate and apply one move, then return the new state. An
// illegal move changes nothing and comes back with ok:false and a reason.
func handleMove(w http.ResponseWriter, r *http.Request) {
	var req moveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}
	writeJSON(w, applyMove(req))
}

// applyMove validates and applies one move. An illegal move changes nothing and
// comes back with ok:false and a reason.
func applyMove(req moveReq) moveResp {
	ev, err := toEvent(req)

	mu.Lock()
	defer mu.Unlock()
	if err != nil {
		return moveResp{OK: false, Error: err.Error(), State: toGameView(game)}
	}
	if err := game.Submit(ev); err != nil {
		return moveResp{OK: false, Error: err.Error(), State: toGameView(game)}
	}
	return moveResp{OK: true, State: toGameView(game)}
}

// toEvent maps a decoded request to the engine's typed Event.
func toEvent(r moveReq) (engine.Event, error) {
	switch r.Type {
	case "SetupPlace":
		return engine.SetupPlace{Player: r.Player, ActiveCardID: r.ActiveCardID, BenchCardIDs: r.BenchCardIDs}, nil
	case "PlayBasic":
		return engine.PlayBasic{Player: r.Player, CardID: r.CardID}, nil
	case "AttachEnergy":
		return engine.AttachEnergy{Player: r.Player, Target: r.Target}, nil
	case "Evolve":
		return engine.Evolve{Player: r.Player, HandCardID: r.HandCardID, Target: r.Target}, nil
	case "Retreat":
		return engine.Retreat{Player: r.Player, BenchIndex: r.BenchIndex}, nil
	case "UseAttack":
		return engine.UseAttack{Player: r.Player, Index: r.Index}, nil
	case "EndTurn":
		return engine.EndTurn{Player: r.Player}, nil
	case "ChooseNewActive":
		return engine.ChooseNewActive{Player: r.Player, BenchIndex: r.BenchIndex}, nil
	case "Concede":
		return engine.Concede{Player: r.Player}, nil
	}
	return nil, fmt.Errorf("unknown move type %q", r.Type)
}
