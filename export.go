package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/store"
	"path/filepath"
)

// Writes a JSON-formatted output of the password store to stdout or file.
func cmdExport() (err error) {
	var file string
	var store *store.Store
	var out *os.File = os.Stdout

	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	fs.StringVar(&file, "file", "", "file to write export data to")
	fs.StringVar(&file, "f", "", "file to write export data to")
	fs.Parse(os.Args[2:])

	passphrase := readPass("Enter passphrase for '%s'", storePath)
	defer crypto.Clear(passphrase)
	if store, err = openStore(passphrase); err != nil {
		return
	}

	if len(file) > 0 {
		out, err = os.Create(file)
		if err != nil {
			return err
		}
		defer out.Close()
	}

	if err = store.Export(out); err != nil {
		return
	}

	if out != os.Stdout {
		if path, err := filepath.Abs(out.Name()); err != nil {
			return err
		} else {
			fmt.Printf("Created export file at '%s'\n", path)
		}
	}

	return
}
