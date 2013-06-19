// Copright (c) 2013 Tijmen van der Burgt
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/howeyc/gopass"
	"os"
	"os/user"
	"path"
	"time"
)

// TODO: clear derived keys etc.

const (
	fileVersion    = 0x00
	filePermission = 0600
)

var (
	// 0x706173736d616e == "passman"
	fileSignature    = []byte{0x70, 0x61, 0x73, 0x73, 0x6d, 0x61, 0x6e}
	defaultStorePath string
)

func init() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	defaultStorePath = path.Join(u.HomeDir, ".passstore")
}

func readPassphrase(prompt string) []byte {
	fmt.Printf("%s ", prompt)
	return gopass.GetPasswd()
}

func readVerifiedPassphrase() []byte {
	for {
		pass1 := readPassphrase("Passphrase:")
		pass2 := readPassphrase("Passphrase verification:")

		if bytes.Equal(pass1, pass2) {
			clear(pass2)
			return pass1
		}

		clear(pass1, pass2)
	}
}

func initStore() error {

	filename := defaultStorePath

	// Read store path
	fmt.Printf("Store location [%s]: ", filename)
	fmt.Scanln(&filename)

	// Read passphrase
	passphrase := readVerifiedPassphrase()
	defer clear(passphrase)

	salt, err := generateRandomSalt()
	if err != nil {
		return err
	}

	cipherKey, hmacKey, err := deriveKeys(passphrase, salt)
	if err != nil {
		return err
	}

	header := Header{Signature: fileSignature, Version: fileVersion, Salt: salt}

	store := NewStore(header)
	store.Entries["tweakers"] = Entry{"cafaro", "jeweetzelf", time.Now().Unix()}
	store.Entries["google"] = Entry{"tvdburgt@gmail.com", "badpasswdtbh", time.Now().Unix()}

	// TODO: check if file already exists!
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePermission)
	if err != nil {
		return err
	}
	defer file.Close()

	stream, err := initStreamCipher(cipherKey)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, hmacKey)
	fmt.Printf("hmacKey=%x\n", hmacKey)

	return store.Serialize(file, stream, mac)
}

func readStore(filename string) (*Store, error) {

	passphrase := readPassphrase("Passphrase:")
	_ = passphrase

	f, err := os.OpenFile(filename, os.O_RDONLY, filePermission)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Get file info with stat
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	s := new(Store)

	if err := s.Header.Deserialize(f); err != nil {
		return nil, err
	}

	cipherKey, hmacKey, err := deriveKeys(passphrase, s.Header.Salt)
	if err != nil {
		return nil, err
	}

	stream, err := initStreamCipher(cipherKey)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, hmacKey)

	// Verify header HMAC
	mac.Write(s.Header.Data[:40])
	if !hmac.Equal(s.Header.HMAC, mac.Sum(nil)) {
		return nil, errors.New("invalid passphrase, try again")
	}

	mac.Write(s.Header.HMAC)

	if err := s.Deserialize(f, int(fi.Size()), stream, mac); err != nil {
		return nil, err
	}

	return s, nil
}

func export() error {
	s, err := readStore("/home/tman/.passstore")
	if err != nil {
		return err
	}

	if err := s.Export(os.Stdout); err != nil {
		return err
	}

	return nil
}

func set() error {
	s, err := readStore("/home/tman/.passstore")
	if err != nil {
		return err
	}

	if len(os.Args) < 3 {
		return errors.New("not enough args")
	}

	e := Entry{Password: "defaultpass", Name: "defaultname"}
	id := os.Args[2]

	s.Set(id, e)

	return nil
}

func main() {

	var err error

	if len(os.Args) < 2 {
		fmt.Println("usage: passman command [pass store]")
		return
	}

	switch os.Args[1] {
	case "init":
		err = initStore()
	case "list":
		var s *Store
		s, err = readStore("/home/tman/.passstore")
		/* pass := []byte("") */
		/* s, err = NewStore("/home/tman/.passstore", os.O_RDONLY, pass) */
		if err == nil {
			fmt.Println(s.Entries)
			s.Print()
		}
	case "add":
		/* if len(os.Args) > 3 { */
		/* 	pass := readPass("Passphrase:") */
		/* 	s, err = NewStore("/home/tman/.passstore", os.O_RDWR, pass) */
		/* 	if err != nil { */
		/* 		fmt.Println("caught error", err) */
		/* 		break */
		/* 	} */
		/* 	err = add(s, os.Args[2], os.Args[3]) */
		/* } */
	case "export":
		err = export()
	case "gen":
		gen()
	case "set":
		err = set()
	}

	if err != nil {
		fmt.Println("error!")
		fmt.Fprintln(os.Stderr, err)
	}
}
