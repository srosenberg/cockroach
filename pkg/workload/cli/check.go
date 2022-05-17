package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	AddSubCmd(func(userFacing bool) *cobra.Command {
		var checkCmd = SetCmdDefaults(&cobra.Command{
			Use:   `check`,
			Short: `check a running cluster's data for consistency`,
		})
		for _, meta := range workload.Registered() {
			gen := meta.New()
			if hooks, ok := gen.(workload.Hookser); !ok || hooks.Hooks().CheckConsistency == nil {
				continue
			}

			var genFlags *pflag.FlagSet
			if f, ok := gen.(workload.Flagser); ok {
				genFlags = f.Flags().FlagSet

				for flagName, meta := range f.Flags().Meta {
					if meta.RuntimeOnly && !meta.CheckConsistencyOnly {
						_ = genFlags.MarkHidden(flagName)
					}
				}
			}

			genCheckCmd := SetCmdDefaults(&cobra.Command{
				Use:  meta.Name + ` [CRDB URI]`,
				Args: cobra.RangeArgs(0, 1),
			})
			genCheckCmd.Flags().AddFlagSet(genFlags)
			genCheckCmd.Run = CmdHelper(gen, check)
			checkCmd.AddCommand(genCheckCmd)
		}
		return checkCmd
	})
}

func check(gen workload.Generator, urls []string, dbName string) error {
	__antithesis_instrumentation__.Notify(693632)
	ctx := context.Background()

	var fn func(context.Context, *gosql.DB) error
	if hooks, ok := gen.(workload.Hookser); ok {
		__antithesis_instrumentation__.Notify(693637)
		fn = hooks.Hooks().CheckConsistency
	} else {
		__antithesis_instrumentation__.Notify(693638)
	}
	__antithesis_instrumentation__.Notify(693633)
	if fn == nil {
		__antithesis_instrumentation__.Notify(693639)
		return errors.Errorf(`no consistency checks are defined for %s`, gen.Meta().Name)
	} else {
		__antithesis_instrumentation__.Notify(693640)
	}
	__antithesis_instrumentation__.Notify(693634)

	sqlDB, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(693641)
		return err
	} else {
		__antithesis_instrumentation__.Notify(693642)
	}
	__antithesis_instrumentation__.Notify(693635)
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil {
		__antithesis_instrumentation__.Notify(693643)
		return err
	} else {
		__antithesis_instrumentation__.Notify(693644)
	}
	__antithesis_instrumentation__.Notify(693636)
	return fn(ctx, sqlDB)
}
