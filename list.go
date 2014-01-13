package main

import (
	"os"
	"regexp"
)

var cmdList = &Command{
	UsageLine: "list [search pattern]",
	Short: "list store entries",
	Long: `
regex is posix?
displays all entries in store, optionally filtered by a regex pattern
	`,
}

func init() {
	cmdList.Run = runList
	addFileFlag(cmdList)
}

func runList(cmd *Command, args []string) {
	s, _ := openStore(false)

	// TODO: posix or not?
	var pattern *regexp.Regexp
	var err error
	if len(args) > 0 {
		pattern, err = regexp.Compile(args[0])
		if err != nil {
			fatalf("invalid pattern: %s", err)
		}
	}
	s.List(os.Stdout, pattern)
}
