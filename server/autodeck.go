package main

import (
	"encoding/json"
	"io"
	"net/http"
)

// autoDeck is a whole draft done for you: the picks and the 20-card deck built
// from them. It is the "Auto-draft" test shortcut, so testing a battle doesn't
// need a full draft every time.
type autoDeck struct {
	Picks []string `json:"picks"`
	Deck  []string `json:"deck"`
}

// autoDeckFor drafts a deck the way the bot does. body is optional JSON
// {"seed": n}; without a seed every call drafts a different deck.
func autoDeckFor(body string) (autoDeck, error) {
	var req newReq
	if body != "" {
		_ = json.Unmarshal([]byte(body), &req) // empty/invalid body -> random seed
	}
	picks, deck, err := draftBotIDs(seedFor(req) ^ yourDraftSalt)
	return autoDeck{Picks: picks, Deck: deck}, err
}

// POST /api/autodeck — draft a legal deck for you.
func handleAutoDeck(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	deck, err := autoDeckFor(string(body))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, deck)
}
