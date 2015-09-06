package main

import (
	"fmt"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/store"
	"os"
)

var cmdInit = &Command{
	Run:       runInit,
	UsageLine: "init [-f <file>]",
	Short:     "create empty passman store file",
	Long: `
JSON-formatted, defaults to stdout.

  -f, -file <store-file>
	override default store file (default file location is $HOME/.pass_store
	or $PASS_STORE, if set)
	`,
}

func init() {
	addFileFlag(cmdInit)
}

func runInit(cmd *Command, args []string) {
	// Read file and make sure it doesn't exist
	if _, err := os.Stat(storeFile); err == nil {
		fatalf("passman init: '%s' already exists", storeFile)
	}
	passphrase := readVerifiedPassphrase()
	defer crypto.Clear(passphrase)
	s := store.NewStore()
	writeStore(s, passphrase)
	fmt.Printf("Initialized empty passman store at '%s'.\n", storeFile)
}
