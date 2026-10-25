// Command server runs the pocket-draft simulator as a single local game over
// plain HTTP. One game lives in memory; the browser asks for its state and posts
// moves. This is the hotseat slice — no rooms, no networking between machines
// yet (that arrives with multiplayer). See ADR-0001/0004/0007.
package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// One game, guarded by a mutex so overlapping HTTP requests can't corrupt it.
// The one in-memory game. Both decks are drafted fresh for every new game unless
// the request supplies them, so they vary by seed (issue #9).
var (
	mu   sync.Mutex
	game *engine.Game
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// GET /api/state — the current game as the browser sees it.
func handleState(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, toGameView(game, humanSeat))
}

// newReq is the optional JSON body for POST /api/new. `you` is your deck as a
// list of card IDs; when present you play it (player 0) against the drafted bot
// deck (player 1). Without it you get a drafted deck too. `energy` is your deck's chosen Energy Zone
// types by name (1–3, spec S2); omitted means the default from its attack costs.
type newReq struct {
	You    []string `json:"you"`
	Seed   *uint64  `json:"seed"`
	Energy []string `json:"energy"`
	// Bot pins the bot's deck (card IDs) instead of drafting one, so tests don't
	// shift when the draft pool does.
	Bot []string `json:"bot"`
	// Preset "fire" deals you the old fixed test deck (fireDeckList), which
	// breaks the 2-copy rule, so browser tests written against its hands keep
	// them. Test-only; the app never sends it.
	Preset string `json:"preset"`
}

// energyChoice parses and validates a requested energy choice; nil when none
// was requested, so the engine applies the deck's default.
func energyChoice(names []string) ([]engine.Energy, error) {
	if len(names) == 0 {
		return nil, nil
	}
	types := make([]engine.Energy, len(names))
	for i, n := range names {
		e, ok := energyByName(n)
		if !ok {
			return nil, fmt.Errorf("unknown energy type %q", n)
		}
		types[i] = e
	}
	return types, engine.ValidateEnergyTypes(types)
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

// seedFor returns the requested seed, or a fresh one so every game differs.
// The seed is narrated at game start, so any game can still be replayed exactly.
func seedFor(req newReq) uint64 {
	if req.Seed != nil {
		return *req.Seed
	}
	var b [8]byte
	rand.Read(b[:])
	return binary.LittleEndian.Uint64(b[:])
}

// startGame builds both decks and replaces the in-memory game. It is the whole
// body of POST /api/new minus the HTTP, so the WebAssembly build can reuse it.
func startGame(req newReq) (gameView, *startErr) {
	seed := seedFor(req)

	bot, err := draftBotDeck(seed)
	if err != nil {
		return gameView{}, &startErr{err, http.StatusInternalServerError}
	}

	if len(req.Bot) > 0 {
		if bot, err = deckFromIDs(req.Bot); err != nil {
			return gameView{}, &startErr{err, http.StatusBadRequest}
		}
	}
	p0, err := yourDeck(req, seed)
	if err != nil {
		return gameView{}, &startErr{err, http.StatusBadRequest}
	}
	p1 := bot
	energy, err := energyChoice(req.Energy)
	if err != nil {
		return gameView{}, &startErr{err, http.StatusBadRequest}
	}

	mu.Lock()
	defer mu.Unlock()
	game = engine.NewGameFromDecks(seed, [2]engine.Deck{{Cards: p0, Energy: energy}, {Cards: p1}})
	return toGameView(game, humanSeat), nil
}

// yourDeck is the deck you asked for, or a drafted one when you named none.
// The draft seed is offset so you and the bot never draft the same deck.
func yourDeck(req newReq, seed uint64) ([]engine.Card, error) {
	if len(req.You) > 0 {
		return deckFromIDs(req.You)
	}
	if req.Preset == "fire" {
		return fireDeckPreset()
	}
	return draftBotDeck(seed ^ yourDraftSalt)
}

const yourDraftSalt = 0x5eed_0f_ea_7e_d0_0d

// startErr carries the status code a failed start should report over HTTP.
type startErr struct {
	error
	status int
}

// moveReq is the superset of fields any move might carry; only the ones relevant
// to the given type are read.
type moveReq struct {
	Type       string   `json:"type"`
	Player     int      `json:"player"`
	CardID     string   `json:"cardId"`
	Slot       int      `json:"slot"`
	HandCardID string   `json:"handCardId"`
	Target     int      `json:"target"`
	BenchIndex int      `json:"benchIndex"`
	Index      int      `json:"index"`
	Energy     []string `json:"energy,omitempty"`
	// Spot is where a Trainer is dropped; absent for one with no target.
	Spot *spotView `json:"spot,omitempty"`
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
		return moveResp{OK: false, Error: err.Error(), State: toGameView(game, humanSeat)}
	}
	if err := game.Submit(ev); err != nil {
		return moveResp{OK: false, Error: err.Error(), State: toGameView(game, humanSeat)}
	}
	return moveResp{OK: true, State: toGameView(game, humanSeat)}
}

