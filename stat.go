package main

import (
	"fmt"
	"github.com/tvdburgt/passman/store"
	"os"
	"text/tabwriter"
	"time"
)

var cmdStat = &Command{
	Run:       runStat,
	UsageLine: "stat",
	Short:     "shows status information (metadata) about the store file",
	Long: `
get displays the content of a single passman entry. The identifier must be an
exact match. To search or display multiple entries, see passman list.
	`,
}

func init() {
	addFileFlag(cmdStat)
}

func runStat(cmd *Command, args []string) {
	file, err := os.Open(storeFile)
	if err != nil {
		fatalf("%s", err)
	}

	h := store.Header{}
	if err := h.Unmarshal(file); err != nil {
		fatalf("%s", err)
	}

	fi, err := file.Stat()
	if err != nil {
		fatalf("%s", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "File\t: %s (%d bytes)\n", storeFile, fi.Size())
	fmt.Fprintf(w, "Last modified\t: %s\n", fi.ModTime().Truncate(time.Second))
	fmt.Fprintf(w, "Signature\t: %x (version 0x%x)\n", h.Signature, h.Version)
	fmt.Fprintf(w, "Salt\t: %x\n", h.Salt)
	// fmt.Fprintf(w, "Inner key\t: %x\n", h.InnerKey)
	fmt.Fprintf(w, "Scrypt params\t: N=%d r=%d p=%d\n",
		1<<h.Params.LogN, h.Params.R, h.Params.P)
	w.Flush()
}
