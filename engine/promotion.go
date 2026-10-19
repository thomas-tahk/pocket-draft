package engine

// HiddenPromotion is a promotion the other player may not see yet (spec E6):
// after a double knockout the first player's pick stays hidden until the
// second player has placed their own new Active.
type HiddenPromotion struct {
	Player     int // who promoted
	BenchIndex int // the bench slot the Pokémon left
	Action     int // index into Game.Actions of the action that carried it out
	Narration  int // index into Game.Narration of the line that names it
}

// HiddenPromotion reports the promotion to hide from the other player, if any.
// It holds only while that other player still owes their own new Active.
func (g *Game) HiddenPromotion() (HiddenPromotion, bool) {
	h, pr := g.S.hidden, g.S.Pending
	if h == nil || pr == nil || pr.Kind != PromptNewActive || pr.Player != 1-h.Player {
		return HiddenPromotion{}, false
	}
	return *h, true
}
