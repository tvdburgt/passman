package store

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	Version = 0x0
)

var (
	byteOrder = binary.LittleEndian
	// TODO: []byte("passman")
	signature = [7]byte{112, 97, 115, 115, 109, 97, 110}
)

type ScryptParams struct {
	LogN byte   `json:"log_n"` // Work factor (iteration count)
	R    uint32 `json:"r"`     // Block size underlying hash
	P    uint32 `json:"p"`     // Parallelization factor
}

type Header struct {
	Signature [7]byte      `json:"-"`
	Version   byte         `json:"version"`
	Params    ScryptParams `json:"params"`
	Salt      [32]byte     `json:"-"`
}

func NewHeader() *Header {
	return &Header{
		Version:   Version,
		Signature: signature,
		Params:    ScryptParams{14, 8, 1},
	}
}

func (h *Header) Marshal(w io.Writer) error {
	return binary.Write(w, byteOrder, h)
}

func (h *Header) Unmarshal(r io.Reader) error {
	if err := binary.Read(r, byteOrder, h); err != nil {
		return err
	}
	if signature != h.Signature {
		return errors.New("invalid store (incorrect signature)")
	}
	if Version != h.Version {
		return fmt.Errorf("file version mismatch (%d, expected %d)",
			h.Version, Version)
	}
	return nil
}
