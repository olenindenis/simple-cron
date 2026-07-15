package modules_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

// Reproduces, in miniature, the goroutine pile-up seen in production: a job
// wired the same way as internal/modules/cron.go (cron.New with no options)
// overlaps with itself when a run takes longer than the tick interval.
func TestCronWiring_OverlappingJobsAccumulate(t *testing.T) {
	scheduler := cron.New()

	var inFlight int32
	var peak int32

	// robfig/cron rounds any @every interval below 1s up to 1s
	// (vendor/github.com/robfig/cron/v3/constantdelay.go:14-17), so the
	// shortest usable tick here is 1s; the job sleeps longer than that to
	// force the next tick to overlap it.
	if _, err := scheduler.AddFunc("@every 1s", func() {
		n := atomic.AddInt32(&inFlight, 1)
		for {
			p := atomic.LoadInt32(&peak)
			if n <= p || atomic.CompareAndSwapInt32(&peak, p, n) {
				break
			}
		}
		time.Sleep(1600 * time.Millisecond)
		atomic.AddInt32(&inFlight, -1)
	}); err != nil {
		t.Fatalf("AddFunc: %v", err)
	}

	scheduler.Start()
	time.Sleep(2200 * time.Millisecond)
	stopCtx := scheduler.Stop()
	<-stopCtx.Done()

	if got := atomic.LoadInt32(&peak); got <= 1 {
		t.Fatalf("expected overlapping ticks to stack up concurrent job executions (peak > 1), got peak=%d", got)
	}
}

// Confirms that adding cron.SkipIfStillRunning to the chain (the fix)
// prevents the overlap demonstrated above. Kept as a regression test.
func TestCronWiring_SkipIfStillRunning_PreventsOverlap(t *testing.T) {
	scheduler := cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))

	var inFlight int32
	var peak int32

	if _, err := scheduler.AddFunc("@every 1s", func() {
		n := atomic.AddInt32(&inFlight, 1)
		for {
			p := atomic.LoadInt32(&peak)
			if n <= p || atomic.CompareAndSwapInt32(&peak, p, n) {
				break
			}
		}
		time.Sleep(1600 * time.Millisecond)
		atomic.AddInt32(&inFlight, -1)
	}); err != nil {
		t.Fatalf("AddFunc: %v", err)
	}

	scheduler.Start()
	time.Sleep(2200 * time.Millisecond)
	stopCtx := scheduler.Stop()
	<-stopCtx.Done()

	if got := atomic.LoadInt32(&peak); got != 1 {
		t.Fatalf("expected SkipIfStillRunning to cap concurrent job executions at 1, got peak=%d", got)
	}
}
