## Decisions that belong to the owner

Bring these to the owner instead of deciding: anything a player sees (board layout,
wording, what a card's text does) · **any rule interpretation**, because a wrong one
is a wrong game, not a bug · adding a dependency (both Go modules are stdlib only
today) · deploying anything, or anything that costs money · changing the draft
format or deck-construction rules · deleting features or scraped data.

The owner's answers on the real Pocket rules live in
`docs/decisions/2026-10-15-rework-answers.md`. Check there before asking again.

### How to run and preview

Two Go modules and a Vite SPA. No external dependencies in either Go module, so
there is no `go.sum` — do not add one.

```sh
# the playable board — this is what a done-gate is exercised against
cd server && go build . && ./server        # http://localhost:8080

# tests
(cd engine && go test ./...)
(cd server && go test ./...)

# the draft SPA
npm ci && npm run lint && npm run build
```

**Start the server in its own shell call.** A process backgrounded in the same
call as the command that probes it dies when that call returns. `go build .`,
then start the binary, then curl — three calls.

Go here is 1.24.7 while both `go.mod` ask for 1.25. That is fine: the toolchain
downloads itself on first build. Do not "fix" the `go.mod`.

The live site runs the Go engine as WebAssembly (`public/pd.wasm`). Any change to
Go sources under `engine/` or `server/` must be followed by `npm run build:wasm`
and the rebuilt binary committed — CI fails the PR otherwise.
