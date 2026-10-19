package engine

// Ability is a Pokémon's Ability. Passive is its always-on part (rule hooks,
// passive.go).
type Ability struct {
	Name    string
	Passive []Passive
}
