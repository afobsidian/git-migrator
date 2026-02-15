// Package progress provides progress reporting for migrations.
package progress

import (
	"sync"
	"time"
)

// Status represents the current migration status
type Status struct {
	Current    int
	Total      int
	Percentage float64
	Operation  string
	ETA        time.Duration
	StartTime  time.Time
}

// Subscriber is a callback for progress updates
type Subscriber func(Status)

// Reporter provides progress reporting with subscriber support
type Reporter struct {
	mu          sync.RWMutex
	current     int
	total       int
	operation   string
	startTime   time.Time
	subscribers []Subscriber
	lastUpdate  time.Time
}

// NewReporter creates a new progress reporter
func NewReporter(total int) *Reporter {
	return &Reporter{
		total: total,
	}
}

// Start begins the progress tracking
func (r *Reporter) Start() {
	r.mu.Lock()
	r.startTime = time.Now()
	r.lastUpdate = time.Now()
	r.mu.Unlock()
	r.notify()
}

// SetCurrent sets the current progress
func (r *Reporter) SetCurrent(current int) {
	r.mu.Lock()
	r.current = current
	r.lastUpdate = time.Now()
	r.mu.Unlock()
	r.notify()
}

// Increment increments the current progress by 1
func (r *Reporter) Increment() {
	r.mu.Lock()
	r.current++
	r.lastUpdate = time.Now()
	r.mu.Unlock()
	r.notify()
}

// SetOperation sets the current operation description
func (r *Reporter) SetOperation(op string) {
	r.mu.Lock()
	r.operation = op
	r.mu.Unlock()
	r.notify()
}

// Current returns the current progress
func (r *Reporter) Current() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.current
}

// Percentage returns the progress percentage
func (r *Reporter) Percentage() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.total == 0 {
		return 0
	}
	return float64(r.current) / float64(r.total) * 100
}

// Operation returns the current operation
func (r *Reporter) Operation() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.operation
}

// ETA estimates time remaining
func (r *Reporter) ETA() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.current == 0 || r.startTime.IsZero() {
		return 0
	}

	elapsed := time.Since(r.startTime)
	if elapsed == 0 {
		return 0
	}

	rate := float64(r.current) / elapsed.Seconds()
	if rate == 0 {
		return 0
	}

	remaining := float64(r.total-r.current) / rate
	return time.Duration(remaining) * time.Second
}

// Subscribe adds a progress subscriber
func (r *Reporter) Subscribe(fn Subscriber) func() {
	r.mu.Lock()
	r.subscribers = append(r.subscribers, fn)
	idx := len(r.subscribers) - 1
	r.mu.Unlock()

	return func() {
		r.mu.Lock()
		r.subscribers[idx] = nil
		r.mu.Unlock()
	}
}

// notify notifies all subscribers
func (r *Reporter) notify() {
	r.mu.RLock()
	current := r.current
	total := r.total
	operation := r.operation
	startTime := r.startTime
	subscribers := make([]Subscriber, len(r.subscribers))
	copy(subscribers, r.subscribers)
	r.mu.RUnlock()

	// Calculate status outside the lock
	percentage := float64(0)
	if total > 0 {
		percentage = float64(current) / float64(total) * 100
	}

	var eta time.Duration
	if current > 0 && !startTime.IsZero() {
		elapsed := time.Since(startTime)
		if elapsed > 0 {
			rate := float64(current) / elapsed.Seconds()
			if rate > 0 {
				remaining := float64(total-current) / rate
				eta = time.Duration(remaining) * time.Second
			}
		}
	}

	status := Status{
		Current:    current,
		Total:      total,
		Percentage: percentage,
		Operation:  operation,
		ETA:        eta,
		StartTime:  startTime,
	}

	for _, fn := range subscribers {
		if fn != nil {
			fn(status)
		}
	}
}
