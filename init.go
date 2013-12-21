package main

import (
	"fmt"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/store"
	"os"
)

var cmdInit = &Command{
	Run:       runInit,
	UsageLine: "init [-f file]",
	Short:     "export passman store",
	Long: `
JSON-formatted, defaults to stdout.
	`,
}

func init() {
	addStoreFlags(cmdInit)
}

func runInit(cmd *Command, args []string) {
	// Read file and make sure it doesn't exist
	if _, err := os.Stat(storeFile); err == nil {
		fatalf("passman init: '%s' already exists", storeFile)
	}

	// Read passphrase
	passphrase := readVerifiedPass()
	defer crypto.Clear(passphrase)

	header := store.NewHeader()
	s := store.NewStore(header) // create default ctor with header defaults?

	err := writeStore(s, passphrase)
	if err != nil {
		fatalf("passman init: %s", err)
	}

	fmt.Printf("Initialized empty passman store: '%s'\n", storeFile)
}
