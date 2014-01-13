package store

import (
	"bytes"
	"fmt"
	"reflect"
	"text/tabwriter"
	"time"
)

const secondsPerMonth = 2.63e+6

type Metadata map[string]string

type Entry struct {
	Name     string `json:"name"`
	Password []byte `json:"password"`

	// Time of previous password modification
	Mtime time.Time `json:"mtime"`

	// Map that holds user-defined fields and values
	Metadata `json:"metadata"`
}

func NewEntry() *Entry {
	return &Entry{Metadata: make(Metadata)}
}

func (e *Entry) Age() time.Duration {
	return time.Since(e.Mtime)
}

// TODO: Create scope for plaintext password
func (e *Entry) PlainPassword(fn func([]byte)) {
	// Decrypt password
	password := e.Password
	fn(password)
	// Clear password
	// Encrypt password
}

func (e *Entry) Touch() {
	e.Mtime = time.Now()
}

func (e *Entry) String() string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 0, ' ', 0)

	// Use reflect to print entry with json tags
	valof := reflect.ValueOf(*e)
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
