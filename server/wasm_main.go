//go:build js && wasm

// The WebAssembly entrypoint. The same engine and the same handler logic that
// main.go serves over HTTP, compiled into the browser instead — so the game
// runs with no server at all and no round-trip per move (ADR-0006: this is the
// client-side predictor, and it is the identical Go code rather than a
// TypeScript reimplementation that drifts).
//
// It exposes one function, __pdCall(path, body) -> json, mirroring the HTTP API
// exactly. The browser's fetch shim routes /api/* here, so nothing in the React
// app has to know which build it is talking to.
package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
)

func main() {
	js.Global().Set("__pdInit", js.FuncOf(pdInit))
	js.Global().Set("__pdCall", js.FuncOf(pdCall))
	select {} // keep the Go runtime alive for the page's lifetime
}

// pdInit(cardsJSON) loads the scraped card data the page already fetched and
// starts an opening game. Returns "" on success, an error string otherwise.
func pdInit(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return "pdInit: expected the card JSON"
	}
	if err := loadCardsFromJSON([]byte(args[0].String())); err != nil {
		return err.Error()
	}
	var err error
	if fireDeck, err = fireDeckPreset(); err != nil {
		return err.Error()
	}
	if err := newGame(1); err != nil {
		return err.Error()
	}
	return ""
}

// pdCall(path, body) runs one API call and returns the response JSON as a
// string. `body` may be empty for calls that take none.
func pdCall(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errJSON(fmt.Errorf("pdCall: expected a path"))
	}
	path := args[0].String()
	body := ""
	if len(args) > 1 && args[1].Type() == js.TypeString {
		body = args[1].String()
	}

	switch path {
	case "/api/state":
		mu.Lock()
		defer mu.Unlock()
		return marshal(toGameView(game))

	case "/api/new":
		var req newReq
		if body != "" {
			_ = json.Unmarshal([]byte(body), &req) // empty/invalid body -> no deck
		}
		view, err := startGame(req)
		if err != nil {
			return errJSON(err)
		}
		return marshal(view)

	case "/api/move":
		var req moveReq
		if err := json.Unmarshal([]byte(body), &req); err != nil {
			return errJSON(fmt.Errorf("bad request body"))
		}
		return marshal(applyMove(req))

	case "/api/bot":
		return marshal(botStep())
	}
	return errJSON(fmt.Errorf("unknown path %q", path))
}

func marshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return errJSON(err)
	}
	return string(b)
}

func errJSON(err error) string {
	b, _ := json.Marshal(map[string]string{"error": err.Error()})
	return string(b)
}
