# Chief state
Vision: docs/VISION.md
Rule: one milestone = one branch (mN-<name>) = one PR into main, in number order. No in-between milestones.
Phase: Review — M4 "Draft runs, start to finish" (branch m4-draft-runs → main, PR https://github.com/thomas-tahk/pocket-draft/pull/129)
2026-11-01 (cloud): built his 2026-11-01 play-test fixes #137–#141 (Rare Candy staple, Trainers-only shop, on-evolve Ability prompt, stale card view after a pick, End Turn glow); all closed with evidence; see "Play-test, 2026-11-01" in docs/milestones/M4.md. Stale .claude/skills/chief and copied global agents deleted (pocket-draft-engine.md kept).
Next action: he plays the PR preview and replies `merge` or a list of fixes on PR #129; then M5 starts.
Waiting on him: PR #129 comment of 2026-11-01 (play the 5 fixes, reply merge / fixes); 2 guesses there (Rare Candy triggers on-evolve; "Not now" keeps the button); Koraidon ex / Miraidon ex tag line in docs/milestones/M4.md
After M4 merges: M5 "Every card works" PLAN APPROVED 2026-10-25 (docs/chief/plan-every-card.html; M5 issues incl. #122, #123, #132–#136 draft-depth (need his calls), #142 missing expansions, #143 Fossil/Stadium buttons). Branch m5-every-card off main. Real opponents = M6.
Cloud notes: dotfiles agents load as agent types; project agent pocket-draft-engine isn't a type in cloud — dispatch backend-engineer told to read .claude/agents/pocket-draft-engine.md. Worktree agents may branch from main: tell them to reset onto the milestone branch first. Workers don't commit pd.wasm; chief rebuilds after each Go merge. 17 e2e fail in the cloud only (card-image CDN blocked / full-game stuck at Ready on base too): full-game ×9, phone-layout ×6, replay:74; a11y-keyboard:369 times out under full-suite load, passes alone. Cloud Playwright pins headless_shell-1148: link /opt/pw-browsers/chromium_headless_shell-1148/chrome-linux/headless_shell to the 1243 chrome-headless-shell binary. Some printings (Eevee B4-131, Frogadier B1-072) are refused by /api/new "can't be played yet".
Open questions: docs/questions/m4-app-flow.md, docs/questions/draft-rework.md
Updated: 2026-11-01
