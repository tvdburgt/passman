package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var cmdExport = &Command{
	UsageLine: "export [-output file]",
	Short: "export passman store",
	Long: `
JSON-formatted, defaults to stdout.
	`,
}

var (
	exportOutput string
)

func init() {
	cmdExport.Run = runExport
	cmdExport.Flag.StringVar(&exportOutput, "o", "", "")
	cmdExport.Flag.StringVar(&exportOutput, "output", "", "")
}


// Writes a JSON-formatted output of the password store to stdout or file.
func runExport(cmd *Command, args []string) {
	var out *os.File = os.Stdout

	store, err := readPassStore()
	if err != nil {
		return
	}

	if len(exportOutput) > 0 {
		out, err = os.Create(exportOutput)
		// TODO: check if file exists
		if err != nil {
			fatalf("passman export: %s", err)
		}
		defer out.Close()
	}

	if err = store.Export(out); err != nil {
		return
	}

	if out != os.Stdout {
		if path, err := filepath.Abs(out.Name()); err != nil {
			fatalf("passman export: %s", err)
		} else {
			fmt.Printf("Created export file at '%s'\n", path)
		}
	}
}
