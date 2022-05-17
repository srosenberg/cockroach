package clierrorplus

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/cli/clierror"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/spf13/cobra"
)

func MaybeShoutError(
	wrapped func(*cobra.Command, []string) error,
) func(*cobra.Command, []string) error {
	__antithesis_instrumentation__.Notify(28429)
	return func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(28430)
		err := wrapped(cmd, args)
		return CheckAndMaybeShout(err)
	}
}

func CheckAndMaybeShout(err error) error {
	__antithesis_instrumentation__.Notify(28431)
	return clierror.CheckAndMaybeLog(err, log.Ops.Shoutf)
}
