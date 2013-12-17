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
	"log"
	"strings"
	// TODO: github.com/seehuhn/password
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

// TODO: check for errors (Ctrl-c)
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
	return writeStore(s, passphrase)
}

func writeStore(s *store.Store, passphrase []byte) (err error) {
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

func readPassStore() (*store.Store, error) {
	passphrase := readPass("Enter passphrase for '%s'", storePath)
	defer crypto.Clear(passphrase)
	return readStore(passphrase)
}

func readStore(passphrase []byte) (s *store.Store, err error) {
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


// Using the subcommand code from the Go source.
// A Command is an implementation of a go command
// like go build or go fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'go help' output.
	Short string

	// Long is the long message shown in the 'go help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet

	// CustomFlags indicates that the command will do its own
	// flag parsing.
	CustomFlags bool
}

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(1)
}

// Commands lists the available commands and help topics.
// The order here is the order in which they are printed by 'passman help'.
var commands = []*Command{
	cmdExport,
	cmdGet,
	cmdList,
	cmdGen,
	cmdRm,
	cmdSet,
}

func usage() {
	fmt.Fprintln(os.Stderr, "passman usage")
	os.Exit(1)
}

func fatalf(format string, args ...interface{}) {
	log.Printf(format, args...)
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	// Don't include date etc. in log output
	log.SetFlags(0)


	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			// TODO: CustomFlags
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			cmd.Run(cmd, args)
			os.Exit(0)
		}
	}

	fmt.Fprintf(os.Stderr, "%[1]s: unknown subcommand %q\nRun '%[1]s help' for usage.\n",
		os.Args[0], args[0])
	os.Exit(1)
}
