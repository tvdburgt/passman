package store

import (
	"bytes"
	"fmt"
	"reflect"
	"text/tabwriter"
	"time"
)

type Metadata map[string]string

type Entry struct {
	Name     string    `json:"name"`
	Password []byte    `json:"password"`
	Ctime    time.Time `json:"ctime"`    // Creation time
	Mtime    time.Time `json:"mtime"`    // Modification time
	Metadata Metadata  `json:"metadata"` // Map for custom fields
}

func NewEntry() *Entry {
	return &Entry{
		Metadata: make(Metadata),
		Ctime:    getCurrentTime(),
		Mtime:    getCurrentTime(),
	}
}

func (e *Entry) Age() time.Duration {
	return time.Since(e.Mtime)
}

func (e *Entry) Touch() {
	e.Mtime = getCurrentTime()
}

func (e Entry) String() string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 0, ' ', 0)

	// Use reflect to print each entry with json tags
	valof := reflect.ValueOf(e)
	for i := 0; i < valof.NumField(); i++ {
		switch val := valof.Field(i).Interface().(type) {
		default:
			tag := valof.Type().Field(i).Tag.Get("json")
			fmt.Fprintf(w, "%s\t : %s\n", tag, val)
		case Metadata:
			for k, v := range val {
				fmt.Fprintf(w, "%s\t : %s\n", k, v)
			}
		}
	}

	w.Flush()
	return b.String()
}

func (m Metadata) String() string {
	var b bytes.Buffer
	for k, v := range m {
		fmt.Fprintf(&b, "%s:%s", k, v)
	}
	return b.String()
}

func getCurrentTime() time.Time {
	return time.Now().Truncate(time.Second)
}
