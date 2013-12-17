package main

import (
	"fmt"
	"github.com/tvdburgt/passman/crypto"
)

var cmdRm = &Command{
	Run: runRm,
	UsageLine: "rm <identifier>",
	Short: "display an individual entry",
	Long: `
get displays the content of a single passman entry. The identifier must be an
exact match. To search or display multiple entries, see passman list.
	`,
}

func runRm(cmd *Command, args []string) {
	if len(args) < 1 {
		fatalf("passman rm: missing identifier")
	}
	id := args[0]

	// TODO: create wrapper fn
	passphrase := readPass("Enter passphrase for '%s'", storePath)
	defer crypto.Clear(passphrase)
	s, err := readStore(passphrase)
	if err != nil {
		return
	}

	// Check existance before deleting
	if _, ok := s.Entries[id]; !ok {
		fatalf("passman get: no such entry %q", id)
	}
	
	delete(s.Entries, id)

	if err = writeStore(s, passphrase); err != nil {
		fatalf("passman rm: %s", err)
	}
	fmt.Printf("Removed entry %q from store\n", id)
}
