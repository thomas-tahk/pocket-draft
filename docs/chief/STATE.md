# Chief state
Vision: docs/VISION.md
Phase: Build — draft rework fixes (branch draft-rework, PR #112), then M4 app flow (branch m4-app-flow, off draft-rework)
Done in M4: #113 phone layout, #114 run record, #115 opening, #116 run screen, #118 end of run, #119 replay, #117 leaving/quitting (merged into m4-app-flow, closed)
Next action (in order):
1. Draft rework fixes from his 2026-10-25 play-test (docs/decisions/2026-10-25-draft-playtest.md), milestone "Draft rework (before M4)". Worker branches off draft-rework, merge into draft-rework:
   #124 picked card never re-offered (pocket-draft-engine) · #125 type lean from 2–3 picks incl. type-support Trainers (pocket-draft-engine) · #126 niche Trainer ↔ Pokémon mutual odds (pocket-draft-engine, after #125: same weights file) · #128 Eevee ex draftable (pocket-draft-engine) · #127 Energy Zone picker at deckbuild (frontend-engineer)
   Also: update docs/rules.md shop section to match src/shop/inventory.ts (staples + same-type ex alts).
   Then merge draft-rework into m4-app-flow, refresh PR #112's report (docs/milestones/draft-rework.md): decisions 1–4 answered, merge still his.
2. M4: #120 Pocket look (frontend-engineer), then #121 accessibility (frontend-engineer).
3. M4 milestone end: full suite → fresh qa-engineer exercises the M4 done-gate (docs/chief/plan-app-flow.html, phone + desktop) → code-reviewer over whole diff → docs/milestones/M4.md (incl. "Learn from it") + row in docs/chief/scorecard.md → PR m4-app-flow → draft-rework
Reviewers: code-reviewer. Any engine/server work: pocket-draft-engine.
Known gaps from #117: "Next battle" hidden while a saved battle settles — untested; closed-past-90 s test ages the stamp instead of the clock
Waiting on him: play-test PR #112 again after the fixes, then merge it. M5 planning (at his Mac): #122, #123 + the 16 effect issues
Open questions: docs/questions/draft-rework.md, docs/questions/m4-app-flow.md
Updated: 2026-10-25
