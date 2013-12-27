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
	UsageLine: "import [-file FILE] [-format FORMAT] file",
	Short:     "import passwords from an export file",
	Long: `
JSON-formatted, defaults to stdout.
	`,
}

var (
	importFormat    = "passman"
	importNormalize = imprt.NormalizeEntries
	importGroups    = imprt.ImportGroups
)

func init() {
	cmdImport.Run = runImport
	cmdImport.Flag.StringVar(&importFormat, "format", importFormat, "")
	cmdImport.Flag.BoolVar(&importNormalize, "normalize", importNormalize, "")
	cmdImport.Flag.BoolVar(&importGroups, "groups", importGroups, "")
	addStoreFlags(cmdImport)
}

func runImport(cmd *Command, args []string) {
	if len(args) < 1 {
		fatalf("passman import: missing file")
	}

	if _, err := os.Stat(storeFile); err == nil {
		fatalf("passman init: '%s' already exists", storeFile)
	}

	filename := args[0]
	file, err := os.Open(filename)
	if err != nil {
		fatalf("passman import: %s", err)
	}
	defer file.Close()

	s, err := importStore(file)
	if err != nil {
		fatalf("passman import: %s", err)
	}

	passphrase := readVerifiedPass()
	defer crypto.Clear(passphrase)

	if err := writeStore(s, passphrase); err != nil {
		fatalf("passman init: %s", err)
	}

	fmt.Printf("Imported entries to '%s'.\n", storeFile)
}

func importStore(file *os.File) (s *store.Store, err error) {
	// Apply flags to import package
	imprt.ImportGroups = importGroups
	imprt.NormalizeEntries = importNormalize

	switch importFormat {
	case "passman":
		s = store.NewStore()
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
