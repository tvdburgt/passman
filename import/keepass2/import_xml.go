package keepass2

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/tvdburgt/passman/import"
	"github.com/tvdburgt/passman/store"
	"io"
	"strings"
	"time"
)

type database struct {
	XMLName   xml.Name `xml:"KeePassFile"`
	Generator *string  `xml:"Meta>Generator"`
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
	Time time.Time `xml:"Times>LastModificationTime"`
}

const fileGenerator = "KeePass"

func ImportXml(in io.Reader) (s *store.Store, err error) {
	var db database
	s = store.NewStore()
	dec := xml.NewDecoder(in)

	if err = dec.Decode(&db); err != nil {
		return
	}
	if db.Generator == nil {
		return nil, errors.New("invalid format: missing 'Generator' tag")
	}
	if *db.Generator != fileGenerator {
		return nil, fmt.Errorf("invalid format: 'Generator' value is '%s' (expected '%s')",
			db.Generator, fileGenerator)
	}
	for _, g := range db.Groups {
		var tree []string
		addGroup(s, &g, tree)
	}
	return
}

func addGroup(s *store.Store, g *group, tree []string) {
	if g.Name == "Recycle Bin" {
		return
	}
	tree = append(tree, g.Name)
	for _, child := range g.Groups {
		addGroup(s, &child, tree)
	}
	for _, e := range g.Entries {
		addEntry(s, &e, tree)
	}
}

func addEntry(s *store.Store, e *entry, tree []string) {
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
		default: // URL, Notes or custom field
			if len(field.Value) > 0 {
				key := field.Key
				if imprt.NormalizeEntries {
					key = strings.ToLower(field.Key)
				}
				ee.Metadata[key] = field.Value
			}
		}
	}

	// Make sure id is non-empty
	if len(id) == 0 {
		id = "unnamed"
	}
	if imprt.ImportGroups {
		tree = append(tree, id)
		id = strings.Join(tree[1:], "/") // Discard root (database) group
	}
	if imprt.NormalizeEntries {
		id = imprt.NormalizeId(id)
	}

	id = imprt.ResolveIdCollisions(s, id)
	ee.Time = e.Time
	s.Entries[id] = ee
}
