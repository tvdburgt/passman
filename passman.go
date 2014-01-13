// Copright (c) 2013 Tijmen van der Burgt
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh/terminal"
	"crypto/cipher"
	"flag"
	"fmt"
	"github.com/tvdburgt/passman/cache"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/store"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
)

// TODO: clear derived keys etc.

const (
	storeFilePerm       os.FileMode = 0600 // Store only requires rw perm for owner
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
		fmt.Println(storeFile)
	}
}

// A Command is an implementation of a subcommand. The subcommand pattern
// code is taken from the Go source ($GOROOT/src/cmd).
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

func readPassphrase(prompt string, args ...interface{}) []byte {
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.GetState(fd)
	if err != nil {
		panic("Could not get state of terminal: " + err.Error())
	}
	defer terminal.Restore(fd, oldState)

	// Restore terminal when SIGINT is caught
	sigch := make(chan os.Signal, 1)
	defer close(sigch)
	signal.Notify(sigch, os.Interrupt)
	defer signal.Stop(sigch)
	go func() {
		for _ = range sigch {
			terminal.Restore(fd, oldState)
			fmt.Println("\nReceived interrupt. Exiting.")
			os.Exit(1)
		}
	}()

	prompt = fmt.Sprintf(prompt+": ", args...)
	fmt.Print(prompt)
	p, err := terminal.ReadPassword(fd)
	if err != nil {
		panic("Failed to read passphrase: " + err.Error())
	}
	fmt.Print("\r", strings.Repeat(" ", len(prompt)), "\r")
	return p
}

func verifyPassphrase() []byte {
	for {
		p1 := readPassphrase("Enter passphrase")
		p2 := readPassphrase("Verify passphrase")
		if bytes.Equal(p1, p2) {
			if len(p1) == 0 {
				// fatalf("Invalid passphrase, aborting.")
			}
			crypto.Clear(p2)
			return p1
		}
		crypto.Clear(p1, p2)
		fmt.Fprintln(os.Stderr, "Passphrases do not match. Try again.")
	}
}

func writeStore(s *store.Store, passphrase []byte) {
	file, err := os.OpenFile(storeFile, storeFileCreateFlag, storeFilePerm)
	if err != nil {
		fatalf("Unable to write to store: %s", err)
	}
	defer file.Close()

	// Generate a new random salt
	err = crypto.Rand(s.Header.Salt[:])
	if err != nil {
		fatalf("Failed to generate salt: %s", err)
	}

	p := s.Header.Params
	stream, mac := crypto.InitStreamParams(passphrase, s.Header.Salt[:],
		int(p.LogN), int(p.R), int(p.P))

	err = s.Encrypt(cipher.StreamWriter{S: stream, W: file}, mac)
	if err != nil {
		fatalf("Failed to encrypt store: %s", err)
	}

	// Try to cache key
	err = cache.CacheKey(passphrase)
	if err != nil {
		// fmt.Println(err)
	}
}

func readStore(passphrase []byte) *store.Store {
	file, err := os.Open(storeFile)
	if err != nil {
		fatalf("Failed to open store: %s", err)
	}
	defer file.Close()

	// Get file info with stat
	fi, err := file.Stat()
	if err != nil {
		panic(err)
	}

	// We need to unmarshal the header before the rest of the store can be
	// decrypted
	s := store.NewStore(&store.Header{})
	if err = s.Header.Unmarshal(file); err != nil {
		fatalf("Failed to marshal header: %s", err)
	}

	// Rewind file offset to origin of file (offset is modified by
	// marshalling the header)
	file.Seek(0, os.SEEK_SET)

	// Attempt to decrypt store
	p := s.Header.Params
	stream, mac := crypto.InitStreamParams(passphrase, s.Header.Salt[:],
		int(p.LogN), int(p.R), int(p.P))
	err = s.Decrypt(cipher.StreamReader{stream, file}, fi.Size(), mac)
	if err != nil {
		fatalf("Failed to decrypt store: %s", err)
	}
	return s
}

// Helper function for reading both passphrase and store.
// Parameter dictates wheter access is read-only (i.e., passphrase needs to be
// retained).
func openStore(ro bool) (s *store.Store, p []byte) {
	p = readPassphrase("Enter passphrase for '%s'", storeFile)
	s = readStore(p)
	if ro {
		crypto.Clear(p)
	}
	return
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
