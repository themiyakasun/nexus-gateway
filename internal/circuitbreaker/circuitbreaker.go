package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

var ErrCircuitOpen = errors.New("circuit breaker is OPEN: service unavailable")

type CircuitBreaker struct {
	state State
	failureThreshold int
	consecutiveFailures int
	cooldown time.Duration
	lastStateChange time.Time
	mux sync.Mutex
}

func NewCircuitBreaker(failureThreshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state: StateClosed,
		failureThreshold: failureThreshold,
		cooldown: cooldown,
		lastStateChange: time.Now(),
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mux.Lock()
	defer cb.mux.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if now.Sub(cb.lastStateChange) > cb.cooldown {
			cb.state = StateHalfOpen
			cb.lastStateChange = now
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mux.Lock()
	defer cb.mux.Unlock()

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.consecutiveFailures = 0
		cb.lastStateChange = time.Now()
		return
	}

	cb.consecutiveFailures = 0
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mux.Lock()
	defer cb.mux.Unlock()

	cb.consecutiveFailures++

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
		return
	}

	if cb.state == StateClosed && cb.consecutiveFailures >= cb.failureThreshold {
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mux.Lock()
	defer cb.mux.Unlock()
	return cb.state
}