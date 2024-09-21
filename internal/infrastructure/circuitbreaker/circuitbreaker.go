package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

type CircuitBreaker struct {
	mutex      sync.Mutex
	state      State
	failures   int
	threshold  int
	timeout    time.Duration
	lastOpened time.Time
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     StateClosed,
		threshold: threshold,
		timeout:   timeout,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case StateClosed:
		return cb.executeClosed(fn)
	case StateHalfOpen:
		return cb.executeHalfOpen(fn)
	default:
		return cb.executeOpen()
	}
}

func (cb *CircuitBreaker) executeClosed(fn func() error) error {
	err := fn()
	if err != nil {
		cb.failures++
		if cb.failures >= cb.threshold {
			cb.tripBreaker()
		}
	} else {
		cb.failures = 0
	}
	return err
}

func (cb *CircuitBreaker) executeHalfOpen(fn func() error) error {
	err := fn()
	if err != nil {
		cb.tripBreaker()
	} else {
		cb.state = StateClosed
		cb.failures = 0
	}
	return err
}

func (cb *CircuitBreaker) executeOpen() error {
	if time.Since(cb.lastOpened) > cb.timeout {
		cb.state = StateHalfOpen
		return cb.executeHalfOpen(fn)
	}
	return errors.New("circuit breaker is open")
}

func (cb *CircuitBreaker) tripBreaker() {
	cb.state = StateOpen
	cb.lastOpened = time.Now()
}