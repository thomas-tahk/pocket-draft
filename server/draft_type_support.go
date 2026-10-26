package main

// Type support for draft offers: which Trainers/Items help one type's Pokémon,
// so leaning a type raises their odds too (owner, 2026-10-25, point 3).
// Mirrored by src/draft/typeSupport.ts; edit both together.
//
// Most Trainers are found by their text: it names a type ("Water Pokémon",
// "[W]", ...), see supportTypes in draft_weights.go. This file holds only what
// the text cannot say, hand-made from public/data/cards.json. It is NOT the
// niche Trainer <-> Pokémon table (#126): that one pairs a Trainer with the
// specific Pokémon it names; this one only says which type a Trainer favours.

// typeSupportTable maps a Trainer's exact name to the types it helps, for
// Trainers whose text names Pokémon (not a type). The comment lists the
// Pokémon the text names; the type is their card type in cards.json, judged
// by hand.
var typeSupportTable = map[string][]string{
	"Blaine":               {"Fire"},      // Ninetales, Rapidash, Magmar
	"Koga":                 {"Darkness"},  // Muk, Weezing
	"Marlon":               {"Water"},     // Carracosta, Jellicent
	"Lana":                 {"Water"},     // Araquanid
	"Hala":                 {"Fighting"},  // Hariyama, Crabominable
	"Nemona":               {"Lightning"}, // Pawmot
	"Clemont":              {"Lightning"}, // Magneton, Heliolisk
	"Clemont's Backpack":   {"Lightning"}, // Magneton, Heliolisk
	"Sophocles":            {"Lightning"}, // Alolan Golem, Vikavolt, Togedemaru
	"Jasmine":              {"Metal"},     // Steelix, Skarmory ex
	"Iris":                 {"Dragon"},    // Haxorus
	"Drayden":              {"Dragon"},    // Draco Meteor attack
	"Piers":                {"Darkness"},  // Galarian Obstagoon
	"Team Galactic Grunt":  {"Darkness"},  // Stunky, Croagunk (Glameow is Colorless)
	"Mallow":               {"Grass"},     // Shiinotic, Tsareena
	"Acerola":              {"Psychic"},   // Mimikyu, Palossand
	"Budding Expeditioner": {"Psychic"},   // Mew ex
}

// notTypeSupport lists Trainers whose text names a type but works against it,
// so the text rule must not count them.
var notTypeSupport = map[string]bool{
	"Squirt Bottle": true, // discards a [R] Energy from the OPPONENT's Pokémon
}
