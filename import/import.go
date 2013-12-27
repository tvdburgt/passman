package imprt

import (
	"fmt"
	"github.com/tvdburgt/passman/store"
	"strings"
)

var (
	ImportGroups     = false
	NormalizeEntries = false
)

// Resolves id collisions by appending a unique number
func ResolveIdCollisions(s *store.Store, id string) string {
	if _, ok := s.Entries[id]; !ok {
		return id
	}
	for i := 0; ; i++ {
		id := fmt.Sprintf("%s%d", id, i)
		if _, ok := s.Entries[id]; !ok {
			return id
		}
	}
	return id
}

// Strips spaces from id and turns all characters lowercase
func NormalizeId(id string) string {
	return strings.Replace(strings.ToLower(id), " ", "-", -1)
}
