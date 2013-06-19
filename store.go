/*
 * Offset	Size		Description
 * ---------------------------------------------------------
 * 0		7		File signature
 * 7		1		File format version
 * 8		32		Salt
 * 40		32		HMAC-SHA256(0..32)
 -----------------------------------------------------------
 * 72		n		Encrypted content
 * 72+n		32		HMAC-SHA256(0..n-32)
*/

package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"text/tabwriter"
)

type Header struct {
	Data      []byte `json:"-"`
	Signature []byte `json:"-"`
	Version   byte
	Salt      []byte `json:"-"`
	HMAC      []byte `json:"-"`
}

type Entry struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Age      int64  `json:"age"`

	/* Metadata map[string]string */
}

type Store struct {
	Header  `json:"header"`
	Entries map[string]Entry `json:"entries"`
}

func NewStore(header Header) *Store {
	entries := make(map[string]Entry)
	return &Store{header, entries}
}

func (h *Header) Serialize(out io.Writer, mac hash.Hash) error {
	fields := [][]byte{
		h.Signature,
		[]byte{h.Version},
		h.Salt,
	}

	// Write data
	if _, err := out.Write(bytes.Join(fields, nil)); err != nil {
		return err
	}

	// Write HMAC
	if _, err := out.Write(mac.Sum(nil)); err != nil {
		return err
	}

	fields = append(fields, mac.Sum(nil))

	return nil
}

func (h *Header) Deserialize(in io.Reader) error {
	h.Data = make([]byte, 72)
	if _, err := in.Read(h.Data); err != nil {
		return err
	}

	h.Signature = h.Data[:7]
	h.Version = h.Data[7]
	h.Salt = h.Data[8:40]
	h.HMAC = h.Data[40:72]

	if !bytes.Equal(fileSignature, h.Signature) {
		return errors.New("invalid pass store")
	}

	if fileVersion != h.Version {
		return fmt.Errorf("pass store version mismatch (%d, expected %d)",
			h.Version, fileVersion)
	}

	// Strip HMAC from header data
	/* h.Data = h.Data[:40] */

	return nil
}

func (s *Store) Serialize(out io.Writer, cipherStream cipher.Stream, mac hash.Hash) error {

	plainWriter := io.MultiWriter(out, mac)
	cipherWriter := cipher.StreamWriter{cipherStream, plainWriter, nil}
	enc := json.NewEncoder(cipherWriter)

	// Write header
	if err := s.Header.Serialize(plainWriter, mac); err != nil {
		return err
	}

	// Go forth, and marshal thyself!
	if err := enc.Encode(s.Entries); err != nil {
		return err
	}

	// Write HMAC of complete file
	out.Write(mac.Sum(nil))

	return nil
}

func (s *Store) Deserialize(in io.Reader, size int, cipherStream cipher.Stream, mac hash.Hash) error {

	plainReader := io.TeeReader(in, mac)
	cipherReader := cipher.StreamReader{cipherStream, plainReader}
	/* dec := json.NewDecoder(cipherReader) */

	content := make([]byte, size-len(s.Data)-mac.Size())
	/* content := make([]byte, size - len(s.Data)) */
	if _, err := cipherReader.Read(content); err != nil {
		return err
	}

	h := make([]byte, mac.Size())
	if _, err := in.Read(h); err != nil {
		return err
	}

	if !hmac.Equal(h, mac.Sum(nil)) {
		return errors.New("invalid pass store")
	}

	if err := json.Unmarshal(content, &s.Entries); err != nil {
		return err
	}

	return nil
}

func (s *Store) Export(out io.Writer) error {
	content, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	out.Write(content)

	return nil
}

func (s *Store) Set(id string, e Entry) error {
	s.Entries[id] = e
	return nil
}

// todo use write!!
/* func (s *Store) String() string { */
/* 	// decrypt... */
/* 	// check indent size */
/* 	json, err := json.MarshalIndent(s, "", "  ") */

/* 	if err != nil { */
/* 		panic("JSON marshal failed") */
/* 	} */

/* 	return string(json) */

/* 	// use defer to encrypt again? */
/* } */

/* func (s Store) MarshalJSON() ([]byte, error) { */
/* 	return json.Marshal(s) */
/* } */

func (s *Store) Print() {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, '\t', 0)

	fmt.Fprintln(w, "Id:\tName:\tPassword:")
	for id, entry := range s.Entries {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			id,
			entry.Name,
			entry.Password)
	}

	w.Flush()
}
