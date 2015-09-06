package main

import (
	"fmt"
	"github.com/tvdburgt/passman/crypto"
)

var cmdDelete = &Command{
	Run:       runDelete,
	UsageLine: "delete id",
	Short:     "display an individual entry",
	Long: `
get displays the content of a single passman entry. The identifier must be an
exact match. To search or display multiple entries, see passman list.
	`,
}

func runDelete(cmd *Command, args []string) {
	if len(args) < 1 {
		fatalf("passman delete: missing identifier")
	}
	id := args[0]

	s, passphrase := openRwStore()
	defer crypto.Clear(passphrase)

	// Check existance before deleting
	if _, ok := s.Entries[id]; !ok {
		fatalf("passman delete: no such entry %q", id)
	}

	delete(s.Entries, id)

	writeStore(s, passphrase)
	fmt.Printf("Removed entry %q from store\n", id)
}
