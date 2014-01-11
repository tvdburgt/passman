package crypto

import (
	"code.google.com/p/go.crypto/scrypt"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	_ "crypto/sha256"
	"hash"
	"io"
)

const (
	KeySize = 32            // 256-bit key for AES-256
	HashFn  = crypto.SHA256 // Use SHA-256 as HMAC hash function
)

// Fill b with entropy taken from the CSPRNG in crypto/rand
func Rand(b []byte) error {
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		return err
	}
	return nil
}

// DeriveKeys returns a set of keys that is derived from the entropy in
// passphrase and salt. Both resulting key lengths are 32 bytes (AES-256 key
// length and SHA-256 hash size respectively).
func DeriveKeys(passphrase, salt []byte, logN, r, p int) (cipherKey, hmacKey []byte) {
	key, err := scrypt.Key(passphrase, salt, 1<<uint(logN), r, p, KeySize+HashFn.Size())
	if err != nil {
		panic(err)
	}
	cipherKey, hmacKey = key[:KeySize], key[KeySize:]
	return
}

func initCTRStream(key []byte) cipher.Stream {
	// TODO: internal state of  block remains in memory?
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err) // KeySizeError
	}
	iv := make([]byte, block.BlockSize()) // IV is always {0, 0, ...}
	return cipher.NewCTR(block, iv)
}

func InitStreamParams(passphrase, salt []byte, logN, r, p int) (stream cipher.Stream, mac hash.Hash) {
	cipherKey, hmacKey := DeriveKeys(passphrase, salt, logN, r, p)
	defer Clear(cipherKey, hmacKey)
	stream = initCTRStream(cipherKey)
	mac = hmac.New(HashFn.New, hmacKey)
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

// TODO: move to crypto or cache
func getChecksum(r io.Reader) (sum []byte, err error) {
	// file, err := os.Open(storeFile)
	// if err != nil {
	// 	return
	// }
	// defer file.Close()

	h := HashFn.New()
	if _, err = io.Copy(h, r); err != nil {
		return
	}
	sum = h.Sum(nil)
	return
}
