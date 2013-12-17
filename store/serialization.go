package store

import (
	"crypto/cipher"
	"io"
	"crypto/hmac"
	"hash"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)


func (h* Header) MarshalBinary() (data []byte, err error) {


	return
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

	if !bytes.Equal(Signature[:], h.Signature[:]) {
		return errors.New("file is not a valid passman store")
	}

	if Version != h.Version {
		return fmt.Errorf("file version mismatch (%d, expected %d)",
			h.Version, Version)
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
