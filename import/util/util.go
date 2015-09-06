package util

import (
	"fmt"
	"github.com/tvdburgt/passman/store"
	"strings"
	"time"
)

const DefaultId = "unnamed"

type ImportSettings struct {
	// NameGroups determines whether entry groups (if applicable) are
	// incorporated into the naming of entries. E.g., if an entry named
	// 'bar' resides in group 'foo', its identifier would become 'foo/bar'.
	NameGroups bool

	// NormalizeEntries determines whether entry values (identifier,
	// metadata keys, etc.) are 'normalized'.  A normalized value does not
	// contain whitespace and uppercase characters.
	NormalizeEntries bool
}

// Resolves id collisions by appending a unique number
func ResolveIdCollisions(s *store.Store, id string) string {
	if s.Entries[id] == nil {
		return id
	}
	for i := 2; ; i++ {
		test := fmt.Sprintf("%s-%d", id, i)
		if s.Entries[test] == nil {
			return test
		}
	}
}

// Normalize returns value as lower cased string with spaces stripped.
func Normalize(value string) string {
	return strings.Replace(strings.ToLower(value), " ", "-", -1)
}

// ParseTime parses time value with current time as fallback.
func ParseTime(value, layout string) time.Time {
	if t, err := time.Parse(layout, value); err == nil {
		return t
	} else {
		return time.Now()
	}
}
