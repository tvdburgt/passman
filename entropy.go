package main

import (
	"math"
	"unicode"
)

func cardinality(password []byte) int {

	var (
		lower,
		upper,
		digit,
		punct,
		symbol,
		space bool
	)

	total := 0

	for _, c := range password {
		r := rune(c)

		if !lower && unicode.IsLower(r) {
			total += 26
		} else if !upper && unicode.IsUpper(r) {
			total += 26
		} else if !digit && unicode.IsDigit(r) {
			total += 10
		} else if !punct && unicode.IsPunct(r) {
			total += 23
		} else if !symbol && unicode.IsSymbol(r) {
			total += 9
		} else if !space && unicode.IsSpace(r) {
			total += 1
		}
	}

	return total
}

// Password entropy approximation in bits
func entropy(password []byte) float64 {
	n := float64(cardinality(password)) // Number of possible symbols
	l := float64(len(password))         // Password length
	return l * math.Log2(n)
}
