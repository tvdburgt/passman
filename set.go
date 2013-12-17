package main

import (
	"fmt"
	"github.com/tvdburgt/passgen"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/store"
	"strconv"
	"unicode"
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
	defaultLength = 64
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

	passphrase := readPass("Enter passphrase for '%s'", storePath)
	defer crypto.Clear(passphrase)
	s, err := readStore(passphrase)
	if err != nil {
		return
	}

	// Fetch entry
	e, existing := s.Entries[id]

	// New entry; create one
	if !existing {
		fmt.Println("Entry %q doesn't exist, creating...", id)
		e = new(store.Entry)
		s.Entries[id] = e
	} else {
		fmt.Println("Found entry %q", id)
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
		if password, err := getPassword(); err != nil {
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

// passman set id [-n name] [-p]
// TODO: metadata
// func runSet(cmd *Command, args []string) {
// 	var flagName string
// 	var flagPassword bool
// 	var flagId string

// 	if len(args) < 1 {
// 		fatalf("passman set: missing identifier")
// 	}
// 	id := args[0]

// 	// TODO: -id (change id)
// 	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
// 	fs.StringVar(&flagName, "name", "", "associated name")
// 	fs.StringVar(&flagName, "n", "", "associated name (shorthand)")
// 	fs.BoolVar(&flagPassword, "password", false, "set password")
// 	fs.BoolVar(&flagPassword, "p", false, "set password")
// 	fs.StringVar(&flagId, "id", "", "set id")

// 	fs.Parse(os.Args[3:])

// 	// fmt.Println(fs.NFlag())

// 	passphrase := readPass("Enter passphrase for '%s'", storePath)
// 	defer crypto.Clear(passphrase)
// 	s, err := openStore(passphrase)
// 	if err != nil {
// 		return
// 	}

// 	// Fetch entry
// 	e, existing := s.Entries[id]

// 	// New entry; create one
// 	if !existing {
// 		fmt.Println("Entry %q doesn't exist, creating...", id)
// 		e = new(store.Entry)
// 		s.Entries[id] = e
// 	} else {
// 		fmt.Println("Found entry %q", id)
// 	}

// 	if existing && fs.NFlag() == 0 {
// 		fatalf("passman set: no arguments to set for %q", id)
// 	}

// 	if len(flagName) > 0 {
// 		e.Name = flagName
// 	}

// 	if len(flagId) > 0 {
// 		if _, ok := s.Entries[flagId]; ok {
// 			fatalf("passman set: entry with id %q already exists", flagId)
// 		}
// 		s.Entries[flagId] = e
// 		delete(s.Entries, id)
// 	}

// 	if !existing || flagPassword {
// 		if password, err := getPassword(); err != nil {
// 			fatalf("passman set: %s", err)
// 		} else {
// 			// TODO: clear password, use []byte
// 			e.Password = password
// 		}
// 		// Update modification time
// 		e.Touch()
// 	}

// 	err = saveStore(s, passphrase)
// 	if err != nil {
// 		return
// 	}

// 	fmt.Println(e)
// 	return
// }

// Helper method for reading numbers from stdin; uses default value (def) if
// input is empty. An error is returned if Atoi can't parse input.
func scanNumber(def int) (n int, err error) {
	var s string
	n = def

	if _, err = fmt.Scanln(&s); err == nil {
		n, err = strconv.Atoi(s)
	} else if len(s) == 0 {
		err = nil
	}

	return
}

func getPassword() ([]byte, error) {

	var method passMethod

	fmt.Printf("\n"+`Password generation methods:
  [%d] manual
  [%d] ascii
  [%d] hex
  [%d] base32
  [%d] diceware`+"\n\n", methodManual, methodAscii, methodHex, methodBase32, methodDiceware)

	for {
		for {
			fmt.Printf("Select method [%d]: ", defaultMethod)
			if n, err := scanNumber(int(defaultMethod)); err == nil {
				method = passMethod(n)
				break
			}
		}

		switch method {
		case methodManual:
			return readVerifiedPass(), nil
		case methodAscii, methodHex, methodBase32:
			return generatePassword(method)
		case methodDiceware:
			panic("not yet implemented :(")
		default:
			// fmt.Fprintln(os.Stderr, "error: invalid method")
		}
	}

	return nil, nil
}

func generatePassword(method passMethod) (password []byte, err error) {
	var length int

	for {
		fmt.Printf("Password length [%d]: ", defaultLength)
		if length, err = scanNumber(defaultLength); err == nil {
			break
		}
	}

	for {
		var m int
		switch method {
		case methodAscii:
			password, err = passgen.Ascii(length, passgen.SetComplete)
			m = passgen.SetComplete.Cardinality()
		case methodHex:
			password, err = passgen.Hex(length)
			m = 16
		case methodBase32:
			password, err = passgen.Base32(length)
			m = 32
		}

		if err != nil {
			return
		}

		fmt.Printf("m=%d\n", m)

		fmt.Println("Generated password:\n")
		fmt.Printf("\t%s (%.2f bits)\n\n",
			password, passgen.Entropy(length, m))

	accept:
		for {
			fmt.Print("Accept? [Y/n] ")
			action := "y"
			fmt.Scanln(&action)
			r := rune(action[0])

			switch unicode.ToLower(r) {
			case 'y':
				return
			case 'n':
				break accept
			}
		}
	}
}
