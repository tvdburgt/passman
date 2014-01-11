package main

import (
	"fmt"
	"github.com/tvdburgt/passman/clipboard"
	"github.com/tvdburgt/passman/store"
	"os"
	"os/signal"
	"strings"
	"time"
)

var cmdClip = &Command{
	UsageLine: "clip [-value value] [-timeout timeout] [-persist] id",
	Short:     "display an individual entry",
	Long: `
get displays the content of a single passman entry. The identifier must be an
exact match. To search or display multiple entries, see passman list.
	`,
}

const closeMsg = "Closing in %s..."

var (
	clipTimeout = 20 * time.Second
	clipPersist = false
	clipField   = "password"
)

func init() {
	cmdClip.Run = runClip
	cmdClip.Flag.DurationVar(&clipTimeout, "timeout", clipTimeout, "")
	cmdClip.Flag.BoolVar(&clipPersist, "persist", clipPersist, "")
	cmdClip.Flag.StringVar(&clipField, "field", clipField, "")
}

func runClip(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Usage()
	}
	id := args[0]

	if clipTimeout <= 0 {
		fatalf("Invalid timeout. Timeout must be a positive duration.")
	}

	s := openStore()
	e, ok := s.Entries[id]
	if !ok {
		fatalf("Entry '%s' does not exist", id)
	}

	value := getValue(e)
	if err := clipboard.Setup(); err != nil {
		fatalf("Clipboard error: %s", err)
	}
	creq, cerr, err := clipboard.Put(value)
	if err != nil {
		fatalf("Clipboard error: %s", err)
	}

	fmt.Println("Listening for clipboard requests...")

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

	ticker := time.Tick(time.Second)

	var msg string
	write := func(s string, args... interface{}) {
		fmt.Print("\r", strings.Repeat(" ", len(msg)+1), "\r")
		fmt.Printf(s, args...)
		msg = s
	}

	write("Closing in %s...", clipTimeout)

	for {
		select {
		case req := <-creq:
			write("Password requested by '%s'\n", req)
			if !clipPersist {
				fmt.Println("Exiting. Add '-persist' flag to allow additional requests.")
				return
			}

		case err := <-cerr:
			write("")
			fatalf("Clipboard error: %s", err)

		case <-ticker:
			clipTimeout -= time.Second
			write("Closing in %s...", clipTimeout)

		case <-time.After(clipTimeout):
			write("Reached timeout. Exiting.\n")
			return

		case <-sigch:
			write("Received interrupt. Exiting.\n")
			os.Exit(1)
		}
	}
}

func validFields(e *store.Entry) []string {
	fields := []string{"'password'", "'name'"}
	for key, _ := range e.Metadata {
		fields = append(fields, fmt.Sprintf("'%s'", key))
	}
	return fields
}

func getValue(e *store.Entry) []byte {
	switch clipField {
	case "password":
		return e.Password
	case "name":
		return []byte(e.Name)
	default:
		if value, ok := e.Metadata[clipField]; ok {
			return []byte(value)
		}
		fatalf("Invalid field '%s' (possible fields: %s)",
			clipField, strings.Join(validFields(e), ", "))
	}
	return nil
}
