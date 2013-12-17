package main

import (
	"flag"
	"fmt"
	"github.com/tvdburgt/passgen"
	"os"
)

func cmdGen() (err error) {

	// /usr/share/dict/words can be changed by symlink or with
	// 'select-default-wordlist' on Debian

	fmt.Println("generating pass")
	fmt.Println(passgen.DicewareDict)


	for i := 0; i < 10; i++ {
		password, m, _ := passgen.Diceware(6, "")
		fmt.Printf("%s (%.2f bits)\n", password, passgen.Entropy(6, m))
	}

	return


	// TODO: clear()
}

func gen() {

	size := flag.Uint("len", 64, "password length")
	n := flag.Uint("n", 1, "number of passwords")

	lower := flag.Bool("lower", false, "lower case characters [a-z]")
	upper := flag.Bool("upper", false, "upper case characters [A-Z]")
	digit := flag.Bool("digit", false, "digits [0-9]")

	os.Args = os.Args[1:]
	flag.Parse()

	charSet := buildCharSet(*lower, *upper, *digit)
	var i uint

	for i = 0; i < *n; i++ {
		pass, err := passgen.Ascii(int(*size), charSet)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%s\n", pass)
	}

}

func buildCharSet(lower, upper, digit bool) passgen.CharSet {
	var set passgen.CharSet = 0

	if lower {
		set |= passgen.SetLower
	}
	if upper {
		set |= passgen.SetUpper
	}
	if digit {
		set |= passgen.SetDigit
	}

	if set == 0 {
		return passgen.SetComplete
	} else {
		return set
	}
}
