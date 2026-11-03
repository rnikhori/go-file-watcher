package summary

import (
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestAddAndTotal(t *testing.T) {
	c := NewCounter()
	c.Add(fsnotify.Create)
	c.Add(fsnotify.Create)
	c.Add(fsnotify.Write)

	if c.Total() != 3 {
		t.Errorf("Total() = %d, want 3", c.Total())
	}

	counts := c.Counts()
	if counts["CREATE"] != 2 {
		t.Errorf("Counts()[CREATE] = %d, want 2", counts["CREATE"])
	}
	if counts["WRITE"] != 1 {
		t.Errorf("Counts()[WRITE] = %d, want 1", counts["WRITE"])
	}
}

func TestStringFormat(t *testing.T) {
	c := NewCounter()
	c.Add(fsnotify.Write)
	c.Add(fsnotify.Create)
	c.Add(fsnotify.Write)

	got := c.String()
	want := "total=3 CREATE=1 WRITE=2"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestEmptyCounter(t *testing.T) {
	c := NewCounter()
	if c.Total() != 0 {
		t.Errorf("Total() = %d, want 0", c.Total())
	}
	if got, want := c.String(), "total=0"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestCountsIsACopy(t *testing.T) {
	c := NewCounter()
	c.Add(fsnotify.Write)

	counts := c.Counts()
	counts["WRITE"] = 100

	if got := c.Counts()["WRITE"]; got != 1 {
		t.Errorf("mutating the returned map affected the Counter; Counts()[WRITE] = %d, want 1", got)
	}
}
