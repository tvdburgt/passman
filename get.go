package main

import (
	"fmt"
)

var cmdGet = &Command{
	Run:       runGet,
	UsageLine: "get entry_id",
	Short:     "show a single entry",
	Long: `
This command shows all fields and corresponding values that belong to an
individual entry, specified by the entry_id argument. To show multiple entries,
see 'passman list'.
	`,
}

func init() {
	addFileFlag(cmdGet)
}

func runGet(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Usage()
	}
	id := args[0]
	s := openStore()
	e, ok := s.Entries[id]
	if !ok {
		fatalf("Entry %q does not exist.", id)
	}
	fmt.Print(e)
}
