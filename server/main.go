// Command server runs the pocket-draft simulator as a single local game over
// plain HTTP. One game lives in memory; the browser asks for its state and posts
// moves. This is the hotseat slice — no rooms, no networking between machines
// yet (that arrives with multiplayer). See ADR-0001/0004/0007.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// One game, guarded by a mutex so overlapping HTTP requests can't corrupt it.
// The two curated decks are built once at startup and reused for every new game.
var (
	mu        sync.Mutex
	game      *engine.Game
	fireDeck  []engine.Card
	waterDeck []engine.Card
)

func newGame(seed uint64) {
	game = engine.NewGame(seed, fireDeck, waterDeck)
}

func main() {
	if err := loadCards(cardsPath); err != nil {
		log.Fatal(err)
	}
	var err error
	if fireDeck, waterDeck, err = curatedDecks(); err != nil {
		log.Fatal(err)
	}

	newGame(1)

	http.HandleFunc("/api/new", handleNew)
	http.HandleFunc("/api/state", handleState)
	http.HandleFunc("/api/move", handleMove)
	http.HandleFunc("/api/bot", handleBot)
	http.Handle("/", http.FileServer(http.Dir("static")))

	const addr = ":8080"
	log.Printf("pocket-draft simulator on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
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
// list of card IDs; when present you play it (player 0) against the curated bot
// deck (player 1). An empty body falls back to the curated demo matchup.
type newReq struct {
	You  []string `json:"you"`
	Seed *uint64  `json:"seed"`
}

// POST /api/new — start a fresh game. With a {you:[cardId...], seed} body you play
// your deck vs the bot preset; with no body it's the curated demo matchup. A fixed
// seed makes a game reproducible (same seed -> same shuffles). ?seed=N still works.
func handleNew(w http.ResponseWriter, r *http.Request) {
	var req newReq
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req) // empty/invalid body -> no deck
	}

	seed := uint64(1)
	if req.Seed != nil {
		seed = *req.Seed
	} else if s := r.URL.Query().Get("seed"); s != "" {
		if n, err := strconv.ParseUint(s, 10, 64); err == nil {
			seed = n
		}
	}

	p0, p1 := fireDeck, waterDeck
	if len(req.You) > 0 {
		d, err := deckFromIDs(req.You)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
		p0, p1 = d, fireDeck
	}

	mu.Lock()
	defer mu.Unlock()
	game = engine.NewGame(seed, p0, p1)
	writeJSON(w, toGameView(game))
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
	ev, err := toEvent(req)

	mu.Lock()
	defer mu.Unlock()
	if err != nil {
		writeJSON(w, moveResp{OK: false, Error: err.Error(), State: toGameView(game)})
		return
	}
	if err := game.Submit(ev); err != nil {
		writeJSON(w, moveResp{OK: false, Error: err.Error(), State: toGameView(game)})
		return
	}
	writeJSON(w, moveResp{OK: true, State: toGameView(game)})
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
