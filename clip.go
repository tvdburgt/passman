package main

import (
	"fmt"
	"github.com/tvdburgt/passman/clipboard"
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
	clipTimeout time.Duration
	clipPersist bool
	// TODO: verbose?
)

func init() {
	cmdClip.Run = runClip
	cmdClip.Flag.DurationVar(&clipTimeout, "timeout", clipTimeoutDefault, "")
	cmdClip.Flag.BoolVar(&clipPersist, "persist", false, "")
}

func runClip(cmd *Command, args []string) {
	if len(args) < 1 {
		fatalf("passman clip: missing identifier")
	}
	id := args[0]

	if clipTimeout <= 0 {
		fatalf("passman clip: negative timeout")
	}

	s, err := readPassStore();
	if err != nil {
		fatalf("passman clip: %s", err)
	}
	e, ok := s.Entries[id]
	if !ok {
		fatalf("passman get: no such entry '%s'", id)
	}

	if err := clipboard.Setup(); err != nil {
		fatalf("passman clip: %s", err)
	}
	creq, cerr, err := clipboard.Put(e.Password)
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
