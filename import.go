package main

import (
	"fmt"
	"github.com/tvdburgt/passman/crypto"
	"github.com/tvdburgt/passman/import"
	"github.com/tvdburgt/passman/import/keepass"
	"github.com/tvdburgt/passman/import/keepass2"
	"github.com/tvdburgt/passman/import/keepassx"
	"github.com/tvdburgt/passman/store"
	"os"
)

var cmdImport = &Command{
	UsageLine: "import [-f file] [-format format] [-normalize] [-groups] import-file",
	Short:     "import passwords from an export file",
	Long: `
JSON-formatted, defaults to stdout.
	`,
}

var (
	importFormat    = "passman"
	importNormalize = &imprt.NormalizeEntries
	importGroups    = &imprt.ImportGroups
)

func init() {
	cmdImport.Run = runImport
	cmdImport.Flag.StringVar(&importFormat, "format", importFormat, "")
	cmdImport.Flag.BoolVar(importNormalize, "normalize", *importNormalize, "")
	cmdImport.Flag.BoolVar(importGroups, "groups", *importGroups, "")
	addFileFlag(cmdImport)
}

func runImport(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Usage()
	}
	if _, err := os.Stat(storeFile); err == nil {
		fatalf("Import target file '%s' already exists", storeFile)
	}

	filename := args[0]
	file, err := os.Open(filename)
	if err != nil {
		fatalf("Failed to open import source file: %s", err)
	}
	defer file.Close()

	s, err := importStore(file)
	if err != nil {
		fatalf("Import failed: %s", err)
	}

	passphrase := verifyPassphrase()
	defer crypto.Clear(passphrase)
	writeStore(s, passphrase)
	fmt.Printf("Imported %d entries to '%s'.\n",
		len(s.Entries), storeFile)
}

func importStore(file *os.File) (s *store.Store, err error) {
	switch importFormat {
	case "passman":
		s = store.NewStore(store.NewHeader())
		err = s.Import(file)
	case "keepass":
		s, err = keepass.ImportXml(file)
	case "keepass2":
		s, err = keepass2.ImportXml(file)
	case "keepassx":
		s, err = keepassx.ImportXml(file)
	default:
		err = fmt.Errorf("unknown format: '%s'", importFormat)
	}
	return
}
