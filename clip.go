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

var cmdClip = &Command{
	Run: runClip,
	UsageLine: "clip identifier",
	Short: "display an individual entry",
	Long: `
get displays the content of a single passman entry. The identifier must be an
exact match. To search or display multiple entries, see passman list.
	`,
}

func runClip(cmd *Command, args []string) {
	// if len(args) < 1 {
	// 	fatalf("passman clip: missing identifier")
	// }
	// id := args[0]

	if err := clipboard.Setup(); err != nil {
		fatalf("passman clip: %s", err)
	}

	// TODO: run event loop in goroutine
	// TODO: add verbose flag
	err := clipboard.Put([]byte("examplepassw0rd"));
	if err != nil {
		fatalf("passman clip: %s", err)
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
