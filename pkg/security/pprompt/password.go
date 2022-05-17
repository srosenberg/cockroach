// Package pprompt provides a facility to prompt a user for a password
// securely (i.e. without echoing the password) from an interactive
// terminal.
//
// This is a separate package to ensure that the CLI code that uses
// this does not indirectly depend on other features from the
// 'security' package.
package pprompt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func PromptForPassword() (string, error) {
	__antithesis_instrumentation__.Notify(186995)
	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		__antithesis_instrumentation__.Notify(186997)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(186998)
	}
	__antithesis_instrumentation__.Notify(186996)

	fmt.Print("\n")

	return string(password), nil
}
