package main

// This file is the data bridge: it loads the real Pocket card data the draft
// tool already scrapes (public/data/cards.json) and converts the cards we use
// into the engine's presentation-free Card model. Effects are deferred (ADR-0002):
// only the fields the engine actually runs — stats, type, weakness, retreat, EX,
// and base-damage attacks — are carried over. An attack's "+"/"x" suffix and its
// effect text are dropped for now, so cards play at base-damage fidelity.
//
// Card art (image URLs) is kept out of the engine and stashed here in cardImages,
// so the view layer can show real card pictures without the engine ever learning
// about presentation.

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/thomas-tahk/pocket-draft/engine"
)

// cardsPath is relative to the server's working directory. Run the server with
// `cd server && go run .` so this resolves to the repo's scraped card data.
const cardsPath = "../public/data/cards.json"

type rawAttack struct {
	Cost   string `json:"cost"`
	Name   string `json:"name"`
	Damage string `json:"damage"`
	Effect string `json:"effect"`
}

type rawCard struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	CardType    string      `json:"cardType"`
	Stage       string      `json:"stage"`
	EvolvesFrom string      `json:"evolvesFrom"`
	HP          int         `json:"hp"`
	ExKind      string      `json:"exKind"`
	Weakness    string      `json:"weakness"`
	Retreat     int         `json:"retreat"`
	Attacks     []rawAttack `json:"attacks"`
	ImageThumb  string      `json:"imageThumb"`
}

// rawByID indexes the scraped cards by ID; cardImages maps card ID -> thumbnail
// URL for the view layer. Both are populated by loadCards.
var (
	rawByID    = map[string]rawCard{}
	cardImages = map[string]string{}
)

// loadCards reads the scraped card JSON into rawByID and cardImages. Called once
// at startup.
func loadCards(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w (run the server from the server/ directory)", path, err)
	}
	var raws []rawCard
	if err := json.Unmarshal(b, &raws); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	for _, rc := range raws {
		rawByID[rc.ID] = rc
		cardImages[rc.ID] = rc.ImageThumb
	}
	return nil
}

// energyByLetter maps Limitless attack-cost letters to engine energies.
var energyByLetter = map[rune]engine.Energy{
	'G': engine.Grass, 'R': engine.Fire, 'W': engine.Water, 'L': engine.Lightning,
	'P': engine.Psychic, 'F': engine.Fighting, 'D': engine.Darkness, 'M': engine.Metal,
	'C': engine.Colorless,
}

// energyByName maps a type/weakness name to an engine energy. Ok is false for
// names the engine does not model (e.g. Dragon, or an empty weakness).
func energyByName(name string) (engine.Energy, bool) {
	switch name {
	case "Grass":
		return engine.Grass, true
	case "Fire":
		return engine.Fire, true
	case "Water":
		return engine.Water, true
	case "Lightning":
		return engine.Lightning, true
	case "Psychic":
		return engine.Psychic, true
	case "Fighting":
		return engine.Fighting, true
	case "Darkness":
		return engine.Darkness, true
	case "Metal":
		return engine.Metal, true
	case "Colorless":
		return engine.Colorless, true
	case "Dragon":
		return engine.Dragon, true
	}
	return engine.NoEnergy, false
}

// parseCost turns a cost string like "RCC" into engine energies. Non-letter
// characters (e.g. the "0" used for a free attack) are skipped.
func parseCost(s string) []engine.Energy {
	var cost []engine.Energy
	for _, ch := range s {
		if e, ok := energyByLetter[ch]; ok {
			cost = append(cost, e)
		}
	}
	return cost
}

// parseDamage reads the leading integer of a damage string, dropping any "+"/"x"
// suffix (whose meaning lives in the deferred effect text). "" -> 0.
func parseDamage(s string) int {
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int(ch-'0')
	}
	return n
}

func stageFromName(s string) engine.Stage {
	switch s {
	case "Stage 1":
		return engine.Stage1
	case "Stage 2":
		return engine.Stage2
	default:
		return engine.Basic
	}
}

