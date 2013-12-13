// Copright (c) 2013 Tijmen van der Burgt
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/howeyc/gopass"
	"os"
	"os/user"
	"path/filepath"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/store"
)

// TODO: clear derived keys etc.

const (
	filePermission   = 0600 // rw for owner only
	storePathKey     = "PASSMAN_STORE"
	defaultStorePath = ".pass_store" // relative to $HOME
)

var storePath string

// file presedence:
//   [1] -file flag (TODO)
//   [2] $PASSMAN_STORE environment variable
//   [3] ~/.pass_store
func init() {
	if path := os.Getenv(storePathKey); path != "" {
		storePath = path
	} else {
		u, err := user.Current()
		if err != nil {
			panic(err)
		}
		storePath = filepath.Join(u.HomeDir, defaultStorePath)
	}
}

func readPass(prompt string, args ...interface{}) []byte {
	fmt.Printf(prompt+": ", args...)
	return gopass.GetPasswd()
}

func readVerifiedPass() []byte {
	for {
		pass1 := readPass("Passphrase")
		pass2 := readPass("Passphrase verification")

		if bytes.Equal(pass1, pass2) {
			crypto.Clear(pass2)
			return pass1
		}

		fmt.Fprintln(os.Stderr, "error: passphrases don't match, try again")
		crypto.Clear(pass1, pass2)
	}
}

func cmdInit() (err error) {
	// Read file and make sure it doesn't exist
	fmt.Printf("Store location [%s]: ", storePath)
	fmt.Scanln(&storePath)
	// if _, err := os.Stat(filename); err == nil {
	// 	return fmt.Errorf("the file '%s' already exists", filename)
	// }

	// Read passphrase
	passphrase := readVerifiedPass()
	defer crypto.Clear(passphrase)

	header := store.NewHeader()
	s := store.NewStore(header) // create default ctor with header defaults?
	return saveStore(s, passphrase)
}

func saveStore(s *store.Store, passphrase []byte) (err error) {
	var salt []byte

	file, err := os.OpenFile(storePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePermission)
	if err != nil {
		return err
	}
	defer file.Close()

	if salt, err = crypto.GenerateRandomSalt(); err != nil {
		return
	}
	copy(s.Header.Salt[:], salt)

	stream, mac, err := crypto.InitStreamParams(passphrase, salt)
	if err != nil {
		return err
	}

	return s.Serialize(file, stream, mac)
}

func openStore(passphrase []byte) (s *store.Store, err error) {
	f, err := os.OpenFile(storePath, os.O_RDONLY, filePermission)
	if err != nil {
		return
	}
	defer f.Close()

	// Get file info with stat
	fi, err := f.Stat()
	if err != nil {
		return
	}

	// We need to serialize the header before the rest of the file can be
	// serialized
	header := new(store.Header)
	if err = header.Deserialize(f); err != nil {
		return
	}

	stream, mac, err := crypto.InitStreamParams(passphrase, header.Salt[:])
	if err != nil {
		return
	}

	s = store.NewStore(header)

	// Rewind file offset to origin of file (offset is modified by
	// header.Deserialize).
	f.Seek(0, os.SEEK_SET)

	// Attempt to deserialize store
	err = s.Deserialize(f, int(fi.Size()), stream, mac)
	return
}

func cmdList() error {
	passphrase := readPass("Enter passphrase for '%s'", storePath)
	defer crypto.Clear(passphrase)
	s, err := openStore(passphrase)
	if err != nil {
		return err
	}

	fmt.Print(s)

	return nil
}

// TODO: -q query
func cmdGet() (err error) {
	if len(os.Args) < 3 {
		return errors.New("missing id argument")
	}
	id := os.Args[2]

	passphrase := readPass("Enter passphrase for '%s'", storePath)
	defer crypto.Clear(passphrase)
	s, err := openStore(passphrase);
	if err != nil {
		return
	}

	e := s.Find(id)
	if e == nil {
		return fmt.Errorf("no such entry '%s'", id)
	}

	fmt.Println(e)
	return
}

func cmdImport() (err error) {
	var flagFormat string
	const usage = "import file format (passman, keepass)"

	if len(os.Args) < 3 {
		return errors.New("missing file argument")
	}
	filename := os.Args[2]

	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	fs.StringVar(&flagFormat, "format", "passman", usage)
	fs.StringVar(&flagFormat, "f", "passman", usage)
	fs.Parse(os.Args[3:])

	switch flagFormat {
	case "passman":
	case "keepass":
	default:
		return fmt.Errorf("unknown import file format '%s'", flagFormat)
	}

	fmt.Println(filename)
	return
}

func cmdRm() (err error) {
	if len(os.Args) < 3 {
		return errors.New("missing id argument")
	}
	id := os.Args[2]

	// TODO: create wrapper fn
	passphrase := readPass("Enter passphrase for '%s'", storePath)
	defer crypto.Clear(passphrase)
	s, err := openStore(passphrase)
	if err != nil {
		return
	}

	// if _, ok := s.Entries[id]; !ok {
	// 	return fmt.Errorf("no such entry '%s'", id)
	// }
	
	// delete(s.Entries, id)
	if err = saveStore(s, passphrase); err != nil {
		return
	}
	fmt.Printf("Removed entry '%s' from store\n", id)

	return
}

func main() {

	var err error
	var cmd string

	if len(os.Args) < 2 {
		fmt.Println("usage: passman command [pass store]")
		return
	}


	// e := new(store.Entry)
	// // fmt.Printf("%v %T %q\n", e, e, e)
	// fmt.Println(e.Entries)
	// e.Entries["foo"] = new(store.Entry)
	// return

	// fmt.Printf("path=%s\n", storePath)

	// Shift subcommand from args
	// cmd, os.Args = os.Args[1], os.Args[:]
	cmd = os.Args[1]
	// os.Args[1], os.Args = os.Args[0], os.Args[1:]

	switch cmd {
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
		err = cmdExport()
	case "import":
		err = cmdImport()
	// case "gen":
	// 	gen()
	case "get":
		err = cmdGet()
	case "set":
		err = cmdSet()
	case "rm":
		err = cmdRm()
	case "clip":
		err = cmdClip()
		// err = xclip()
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		// Exit code = 1?
	}
}
