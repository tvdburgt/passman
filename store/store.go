package store

import (
	"bytes"
	"crypto/cipher"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"sort"
	"text/tabwriter"
	"unsafe"
)

// TODO: expose?
const fileVersion = 0x00
const monthDuration = 2.63e+6

// TODO: check endianness
var fileSignature = [7]byte{0x6e, 0x50, 0x41, 0x53, 0x4d, 0x41, 0x4e}

//var fileSignature = []byte("passman")

/* File format
* ---------------------------------------------------------
* Offset	Size		Description
* ---------------------------------------------------------
* 0		7		File signature / magic number
* 7		1		File format version
* 8		32		Salt
* 40		32		HMAC-SHA256(0 .. 32)
-----------------------------------------------------------
* 72		n		Encrypted data
* 72+n		32		HMAC-SHA256(0 .. 72 + (n - 1))
*/

type Header struct {
	Signature [7]byte  `json:"-"`
	Version   byte     `json:"version"`
	Salt      [32]byte `json:"-"`
	HMAC      [32]byte `json:"-"`
}

type Store struct {
	Header  `json:"header"`
	Entries map[string]*Entry `json:"entries"`
}

func NewHeader() *Header {
	return &Header{Version: fileVersion, Signature: fileSignature}
}

func NewStore(header *Header) *Store {
	// required to make map?
	return &Store{*header, make(map[string]*Entry)}
}

func (h *Header) Size() int {
	return int(unsafe.Sizeof(*h))
}

func (h *Header) Serialize(out io.Writer, mac hash.Hash) error {
	buf := new(bytes.Buffer)
	buf.Write(h.Signature[:])
	buf.WriteByte(h.Version)
	buf.Write(h.Salt[:])

	if _, err := buf.WriteTo(out); err != nil {
		return err
	}

	// fmt.Printf("salt=%x\n", h.Salt)
	// fmt.Printf("h_hmac=%x\n", mac.Sum(nil))

	// Write HMAC
	if _, err := out.Write(mac.Sum(nil)); err != nil {
		return err
	}

	return nil
}

func (s *Store) Serialize(out io.Writer, cipherStream cipher.Stream, mac hash.Hash) error {
	mac.Reset()
	plainWriter := io.MultiWriter(out, mac)
	cipherWriter := cipher.StreamWriter{cipherStream, plainWriter, nil}
	enc := json.NewEncoder(cipherWriter)

	// Write plaintext header
	if err := s.Header.Serialize(plainWriter, mac); err != nil {
		return err
	}

	// Encrypt and write JSON-encoded entries
	if err := enc.Encode(s.Entries); err != nil {
		return err
	}

	// fmt.Printf("hmac=%x\n", mac.Sum(nil))

	// Write plaintext HMAC of preceeding data
	out.Write(mac.Sum(nil))

	return nil
}

func (h *Header) Deserialize(in io.Reader) error {
	data := make([]byte, h.Size())
	if _, err := in.Read(data); err != nil {
		return err
	}

	// Sequentially read header elements
	r := bytes.NewReader(data)
	r.Read(h.Signature[:])
	h.Version, _ = r.ReadByte()
	r.Read(h.Salt[:])
	r.Read(h.HMAC[:])

	if !bytes.Equal(fileSignature[:], h.Signature[:]) {
		return errors.New("file is not a valid passman store")
	}

	if fileVersion != h.Version {
		return fmt.Errorf("file version mismatch (%d, expected %d)",
			h.Version, fileVersion)
	}

	// fmt.Printf("salt=%x\n", h.Salt)
	// fmt.Printf("h_hmac=%x\n", h.HMAC)

	return nil
}

func (s *Store) Deserialize(in io.Reader, size int, cipherStream cipher.Stream, mac hash.Hash) error {
	mac.Reset()
	plainReader := io.TeeReader(in, mac)
	plainReader.Read(make([]byte, s.Header.Size()-mac.Size()))

	if !hmac.Equal(s.Header.HMAC[:], mac.Sum(nil)) {
		return errors.New("incorrect passphrase")
	}

	plainReader.Read(make([]byte, mac.Size()))

	cipherReader := cipher.StreamReader{cipherStream, plainReader}
	content := make([]byte, size-s.Header.Size()-mac.Size())

	if _, err := cipherReader.Read(content); err != nil {
		return err
	}

	h := make([]byte, mac.Size())
	if _, err := in.Read(h); err != nil {
		return err
	}

	// fmt.Printf("hmac=%x\n", mac.Sum(nil))

	if !hmac.Equal(h, mac.Sum(nil)) {
		return errors.New("incorrect passphrase")
	}

	if err := json.Unmarshal(content, &s.Entries); err != nil {
		return err
	}

	return nil
}

func (s *Store) Export(out io.Writer) (err error) {
	var content []byte
	if content, err = json.MarshalIndent(s, "", "  "); err != nil {
		return
	}
	out.Write(content)
	fmt.Fprintln(out)
	return
}

func (s *Store) String() string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 8, 2, '\t', 0)

	// Copy entries to slice and sort by id
	keys := make([]string, 0, len(s.Entries))
	for id := range s.Entries {
		keys = append(keys, id)
	}
	sort.Strings(keys)

	// Print all entries using a tabular format
	fmt.Fprintln(w, "Id:\tName:\tPassword:")
	for _, id := range keys {
		entry := s.Entries[id]
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			id,
			entry.Name,
			entry.Password)
	}

	w.Flush()
	return b.String()
}
