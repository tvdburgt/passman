package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/store"
	"github.com/tvdburgt/passman/term"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	storeFilePerm       os.FileMode = 0600 // Store only requires rw perms for owner
	storeFileCreateFlag             = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	storeFileEnvKey                 = "PASSMAN_STORE"
	storeFileDefault                = "$HOME/.pass_store"
)

// Global variable for filename of the passman store. This value defaults
// to the storeFileDefault constant or the value of the environment variable with the
// key in storeFileEnvKey (if set). The filename can be overridden with the
// global -f or -file flag.
var storeFile string

// Commands lists the available commands and help topics.
// The order here is the order in which they are printed by 'passman help'.
var commands = []*Command{
	cmdGet,
	cmdSet,
	cmdClip,
	cmdInit,
	cmdImport,
	cmdExport,
	cmdList,
	cmdStat,
	cmdGen,
	cmdDelete,
}

// Set default store file
func init() {
	if v := os.Getenv(storeFileEnvKey); v != "" {
		storeFile = v
	} else {
		u, err := user.Current()
		if err != nil {
			panic("failed to get current user: " + err.Error())
		}
		mapping := func(s string) string {
			switch s {
			case "HOME":
				return u.HomeDir
			default:
				return ""
			}
		}
		storeFile = filepath.FromSlash(storeFileDefault)
		storeFile = os.Expand(storeFile, mapping)
		// fmt.Println(storeFile)
	}
}

// A Command is an implementation of a subcommand. The subcommand pattern
// code is taken from the Go source ($GOROOT/src/cmd/go).
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
}

func readVerifiedPassphrase() []byte {
	for {
		p1 := term.ReadPassphrase("Enter passphrase: ")
		p2 := term.ReadPassphrase("Verify passphrase: ")
		if bytes.Equal(p1, p2) {
			if len(p1) == 0 {
				// fatalf("Invalid passphrase, aborting.")
			}
			crypto.Clear(p2)
			return p1
		}
		crypto.Clear(p1)
		crypto.Clear(p2)
		fmt.Fprintln(os.Stderr, "Passphrases do not match, try again.")
	}
}

func writeStore(s *store.Store, passphrase []byte) {
	file, err := os.OpenFile(storeFile, storeFileCreateFlag, storeFilePerm)
	if err != nil {
		fatalf("Unable to write to store: %s", err)
	}
	defer file.Close()

	// Generate a new random salt
	err = crypto.ReadRand(s.Header.Salt[:])
	if err != nil {
		fatalf("Failed to generate salt: %s", err)
	}

	err = crypto.WriteStore(file, s, passphrase)
	if err != nil {
		fatalf("Failed to write to store: %s", err)
	}
}

func readStore(passphrase []byte) (s *store.Store, err error) {
	file, err := os.Open(storeFile)
	if err != nil {
		return
	}
	defer file.Close()

	s, err = crypto.ReadStore(file, passphrase)
	if err != nil {
		return
	}

	return
}

// Helper function for reading both passphrase and store.
func openRwStore() (s *store.Store, passphrase []byte) {
	for {
		passphrase = term.ReadPassphrase("Enter passphrase for %q: ", storeFile)
		s, err := readStore(passphrase)
		if err == nil {
			return s, passphrase
		}
		crypto.Clear(passphrase)
		if err == crypto.ErrWrongPass {
			fmt.Fprintln(os.Stderr, "Incorrect passphrase. Try again.")
			continue
		}
		fatalf("Failed to open store: %s", err)
	}
}

func openStore() *store.Store {
	s, passphrase := openRwStore()
	crypto.Clear(passphrase)
	return s
}

// Name returns the command's name: the first word in the usage line.
// TODO: 'go help cmd' should suffix cmd.Short with 'passman' (see go source)
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

func usage() {
	fmt.Fprintln(os.Stderr, "passman usage")
	os.Exit(1)
}

func fatalf(format string, args ...interface{}) {
	log.Printf(format, args...)
	os.Exit(1)
}

// Adds global flags for store-specific commands.
func addFileFlag(cmd *Command) {
	cmd.Flag.StringVar(&storeFile, "f", storeFile, "")
	cmd.Flag.StringVar(&storeFile, "file", storeFile, "")
}

// Makes sure the store file path is absolute
func fixStoreFile() {
	var err error
	storeFile, err = filepath.Abs(storeFile)
	if err != nil {
		panic(err)
	}
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
			// cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Usage = cmd.Usage
			cmd.Flag.Parse(args[1:])
			fixStoreFile()
			args = cmd.Flag.Args()
			cmd.Run(cmd, args)
			os.Exit(0)
		}
	}

	fmt.Fprintf(os.Stderr, "%[1]s: unknown subcommand %q\nRun '%[1]s help' for usage.\n",
		os.Args[0], args[0])
	os.Exit(1)
}
