package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var cmdExport = &Command{
	UsageLine: "export file",
	Short:     "export passman store",
	Long: `
JSON-formatted, defaults to stdout.
	`,
}

func init() {
	cmdExport.Run = runExport
	// cmdExport.Flag.StringVar(&exportOutput, "o", "", "")
	// cmdExport.Flag.StringVar(&exportOutput, "output", "", "")
}

// Writes a JSON-formatted output of the password store to stdout or file.
func runExport(cmd *Command, args []string) {
	var err error
	var out *os.File = os.Stdout

	s, _ := openStore(false)

	if len(args) > 0 {
		filename := args[0]
		if _, err := os.Stat(filename); err == nil {
			fatalf("passman init: '%s' already exists", filename)
		}
		out, err = os.OpenFile(filename, storeFileCreateFlag, storeFilePerm)
		if err != nil {
			fatalf("passman export: %s", err)
		}
		defer out.Close()
	}


	if err = s.Export(out); err != nil {
		return
	}

	if out != os.Stdout {
		if path, err := filepath.Abs(out.Name()); err != nil {
			panic(err)
		} else {
			fmt.Printf("Created export file at '%s'\n", path)
		}
	}
}
