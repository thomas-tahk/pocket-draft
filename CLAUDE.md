
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

**Brought to the owner** (`your-call`): anything a user sees — UX flow, visuals,
wording, feature scope · anything that costs money or adds an outside service ·
hard-to-undo data or schema changes · auth and security · deleting features or data.

**Decided by the run:** internal implementation · file layout · tests · refactors ·
choosing between equivalent libraries · fixing its own bugs.

_(Adjust these two lists for this repo. They override the chief's defaults.)_

### How to run and preview

```
<!-- install/dev/test/build commands, and where the preview shows up -->
```
<!-- chief:end -->
