# Draft rework — interpretations

- #106 · run end 3L/7W is not enforced anywhere in code (W/L are manual History entries) · moved to app-flow (M4) run screen
- #108 · type weight coefficient = 1.0 (pre-May value; 8 Fire picks ≈ 6× baseline Fire share) · strong; tune after play-test
- #108 · pack-1 ex = forced last slot, mega-ex counts · matches the pre-removal anchor
- noticed · hovering a draft card opens a full zoom overlay over the offers (his card-zoom complaint, 10-15 answers) · not in any issue yet
- #107 · Poké Ball + Professor's Research no longer excluded from the draft pool — draftable/buyable like any Trainer · otherwise unobtainable · DECISION for owner (revert = restore isDraftCandidate exclusion)
- #107 · bot-play test relaxed (no "held a Trainer, played none" check; benched ≥90% of seats) · bot now holds conditional Trainers correctly · #110 must restore strict checks
- #109 · evolution-line synergy dropped: draft pool is final evolutions only, so it can never fire · replaced by type-reference synergy (card text names a type you picked, e.g. Misty→Water, or a pick's text names the candidate's type) + Trainers naming a picked Pokémon; ×3, not stacked · DECISION for owner on play-test
