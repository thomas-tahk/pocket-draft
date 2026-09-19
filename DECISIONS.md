# Decisions

Newest first. **One line per entry, hard rule.**

`date · what was chosen · why in one clause · issue · who`

2026-09-19 · built #5 as a new `EffectOp` verb, not the per-card `Effect` struct its body asked for · an owner comment on the issue declined that shape and locked Approach A, which slice 1 already shipped · #5 · run
2026-09-19 · added `BonusIfInDiscard` as a flat verb rather than opening a condition DSL · Ire-Fly needs one discard-pile read, and ADR-0002 grows coverage a building block at a time · #5 · run
2026-09-19 · the discard condition matches on card Name, not card ID · the printed text names the Pokémon and Volbeat has two printings in the scrape · #5 · run
2026-09-19 · the +60 applies once however many Volbeat sit in the discard · the text is a condition, not a per-copy multiplier · #5 · run
2026-09-19 · exercised the done-gate with a format-legal deck posted to `/api/new`, not by editing a preset deck · preset decks are player-visible and the API already takes a deck · #5 · run
2026-09-18 · onboarded this repo to the chief · a labelled issue should start a cloud build without a laptop · — · owner
2026-09-18 · CI runs go test, go vet and the web build on every PR · a PR carried no signal before, and auto-merge only arms behind a required check · #14 · owner

<!-- entries go above this line -->