// toEvent maps a decoded request to the engine's typed Event.
func toEvent(r moveReq) (engine.Event, error) {
	switch r.Type {
	case "SetupActive":
		return engine.SetupActive{Player: r.Player, CardID: r.CardID}, nil
	case "SetupBench":
		return engine.SetupBench{Player: r.Player, CardID: r.CardID, Slot: r.Slot}, nil
	case "SetupReady":
		return engine.SetupReady{Player: r.Player}, nil
	case "SetupReturn":
		return engine.SetupReturn{Player: r.Player, Pos: r.Target}, nil
	case "SetupMove": // from in `index`, to in `target`
		return engine.SetupMove{Player: r.Player, From: r.Index, To: r.Target}, nil
	case "PlayBasic":
		return engine.PlayBasic{Player: r.Player, CardID: r.CardID, Slot: r.Slot}, nil
	case "AttachEnergy":
		return engine.AttachEnergy{Player: r.Player, Target: r.Target}, nil
	case "Evolve":
		return engine.Evolve{Player: r.Player, HandCardID: r.HandCardID, Target: r.Target}, nil
	case "Retreat":
		return engine.Retreat{Player: r.Player, BenchIndex: r.BenchIndex}, nil
	case "PlayTrainer":
		return engine.PlayTrainer{Player: r.Player, CardID: r.CardID, Target: toSpot(r.Spot)}, nil
	case "DiscardFossil":
		return engine.DiscardFossil{Player: r.Player, Pos: r.Target}, nil
	case "UseAbility":
		return engine.UseAbility{Player: r.Player, Pos: r.Target, Target: abilitySpot(r.Spot)}, nil
	case "UseStadium":
		return engine.UseStadium{Player: r.Player, CardID: r.CardID}, nil
	case "UseAttack":
		return engine.UseAttack{Player: r.Player, Index: r.Index}, nil
	case "EndTurn":
		return engine.EndTurn{Player: r.Player}, nil
	case "ChooseNewActive":
		return engine.ChooseNewActive{Player: r.Player, BenchIndex: r.BenchIndex}, nil
	case "RetreatDiscard":
		return retreatDiscard(r)
	case "ChooseOption":
		return engine.ChooseOption{Player: r.Player, Index: r.Index}, nil
	case "Concede":
		return engine.Concede{Player: r.Player}, nil
	}
	return nil, fmt.Errorf("unknown move type %q", r.Type)
}

// retreatDiscard decodes the energy named in a retreat-energy answer.
func retreatDiscard(r moveReq) (engine.Event, error) {
	energy := make([]engine.Energy, len(r.Energy))
	for i, n := range r.Energy {
		e, ok := energyByName(n)
		if !ok {
			return nil, fmt.Errorf("unknown energy type %q", n)
		}
		energy[i] = e
	}
	return engine.RetreatDiscard{Player: r.Player, Energy: energy}, nil
}

// fromEvent is the inverse of toEvent: the request that would submit ev.
func fromEvent(ev engine.Event) moveReq {
	switch e := ev.(type) {
	case engine.SetupActive:
		return moveReq{Type: "SetupActive", Player: e.Player, CardID: e.CardID}
	case engine.SetupBench:
		return moveReq{Type: "SetupBench", Player: e.Player, CardID: e.CardID, Slot: e.Slot}
	case engine.SetupReady:
		return moveReq{Type: "SetupReady", Player: e.Player}
	case engine.SetupReturn:
		return moveReq{Type: "SetupReturn", Player: e.Player, Target: e.Pos}
	case engine.SetupMove:
		return moveReq{Type: "SetupMove", Player: e.Player, Index: e.From, Target: e.To}
	case engine.PlayBasic:
		return moveReq{Type: "PlayBasic", Player: e.Player, CardID: e.CardID, Slot: e.Slot}
	case engine.AttachEnergy:
		return moveReq{Type: "AttachEnergy", Player: e.Player, Target: e.Target}
	case engine.Evolve:
		return moveReq{Type: "Evolve", Player: e.Player, HandCardID: e.HandCardID, Target: e.Target}
	case engine.Retreat:
		return moveReq{Type: "Retreat", Player: e.Player, BenchIndex: e.BenchIndex}
	case engine.PlayTrainer:
		return moveReq{Type: "PlayTrainer", Player: e.Player, CardID: e.CardID, Spot: fromSpot(e.Target)}
	case engine.DiscardFossil:
		return moveReq{Type: "DiscardFossil", Player: e.Player, Target: e.Pos}
	case engine.UseAbility:
		return moveReq{Type: "UseAbility", Player: e.Player, Target: e.Pos, Spot: &spotView{e.Target.Player, e.Target.Pos}}
	case engine.UseStadium:
		return moveReq{Type: "UseStadium", Player: e.Player, CardID: e.CardID}
	case engine.UseAttack:
		return moveReq{Type: "UseAttack", Player: e.Player, Index: e.Index}
	case engine.EndTurn:
		return moveReq{Type: "EndTurn", Player: e.Player}
	case engine.ChooseNewActive:
		return moveReq{Type: "ChooseNewActive", Player: e.Player, BenchIndex: e.BenchIndex}
	case engine.RetreatDiscard:
		return moveReq{Type: "RetreatDiscard", Player: e.Player, Energy: energyLabels(e.Energy)}
	case engine.ChooseOption:
		return moveReq{Type: "ChooseOption", Player: e.Player, Index: e.Index}
	case engine.Concede:
		return moveReq{Type: "Concede", Player: e.Player}
	}
	panic(fmt.Sprintf("fromEvent: no request shape for %T", ev))
}

// abilitySpot is the spot an Ability is aimed at; the zero Spot when it has no
// target (the engine's convention for abilities, unlike Trainers' NoSpot).
func abilitySpot(s *spotView) engine.Spot {
	if s == nil {
		return engine.Spot{}
	}
	return engine.Spot{Player: s.Player, Pos: s.Pos}
}

func toSpot(s *spotView) engine.Spot {
	if s == nil {
		return engine.NoSpot
	}
	return engine.Spot{Player: s.Player, Pos: s.Pos}
}

func fromSpot(s engine.Spot) *spotView {
	if s == engine.NoSpot {
		return nil
	}
	return &spotView{s.Player, s.Pos}
}
