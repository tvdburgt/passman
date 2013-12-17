package main

import (
	"fmt"
	"os"
	"regexp"
)

var cmdList = &Command{
	Run: runList,
	UsageLine: "list [pattern]",
	Short: "list store entries",
	Long: `
displays all entries in store, optionally filtered by a regex pattern
	`,
}

func runList(cmd *Command, args []string) {
	s, err := readPassStore()
	if err != nil {
		fatalf("passman list: %s", err)
	}

	var pattern *regexp.Regexp
	if len(args) > 0 {
		pattern, err = regexp.Compile(args[0])
		fmt.Println(err)
		if err != nil {
			fatalf("invalid pattern: %s", err)
		}
	}
	s.List(os.Stdout, pattern)
}
