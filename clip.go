package main

import (
	"fmt"
	"github.com/tvdburgt/passman/clipboard"
	"github.com/tvdburgt/passman/store"
	"strings"
	"time"
)

var cmdClip = &Command{
	UsageLine: "clip identifier",
	Short:     "display an individual entry",
	Long: `
get displays the content of a single passman entry. The identifier must be an
exact match. To search or display multiple entries, see passman list.
	`,
}

const clipTimeoutDefault = 20 * time.Second

var (
	clipTimeout = clipTimeoutDefault
	clipPersist = false
	clipValue   = "password"
	// TODO: -verbose?
)

func init() {
	cmdClip.Run = runClip
	cmdClip.Flag.DurationVar(&clipTimeout, "timeout", clipTimeout, "")
	cmdClip.Flag.BoolVar(&clipPersist, "persist", clipPersist, "")
	cmdClip.Flag.StringVar(&clipValue, "value", clipValue, "")
}

func runClip(cmd *Command, args []string) {
	if len(args) < 1 {
		fatalf("passman clip: missing identifier")
	}
	id := args[0]

	if clipTimeout <= 0 {
		fatalf("passman clip: nonpositive timeout")
	}

	s, err := readPassStore()
	if err != nil {
		fatalf("passman clip: %s", err)
	}
	e, ok := s.Entries[id]
	if !ok {
		fatalf("passman get: no such entry '%s'", id)
	}

	value := getValue(e)

	if err := clipboard.Setup(); err != nil {
		fatalf("passman clip: %s", err)
	}
	creq, cerr, err := clipboard.Put(value)
	if err != nil {
		fatalf("passman clip: %s", err)
	}

	fmt.Println("Listening for clipboard requests...")

	for {
		select {
		case req := <-creq:
			fmt.Printf("Password requested by '%s'\n", req)
			if !clipPersist {
				fmt.Println("Exiting... Add '-persist' flag to allow additional requests.")
				return
			}
		case err := <-cerr:
			fatalf("passman clip: %s", err)
		case <-time.After(clipTimeout):
			fmt.Printf("Reached timeout (%s), exiting.", clipTimeout)
			if clipTimeout == clipTimeoutDefault {
				fmt.Print(" Duration can be changed with the '-timeout' flag.")
			}
			fmt.Println()
			return
		}
	}
}

func validValues(e *store.Entry) (values []string) {
	values = make([]string, 2, len(e.Metadata)+2)
	values[0] = "'password'"
	values[1] = "'name'"
	for key, _ := range e.Metadata {
		values = append(values, fmt.Sprintf("'%s'", key))
	}
	return
}

func getValue(e *store.Entry) []byte {
	switch clipValue {
	case "password":
		return e.Password
	case "name":
		return []byte(e.Name)
	default:
		if value, ok := e.Metadata[clipValue]; ok {
			return []byte(value)
		}
		// TODO: fatalf here or bubble error to runClip?
		fatalf("passman clip: invalid value '%s' (valid values: %s)",
			clipValue, strings.Join(validValues(e), ", "))
	}
	return nil
}
