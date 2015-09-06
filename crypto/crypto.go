package crypto

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/tvdburgt/passman/store"
	"golang.org/x/crypto/scrypt"
	"hash"
	"io"
)

const (
	keySize  = 32            // 256-bit key for AES-256
	hashFunc = crypto.SHA256 // Use SHA-256 as HMAC hash function
)

var ErrWrongPass = errors.New("incorrect passphrase")

// WriteStore encrypts and writes a password store object to an output stream.
func WriteStore(out io.Writer, s *store.Store, passphrase []byte) (err error) {
	// Initialize the stream (AES-CTR stream and HMAC)
	stream, mac := initStream(passphrase, &s.Header)

	// Redirect cipher stream to output (and HMAC writer)
	ct := cipher.StreamWriter{S: stream, W: io.MultiWriter(out, mac)}

	// Combine plaintext with HMAC writer
	pt := io.MultiWriter(out, mac)

	// Write header (plaintext)
	if err = s.Header.Marshal(pt); err != nil {
		return
	}

	// Write header HMAC (plaintext)
	if _, err = pt.Write(mac.Sum(nil)); err != nil {
		return
	}

	// Write JSON-encoded entries (ciphertext)
	enc := json.NewEncoder(ct)
	if err = enc.Encode(s.Entries); err != nil {
		return
	}

	// Write file HMAC (plaintext)
	if _, err = out.Write(mac.Sum(nil)); err != nil {
		return
	}

	return
}

// ReadStore decrypts an input stream and returns a constructed password store
// object.
func ReadStore(in io.Reader, passphrase []byte) (s *store.Store, err error) {
	s = store.NewStore()
	buf := new(bytes.Buffer)

	// Marshal header and redirect data to buffer
	r := io.TeeReader(in, buf)
	if err = s.Header.Unmarshal(r); err != nil {
		return nil, err
	}

	// Initialize crypto stream and HMAC
	stream, mac := initStream(passphrase, &s.Header)

	// Update HMAC by feeding previously read header
	io.Copy(mac, buf)

	// Check header HMAC
	if ok, err := checkHMAC(io.TeeReader(in, mac), mac); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrWrongPass
	}

	// Copy remainder of input stream to buffer
	if _, err = io.Copy(buf, in); err != nil {
		return nil, err
	}

	// Construct ciphertext reader for entry data block
	n := int64(buf.Len() - mac.Size())
	r = io.LimitReader(io.TeeReader(buf, mac), n)
	r = cipher.StreamReader{S: stream, R: r}

	// Read JSON-encoded entries
	dec := json.NewDecoder(r)
	if err := dec.Decode(&s.Entries); err != nil {
		return nil, err
	}

	// Check store HMAC
	if ok, err := checkHMAC(buf, mac); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrWrongPass
	}

	return
}

// Clear removes sensitive data from memory (useful for plaintext passwords
// etc.).
func Clear(secret []byte) {
	for i := range secret {
		secret[i] = 0 // Clear each byte
	}
	secret = nil // Reset data slice
}

// ReadRand fills b with entropy taken from the CSPRNG in crypto/rand.
func ReadRand(b []byte) error {
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return err
	}
	return nil
}

// checkHMAC reads HMAC bytes from r and performs a constant-time comparison
// with mac.
func checkHMAC(r io.Reader, mac hash.Hash) (ok bool, err error) {
	mac1 := mac.Sum(nil)             // Calculated HMAC
	mac2 := make([]byte, mac.Size()) // Read HMAC
	if _, err = io.ReadFull(r, mac2); err != nil {
		return false, err
	}
	return hmac.Equal(mac1, mac2), nil
}

// deriveKeys returns a set of keys that is derived from the entropy in
// passphrase and salt. Both resulting key lengths are 32 bytes (AES-256 key
// length and SHA-256 hash size respectively).
func deriveKeys(passphrase, salt []byte, logN, r, p int) (cipherKey, hmacKey []byte) {
	keyLen := keySize + hashFunc.Size()
	key, err := scrypt.Key(passphrase, salt, 1<<uint(logN), r, p, keyLen)
	if err != nil {
		panic(err)
	}
	cipherKey, hmacKey = key[:keySize], key[keySize:]
	return
}

func getStream(key []byte) cipher.Stream {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err) // KeySizeError
	}
	iv := make([]byte, block.BlockSize()) // IV is always {0, 0, ...}
	return cipher.NewCTR(block, iv)
}

func initStream(passphrase []byte, h *store.Header) (stream cipher.Stream, mac hash.Hash) {
	salt := h.Salt[:]
	logN, r, p := int(h.Params.LogN), int(h.Params.R), int(h.Params.P)
	cipherKey, hmacKey := deriveKeys(passphrase, salt, logN, r, p)
	defer Clear(cipherKey)
	defer Clear(hmacKey)
	stream = getStream(cipherKey)
	mac = hmac.New(hashFunc.New, hmacKey)
	return
}
