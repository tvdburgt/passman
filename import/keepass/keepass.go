package keepass

import (
	"encoding/xml"
	"github.com/tvdburgt/passman/import/util"
	"github.com/tvdburgt/passman/store"
	"io"
	"strings"
)

const timeLayout = "2006-01-02T15:04:05"

var importSettings *util.ImportSettings

func Import(r io.Reader, settings *util.ImportSettings) (s *store.Store, err error) {
	var db list
	importSettings = settings
	s = store.NewStore()
	dec := xml.NewDecoder(r)

	if err = dec.Decode(&db); err != nil {
		return
	}
	for _, e := range db.Entries {
		e.sync(s)
	}
	return
}

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
	Ctime    string `xml:"creationtime"`
	Mtime    string `xml:"lastmodtime"`
}

type group struct {
	Tree string `xml:"tree,attr"`
	Name string `xml:",chardata"`
}

func (e *entry) sync(s *store.Store) {
	// Skip entries from trash
	if e.Group.Name == "Recycle Bin" {
		return
	}

	// Resolve id
	id := e.id(s)

	// Build entry
	ee := &store.Entry{
		Name:     e.Username,
		Password: e.Password,
		Ctime:    util.ParseTime(e.Ctime, timeLayout),
		Mtime:    util.ParseTime(e.Mtime, timeLayout),
		Metadata: make(store.Metadata),
	}
	if len(e.Url) > 0 {
		ee.Metadata["url"] = e.Url
	}
	if len(e.Notes) > 0 {
		ee.Metadata["notes"] = e.Notes
	}
	s.Entries[id] = ee
}

func (e *entry) id(s *store.Store) (id string) {
	id = e.Title
	if len(id) == 0 {
		id = util.DefaultId
	}
	if importSettings.NameGroups && len(e.Group.Name) > 0 {
		var tree []string
		if len(e.Group.Tree) > 0 {
			tree = strings.Split(e.Group.Tree, "\\")
		}
		tree = append(tree, e.Group.Name)
		tree = append(tree, id)
		id = strings.Join(tree, "/")
	}
	if importSettings.NormalizeEntries {
		id = util.Normalize(id)
	}
	id = util.ResolveIdCollisions(s, id)
	return
}
