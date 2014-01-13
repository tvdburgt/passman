package main

import (
	"errors"
	"fmt"
	"github.com/tvdburgt/passman/clipboard"
	"github.com/tvdburgt/passman/store"
	"os"
	"os/signal"
	"strings"
	"time"
)

var cmdClip = &Command{
	UsageLine: "clip [flags] entry_id",
	Short:     "make entry data available for clipboard requests",
	Long: `
The clip command makes information associated with an entry, available for
clipboard request from other applications. Both CLIPBOARD and PRIMARY X11
selection types are used to convey data.

Available flags:

    -fields field_list
	Specifies which entry values should be copied to the clipboard.  If
	multiple fields are supplied, the associated values are copied
	consecutively after each selection request (in the same order as they
	are provided). Possible fields are "password", "name" and any of the
	metadata keys that are set for that particular entry. The default value
	is "password". Multiple values are separated with a comma (without
	spaces). If a metadata field is provided that collides with another
	field, the non-metadata field will be used.

    -timeout duration
	Sets timeout (delay before passman will exit and clear clipboard
	selection). Duration must be a valid duration string that ends with a
	valid time unit ("s" for seconds, "m" for minutes, etc.).  To disable
	the timeout, use a nonpositive duration. See 'godoc time ParseDuration'
	for more info about the duration string format.

    -persist
	By default, passman will automatically exit after each value in -fields
	has been delivered through the clipboard. By using this flag, entry data
	will remain in the clipboard upon successive selection requests (as long
	as passman remains running). Keep in mind that for a multivalued -fields
	flag, only the first field is made available for selection requests.
	`,
}

type fieldSlice []string

func (fs *fieldSlice) String() string {
	return strings.Join(*fs, ",")
}

func (fs *fieldSlice) Set(value string) error {
	if len(value) == 0 {
		return errors.New("value is empty")
	}
	*fs = strings.Split(value, ",")
	return nil
}

var currentMessage string
func printMessage(s string, args ...interface{}) {
	fmt.Print("\r", strings.Repeat(" ", len(currentMessage)+1), "\r")
	fmt.Printf(s, args...)
	currentMessage = s
}

var (
	clipTimeout = 20 * time.Second
	clipPersist = false
	clipFields  = fieldSlice{"password"}
)

func init() {
	cmdClip.Run = runClip
	cmdClip.Flag.DurationVar(&clipTimeout, "timeout", clipTimeout, "")
	cmdClip.Flag.BoolVar(&clipPersist, "persist", clipPersist, "")
	cmdClip.Flag.Var(&clipFields, "fields", "")
	addFileFlag(cmdClip)
}

func runClip(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.Usage()
	}
	id := args[0]

	s, _ := openStore(false)
	e, ok := s.Entries[id]
	if !ok {
		fatalf("Entry %q does not exist.", id)
	}

	if err := clipboard.Setup(); err != nil {
		fatalf("Clipboard error: %s", err)
	}

	// Call getValue for each field to trigger possible fatal errors for
	// invalid fields.
	for _, f := range clipFields {
		getValue(e, f)
	}

	// Disallow -persist flag if more than one field is provided
	// if len(clipFields) > 1 && clipPersist {
	// 	clipPersist = false
	// 	fmt.Println("Ignoring -persist flag (disallowed in combination with multivalued -field flag)")
	// }

	fmt.Println("Listening for clipboard requests...")

	// Only set timeout/ticker if timeout duration is positive
	var timeout, tick <-chan time.Time
	if clipTimeout > 0 {
		timeout = time.After(clipTimeout)
		tick = time.Tick(time.Second)
		printMessage("Closing in %s...", clipTimeout)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	for _, field := range clipFields {
		value := getValue(e, field)
		clipValue(field, value, timeout, tick, sig)
	}

	fmt.Println("All field values are copied. Exiting.")
}

func clipValue(field string, value []byte, timeout, tick <-chan time.Time,
	sig <-chan os.Signal) {

	request, err := clipboard.Put(value)
	for {
		select {
		case name := <-request:
			printMessage("Field %q requested by %q\n", field, name)
			if !clipPersist {
				clipboard.Clear()
				return
			}
		case e := <-err:
			printMessage("Clipboard error: %s\n", e)
			os.Exit(1)
		case s := <-sig:
			printMessage("Received %s signal. Exiting.\n", s)
			os.Exit(1)
		case <-tick:
			clipTimeout -= time.Second
			printMessage("Closing in %s...", clipTimeout)
		case <-timeout:
			printMessage("Reached timeout. Exiting.\n")
			os.Exit(0)
		}
	}
}

func validFields(e *store.Entry) []string {
	fields := []string{`"password"`, `"name"`}
	for key, _ := range e.Metadata {
		fields = append(fields, fmt.Sprintf("%q", key))
	}
	return fields
}

func getValue(e *store.Entry, field string) []byte {
	switch field {
	case "password":
		return e.Password
	case "name":
		return []byte(e.Name)
	default:
		if value, ok := e.Metadata[field]; ok {
			return []byte(value)
		}
		fatalf("Invalid field %q (possible fields: %s)",
			field, strings.Join(validFields(e), ", "))
	}
	return nil
}
