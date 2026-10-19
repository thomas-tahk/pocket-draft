# How agents work the battle rebuild

The owner reviews finished milestones, not steps. Spec: `docs/battle-spec.html` (rows S1…X1). Plan: `docs/battle-plan.html`. Issues are titled `[M.N] …`; their bodies quote the spec rows they implement.

## Taking an issue
1. Pick an open issue in the current milestone whose "After" issues are all closed. Comment `taking` on it.
2. Work in your own git worktree, branched from the milestone branch (e.g. `m1-plain-battle`).
3. **Test first.** Write the Go test named in "Done when". Run it and see it fail. Then make it pass.
   Screen issues: the Go part (what the engine says is lit, prompted, shown) gets a Go test; drag, drop and layout get a Playwright test in `e2e/` at 390×844.
4. All of these must pass before merging:
   `(cd engine && go test ./...)`, `(cd server && go test ./...)`, `npm run lint`, `npm run build`, and `npx playwright test` once `e2e/` exists.
5. Merge yourself: `git fetch && git rebase origin/<milestone-branch>`, rerun the tests, then `git push origin HEAD:<milestone-branch>`. If the push is rejected, rebase and retry. Close the issue with a comment naming the commit.

## Rules
- Go modules stay stdlib only. No `go.sum`. Don't touch `go.mod` versions.
- **Don't commit `public/pd.wasm`.** It is rebuilt once at milestone end.
- One behaviour per issue. Don't fix neighbouring things; note them in the questions file instead.
- The spec is the authority. CLAUDE.md lists what belongs to the owner; anything you see players see beyond the spec is the owner's.
- **Real ambiguity** (the spec doesn't decide it and a wrong guess would make a wrong game): append it to `docs/questions/M<n>.md` as `- [M.N] question — what you did meanwhile`, push that, and move to another unblocked issue. Don't block on it.
- Card-text issues: every text in the batch maps to an effect, and each distinct wording gets a test on a real card from `public/data/cards.json`.
- Motion on screen is under 150 ms (spec F16).
- Commit messages: plain imperative, no Co-Authored-By trailer.

## Milestone end
When every issue in the milestone is closed: rebuild wasm, open a PR from the milestone branch to `main`, and write `docs/milestones/M<n>.md` with what was built, the batched questions, the play-test card list, and exact play steps. Then stop. Merging and deploying are the owner's.
