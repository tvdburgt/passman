


package main

import (
	"fmt"
	"encoding/json"
	"os"
	"crypto/aes"
	"crypto/rand"
	/* "crypto/sha512" */
	"crypto/cipher"
	"code.google.com/p/go.crypto/scrypt"
	"io"
)

/*
Offset		Size		Description
---------------------------------------------------------
0		32		Salt
32		7		File signature
39		1		File format version
40		x		Data
40+x		32		HMAC-SHA256(0 .. x-32)
*/

const (
	keySize = 32
	fileSignature = "passman" // 0x706173736d616e
	fileVersion uint8 = 0
)

/* type Header struct { */
/* 	Salt string */
/* 	Key string */
/* 	Checksum string // TODO: crc, md5 or sha? */
/* 	entryCount string */
/* } */

type Entry struct {
	/* Id string */
	Name string
	Password string
}

// return []byte, or use ref param
// handle error in method, or return error?
func getSalt() ([]byte, error) {
	salt := make([]byte, keySize)
	n, err := io.ReadFull(rand.Reader, salt)

	// Reader.Read always returns an error when insufficient bytes are
	// available?
	if n != keySize || err != nil {
		return nil, err
	}

	return salt, nil
}

func deriveKey(passphrase, salt []byte) (encKey, macKey []byte, err error) {
	key, err := scrypt.Key(passphrase, salt, 16384, 8, 1, 64)
	encKey = key[:keySize]
	macKey = key[keySize:]
	return
}

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "missing file param")
		return
	}

	filename := os.Args[1]

	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer f.Close()

	salt, err := getSalt()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	f.Write(salt)

	var pass []byte
	/* fmt.Print("passphrase: ") */
	/* fmt.Scan(&pass) */
	pass = []byte("testje")

	/* fmt.Printf("%s %d : %#x\nenc %d : %x\nmac %d : %x\n", */
	/* pass, len(key), key, */
	/* len(encKey), encKey, */
	/* len(macKey), macKey) */

	encKey, _, err := deriveKey(pass, salt)

	block, _ := aes.NewCipher(encKey)
	iv := make([]byte, block.BlockSize())
	stream := cipher.NewCTR(block, iv)

	writer := &cipher.StreamWriter{stream, f, nil} // add err?

	writer.Write([]byte(fileSignature))
	writer.Write([]byte{fileVersion})

	entries := make(map[string]Entry)

	entries["gmail"] = Entry{"foo@gmail.com", "somecrappypassword"}
	entries["tweakers.net"] = Entry{"noobie", "anothershittypassword"}

	content, _ := json.Marshal(entries)

	fmt.Printf("%s\n", content)
	writer.Write(content)
}
