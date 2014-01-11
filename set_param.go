package main

import (
	"fmt"
	"github.com/tvdburgt/passman/crypto"
	"time"
)

var cmdSetParam = &Command{
	Run:       runSetParam,
	UsageLine: "set-param",
	Short:     "shows status information about the store file",
	Long: `
get displays the content of a single passman entry. The identifier must be an
exact match. To search or display multiple entries, see passman list.
	`,
}

func runSetParam(cmd *Command, args []string) {
	if len(args) < 2 {
		// TODO: show usage
		fatalf("passman set-param: missing identifier")
	}
	name, value := args[0], args[1]

	passphrase := readPass("Enter passphrase for '%s'", storeFile)
	defer crypto.Clear(passphrase)
	s, err := decryptStore(passphrase)
	if err != nil {
		fatalf("passman set-param: %s", err)
	}

	p := &s.Header.Params
	params := map[string]interface{}{
		"n": &p.LogN,
		"r": &p.R,
		"p": &p.P,
	}

	field, ok := params[name]
	if !ok {
		fatalf("passman set-param: unknown parameter name '%s'", name)
	}

	// TODO: type conversion from string to byte/uint32?
	field = value
	_ = field // Go complains about unused variable

	fmt.Println("Verifying parameter...")

	// Test new parameter before finalizing the change
	defer func() {
		if err := recover(); err != nil {
			fatalf("passman set-param:", err)
		}
	}()
	before := time.Now()
	crypto.DeriveKeys([]byte{}, []byte{}, int(p.LogN), int(p.R), int(p.P))
	after := time.Now()

	d := before.Sub(after)
	fmt.Println("Key derivation took", d)

	writeStore(s, passphrase)

	fmt.Printf("Changed '%s' to '%s'\n", name, value)
}
