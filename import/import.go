package imprt

import (
	"fmt"
	"github.com/tvdburgt/passman/import/keepass"
	"github.com/tvdburgt/passman/import/keepass2"
	"github.com/tvdburgt/passman/import/keepassx"
	"github.com/tvdburgt/passman/import/util"
	"github.com/tvdburgt/passman/store"
	"io"
)

type Settings util.ImportSettings

type importFunc func(r io.Reader, settings *util.ImportSettings) (s *store.Store, err error)

var importers = map[string]importFunc{
	"keepass":  keepass.Import,
	"keepass2": keepass2.Import,
	"keepassx": keepassx.Import,
}

func ImportStore(r io.Reader, format string, settings *Settings) (s *store.Store, err error) {
	if fn, ok := importers[format]; ok {
		utilSettings := util.ImportSettings(*settings)
		return fn(r, &utilSettings)
	} else {
		return nil, fmt.Errorf("no importer available for format '%s'", format)
	}
}
