package crypto

import (
	"code.google.com/p/go.crypto/scrypt"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"hash"
	"io"
)

const keySize = 32       // 256-bit key for AES-256
var macHash = sha256.New // Use SHA-256 as HMAC

// Generates a cryptographically secure salt with length equal to key size.
func GenerateRandomSalt() (salt []byte, err error) {
	salt = make([]byte, keySize)
	if _, err = io.ReadFull(rand.Reader, salt); err != nil {
		return
	}
	return
}

func deriveKeys(passphrase, salt []byte) (cipherKey, hmacKey []byte) {
	// Use sensible work factor defaults for now...
	key, err := scrypt.Key(passphrase, salt, 16384, 8, 1, 64) // keySize*2
	if err != nil {
		panic(err) // scrypt params should be guarded elsewhere
	}
	if err == nil {
		cipherKey, hmacKey = key[:keySize], key[keySize:]
	}
	return
}

func initCTRStream(key []byte) cipher.Stream {
	block, err := aes.NewCipher(key)
	// TODO: internal state of block remains in memory?
	if err != nil {
		panic(err) // KeySizeError
	}
	iv := make([]byte, block.BlockSize()) // IV is always {0, 0, ...}
	return cipher.NewCTR(block, iv)
}

func InitStreamParams(passphrase, salt []byte) (stream cipher.Stream, mac hash.Hash) {
	cipherKey, hmacKey := deriveKeys(passphrase, salt)
	defer Clear(cipherKey, hmacKey)
	stream = initCTRStream(cipherKey)
	mac = hmac.New(macHash, hmacKey)
	return
}

// Clears sensitive data from memory (useful for plaintext passwords etc.)
func Clear(s ...[]byte) {
	for _, secret := range s {
		for i := range secret {
			secret[i] = 0
		}
	}
}
