package keepass

import (
	"encoding/xml"
	"github.com/tvdburgt/passman/import"
	"github.com/tvdburgt/passman/store"
	"io"
	"strings"
	"time"
)

type list struct {
	XMLName xml.Name `xml:"pwlist"`
	Entries []entry  `xml:"pwentry"`
}

type entry struct {
	Group    group  `xml:"group"`
	Title    string `xml:"title"`
	Username string `xml:"username"`
	Url      string `xml:"url"`
	Password []byte `xml:"password"`
	Notes    string `xml:"notes"`
	Time     string `xml:"lastmodtime"`
}

type group struct {
	Tree string `xml:"tree,attr"`
	Name string `xml:",chardata"`
}

func (e *entry) id(s *store.Store) (id string) {
	id = e.Title
	if len(id) == 0 {
		id = "unnamed"
	}
	if imprt.ImportGroups && len(e.Group.Name) > 0 {
		var tree []string
		if len(e.Group.Tree) > 0 {
			tree = strings.Split(e.Group.Tree, "\\")
		}
		tree = append(tree, e.Group.Name)
		tree = append(tree, id)
		id = strings.Join(tree, "/")
	}
	if imprt.NormalizeEntries {
		id = imprt.NormalizeId(id)
	}
	id = imprt.ResolveIdCollisions(s, id)
	return
}

// Custom time format that is used to parse XML strings to time.Time
const layout = "2006-01-02T15:04:05"

func ImportXml(in io.Reader) (s *store.Store, err error) {
	var db list
	s = store.NewStore()
	dec := xml.NewDecoder(in)

	if err = dec.Decode(&db); err != nil {
		return
	}
	for _, e := range db.Entries {
		addEntry(s, &e)
	}
	return
}

func addEntry(s *store.Store, e *entry) {
	// Skip entries from trash
	if e.Group.Name == "Recycle Bin" {
		return
	}

	// Resolve id
	id := e.id(s)

	// Parse time using a designated layout
	t, err := time.Parse(layout, e.Time)
	if err != nil {
		t = time.Now()
	}

	// Build entry
	ee := store.NewEntry()
	ee.Name = e.Username
	ee.Password = e.Password
	ee.Time = t
	if len(e.Url) > 0 {
		ee.Metadata["url"] = e.Url
	}
	if len(e.Notes) > 0 {
		ee.Metadata["notes"] = e.Notes
	}
	s.Entries[id] = ee
}
