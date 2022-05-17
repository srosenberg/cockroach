package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func capture(args ...string) (string, error) {
	__antithesis_instrumentation__.Notify(52662)
	var cmd *exec.Cmd
	if len(args) == 0 {
		__antithesis_instrumentation__.Notify(52665)
		panic("capture called with no arguments")
	} else {
		__antithesis_instrumentation__.Notify(52666)
		if len(args) == 1 {
			__antithesis_instrumentation__.Notify(52667)
			cmd = exec.Command(args[0])
		} else {
			__antithesis_instrumentation__.Notify(52668)
			cmd = exec.Command(args[0], args[1:]...)
		}
	}
	__antithesis_instrumentation__.Notify(52663)
	out, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(52669)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			__antithesis_instrumentation__.Notify(52671)
			err = fmt.Errorf("%w: %s", err, exitErr.Stderr)
		} else {
			__antithesis_instrumentation__.Notify(52672)
		}
		__antithesis_instrumentation__.Notify(52670)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(52673)
	}
	__antithesis_instrumentation__.Notify(52664)
	return string(bytes.TrimSpace(out)), err
}

func spawn(args ...string) error {
	__antithesis_instrumentation__.Notify(52674)
	var cmd *exec.Cmd
	if len(args) == 0 {
		__antithesis_instrumentation__.Notify(52676)
		panic("spawn called with no arguments")
	} else {
		__antithesis_instrumentation__.Notify(52677)
		if len(args) == 1 {
			__antithesis_instrumentation__.Notify(52678)
			cmd = exec.Command(args[0])
		} else {
			__antithesis_instrumentation__.Notify(52679)
			cmd = exec.Command(args[0], args[1:]...)
		}
	}
	__antithesis_instrumentation__.Notify(52675)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
