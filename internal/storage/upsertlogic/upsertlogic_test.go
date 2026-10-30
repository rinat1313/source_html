package upsertlogic

import (
	"testing"
	"time"
)

func TestDecideAction(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	older := base.Add(-24 * time.Hour)
	newer := base.Add(24 * time.Hour)

	if got := DecideAction(nil, base); got != ActionCreated {
		t.Fatalf("nil existing: got %q", got)
	}
	if got := DecideAction(&base, newer); got != ActionUpdated {
		t.Fatalf("newer: got %q", got)
	}
	if got := DecideAction(&base, base); got != ActionSkipped {
		t.Fatalf("equal: got %q", got)
	}
	if got := DecideAction(&base, older); got != ActionSkipped {
		t.Fatalf("older: got %q", got)
	}
}

func TestCountsTowardLimit(t *testing.T) {
	if CountsTowardLimit(ActionSkipped, false) {
		t.Fatal("skip should not count")
	}
	if !CountsTowardLimit(ActionCreated, false) {
		t.Fatal("created should count")
	}
	if !CountsTowardLimit(ActionUpdated, false) {
		t.Fatal("updated should count")
	}
	if CountsTowardLimit(ActionCreated, true) {
		t.Fatal("already attached should not count")
	}
}
