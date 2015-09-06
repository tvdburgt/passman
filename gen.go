package main

import (
	"fmt"
	"github.com/tvdburgt/passgen"
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
	defaultMethod       = methodAscii
	defaultLen          = 64
	defaultDicewareLen  = 6
	defaultDicewareDict = "/usr/share/dict/words"
)

var cmdGen = &Command{
	Run:       runGen,
	UsageLine: "gen",
	Short:     "generates a password",
	Long: `
long description

	-n -name <name>		set name
	-p -password		prompt for password
	-id <identifier>	change id of existing entry
	`,
}

func runGen(cmd *Command, args []string) {
	_, err := readPassword()
	if err != nil {
		fatalf("passman gen: %s", err)
	}
}

// Helper method for reading numbers from stdin; uses default value if
// input is empty. An error is returned if Atoi can't parse input.
func scanNumber(value int) (n int, err error) {
	var s string
	n = value
	if _, err = fmt.Scanln(&s); err == nil {
		n, err = strconv.Atoi(s)
	} else if len(s) == 0 {
		err = nil
	}
	return
}

// TODO: scheme
func readPassword() (password []byte, err error) {

	var method passMethod

	fmt.Printf(`Password generation methods:
  [%d] manual
  [%d] ascii
  [%d] hex
  [%d] base32
  [%d] diceware
  
`, methodManual, methodAscii, methodHex, methodBase32, methodDiceware)

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
			return readVerifiedPassphrase(), nil
		case methodAscii, methodHex, methodBase32, methodDiceware:
			password, err := generatePassword(method)
			switch {
			case err != nil:
				return nil, err
			case password == nil:
				return readPassword()
			default:
				return password, nil
			}
		}
	}
	return
}

func generatePassword(method passMethod) (password []byte, err error) {
	var n int // Password length
	var m int // Password symbol space

	for {
		switch method {
		case methodDiceware:
			n = defaultDicewareLen
			fmt.Printf("Number of words [%d]: ", n)
		default:
			n = defaultLen
			fmt.Printf("Password length [%d]: ", n)
		}
		if n, err = scanNumber(n); err == nil {
			break
		}
	}

	for {
		switch method {
		case methodAscii:
			password, err = passgen.Ascii(n, passgen.SetComplete)
			m = passgen.SetComplete.Cardinality()
		case methodHex:
			password, err = passgen.Hex(n)
			m = 16
		case methodBase32:
			password, err = passgen.Base32(n)
			m = 32
		case methodDiceware:
			// TODO: prompt for diceware dict location
			dict, err := os.Open(defaultDicewareDict)
			if err != nil {
				password, m, err = passgen.Diceware(dict, n, " ")
			}
		}

		if err != nil {
			return
		}

		fmt.Println("Generated password:\n")
		fmt.Printf("\t%q (%.2f bits)\n\n",
			password, passgen.Entropy(n, m))

	accept:
		for {
			fmt.Print("Accept? [Y/n/u/q] ")
			action := "y"
			fmt.Scanln(&action)
			r := rune(action[0])

			switch unicode.ToLower(r) {
			case 'y':
				return
			case 'n':
				break accept
			case 'u':
				return nil, nil
			case 'q':
				// Perhaps exit gracefully, so deferred functions can run...
				os.Exit(0)
			}
		}
	}
}
