package imprt

// import (
// 	"encoding/xml"
// 	"fmt"
// 	"github.com/tvdburgt/passman/store"
// 	"io"
// 	"strings"
// 	"time"
// )

// type kp_list struct {
// 	XMLName xml.Name `xml:"pwlist"`
// 	Entries []kp_entry  `xml:"pwentry"`
// }

// type kp_entry struct {
// 	Group    kp_group  `xml:"group"`
// 	Title    string `xml:"title"`
// 	Username string `xml:"username"`
// 	Url      string `xml:"url"`
// 	Password []byte `xml:"password"`
// 	Notes    string `xml:"notes"`
// 	Time     string `xml:"lastmodtime"`
// }

// type kp_group struct {
// 	Tree string `xml:"tree,attr"`
// 	Name string `xml:",chardata"`
// }

// func (e *kp_entry) id(s *store.Store) (id string) {
// 	id = e.Title
// 	if len(id) == 0 {
// 		id = "unnamed"
// 	}
// 	if ImportGroups && len(e.Group.Name) > 0 {
// 		var tree []string
// 		if len(e.Group.Tree) > 0 {
// 			tree = strings.Split(e.Group.Tree, `\`)
// 		}
// 		tree = append(tree, e.Group.Name)
// 		tree = append(tree, id)
// 		id = strings.Join(tree, "/")
// 	}
// 	if NormalizeEntries {
// 		id = NormalizeId(id)
// 	}
// 	id = ResolveIdCollisions(s, id)
// 	return
// }

// type KeepassFormat kp_list

// func (kp *KeepassFormat) Import(r io.Reader) (s *store.Store, err error) {
// 	dec := xml.NewDecoder(r)
// 	dec.Decode(kp)
// 	fmt.Println(kp)
// 	return
// }

// func ImportXml(in io.Reader) (s *store.Store, err error) {
// 	var list kp_list
// 	s = store.NewStore(store.NewHeader())
// 	dec := xml.NewDecoder(in)

// 	if err = dec.Decode(&list); err != nil {
// 		return
// 	}
// 	for _, e := range list.Entries {
// 		addEntry(s, &e)
// 	}
// 	return
// }

// func addEntry(s *store.Store, e *kp_entry) {
// 	// Skip entries from trash
// 	if e.Group.Name == "Recycle Bin" {
// 		return
// 	}

// 	// Resolve id
// 	id := e.id(s)

// 	// Parse time using a designated layout
// 	const layout = "2006-01-02T15:04:05"
// 	t, err := time.Parse(layout, e.Time)
// 	if err != nil {
// 		t = time.Now()
// 	}

// 	// Build entry
// 	ee := store.NewEntry()
// 	ee.Name = e.Username
// 	ee.Password = e.Password
// 	ee.Mtime = t
// 	if len(e.Url) > 0 {
// 		ee.Metadata["url"] = e.Url
// 	}
// 	if len(e.Notes) > 0 {
// 		ee.Metadata["notes"] = e.Notes
// 	}
// 	s.Entries[id] = ee
// }
