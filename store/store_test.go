package store

import (
	"testing"
	"fmt"
)

// - foo
// - bar
// -   baz

func TestCreate(t *testing.T) {
	var err error
	h := NewHeader()
	s := NewStore(h)

	err = s.Insert("foo", new(Entry))
	if err != nil {
		t.Errorf("Create error: %s\n", err)
	}

	err = s.Insert("bar/baz", new(Entry))
	if err != nil {
		t.Errorf("Create error: %s\n", err)
	}

	bar, ok := s.entries["bar"]
	if !ok {
		t.Errorf("Can't find created entry")
	}

	baz, ok := bar.entries["baz"]
	if !ok {
		t.Errorf("Can't find nested entry")
	}

	if s.Find("bar/baz") != baz {
		t.Errorf("Find() returns incorrent entry pointer")
	}

	err = s.Insert("", new(Entry))
	if err == nil {
		t.Errorf("Insert() with empty id doesn't return error")
	}


	// Check entry type
	if !bar.IsContainer() {
		t.Errorf("IsContainer() is false for container entry")
	}
	if baz.IsContainer() {
		t.Errorf("IsContainer() is true for leaf entry")
	}


	fmt.Println(s.entries)
}
