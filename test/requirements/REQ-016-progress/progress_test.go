package requirements

import (
	"sync"
	"testing"
	"time"

	"github.com/adamf123git/git-migrator/internal/progress"
)

// TestProgressReport tests basic progress reporting
func TestProgressReport(t *testing.T) {
	reporter := progress.NewReporter(100)

	// Initial state
	if reporter.Percentage() != 0 {
		t.Errorf("Expected 0%%, got %f%%", reporter.Percentage())
	}

	// Update progress
	reporter.SetCurrent(50)
	if reporter.Percentage() != 50 {
		t.Errorf("Expected 50%%, got %f%%", reporter.Percentage())
	}

	// Complete
	reporter.SetCurrent(100)
	if reporter.Percentage() != 100 {
		t.Errorf("Expected 100%%, got %f%%", reporter.Percentage())
	}
}

// TestProgressCurrentOperation tests current operation tracking
func TestProgressCurrentOperation(t *testing.T) {
	reporter := progress.NewReporter(100)

	reporter.SetOperation("Processing commit abc123")
	if reporter.Operation() != "Processing commit abc123" {
		t.Errorf("Unexpected operation: %s", reporter.Operation())
	}

	reporter.SetOperation("Writing file: main.c")
	if reporter.Operation() != "Writing file: main.c" {
		t.Errorf("Unexpected operation: %s", reporter.Operation())
	}
}

// TestProgressETA tests ETA calculation
func TestProgressETA(t *testing.T) {
	reporter := progress.NewReporter(100)

	// No ETA at start
	reporter.Start()

	// Simulate some work with enough delay to calculate rate
	time.Sleep(50 * time.Millisecond)
	reporter.SetCurrent(10)

	eta := reporter.ETA()
	if eta < 0 {
		t.Error("ETA should not be negative")
	}

	// ETA should be calculable after progress with time elapsed
	// The exact value depends on timing, so we just verify it's positive
	t.Logf("ETA after 10%%: %v", eta)
}

// TestProgressSubscriber tests progress subscription
func TestProgressSubscriber(t *testing.T) {
	reporter := progress.NewReporter(100)

	var received []progress.Status
	var mu sync.Mutex

	// Subscribe
	unsubscribe := reporter.Subscribe(func(s progress.Status) {
		mu.Lock()
		received = append(received, s)
		mu.Unlock()
	})
	defer unsubscribe()

	reporter.Start()
	reporter.SetCurrent(50)
	reporter.SetCurrent(100)

	// Wait for updates
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(received) == 0 {
		t.Error("Expected to receive progress updates")
	}
}

// TestProgressMultipleSubscribers tests multiple subscribers
func TestProgressMultipleSubscribers(t *testing.T) {
	reporter := progress.NewReporter(100)

	var count1, count2 int
	var mu sync.Mutex

	reporter.Subscribe(func(s progress.Status) {
		mu.Lock()
		count1++
		mu.Unlock()
	})

	reporter.Subscribe(func(s progress.Status) {
		mu.Lock()
		count2++
		mu.Unlock()
	})

	reporter.Start()
	reporter.SetCurrent(50)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if count1 == 0 || count2 == 0 {
		t.Error("Expected both subscribers to receive updates")
	}
}

// TestProgressThreadSafety tests concurrent updates
func TestProgressThreadSafety(t *testing.T) {
	reporter := progress.NewReporter(1000)
	reporter.Start()

	var wg sync.WaitGroup

	// Concurrent updates - fewer iterations to avoid race detector slowdown
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				reporter.Increment()
			}
		}()
	}

	wg.Wait()

	if reporter.Current() != 1000 {
		t.Errorf("Expected 1000, got %d", reporter.Current())
	}
}

// TestProgressStatus tests status structure
func TestProgressStatus(t *testing.T) {
	status := progress.Status{
		Current:   50,
		Total:     100,
		Percentage: 50.0,
		Operation: "Testing",
		ETA:       time.Minute,
	}

	if status.Percentage != 50.0 {
		t.Errorf("Expected 50%%, got %f%%", status.Percentage)
	}

	if status.Operation != "Testing" {
		t.Errorf("Expected 'Testing', got %q", status.Operation)
	}
}
