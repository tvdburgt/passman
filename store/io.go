package store

import (
	"crypto/cipher"
	"crypto/hmac"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
)

var ErrWrongPass = errors.New("incorrect passphrase")

var byteOrder = binary.LittleEndian

func (h *Header) Marshal(w io.Writer) error {
	return binary.Write(w, byteOrder, h)
}

func (h *Header) Unmarshal(r io.Reader) error {
	if err := binary.Read(r, byteOrder, h); err != nil {
		return err
	}
	if Signature != h.Signature {
		return errors.New("invalid store (incorrect signature)")
	}
	if Version != h.Version {
		return fmt.Errorf("file version mismatch (%d, expected %d)",
			h.Version, Version)
	}
	return nil
}

func (s *Store) Encrypt(sw cipher.StreamWriter, mac hash.Hash) error {
	plainWriter := io.MultiWriter(sw.W, mac)
	cipherWriter := io.MultiWriter(sw, mac)

	// Write header (plaintext)
	if err := s.Header.Marshal(plainWriter); err != nil {
		return err
	}

	// Write header HMAC (plaintext)
	if _, err := plainWriter.Write(mac.Sum(nil)); err != nil {
		return err
	}

	// Write JSON-encoded entries (ciphertext)
	enc := json.NewEncoder(cipherWriter)
	if err := enc.Encode(s.Entries); err != nil {
		return err
	}

	// Write file HMAC (plaintext)
	if _, err := plainWriter.Write(mac.Sum(nil)); err != nil {
		return err
	}

	return nil
}

func (s *Store) Decrypt(sr cipher.StreamReader, n int64, mac hash.Hash) error {
	lr := &io.LimitedReader{sr.R, n - int64(mac.Size())}
	sr.R = lr
	plainReader := io.TeeReader(sr.R, mac)
	cipherReader := io.TeeReader(sr, mac)

	// Read header
	if err := binary.Read(plainReader, byteOrder, &s.Header); err != nil {
		return err
	}

	// Verify header HMAC
	if !checkHMAC(plainReader, mac) {
		return ErrWrongPass
	}

	// Read and decrypt entries
	dec := json.NewDecoder(cipherReader)
	if err := dec.Decode(&s.Entries); err != nil {
		return err
	}

	// Verify file HMAC
	if !checkHMAC(lr.R, mac) {
		return ErrWrongPass
	}

	return nil
}

// Reads HMAC from r and compares it with mac
func checkHMAC(r io.Reader, mac hash.Hash) bool {
	macExpected := mac.Sum(nil)
	macRead := make([]byte, mac.Size())
	if _, err := r.Read(macRead); err != nil {
		return false
	}
	return hmac.Equal(macRead, macExpected)
}
