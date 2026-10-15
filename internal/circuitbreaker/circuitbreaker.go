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

type requestRecord struct {
	timestamp time.Time
	isFailure bool
}

type CircuitBreaker struct {
	state State
	window time.Duration
	rateThreshold float64
	minRequests int
	cooldown time.Duration
	lastStateChange time.Time
	records []requestRecord
	mux sync.Mutex
}

func NewCircuitBreaker(window time.Duration, rateThreshold float64, minRequests int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state: StateClosed,
		window: window,
		rateThreshold: rateThreshold,
		minRequests: minRequests,
		cooldown: cooldown,
		lastStateChange: time.Now(),
		records: make([]requestRecord, 0),
	}
}

func (cb *CircuitBreaker) cleanOldRecords(now time.Time) {
	cutoff := now.Add(-cb.window)
	valid := cb.records[:0]
	for _, rec := range cb.records {
		if rec.timestamp.After(cutoff) {
			valid = append(valid, rec)
		}
	}

	cb.records = valid
}

func (cb *CircuitBreaker) checkRate(now time.Time) {
	cb.cleanOldRecords(now)

	total := len(cb.records)
	if total < cb.minRequests {
		return
	}

	failures := 0
	for _, rec := range cb.records {
		if rec.isFailure {
			failures++
		}
	}

	failureRate := float64(failures) / float64(total)

	if failureRate >= cb.rateThreshold {
		cb.state = StateOpen
		cb.lastStateChange = now
		cb.records = nil
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

	now := time.Now()

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.lastStateChange = now
		cb.records = nil
		return
	}

	cb.records = append(cb.records, requestRecord{timestamp: now, isFailure: false})
	cb.cleanOldRecords(now)
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mux.Lock()
	defer cb.mux.Unlock()

	now := time.Now()

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.lastStateChange = now
		cb.records = nil
		return
	}

	cb.records = append(cb.records, requestRecord{timestamp: now, isFailure: true})

	if cb.state == StateClosed {
		cb.checkRate(now)
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mux.Lock()
	defer cb.mux.Unlock()
	return cb.state
}