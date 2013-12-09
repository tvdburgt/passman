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
func GenerateRandomSalt() ([]byte, error) {
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

func initCTRStream(key []byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, block.BlockSize()) // IV is always {0, 0, ...}
	return cipher.NewCTR(block, iv), nil
}

func InitStreamParams(passphrase, salt []byte) (stream cipher.Stream, mac hash.Hash, err error) {
	cipherKey, hmacKey, err := deriveKeys(passphrase, salt)
	if err != nil {
		return
	}
	defer Clear(cipherKey, hmacKey)

	if stream, err = initCTRStream(cipherKey); err != nil {
		return
	}

	mac = hmac.New(macHash, hmacKey)
	return
}

// Clears sensitive data from memory (useful for plaintext passwords etc.)
// https://groups.google.com/forum/?fromgroups=#!topic/golang-nuts/sKQtvluD_So
// https://groups.google.com/forum/#!msg/golang-nuts/KvgjNbCXTY4/uigWOtc6bJcJ
func Clear(s ...[]byte) {
	for _, secret := range s {
		for i := range secret {
			secret[i] = 0
		}
	}
}
