package main

import "testing"

func TestSeedForUsesRequestedSeed(t *testing.T) {
	want := uint64(42)
	if got := seedFor(newReq{Seed: &want}); got != want {
		t.Fatalf("seedFor = %d, want %d", got, want)
	}
}

func TestSeedForVariesWithoutRequestedSeed(t *testing.T) {
	a := seedFor(newReq{})
	b := seedFor(newReq{})
	if a == b {
		t.Fatalf("two unseeded games got the same seed %d", a)
	}
}
