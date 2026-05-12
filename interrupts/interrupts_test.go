package interrupts

import (
	"testing"
)

func TestSignalsDefault(t *testing.T) {
	// signals is a package var, verify it returns a channel
	if signals == nil {
		t.Error("signals should not be nil")
	}
}

func TestManager(t *testing.T) {
	if single == nil {
		t.Error("single manager should be initialized by init()")
	}
}

// Test the wait function with a direct cancel (seenSignal already true)
func TestWaitWithSeenSignal(t *testing.T) {
	cancelCalled := false
	cancel := func() {
		cancelCalled = true
	}

	// Simulate already seen signal
	single.c.L.Lock()
	seenBefore := single.seenSignal
	single.seenSignal = true
	single.c.L.Unlock()

	wait(cancel)

	if !cancelCalled {
		t.Error("expected cancel to be called")
	}

	// Restore
	single.c.L.Lock()
	single.seenSignal = seenBefore
	single.c.L.Unlock()
}
