//go:build !js

// The HTTP entrypoint. The same handlers also back the WebAssembly build
// (wasm_main.go), which has no network stack — so everything they need lives in
// main.go and this file holds only the part that is specific to running as a
// local server process.
package main

import (
	"log"
	"net/http"
)

func main() {
	if err := loadCards(cardsPath); err != nil {
		log.Fatal(err)
	}
	var err error
	if fireDeck, err = fireDeckPreset(); err != nil {
		log.Fatal(err)
	}

	if err := newGame(1); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/new", handleNew)
	http.HandleFunc("/api/state", handleState)
	http.HandleFunc("/api/move", handleMove)
	http.HandleFunc("/api/bot", handleBot)
	http.Handle("/", http.FileServer(http.Dir("static")))

	const addr = ":8080"
	log.Printf("pocket-draft simulator on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
