package runner

import (
	"context"
	"testing"
	"time"
)

// Verifies that a context deadline kills the job's process group via
// cmd.Cancel (SIGTERM) and Exec returns promptly with an error.
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
	if elapsed > 2*time.Second {
		t.Fatalf("Exec took %v to return after a %v timeout; SIGTERM/cancel does not appear to be taking effect promptly", elapsed, timeout)
	}
}
