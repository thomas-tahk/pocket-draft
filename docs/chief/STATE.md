# Chief state
Vision: docs/VISION.md
Rule: one milestone = one branch (mN-<name>) = one PR into main, in number order. No in-between milestones.
Phase: Review — M4 "Draft runs, start to finish" (branch m4-draft-runs → main, PR https://github.com/thomas-tahk/pocket-draft/pull/129)
Play-test 2026-10-27: Stadium zoom fixed (#130, e2e/stadium-zoom.spec.ts). Bold look: 3 directions in docs/chief/look/m4-bold/ (#131), not applied.
Next action: when he picks A/B/C (or a mix) for #131, restyle every non-battle screen through src/tokens.css + shared components on m4-draft-runs, with fresh 390×844 screenshots and contrast checks; then answers to the remaining decisions in docs/milestones/M4.md.
Waiting on him: the #131 look pick + the decisions in docs/milestones/M4.md + merging PR #129
After M4 merges: M5 "Every card works" PLAN APPROVED 2026-10-25 (docs/chief/plan-every-card.html; M5 issues incl. #122, #123). Branch m5-every-card off main. Real opponents = M6.
Cloud notes: .claude/agents/* don't load as agent types in a cloud session started on another branch; dispatch general-purpose told to read the specialist file. Workers don't commit pd.wasm; chief rebuilds after each merge. 16 e2e (full-game ×9, phone-layout ×6, replay:74) fail in the cloud only: card-image CDN blocked. Cloud Playwright pins headless_shell-1148; symlink it to /opt/pw-browsers/chromium_headless_shell-1243.
Open questions: docs/questions/m4-app-flow.md, docs/questions/draft-rework.md
Updated: 2026-10-27
