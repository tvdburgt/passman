package main

import (
	_ "errors"
	_ "flag"
	"fmt"
	_ "github.com/atotto/clipboard"
	_ "os"
	"github.com/tvdburgt/passman/clipboard"
	"time"
)

func cmdClip() (err error) {
	if err = clipboard.Setup(); err != nil {
		return
	}

	// TODO: run event loop in goroutine
	// TODO: add verbose flag
	if err = clipboard.Put([]byte("dit is een testje\n")); err != nil {
		return
	}

	time.Sleep(5 * time.Second)

	fmt.Println("Done sleeping; exiting...")

	return


	// var store *Store
	// var entry *Entry
	// var ok bool

	// // fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	// // fs.Parse(os.Args[2:])
	// if len(os.Args) < 3 {
	// 	return errors.New("missing id argument")
	// }
	// id := os.Args[2]

	// passphrase := readPass("Enter passphrase for '%s'", storePath)
	// defer clear(passphrase)
	// if store, err = openStore(passphrase); err != nil {
	// 	return
	// }

	// if entry, ok = store.Entries[id]; !ok {
	// 	return fmt.Errorf("no such entry '%s'", id)
	// }

	// if err = clipboard.WriteAll(entry.Password); err != nil {
	// 	return
	// }

	// fmt.Println("Copied password to clipboard")
	// fmt.Println(entry.Password)

	// return
}
