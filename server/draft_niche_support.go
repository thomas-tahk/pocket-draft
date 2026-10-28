package main

import (
	"slices"
	"sort"
)

// Niche support for draft offers (#126): archetype Trainers whose text does
// not name the Pokémon they work with, paired with those Pokémon, so picking
// either raises the other's odds (owner, 2026-10-25, point 2). Mirrored by
// src/draft/nicheSupport.ts; edit both together.
//
// A Trainer that names a Pokémon in its text needs no entry here: the synergy
// rules read the text (rule 1 / rule 4 in draft_synergy.go). This file is only
// for what the text cannot say. Hand-made from public/data/cards.json; the
// owner has not reviewed the groups, so they are guesses.
//
// Names are exact. A Pokémon matches an entry equal to its name or to its name
// minus a trailing " ex" (the same stripping rule 1 applies to picked names),
// so "Aerodactyl" covers "Aerodactyl ex" but never a longer name, and an entry
// "Koraidon ex" covers only the ex card, never "Koraidon". Pokémon not in
// today's draft pool are listed too: the pool grows as card text lands, and an
// entry with no card in the pool does nothing. Eevee ex is not listed (#128).

// nicheGroup pairs Trainers with the Pokémon they support. Cross-kind only: a
// picked Trainer raises the group's Pokémon and a picked Pokémon raises the
// group's Trainers; two Pokémon, or two Trainers, of one group do not raise
// each other.
type nicheGroup struct {
	label       string
	trainers    []string // exact Trainer names
	pokemon     []string // Pokémon names, with or without " ex" (see above)
	evolvesFrom []string // any Pokémon whose card says it evolves from one of these
}

var nicheGroups = []nicheGroup{
	// Fossils: each Fossil/Old Amber supports the final evolution of its own line.
	{label: "Helix Fossil", trainers: []string{"Helix Fossil"}, pokemon: []string{"Omastar"}},
	{label: "Dome Fossil", trainers: []string{"Dome Fossil"}, pokemon: []string{"Kabutops"}},
	{label: "Old Amber", trainers: []string{"Old Amber"}, pokemon: []string{"Aerodactyl"}},
	{label: "Skull Fossil", trainers: []string{"Skull Fossil"}, pokemon: []string{"Rampardos"}},
	{label: "Armor Fossil", trainers: []string{"Armor Fossil"}, pokemon: []string{"Bastiodon"}},
	{label: "Plume Fossil", trainers: []string{"Plume Fossil"}, pokemon: []string{"Archeops"}},
	{label: "Cover Fossil", trainers: []string{"Cover Fossil"}, pokemon: []string{"Carracosta"}},
	{label: "Root Fossil", trainers: []string{"Root Fossil"}, pokemon: []string{"Cradily"}},
	{label: "Claw Fossil", trainers: []string{"Claw Fossil"}, pokemon: []string{"Armaldo"}},
	{label: "Jaw Fossil", trainers: []string{"Jaw Fossil"}, pokemon: []string{"Tyrantrum"}},
	{label: "Sail Fossil", trainers: []string{"Sail Fossil"}, pokemon: []string{"Aurorus"}},
	// Eevee Bag's text says "Eevee"; the Pokémon it helps are the ones that
	// evolve from it (Eevee ex itself is #128's).
	{label: "Eeveelutions", trainers: []string{"Eevee Bag"}, evolvesFrom: []string{"Eevee"}},
	// Ancient / Future: a card counts only if the card itself says so (owner
	// 2026-10-28, decision 3). cards.json has no such tag, so the members are
	// the names of the engine's tagged printings (ancientIDs / futureIDs in
	// texts_3_11.go), exact: "Koraidon ex" (B3a) is Ancient, the B2a "Koraidon"
	// is not. TestNicheAncientFutureAreTheEnginesPrintings keeps them equal.
	{
		label:    "Ancient",
		trainers: []string{"Ancient Booster Energy Capsule", "Professor Sada"},
		pokemon: []string{"Brute Bonnet", "Flutter Mane", "Flutter Mane ex", "Gouging Fire", "Great Tusk", "Koraidon ex",
			"Raging Bolt", "Roaring Moon", "Sandy Shocks", "Scream Tail", "Slither Wing", "Walking Wake"},
	},
	{
		label:    "Future",
		trainers: []string{"Future Booster Energy Capsule", "Professor Turo"},
		pokemon: []string{"Iron Boulder", "Iron Bundle", "Iron Bundle ex", "Iron Crown", "Iron Hands", "Iron Jugulis",
			"Iron Leaves", "Iron Moth", "Iron Thorns", "Iron Treads", "Iron Valiant", "Miraidon ex"},
	},
	// Ultra Beast: the engine's own list (texts_3_12.go), not a second copy.
	{
		label:    "Ultra Beast",
		trainers: []string{"Lusamine", "Beast Wall", "Beastite"},
		pokemon:  ultraBeastNames(),
	},
}

// ultraBeastNames is the engine's Ultra Beast list (texts_3_12.go), sorted.
func ultraBeastNames() []string {
	out := make([]string, 0, len(ultraBeasts))
	for n := range ultraBeasts {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Lookup tables over nicheGroups: name -> indexes of the groups it is in.
var (
	nicheByTrainer     = map[string][]int{}
	nicheByPokemon     = map[string][]int{}
	nicheByEvolvesFrom = map[string][]int{}
)

func init() {
	for i, g := range nicheGroups {
		for _, n := range g.trainers {
			nicheByTrainer[n] = append(nicheByTrainer[n], i)
		}
		for _, n := range g.pokemon {
			nicheByPokemon[n] = append(nicheByPokemon[n], i)
		}
		for _, n := range g.evolvesFrom {
			nicheByEvolvesFrom[n] = append(nicheByEvolvesFrom[n], i)
		}
	}
}

// nicheGroupsOf is the niche groups c belongs to: a Trainer by exact name, a
// Pokémon by its name, its name minus " ex", or what it evolves from. Each
// group at most once.
func nicheGroupsOf(c draftCard) []int {
	if c.CardType == "Trainer" {
		return nicheByTrainer[c.Name]
	}
	var out []int
	add := func(groups []int) {
		for _, g := range groups {
			if !slices.Contains(out, g) {
				out = append(out, g)
			}
		}
	}
	add(nicheByPokemon[c.Name])
	add(nicheByPokemon[pokemonBaseName(c.Name)])
	if c.EvolvesFrom != "" {
		add(nicheByEvolvesFrom[c.EvolvesFrom])
	}
	return out
}
