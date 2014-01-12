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
		cmd.Usage()
	}
	id := args[0]
	s := openStore()
	e, ok := s.Entries[id]
	if !ok {
		fatalf("Entry '%s' does not exist", id)
	}
	fmt.Print(e)
}

