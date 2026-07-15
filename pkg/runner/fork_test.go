package runner

import (
	"context"
	"testing"
	"time"
)

// Mirrors the production wiring in internal/modules/cron.go, where the
// context passed into Runner.Exec is a single context.Background() created
// once for the lifetime of the app and never scoped per job. It proves Fork
// has no built-in timeout of its own: a hanging child blocks Exec for its
// full duration regardless of how long that is.
func TestFork_Exec_NoTimeoutContext_BlocksOnHangingChild(t *testing.T) {
	fork := NewFork()

	const sleep = 700 * time.Millisecond
	start := time.Now()

	if err := fork.Exec(context.Background(), "sleep 0.7"); err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < sleep {
		t.Fatalf("Exec returned after %v, before the child's own sleep of %v elapsed; expected no independent timeout to cut it short", elapsed, sleep)
	}
}

// Proves the SIGTERM/cancel mechanism in Fork.Exec (fork.go's cmd.Cancel)
// works correctly when the caller does supply a deadline. This isolates the
// defect to "the caller never provides one" rather than "the kill mechanism
// is broken".
func TestFork_Exec_RespectsContextTimeout(t *testing.T) {
	fork := NewFork()

	const timeout = 200 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	err := fork.Exec(ctx, "sleep 5")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error from a command killed by context timeout, got nil")
	}
	if elapsed >= 5*time.Second {
		t.Fatalf("Exec took %v; expected it to be cut short by the %v context timeout", elapsed, timeout)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("Exec took %v to return after a %v timeout; SIGTERM/cancel does not appear to be taking effect promptly", elapsed, timeout)
	}
}
