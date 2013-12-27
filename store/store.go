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
	"github.com/tvdburgt/passman/crypto"
	"io"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"unsafe"
)

const (
	Version = 0x00
)

// TODO: check endianness
var Signature = [7]byte{0x6e, 0x50, 0x41, 0x53, 0x4d, 0x41, 0x4e}

//var fileSignature = []byte("passman")

// Use nested struct?
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

// func NewHeader() *Header {
// 	return &Header{Version: Version, Signature: Signature}
// }

func NewStore() *Store {
	return &Store{
		Header:  Header{Version: Version, Signature: Signature},
		Entries: make(map[string]*Entry),
	}
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

func (s *Store) Import(in io.Reader) (err error) {
	dec := json.NewDecoder(in)
	if err = dec.Decode(s); err != nil {
		return
	}
	if s.Version != Version {
		return fmt.Errorf("incorrect store version %d (expected %d)",
			s.Version, Version)
	}
	return
}

func (s *Store) Close() {
	for _, e := range s.Entries {
		crypto.Clear(e.Password) // crypto dependency
	}
}

// format: %#v
func (s *Store) GoString() string {
	return ""
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

func (s *Store) List(out io.Writer, pattern *regexp.Regexp) {
	// check $COLUMNS and $LINES
	w := tabwriter.NewWriter(out, 0, 8, 2, '\t', 0)
	n, _ := fmt.Fprintln(w, "id\tname\tage")
	fmt.Fprintln(w, strings.Repeat("-", 80))
	_ = strings.Repeat("-", n)
	for _, id := range s.ids(pattern) {
		e := s.Entries[id]
		months := e.Age().Seconds() / monthDuration
		fmt.Fprintf(w, "%s\t%s\t%.1f months\n",
			id,
			e.Name,
			months)
	}
	fmt.Fprintln(w, strings.Repeat("-", 80))
	w.Flush()
}

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