// toEngineCard converts a scraped card into the engine's Card. It errors only if
// the Pokémon's type is one the engine does not model (Dragon), which keeps
// unsupported cards out of a deck rather than silently mistyping them.
func toEngineCard(rc rawCard) (engine.Card, error) {
	typ, ok := energyByName(rc.CardType)
	if !ok {
		return engine.Card{}, fmt.Errorf("card %s (%s): unsupported type %q", rc.ID, rc.Name, rc.CardType)
	}
	weak, _ := energyByName(rc.Weakness) // unknown/empty weakness -> NoEnergy

	attacks := make([]engine.Attack, 0, len(rc.Attacks))
	for _, a := range rc.Attacks {
		attacks = append(attacks, engine.Attack{
			Name:   a.Name,
			Cost:   parseCost(a.Cost),
			Damage: parseDamage(a.Damage),
		})
	}

	return engine.Card{
		ID:          rc.ID,
		Name:        rc.Name,
		Stage:       stageFromName(rc.Stage),
		EvolvesFrom: rc.EvolvesFrom,
		HP:          rc.HP,
		Type:        typ,
		Attacks:     attacks,
		Weakness:    weak,
		RetreatCost: rc.Retreat,
		IsEX:        rc.ExKind != "regular",
	}, nil
}

// deckEntry is a curated card plus how many copies to include.
type deckEntry struct {
	id    string
	count int
}

// Two curated "preset AI" decks of real cards — a stepping stone to playing
// drafted decks. Each is single-type Basics with a base-damage attack so the
// Energy Zone always feeds a cost and auto-play reliably reaches a winner.
var (
	fireDeckList = []deckEntry{
		{"B2a-126", 6}, // Entei ex   — Blazing Beatdown [RR] 60
		{"A4b-069", 7}, // Torkoal    — Flamethrower [RR] 70
		{"A4b-076", 7}, // Heatran    — Ragin' Mad Strike [RR] 40
	}
	waterDeckList = []deckEntry{
		{"A4b-101", 6}, // Articuno ex — Ice Wing [WC] 40
		{"A3-044", 7},  // Lapras      — Surf [WWC] 70
		{"A3b-021", 7}, // Alomomola   — Water Pulse [WCC] 50
	}
)

// buildDeck expands a curated list into engine cards, erroring if a card ID is
// missing from the loaded data or cannot be modeled.
func buildDeck(list []deckEntry) ([]engine.Card, error) {
	var deck []engine.Card
	for _, e := range list {
		rc, ok := rawByID[e.id]
		if !ok {
			return nil, fmt.Errorf("curated card %s not found in %s", e.id, cardsPath)
		}
		card, err := toEngineCard(rc)
		if err != nil {
			return nil, err
		}
		for i := 0; i < e.count; i++ {
			deck = append(deck, card)
		}
	}
	return deck, nil
}

// deckFromIDs expands a list of card IDs into engine cards. The input must be a
// format-legal 20 (at most 2 copies of any card by name, at least one Basic
// Pokémon) — the deckbuild UI already enforces this. Cards the engine cannot
// model yet (Trainers, whose effects are deferred per ADR-0002) are skipped, so
// the deck plays as its Pokémon subset rather than being rejected outright. Card
// access is unrestricted — only the format is checked.
func deckFromIDs(ids []string) ([]engine.Card, error) {
	if len(ids) != 20 {
		return nil, fmt.Errorf("deck must be exactly 20 cards, got %d", len(ids))
	}
	deck := make([]engine.Card, 0, 20)
	countByName := map[string]int{}
	basics := 0
	for _, id := range ids {
		rc, ok := rawByID[id]
		if !ok {
			return nil, fmt.Errorf("unknown card id %q", id)
		}
		card, err := toEngineCard(rc)
		if err != nil {
			// A card the engine can't model yet (e.g. a Trainer). Skip it so the
			// rest of the deck is still playable; its behavior lands with effects.
			continue
		}
		countByName[card.Name]++
		if countByName[card.Name] > 2 {
			return nil, fmt.Errorf("at most 2 copies of %q allowed", card.Name)
		}
		if card.Stage == engine.Basic {
			basics++
		}
		deck = append(deck, card)
	}
	if basics == 0 {
		return nil, fmt.Errorf("deck needs at least 1 Basic Pokémon")
	}
	return deck, nil
}

// curatedDecks builds the Fire and Water preset decks from loaded card data.
func curatedDecks() (fire, water []engine.Card, err error) {
	if fire, err = buildDeck(fireDeckList); err != nil {
		return nil, nil, err
	}
	if water, err = buildDeck(waterDeckList); err != nil {
		return nil, nil, err
	}
	return fire, water, nil
}
