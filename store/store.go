// Package store contains the main data structure for passman.
// Store consists of a header and a map of entries. The store is used for
// (de)serialization as part of the encryption/decryption process.
//
// File format
// ---------------------------------------------------------
// Offset	Length		Description
// ---------------------------------------------------------
// 0		7		signature / magic number
// 7		1		file format version
// 8		1		scrypt param: log2(N)
// 9		4		scrypt param: r
// 13		4		scrypt param: n
// 17		32		salt
// 49		32		HMAC-SHA256(0 .. 32)
// ---------------------------------------------------------
// 81		n		encrypted entry data
// 81+n		32		HMAC-SHA256(0 .. 81 + (n - 1))
package store

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
)

type EntryMap map[string]*Entry

type Store struct {
	Header  `json:"header"`
	Entries EntryMap `json:"entries"`
}

// func NewStore(h *Header) *Store {
// 	return &Store{*h, make(EntryMap)}
// }

func NewStore() *Store {
	return &Store{*NewHeader(), make(EntryMap)}
}

// TODO: move to list.go?
func (s *Store) List(out io.Writer, pattern *regexp.Regexp) {
	ids := s.ids(pattern)
	if len(ids) == 0 {
		fmt.Fprintln(out, "No entries found.")
		return
	}

	// check $COLUMNS and $LINES
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "id\tname\tmetadata")
	for _, id := range ids {
		e := s.Entries[id]
		//const secondsPerMonth = 2.63e+6 // TODO: move to inside fn
		// months := e.Age().Seconds() / secondsPerMonth
		// fmt.Fprintf(w, "%s\t%s\t%.1f\n",
		// 	id,
		// 	e.Name,
		// 	months)

		fmt.Fprintf(w, "%s\t%s\t%s\n",
			id,
			e.Name,
			e.Metadata)
	}

	w.Flush()
	header, _ := b.ReadString('\n')
	hr := strings.Repeat("-", len(header)-1) + "\n"
	fmt.Fprint(out, hr, header, hr, b.String(), hr)
}

// func maxWidth(b bytes.Buffer) int {
// 	var n int
// 	for {
// 		s, err := b.ReadString('\n')
// 		if err != nil {
// 			return n
// 		}
// 		if len(s) > n {
// 			n = len(s)
// 		}
// 	}
// }

func (s *Store) String() string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 8, 2, '\t', 0)

	fmt.Fprintln(w, "Id:\tName:\tPassword:")

	for _, id := range s.ids(nil) {
		e := s.Entries[id]
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			id,
			e.Name,
			e.Password)
	}

	w.Flush()
	return b.String()
}

// Returns entry ids with given prefix in sorted order
func (s *Store) ids(pattern *regexp.Regexp) []string {
	ids := make([]string, 0, len(s.Entries))
	for id := range s.Entries {
		if pattern == nil || pattern.MatchString(id) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
