package debounce

import (
	"testing"
	"time"
)

func TestAllowSuppressesWithinWindow(t *testing.T) {
	c := New(100 * time.Millisecond)
	base := time.Unix(0, 0)

	if !c.Allow("a", base) {
		t.Fatal("first event for a key should always be allowed")
	}
	if c.Allow("a", base.Add(50*time.Millisecond)) {
		t.Fatal("second event within the window should be suppressed")
	}
	if c.Allow("a", base.Add(99*time.Millisecond)) {
		t.Fatal("event just under the window boundary should be suppressed")
	}
}

func TestAllowAfterWindowElapses(t *testing.T) {
	c := New(100 * time.Millisecond)
	base := time.Unix(0, 0)

	if !c.Allow("a", base) {
		t.Fatal("first event should be allowed")
	}
	if !c.Allow("a", base.Add(150*time.Millisecond)) {
		t.Fatal("event after the window elapses should be allowed")
	}
}

func TestAllowIsPerKey(t *testing.T) {
	c := New(100 * time.Millisecond)
	base := time.Unix(0, 0)

	if !c.Allow("a", base) {
		t.Fatal("first event for a should be allowed")
	}
	if !c.Allow("b", base.Add(1*time.Millisecond)) {
		t.Fatal("first event for a different key b should be allowed independently")
	}
}

func TestAllowBurstCoalescesToOne(t *testing.T) {
	c := New(50 * time.Millisecond)
	base := time.Unix(0, 0)

	allowedCount := 0
	for i := 0; i < 10; i++ {
		if c.Allow("a", base.Add(time.Duration(i)*5*time.Millisecond)) {
			allowedCount++
		}
	}
	if allowedCount != 1 {
		t.Fatalf("expected a dense burst within the window to coalesce to 1 allowed event, got %d", allowedCount)
	}
}

func TestZeroWindowDisablesDebouncing(t *testing.T) {
	c := New(0)
	base := time.Unix(0, 0)

	for i := 0; i < 5; i++ {
		if !c.Allow("a", base) {
			t.Fatal("zero window should allow every event")
		}
	}
}
