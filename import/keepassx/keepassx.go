package keepassx

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
	var db database
	importSettings = settings
	s = store.NewStore()
	dec := xml.NewDecoder(r)

	if err = dec.Decode(&db); err != nil {
		return
	}
	for _, g := range db.Groups {
		var tree []string
		g.sync(s, tree)
	}
	return
}

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
	Ctime    string `xml:"creation"`
	Mtime    string `xml:"lastmod"`
}

func (e *entry) id(s *store.Store, tree []string) (id string) {
	id = e.Title
	if len(id) == 0 {
		id = util.DefaultId
	}
	if importSettings.NameGroups {
		id = strings.Join(tree, "/")
	}
	if importSettings.NormalizeEntries {
		id = util.Normalize(id)
	}
	id = util.ResolveIdCollisions(s, id)
	return
}

func (g *group) sync(s *store.Store, tree []string) {
	if g.Title == "Backup" || g.Title == "Recycle Bin" {
		return
	}
	tree = append(tree, g.Title)
	for _, child := range g.Groups {
		child.sync(s, tree)
	}
	for _, e := range g.Entries {
		e.sync(s, append(tree, e.Title))
	}
}

func (e *entry) sync(s *store.Store, tree []string) {
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
	if len(e.Comment) > 0 {
		ee.Metadata["comment"] = e.Comment
	}

	id := e.id(s, tree)
	s.Entries[id] = ee
}
