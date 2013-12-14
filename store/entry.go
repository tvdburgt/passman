package store

import (
	"fmt"
	"time"
)

type Entry struct {
	Name     string            `json:"name,omitempty"`     // remove omitempty
	Password string            `json:"password,omitempty"` // use []byte to prevent json errors
	Time     int64             `json:"time,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Entries  container         `json:"entries,omitempty"`
}

// func NewEntry(name, password string) (e *Entry) {
// 	e = &Entry{Name: name, Password: password}
// 	e.Touch()
// 	return
// }

func (e *Entry) Age() time.Duration {
	t := time.Unix(e.Time, 0)
	return time.Since(t)
}

func (e *Entry) Touch() {
	e.Time = time.Now().Unix()
}

// func (e *Entry) Id() {
// 	e.Time = time.Now().Unix()
// }

func (e *Entry) String() string {
	months := e.Age().Seconds() / monthDuration
	return fmt.Sprintf("%s => %s (%.1f months)", e.Name, e.Password, months)
}

// Determines if entry is empty. Empty entries are used as placeholders that
// contain non-empty sub-entries.
// TODO: do Name and Time need to be empty?
// func (e *Entry) IsEmpty() bool {
// 	return len(e.Password) == 0
// }

func (e *Entry) IsContainer() bool {
	return e.Entries != nil
}
