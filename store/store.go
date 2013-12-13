// File format
// ---------------------------------------------------------
// Offset	Size		Description
// ---------------------------------------------------------
// 0		7		File signature / magic number
// 7		1		File format version
// 8		32		Salt
// 40		32		HMAC-SHA256(0 .. 32)
// ---------------------------------------------------------
// 72		n		Encrypted data
// 72+n		32		HMAC-SHA256(0 .. 72 + (n - 1))
package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"unsafe"
)

const (
	Version       = 0x00
	monthDuration = 2.63e+6
)

// TODO: check endianness
var Signature = [7]byte{0x6e, 0x50, 0x41, 0x53, 0x4d, 0x41, 0x4e}

//var fileSignature = []byte("passman")

type container map[string]*Entry

// Returns entry ids in sorted order
func (c container) ids() []string {
	ids := make([]string, 0, len(c))
	for id := range c {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}


// Nested struct?
type Header struct {
	Signature [7]byte  `json:"-"`
	Version   byte     `json:"version"`
	Salt      [32]byte `json:"-"`
	HMAC      [32]byte `json:"-"`
}

type Store struct {
	Header `json:"header"`
	// Entries map[string]*Entry `json:"entries"`
	Entries container `json:"entries"`
}

func NewHeader() *Header {
	return &Header{Version: Version, Signature: Signature}
}

func NewStore(header *Header) *Store {
	return &Store{*header, make(container)}
}

func (h *Header) Size() int {
	return int(unsafe.Sizeof(*h))
}

func (s *Store) find(ids []string) (e *Entry) {
	entries := s.Entries
	var ok bool
	for _, id := range ids {
		if entries == nil {
			return nil
		}
		if e, ok = entries[id]; !ok {
			return
		}
		entries = e.Entries
	}
	return
}

func (s *Store) Find(id string) (e *Entry) {
	return s.find(strings.Split(id, "/"))
}

// Returns direct parent of entry. Does not check for existance!
func (s *Store) Parent(id string) (e *Entry) {
	ids := strings.Split(id, "/")
	n := len(ids)
	if n < 2 {
		return
	}
	return s.find(ids[:n-1])
}

func (s *Store) Remove(id string) (err error) {
	e := s.Find(id)
	if e != nil {
		return fmt.Errorf("no such entry with id '%s'")
	}

	ids := strings.Split(id, "/")
	if len(ids) == 1 {
		id = ids[len(ids)-1]
		delete(s.Entries, id)
	} else {
		parent := s.Parent(id)
		delete(parent.Entries, id)
	}

	return
}

func (s *Store) Insert(id string, e *Entry) (err error) {
	ids := strings.Split(id, "/")
	n := len(ids)

	// Check correctness of id segments
	for _, id := range ids {
		if len(id) == 0 {
			return fmt.Errorf("invalid id")
		}
	}

	// Traverse all individual parents. Create new entries if needed.
	// All existing parents need to be a container.
	entries := s.Entries
	for _, id := range ids[:n-1] {
		parent, ok := entries[id]
		if !ok {
			parent = newContainer()
			entries[id] = parent
		} else if !e.IsContainer() {
			return fmt.Errorf("path contains invalid parent '%s'", id)
		}
		entries = parent.Entries
	}

	id = ids[n-1]
	entries[id] = e
	return
}

func (s *Store) Export(out io.Writer) (err error) {
	content, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	out.Write(content)
	fmt.Fprintln(out)
	return
}

// format: %#v
func (s *Store) GoString() string {
	return ""
}

func (s *Store) String() string {
	var dfs func(c container, prefix string)
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 8, 2, '\t', 0)

	fmt.Fprintln(w, "Id:\tName:\tPassword:")

	dfs = func(c container, prefix string) {
		for _, id := range c.ids() {
			e := c[id]
			if e.IsContainer() {
				dfs(e.Entries, prefix + id + "/")
			} else {
				fmt.Fprintf(w,"%s\t%s\t%s\n",
				prefix + id,
				e.Name,
				e.Password)
			}
		}
	}

	dfs(s.Entries, "")
	w.Flush()
	return b.String()
}
