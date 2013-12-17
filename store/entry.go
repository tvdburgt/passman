package store

import (
	"fmt"
	"time"
)

type Entry struct {
	Name     string            `json:"name,omitempty"`
	Password []byte            `json:"password,omitempty"`
	Time     int64             `json:"time,omitempty"`
	// Entropy  int             `json:"entropy,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
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

func (e *Entry) String() string {
	months := e.Age().Seconds() / monthDuration
	return fmt.Sprintf("%s => %s (%.1f months)", e.Name, e.Password, months)
}

