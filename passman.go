// Copright (c) 2013 Tijmen van der Burgt
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

package main

import (
	"bytes"
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
	filePermission = 0600
)

var defaultStorePath string

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

func cmdInit() error {

	pathname := defaultStorePath

	// Read store path
	fmt.Printf("Store location [%s]: ", pathname)
	fmt.Scanln(&pathname)

	// Read passphrase
	passphrase := readVerifiedPassphrase()
	defer clear(passphrase)

	// TODO: check if file already exists!
	file, err := os.OpenFile(pathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePermission)
	if err != nil {
		return err
	}
	defer file.Close()

	salt, err := generateRandomSalt()
	if err != nil {
		return err
	}

	stream, mac, err := initStreamParams(passphrase, salt)
	if err != nil {
		return err
	}

	header := NewHeader()
	copy(header.Salt[:], salt[:])
	store := NewStore(header)
	store.Entries["tweakers"] = Entry{"cafaro", "jeweetzelf", time.Now().Unix()}
	store.Entries["google"] = Entry{"tvdburgt@gmail.com", "badpasswdtbh", time.Now().Unix()}


	return store.Serialize(file, stream, mac)
}

func readStore(filename string) (*Store, error) {

	passphrase := readPassphrase("Passphrase:")

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

	header := new(Header)
	if err := header.Deserialize(f); err != nil {
		return nil, err
	}

	stream, mac, err := initStreamParams(passphrase, header.Salt[:])
	if err != nil {
		return nil, err
	}

	s := NewStore(header)

	f.Seek(0, 0) // Make sure stream starts at beggining of file
	if err := s.Deserialize(f, int(fi.Size()), stream, mac); err != nil {
		return nil, err
	}

	return s, nil
}

func cmdList() error {
	s, err := readStore(defaultStorePath)
	if err != nil {
		return err
	}

	s.Print(os.Stdout)

	return nil
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

func cmdSet() error {

	// passman set <id> <name> [-p password]

	s, err := readStore(defaultStorePath)
	if err != nil {
		return err
	}

	if len(os.Args) < 3 {
		return errors.New("not enough args")
	}

	id := os.Args[2]
	name := os.Args[3]
	os.Args = os.Args[3:]

	s.Entries[id] = Entry{Password: "defaultpass", Name: name}
	s.Print(os.Stdout)

	return nil
}

func cmdDel() error {
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
		err = cmdInit()
	case "list":
		err = cmdList()
		/* pass := []byte("") */
		/* s, err = NewStore("/home/tman/.passstore", os.O_RDONLY, pass) */
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
		err = cmdSet()
	}

	if err != nil {
		fmt.Println("error!")
		fmt.Fprintln(os.Stderr, err)
	}
}
