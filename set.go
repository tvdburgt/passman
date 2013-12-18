package main

import (
	"fmt"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/store"
)

var cmdSet = &Command{
	UsageLine: "set [options] <id>",
	Short: "create or modify a passman entry",
	Long: `
long description

	-n -name <name>		set name
	-p -password		prompt for password
	-id <identifier>	change id of existing entry
	`,
}

func init() {
	cmdSet.Run = runSet
	cmdSet.Flag.StringVar(&setName, "n", "", "")
	cmdSet.Flag.StringVar(&setName, "name", "", "")
	cmdSet.Flag.BoolVar(&setPassword, "p", false, "")
	cmdSet.Flag.BoolVar(&setPassword, "password", false, "")
	cmdSet.Flag.StringVar(&setId, "id", "", "")
}

type passMethod int

const (
	methodManual passMethod = iota + 1
	methodAscii
	methodHex
	methodBase32
	methodDiceware
)

const (
	defaultMethod = methodAscii
	defaultLen = 64
	defaultWordCount = 6
)

var (
	setName string
	setPassword bool
	setId string
)

func runSet(cmd *Command, args []string) {
	if len(args) < 1 {
		fatalf("passman set: missing id")
	}
	id := args[0]

	passphrase := readPass("Enter passphrase for '%s'", storeFile)
	defer crypto.Clear(passphrase)
	s, err := readStore(passphrase)
	if err != nil {
		return
	}

	// Fetch entry
	e, existing := s.Entries[id]

	// New entry; create one
	if !existing {
		fmt.Printf("Entry %q doesn't exist, creating...\n", id)
		e = new(store.Entry)
		s.Entries[id] = e
	} else {
		fmt.Printf("Found entry %q\n", id)
	}

	if existing && cmd.Flag.NFlag() == 0 {
		fatalf("passman set: no arguments to set for %q", id)
	}

	if len(setName) > 0 {
		e.Name = setName
	}

	if len(setId) > 0 {
		if _, ok := s.Entries[setId]; ok {
			fatalf("passman set: entry with id %q already exists", setId)
		}
		s.Entries[setId] = e
		delete(s.Entries, id)
	}

	if !existing || setPassword {
		if password, err := readPassword(); err != nil {
			fatalf("passman set: %s", err)
		} else {
			// TODO: clear password, use []byte
			e.Password = password
		}
		// Update modification time
		e.Touch()
	}

	err = writeStore(s, passphrase)
	if err != nil {
		return
	}

	fmt.Println(e)
}
