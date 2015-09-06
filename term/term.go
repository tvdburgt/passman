package term

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"os/signal"
	"strings"
)

// ReadPassphrase prints a prompt and interactively reads a passphrase from the
// terminal. The entered passphrase is not echoed.
func ReadPassphrase(prompt string, args ...interface{}) []byte {
	fd := int(os.Stdin.Fd())
	state, err := terminal.GetState(fd)
	if err != nil {
		panic("could not get state of terminal: " + err.Error())
	}

	prompt = fmt.Sprintf(prompt, args...)
	fmt.Print(prompt)
	defer clear(prompt)

	// Restore terminal when SIGINT is caught
	sigch := make(chan os.Signal, 1)
	defer close(sigch)
	signal.Notify(sigch, os.Interrupt)
	defer signal.Stop(sigch)
	go func() {
		if _, ok := <-sigch; ok {
			clear(prompt)
			terminal.Restore(fd, state)
			os.Exit(1)
		}
	}()

	phrase, err := terminal.ReadPassword(fd)
	if err != nil {
		panic("failed to read passphrase: " + err.Error())
	}

	return phrase
}

func clear(line string) {
	fmt.Print("\r", strings.Repeat(" ", len(line)), "\r")
}
