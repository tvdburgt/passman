package main

import (
	"fmt"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/import"
	"os"
)

var cmdImport = &Command{
	UsageLine: "import [-f file] [-format format] [-normalize] [-groups] import-file",
	Short:     "import passwords from an export file",
	Long: `
JSON-formatted, defaults to stdout.


	-format [format]
	The following formats are available:
		- keepassx (XML export)
	`,
}

var (
	importFormat    = "passman"
	importNormalize = false
	importGroups    = false
)

func init() {
	cmdImport.Run = runImport
	cmdImport.Flag.StringVar(&importFormat, "format", importFormat, "")
	cmdImport.Flag.BoolVar(&importNormalize, "normalize", importNormalize, "")
	cmdImport.Flag.BoolVar(&importGroups, "groups", importGroups, "")
	addFileFlag(cmdImport)
}

func runImport(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Usage()
	}
	if _, err := os.Stat(storeFile); err == nil {
		fatalf("Output file '%s' already exists", storeFile)
	}

	filename := args[0]
	file, err := os.Open(filename)
	if err != nil {
		fatalf("Failed to open import file: %s", err)
	}
	defer file.Close()

	settings := &imprt.Settings{
		NameGroups:       importGroups,
		NormalizeEntries: importNormalize,
	}

	s, err := imprt.ImportStore(file, importFormat, settings)

	if err != nil {
		fatalf("Import failed: %s", err)
	}

	passphrase := readVerifiedPassphrase()
	defer crypto.Clear(passphrase)
	writeStore(s, passphrase)
	fmt.Printf("Imported %d entries to '%s'.\n",
		len(s.Entries), storeFile)
}
