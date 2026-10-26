# Chief state
Vision: docs/VISION.md
Rule: one milestone = one branch (mN-<name>) = one PR into main, in number order. No in-between milestones.
Phase: Build — M4 "Draft runs, start to finish" (branch m4-draft-runs → main). Includes the draft rework (old PR #112, closed and folded in on 2026-10-25).
Done in M4: draft rework #106–#110 · phone layout #113 · run record #114 · opening #115 · run screen #116 · leaving/quitting #117 · end of run #118 · replay #119
Next action (in order; every worker branches off m4-draft-runs and merges back into it):
1. Draft fixes from his 2026-10-25 play-test (docs/decisions/2026-10-25-draft-playtest.md): MERGED #124 no repeat offers · #125 type lean · #128 Eevee ex · rules.md shop section. IN FIX ROUND after review: #126 niche odds (branch w126-niche-odds) · #127 Energy Zone picker (branch w127-energy-picker; default must mirror engine DefaultEnergyTypes).
   Cloud note: .claude/agents/* didn't load as agent types here; dispatch general-purpose told to read the specialist file. Workers don't commit pd.wasm; chief rebuilds after each merge.
2. #120 Pocket look (frontend-engineer), then #121 accessibility (frontend-engineer).
3. Milestone end: full suite → fresh qa-engineer exercises the M4 done-gate (docs/chief/plan-app-flow.html, phone + desktop) → code-reviewer over the whole m4-draft-runs diff vs main → docs/milestones/M4.md (fold in docs/milestones/draft-rework.md; incl. "Learn from it") + row in docs/chief/scorecard.md → one PR m4-draft-runs → main
Reviewers: code-reviewer. Any engine/server work: pocket-draft-engine.
Known gaps from #117: "Next battle" hidden while a saved battle settles — untested; closed-past-90 s test ages the stamp instead of the clock
Waiting on him: M4 PR play-test + merge (at milestone end)
M5 "Every card works" PLAN APPROVED 2026-10-25 (docs/chief/plan-every-card.html; M5 issues incl. #122, #123). Build after M4 merges: branch m5-every-card off main. Real opponents = M6.
Open questions: docs/questions/draft-rework.md, docs/questions/m4-app-flow.md
Updated: 2026-10-26
