package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"os"

	"github.com/cockroachdb/cockroach/pkg/cli/exit"
)

var drainSignals = []os.Signal{os.Interrupt}

var termSignal os.Signal = nil

var quitSignal os.Signal = nil

var debugSignal os.Signal = nil

const backgroundFlagDefined = false

func handleSignalDuringShutdown(os.Signal) {
	__antithesis_instrumentation__.Notify(34362)

	exit.WithCode(exit.UnspecifiedError())
}

func maybeRerunBackground() (bool, error) {
	__antithesis_instrumentation__.Notify(34363)
	return false, nil
}

func disableOtherPermissionBits() {
	__antithesis_instrumentation__.Notify(34364)

}
