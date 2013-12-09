package store

import (
	"fmt"
	"time"
)

// TODO: create ctor that sets time
type Entry struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Time     int64  `json:"time"`
	// Metadata map[string]string
}

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
