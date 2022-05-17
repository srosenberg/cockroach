//go:build !windows
// +build !windows

package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/sdnotify"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"golang.org/x/sys/unix"
)

var drainSignals = []os.Signal{unix.SIGINT, unix.SIGTERM}

var termSignal os.Signal = unix.SIGTERM

var quitSignal os.Signal = unix.SIGQUIT

var debugSignal os.Signal = unix.SIGUSR2

func handleSignalDuringShutdown(sig os.Signal) {
	__antithesis_instrumentation__.Notify(34344)

	signal.Reset(sig)

	if err := unix.Kill(unix.Getpid(), sig.(sysutil.Signal)); err != nil {
		__antithesis_instrumentation__.Notify(34346)

		log.Fatalf(context.Background(), "unable to forward signal %v: %v", sig, err)
	} else {
		__antithesis_instrumentation__.Notify(34347)
	}
	__antithesis_instrumentation__.Notify(34345)

	select {}
}

const backgroundFlagDefined = true

func maybeRerunBackground() (bool, error) {
	__antithesis_instrumentation__.Notify(34348)
	if startBackground {
		__antithesis_instrumentation__.Notify(34350)
		args := make([]string, 0, len(os.Args))
		foundBackground := false
		for _, arg := range os.Args {
			__antithesis_instrumentation__.Notify(34353)
			if arg == "--background" || func() bool {
				__antithesis_instrumentation__.Notify(34355)
				return strings.HasPrefix(arg, "--background=") == true
			}() == true {
				__antithesis_instrumentation__.Notify(34356)
				foundBackground = true
				continue
			} else {
				__antithesis_instrumentation__.Notify(34357)
			}
			__antithesis_instrumentation__.Notify(34354)
			args = append(args, arg)
		}
		__antithesis_instrumentation__.Notify(34351)
		if !foundBackground {
			__antithesis_instrumentation__.Notify(34358)
			args = append(args, "--background=false")
		} else {
			__antithesis_instrumentation__.Notify(34359)
		}
		__antithesis_instrumentation__.Notify(34352)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = stderr

		_ = os.Setenv(backgroundEnvVar, "1")

		return true, sdnotify.Exec(cmd)
	} else {
		__antithesis_instrumentation__.Notify(34360)
	}
	__antithesis_instrumentation__.Notify(34349)
	return false, nil
}

func disableOtherPermissionBits() {
	__antithesis_instrumentation__.Notify(34361)
	mask := unix.Umask(0000)
	mask |= 00007
	_ = unix.Umask(mask)
}
