package keepass2

import (
	"encoding/xml"
	"fmt"
	"github.com/tvdburgt/passman/import/util"
	"github.com/tvdburgt/passman/store"
	"io"
	"strings"
	"time"
)

const fileGenerator = "KeePass"

var importSettings *util.ImportSettings

func Import(r io.Reader, settings *util.ImportSettings) (s *store.Store, err error) {
	var db database
	importSettings = settings
	s = store.NewStore()

	dec := xml.NewDecoder(r)

	if err = dec.Decode(&db); err != nil {
		return
	}

	// Sanity check on <Generator> value
	if db.Generator != fileGenerator {
		return nil, fmt.Errorf("invalid format: <Generator> contains %q (expecting %q)",
			db.Generator, fileGenerator)
	}

	for _, g := range db.Groups {
		var tree []string
		g.sync(s, tree)
	}
	return
}

type database struct {
	XMLName   xml.Name `xml:"KeePassFile"`
	Generator string   `xml:"Meta>Generator"`
	Groups    []group  `xml:"Root>Group"`
}

type group struct {
	Name    string
	Groups  []group `xml:"Group"`
	Entries []entry `xml:"Entry"`
}

type entry struct {
	Data []struct {
		Key, Value string
	} `xml:"String"`
	Ctime time.Time `xml:"Times>CreationTime"`
	Mtime time.Time `xml:"Times>LastModificationTime"`
}

func (g *group) sync(s *store.Store, tree []string) {
	if g.Name == "Recycle Bin" {
		return
	}
	tree = append(tree, g.Name)
	for _, child := range g.Groups {
		child.sync(s, tree)
	}
	for _, e := range g.Entries {
		e.sync(s, tree)
	}
}

func (e *entry) sync(s *store.Store, tree []string) {
	var id string
	ee := store.NewEntry()

	// Process each entry field
	for _, field := range e.Data {
		switch field.Key {
		case "Title":
			id = field.Value
		case "Password":
			ee.Password = []byte(field.Value)
		case "UserName":
			ee.Name = field.Value
		default: // Arbitrary metadata fields
			if len(field.Value) > 0 {
				key := field.Key
				if importSettings.NormalizeEntries {
					key = util.Normalize(key)
				}
				ee.Metadata[key] = field.Value
			}
		}
	}

	// Make sure id is non-empty
	if len(id) == 0 {
		id = util.DefaultId
	}
	if importSettings.NameGroups {
		tree = append(tree, id)
		id = strings.Join(tree[1:], "/") // Discard root (database) group
	}
	if importSettings.NormalizeEntries {
		id = util.Normalize(id)
	}

	id = util.ResolveIdCollisions(s, id)
	ee.Ctime = e.Ctime
	ee.Mtime = e.Mtime
	s.Entries[id] = ee
}
