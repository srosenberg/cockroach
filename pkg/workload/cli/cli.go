package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

func WorkloadCmd(userFacing bool) *cobra.Command {
	__antithesis_instrumentation__.Notify(693645)
	rootCmd := SetCmdDefaults(&cobra.Command{
		Use:     `workload`,
		Short:   `generators for data and query loads`,
		Version: "details:\n" + build.GetInfo().Long(),
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   `version`,
		Short: `print version information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			__antithesis_instrumentation__.Notify(693649)
			info := build.GetInfo()
			fmt.Println(info.Long())
			return nil
		}},
	)
	__antithesis_instrumentation__.Notify(693646)

	for _, subCmdFn := range subCmdFns {
		__antithesis_instrumentation__.Notify(693650)
		rootCmd.AddCommand(subCmdFn(userFacing))
	}
	__antithesis_instrumentation__.Notify(693647)
	if userFacing {
		__antithesis_instrumentation__.Notify(693651)
		allowlist := map[string]struct{}{
			`workload`: {},
			`init`:     {},
			`run`:      {},
		}
		for _, m := range workload.Registered() {
			__antithesis_instrumentation__.Notify(693654)
			allowlist[m.Name] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(693652)
		var hideNonPublic func(c *cobra.Command)
		hideNonPublic = func(c *cobra.Command) {
			__antithesis_instrumentation__.Notify(693655)
			if _, ok := allowlist[c.Name()]; !ok {
				__antithesis_instrumentation__.Notify(693657)
				c.Hidden = true
			} else {
				__antithesis_instrumentation__.Notify(693658)
			}
			__antithesis_instrumentation__.Notify(693656)
			for _, sub := range c.Commands() {
				__antithesis_instrumentation__.Notify(693659)
				hideNonPublic(sub)
			}
		}
		__antithesis_instrumentation__.Notify(693653)
		hideNonPublic(rootCmd)
	} else {
		__antithesis_instrumentation__.Notify(693660)
	}
	__antithesis_instrumentation__.Notify(693648)
	return rootCmd
}

var subCmdFns []func(userFacing bool) *cobra.Command

func AddSubCmd(fn func(userFacing bool) *cobra.Command) {
	__antithesis_instrumentation__.Notify(693661)
	subCmdFns = append(subCmdFns, fn)
}

func HandleErrs(
	f func(cmd *cobra.Command, args []string) error,
) func(cmd *cobra.Command, args []string) {
	__antithesis_instrumentation__.Notify(693662)
	return func(cmd *cobra.Command, args []string) {
		__antithesis_instrumentation__.Notify(693663)
		err := f(cmd, args)
		if err != nil {
			__antithesis_instrumentation__.Notify(693664)
			hint := errors.FlattenHints(err)
			cmd.Println("Error:", err.Error())
			if hint != "" {
				__antithesis_instrumentation__.Notify(693666)
				cmd.Println("Hint:", hint)
			} else {
				__antithesis_instrumentation__.Notify(693667)
			}
			__antithesis_instrumentation__.Notify(693665)
			exit.WithCode(exit.UnspecifiedError())
		} else {
			__antithesis_instrumentation__.Notify(693668)
		}
	}
}
