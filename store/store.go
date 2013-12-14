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
	Entries map[string]*Entry `json:"entries"`
}

func NewHeader() *Header {
	return &Header{Version: Version, Signature: Signature}
}

func NewStore(header *Header) *Store {
	return &Store{*header, make(map[string]*Entry)}
}

func (h *Header) Size() int {
	return int(unsafe.Sizeof(*h))
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

// Returns entry ids in sorted order
func (s *Store) ids() []string {
	ids := make([]string, 0, len(s.Entries))
	for id := range s.Entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (s *Store) String() string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 8, 2, '\t', 0)

	fmt.Fprintln(w, "Id:\tName:\tPassword:")

	for _, id := range s.ids() {
		e := s.Entries[id]
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			id,
			e.Name,
			e.Password)
	}

	w.Flush()
	return b.String()
}
