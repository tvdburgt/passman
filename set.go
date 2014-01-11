package main

import (
	"errors"
	"fmt"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/store"
	"strings"
)

var cmdSet = &Command{
	UsageLine: "set [options] <id>",
	Short:     "create or modify a passman entry",
	Long: `
long description

	-n -name <name>		set name
	-p -password		prompt for password
	-id <identifier>	change id of existing entry
	`,
}

var (
	setName     string
	setId       string
	setPassword = false
	setMeta     = make(metadata)
)

func init() {
	cmdSet.Run = runSet
	cmdSet.Flag.StringVar(&setName, "name", "", "")
	cmdSet.Flag.BoolVar(&setPassword, "password", setPassword, "")
	cmdSet.Flag.StringVar(&setId, "id", "", "")
	cmdSet.Flag.Var(setMeta, "meta", "")
}

type metadata store.Metadata

// This only gets called when using the default flag.Usage.
// Only required for satisfying the the flag.Value interface.
func (m metadata) String() string {
	return fmt.Sprintf("%#v", m)
}

func (m metadata) Set(data string) (err error) {
	pair := strings.Split(data, "=")
	if len(pair) == 1 {
		return errors.New("invalid key-value format")
	}
	// Use first '=' symbol as separator, concatenate the rest if needed
	key, val := pair[0], strings.Join(pair[1:], "")
	m[key] = val
	return
}

func runSet(cmd *Command, args []string) {
	if len(args) < 1 {
		fatalf("passman set: missing id")
	}
	id := args[0]

	passphrase := readPass("Enter passphrase for '%s'", storeFile)
	defer crypto.Clear(passphrase)
	s, err := decryptStore(passphrase)
	if err != nil {
		fatalf("passman set: %s", err)
	}

	// Fetch entry
	e, ok := s.Entries[id]
	if !ok {
		fmt.Printf("Entry %q doesn't exist, creating...\n", id)
		e = store.NewEntry()
		s.Entries[id] = e
		setPassword = true // Always prompt for password for new entries
	} else {
		fmt.Printf("Found entry %q\n", id)
		if cmd.Flag.NFlag() == 0 {
			fatalf("passman set: no arguments to set for %q", id)
		}
	}

	if setName != "" {
		e.Name = setName
	}

	if setId != "" {
		if _, ok := s.Entries[setId]; ok {
			fatalf("passman set: entry with id %q already exists", setId)
		}
		s.Entries[setId] = e
		delete(s.Entries, id)
	}

	// Merge metadata modifications with entry
	for key, val := range setMeta {
		if len(val) == 0 {
			delete(e.Metadata, key)
		} else {
			e.Metadata[key] = val
		}
	}

	if setPassword {
		if password, err := readPassword(); err != nil {
			fatalf("passman set: %s", err)
		} else {
			// TODO: clear password, use []byte
			e.Password = password
		}
		// Update modification time
		e.Touch()
	}

	writeStore(s, passphrase)

	fmt.Print(e)
}
