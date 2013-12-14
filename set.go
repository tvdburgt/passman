package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/tvdburgt/passgen"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/store"
	"os"
	"strconv"
	"unicode"
)

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

// passman set id [-n name] [-p]
// TODO: metadata
func cmdSet() (err error) {
	var flagName string
	var flagPassword bool
	var flagId string

	if len(os.Args) < 3 {
		return errors.New("missing id argument")
	}
	id := os.Args[2]

	// TODO: -id (change id)
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	fs.StringVar(&flagName, "name", "", "associated name")
	fs.StringVar(&flagName, "n", "", "associated name (shorthand)")
	fs.BoolVar(&flagPassword, "password", false, "set password")
	fs.BoolVar(&flagPassword, "p", false, "set password")
	fs.StringVar(&flagId, "id", "", "set id")

	fs.Parse(os.Args[3:])

	fmt.Println(fs.NFlag())

	passphrase := readPass("Enter passphrase for '%s'", storePath)
	defer crypto.Clear(passphrase)
	s, err := openStore(passphrase)
	if err != nil {
		return
	}

	// Fetch entry
	e, existing := s.Entries[id]

	// New entry; create one
	if !existing {
		e = new(store.Entry)
		s.Entries[id] = e
	}

	if existing && fs.NFlag() == 0 {
		return fmt.Errorf("no arguments to set for '%s'", id)
	}

	if len(flagName) > 0 {
		e.Name = flagName
	}

	if len(flagId) > 0 {
		if _, ok := s.Entries[flagId]; ok {
			return fmt.Errorf("Entry '%s' already exists", flagId)
		}
		s.Entries[flagId] = e
		delete(s.Entries, id)
	}

	if !existing || flagPassword {
		if password, err := getPassword(); err != nil {
			return err
		} else {
			e.Password = string(password)
		}
		// Update modification time
		e.Touch()
	}

	err = saveStore(s, passphrase)
	if err != nil {
		return
	}

	fmt.Println(e)
	return
}

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
		switch method {
		case methodAscii:
			password, err = passgen.Generate(length, passgen.SetComplete)
		case methodHex:
			password, err = passgen.GenerateHex(length)
		case methodBase32:
			password, err = passgen.GenerateBase32(length)
		}

		if err != nil {
			return
		}

		fmt.Println("Generated password:\n")
		fmt.Printf("\t%s\n\n", password)

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
