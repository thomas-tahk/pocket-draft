module github.com/thomas-tahk/pocket-draft/server

go 1.25

require github.com/thomas-tahk/pocket-draft/engine v0.0.0

// The engine lives in a sibling module in this same repo. The replace directive
// points the import at the local source instead of trying to download it.
replace github.com/thomas-tahk/pocket-draft/engine => ../engine
