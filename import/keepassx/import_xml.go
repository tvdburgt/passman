package keepassx

import (
	"encoding/xml"
	"github.com/tvdburgt/passman/import"
	"github.com/tvdburgt/passman/store"
	"io"
	"strings"
	"time"
)

type database struct {
	XMLName xml.Name `xml:"database"`
	Groups  []group  `xml:"group"`
}

type group struct {
	Title   string  `xml:"title"`
	Entries []entry `xml:"entry"`
	Groups  []group `xml:"group"`
}

type entry struct {
	Title    string `xml:"title"`
	Username string `xml:"username"`
	Url      string `xml:"url"`
	Password []byte `xml:"password"`
	Comment  string `xml:"comment"`
	Time     string `xml:"lastmod"`
}

func (e *entry) id(s *store.Store, tree []string) (id string) {
	id = e.Title
	if len(id) == 0 {
		id = "unnamed"
	}
	if imprt.ImportGroups {
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
	var db database
	s = store.NewStore(store.NewHeader())
	dec := xml.NewDecoder(in)

	if err = dec.Decode(&db); err != nil {
		return
	}
	for _, g := range db.Groups {
		var tree []string
		addGroup(s, &g, tree)
	}
	return
}

func addGroup(s *store.Store, g *group, tree []string) {
	if g.Title == "Backup" || g.Title == "Recycle Bin" {
		return
	}
	tree = append(tree, g.Title)
	for _, child := range g.Groups {
		addGroup(s, &child, tree)
	}
	for _, e := range g.Entries {
		addEntry(s, &e, append(tree, e.Title))
	}
}

func addEntry(s *store.Store, e *entry, tree []string) {
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
	if len(e.Comment) > 0 {
		ee.Metadata["comment"] = e.Comment
	}
	id := e.id(s, tree)
	s.Entries[id] = ee
}
