package circuitbreaker

import (
	"testing"
	"time"
)

func TestCircuitBreaker_FullLifecyle(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	if cb.State() != StateClosed {
		t.Fatalf("Expected initial state CLOSED, got %s", cb.state)
	}

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateClosed {
		t.Fatalf("Expected state to stay CLOSED after 2 failures, got %s", cb.State())
	}

	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("Expected state to TRIP to OPEN, got %s", cb.State())
	}

	if cb.Allow() {
		t.Fatalf("Expected Allow() to be FALSE while breaker is OPEN")
	}

	time.Sleep(120 * time.Millisecond)

	if !cb.Allow() {
		t.Fatalf("Expected Allow() to be TRUE for canary probe")
	}
	if cb.State() != StateHalfOpen {
		t.Fatalf("Expected state to be HALF-OPEN, got %s", cb.State())
	}

	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Fatalf("Expected state to reset to CLOSED, got %s", cb.State())
	}
}