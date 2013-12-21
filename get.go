package main

import (
	"fmt"
)

var cmdGet = &Command{
	Run: runGet,
	UsageLine: "get identifier",
	Short: "display an individual entry",
	Long: `
get displays the content of a single passman entry. The identifier must be an
exact match. To search or display multiple entries, see passman list.
	`,
}

func runGet(cmd *Command, args []string) {
	if len(args) < 1 {
		fatalf("passman get: missing identifier")
	}
	id := args[0]

	s, err := readPassStore();
	if err != nil {
		return
	}
	defer s.Close()

	e, ok := s.Entries[id]
	if !ok {
		fatalf("passman get: no such entry %q", id)
	}

	fmt.Print(e)
}

