package main

import (
	"code.google.com/p/go.crypto/scrypt"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	_ "fmt"
	"io"
)

const keySize = 32

// Generates a cryptocraphically secure salt, with length equal to key size.
func generateRandomSalt() ([]byte, error) {
	salt := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

func deriveKeys(passphrase, salt []byte) (cipherKey, hmacKey []byte, err error) {
	// Use sensible work factor defaults for now...
	key, err := scrypt.Key(passphrase, salt, 16384, 8, 1, 64) // keySize*2

	if err == nil {
		cipherKey, hmacKey = key[:keySize], key[keySize:]
	}

	return
}

func initStreamCipher(key []byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, block.BlockSize()) // IV is always {0, 0, ...}
	return cipher.NewCTR(block, iv), nil
}

// https://groups.google.com/forum/?fromgroups=#!topic/golang-nuts/sKQtvluD_So
// https://groups.google.com/forum/#!msg/golang-nuts/KvgjNbCXTY4/uigWOtc6bJcJ
func clear(s ...[]byte) {
	for _, secret := range s {
		/* fmt.Printf("clearing '%s'\n", secret) */
		for i := range secret {
			secret[i] = 0
		}
	}
}
