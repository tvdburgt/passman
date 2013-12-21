package store

import (
	"bytes"
	"fmt"
	"text/tabwriter"
	"time"
)

const monthDuration = 2.63e+6

type Entry struct {
	Name     string            `json:"name,omitempty"`
	Password []byte            `json:"password,omitempty"` // TODO: make private?
	Time     int64             `json:"time,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"` // Allow metadata to be null
	// Entropy  int             `json:"entropy,omitempty"`
}

// func NewEntry() *Entry {
// 	return &Entry{Metadata: make(map[string]string)}
// }

func (e *Entry) Age() time.Duration {
	t := time.Unix(e.Time, 0)
	return time.Since(t)
}

// TODO: investigate custom function scoping
func (e *Entry) ProcessPassword(fn func([]byte)) {
	// Decrypt password
	password := e.Password
	fn(password)
	// clear password
	// Encrypt password
}

func (e *Entry) Touch() {
	e.Time = time.Now().Unix()
}

func (e *Entry) String() string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 8, 0, '\t', 0)

	months := e.Age().Seconds() / monthDuration

	fmt.Fprintf(w, "name:\t%s\n", e.Name)
	fmt.Fprintf(w, "pass:\t%q\n", e.Password)
	fmt.Fprintf(w, "age:\t%.1f months\n", months)

	for key, val := range e.Metadata {
		fmt.Fprintf(w, "%s:\t%s\n", key, val)
	}

	w.Flush()
	return b.String()
}
