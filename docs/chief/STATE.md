# Chief state
Vision: docs/VISION.md
Rule: one milestone = one branch (mN-<name>) = one PR into main, in number order. No in-between milestones.
Phase: Review — M4 "Draft runs, start to finish" (branch m4-draft-runs → main, PR https://github.com/thomas-tahk/pocket-draft/pull/129)
Done in M4: #106–#110 draft rework · #113–#119 app flow · #124–#128 draft play-test fixes · #120 Pocket look · #121 accessibility · milestone QA + whole-diff review fixes (resume race 4fe3962, double-tap pick 761f55a)
Next action: wait for his play-test and answers to the 9 decisions in docs/milestones/M4.md; then write each answer where it belongs, fill escaped bugs / guesses kept in docs/chief/scorecard.md, route fixes to issues. Watch PR #129 CI (first run of the 16 image-dependent e2e with real images).
Waiting on him: M4 play-test + the 9 decisions in docs/milestones/M4.md + merging PR #129
After M4 merges: M5 "Every card works" PLAN APPROVED 2026-10-25 (docs/chief/plan-every-card.html; M5 issues incl. #122, #123). Branch m5-every-card off main. Real opponents = M6.
Cloud notes: .claude/agents/* don't load as agent types in a cloud session started on another branch; dispatch general-purpose told to read the specialist file. Workers don't commit pd.wasm; chief rebuilds after each merge. 16 e2e (full-game ×9, phone-layout ×6, replay:74) fail in the cloud only: card-image CDN blocked.
Open questions: docs/questions/m4-app-flow.md, docs/questions/draft-rework.md
Updated: 2026-10-26
