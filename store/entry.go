package store

import (
	"bytes"
	"fmt"
	"text/tabwriter"
	"time"
)

const secondsPerMonth = 2.63e+6

type Metadata map[string]string

type Entry struct {
	Name     string    `json:"name"`
	Password []byte    `json:"password"`
	Time     time.Time `json:"time"`
	Metadata `json:"metadata"`
}

func NewEntry() *Entry {
	return &Entry{Metadata: make(Metadata)}
}

func (e *Entry) Age() time.Duration {
	return time.Since(e.Time)
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
	e.Time = time.Now()
}

func (e *Entry) String() string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 8, 0, '\t', 0)

	age := e.Age().Seconds() / secondsPerMonth

	fmt.Fprintf(w, "name:\t%s\n", e.Name)
	fmt.Fprintf(w, "pass:\t%q\n", e.Password)
	fmt.Fprintf(w, "age:\t%.1f months\n", age)

	for key, val := range e.Metadata {
		fmt.Fprintf(w, "%s:\t%s\n", key, val)
	}

	w.Flush()
	return b.String()
}
