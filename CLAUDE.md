
<!-- chief:begin — managed by github.com/thomas-tahk/chief. Edit freely; the
     installer will not overwrite a block you have changed. -->
## Working with the chief

Issues in this repo can be labelled `build`, which starts an autonomous cloud run
that opens a PR. `DECISIONS.md` is the audit trail; `STATUS.md` is where the
project actually is. Both are updated by every chief PR.

**Every issue needs a done-gate** — one sentence describing a single
user-observable transaction that proves it works. *"I sign in with my real
account and see my real data."* Not "tests pass". A run with no done-gate stops
and asks rather than guessing.

### The decision line for this repo

**Brought to the owner** (`your-call`): anything a player sees — board layout,
wording, what a card's text does · **any rule interpretation**, because a wrong
one is a wrong game, not a bug · adding a dependency (both Go modules are stdlib
only today) · deploying anything, or anything that costs money · changing the
draft format or deck-construction rules · deleting features or scraped data.

**Decided by the run:** internal implementation · file layout · tests · refactors ·
which `EffectOp` verbs to add and how to shape them · fixing its own bugs.

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
<!-- chief:end -->
