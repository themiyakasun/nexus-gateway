package circuitbreaker

import (
	"testing"
	"time"
)

func TestRateCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(10*time.Second, 0.50, 6, 100*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateClosed {
		t.Fatalf("Should stay CLOSED because minRequests (6) not reached, got %s", cb.State())
	}

	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordSuccess()

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatalf("Expected breaker to TRIP to OPEN at 60%% failure rate, got %s", cb.State())
	}

	if cb.Allow() {
		t.Fatalf("Expected Allow() to be False while OPEN")
	}

	time.Sleep(120 * time.Millisecond)

	if !cb.Allow() {
		t.Fatalf("Expected Allow() to be TRUE for canary probe")
	}

	if cb.State() != StateHalfOpen {
		t.Fatalf("Expected HALF-OPEN, got %s", cb.State())
	}

	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Fatalf("Expected state to recover to CLOSED, got %s", cb.State())
	}
}