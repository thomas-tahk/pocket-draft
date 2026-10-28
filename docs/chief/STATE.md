# Chief state
Vision: docs/VISION.md
Rule: one milestone = one branch (mN-<name>) = one PR into main, in number order. No in-between milestones.
Phase: Review — M4 "Draft runs, start to finish" (branch m4-draft-runs → main, PR https://github.com/thomas-tahk/pocket-draft/pull/129)
2026-10-28 (cloud): built his review answers (docs/decisions/2026-10-28-m4-review-answers.md): blended look + light/dark toggle (#131, docs/chief/look/m4-blend/), type-lean cap 4, text-decided type support, tagged Ancient/Future, Eevee ex both ways, Muk test, shorter Energy Zone picker. Draft-depth issues #132–#136 filed in M5 with evidence (docs/chief/research/2026-10-28-draft-depth-evidence.md).
Next action: he plays the PR preview (look in light and dark) and merges PR #129; then M5 starts.
Waiting on him: playing + merging PR #129; the one open item in docs/milestones/M4.md (Koraidon ex / Miraidon ex tag)
After M4 merges: M5 "Every card works" PLAN APPROVED 2026-10-25 (docs/chief/plan-every-card.html; M5 issues incl. #122, #123, and the new #132–#136 draft-depth issues, which need his calls before building). Branch m5-every-card off main. Real opponents = M6.
Cloud notes: .claude/agents/* don't load as agent types in a cloud session started on another branch; dispatch general-purpose told to read the specialist file. Worktree agents may branch from main: tell them to reset their own branch onto the milestone branch first. Workers don't commit pd.wasm; chief rebuilds after each merge. 16 e2e (full-game ×9, phone-layout ×6, replay:74) fail in the cloud only: card-image CDN blocked; a11y-keyboard:369 and draft-tap:63 can time out under full-suite load and pass alone. Cloud Playwright pins headless_shell-1148: make /opt/pw-browsers/chromium_headless_shell-1148/chrome-linux/headless_shell a link to the 1243 chrome-headless-shell binary.
Open questions: docs/questions/m4-app-flow.md, docs/questions/draft-rework.md
Updated: 2026-10-28
