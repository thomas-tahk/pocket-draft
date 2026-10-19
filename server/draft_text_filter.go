package main

// TEMPORARY (approved plan, "Draft pool during milestones 1–3"): until
// milestone 4 finishes card text, the draft only offers cards whose text the
// engine already plays, so every drafted deck is a legal 20. Milestone 4's last
// issue (4.15) deletes this file, the filter line in toDraftablePool,
// public/data/draft-pool.json and its reader in src/data/loader.ts.
//
// This Go code is the one source of truth. The SPA reads the same answer from
// public/data/draft-pool.json, which TestDraftPoolFileMatchesTheEngine keeps
// in step (regenerate with PD_WRITE_DRAFT_POOL=1).

// onlyWorkingTextInDraft is the switch.
const onlyWorkingTextInDraft = true

// draftPoolFile is relative to the server's working directory, like cardsPath.
const draftPoolFile = "../public/data/draft-pool.json"

// textWorks reports whether the engine can play rc as printed: it converts to
// an engine card (Trainers don't yet), it has no Ability (the engine models
// none), and every attack with text has an authored effect in effectsByID.
func textWorks(rc rawCard) bool {
	if _, err := toEngineCard(rc); err != nil {
		return false
	}
	if rc.Ability != nil {
		return false
	}
	for _, a := range rc.Attacks {
		if a.Effect != "" && effectsByID[rc.ID][a.Name] == nil {
			return false
		}
	}
	return true
}

// keepWorkingText drops the cards whose text doesn't work yet, when the switch
// is on.
func keepWorkingText(pool []draftCard) []draftCard {
	if !onlyWorkingTextInDraft {
		return pool
	}
	out := make([]draftCard, 0, len(pool))
	for _, c := range pool {
		if textWorks(c.rawCard) {
			out = append(out, c)
		}
	}
	return out
}

// draftPoolFileJSON is what the SPA reads: the switch, and the ids the draft
// may offer.
type draftPoolFileJSON struct {
	OnlyWorkingText bool     `json:"onlyWorkingText"`
	IDs             []string `json:"ids"`
}

func draftPoolFileContents() draftPoolFileJSON {
	pool := toDraftablePool()
	ids := make([]string, len(pool))
	for i, c := range pool {
		ids[i] = c.ID
	}
	return draftPoolFileJSON{OnlyWorkingText: onlyWorkingTextInDraft, IDs: ids}
}
