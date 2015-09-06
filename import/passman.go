package imprt

import (
	"encoding/json"
	"fmt"
	"github.com/tvdburgt/passman/store"
	"io"
)

type PassmanFormat store.Store

func (pf *PassmanFormat) Import(r io.Reader) (s *store.Store, err error) {
	// TODO: will this work?
	// s = store.NewStore()
	s = &store.Store{}
	dec := json.NewDecoder(r)
	if err = dec.Decode(s); err != nil {
		return
	}
	if s.Version != store.Version {
		return nil, fmt.Errorf("incorrect store version %d (expected %d)",
			s.Version, store.Version)
	}
	return
}
