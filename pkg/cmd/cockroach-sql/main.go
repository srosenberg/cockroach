// cockroach-sql is an entry point for a CockroachDB binary that only
// includes the SQL shell and does not include any server components.
// It also does not include CCL features.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"

	"github.com/cockroachdb/cockroach/pkg/cli/clicfg"
	"github.com/cockroachdb/cockroach/pkg/cli/clierror"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlcfg"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var url = "postgresql://localhost:26257/defaultdb"
var cfg = func() *clisqlcfg.Context {
	__antithesis_instrumentation__.Notify(38144)
	cliCtx := &clicfg.Context{}
	c := &clisqlcfg.Context{
		CliCtx:  cliCtx,
		ConnCtx: &clisqlclient.Context{CliCtx: cliCtx},
		ExecCtx: &clisqlexec.Context{CliCtx: cliCtx},
	}
	c.LoadDefaults(os.Stdout, os.Stderr)
	return c
}()

func main() {
	__antithesis_instrumentation__.Notify(38145)

	cfg.Database = "defaultdb"

	cfg.User = "root"
	cfg.ApplicationName = "$ cockroach sql"
	cfg.ConnectTimeout = 15

	f := sqlCmd.PersistentFlags()
	f.StringVar(&url, cliflags.URL.Name, url, "Connection URL.")
	f.StringVarP(&cfg.Database, cliflags.Database.Name, cliflags.Database.Shorthand, cfg.Database, cliflags.Database.Description)
	f.StringVarP(&cfg.User, cliflags.User.Name, cliflags.User.Shorthand, cfg.User, cliflags.User.Description)
	f.VarP(&cfg.ShellCtx.ExecStmts, cliflags.Execute.Name, cliflags.Execute.Shorthand, cliflags.Execute.Description)
	f.StringVarP(&cfg.InputFile, cliflags.File.Name, cliflags.File.Shorthand, cfg.InputFile, cliflags.File.Description)
	f.DurationVar(&cfg.ShellCtx.RepeatDelay, cliflags.Watch.Name, cfg.ShellCtx.RepeatDelay, cliflags.Watch.Description)
	f.Var(&cfg.SafeUpdates, cliflags.SafeUpdates.Name, cliflags.SafeUpdates.Description)
	f.BoolVar(&cfg.ReadOnly, cliflags.ReadOnly.Name, cfg.ReadOnly, cliflags.ReadOnly.Description)
	f.Lookup(cliflags.SafeUpdates.Name).NoOptDefVal = "true"
	f.BoolVar(&cfg.ConnCtx.DebugMode, cliflags.CliDebugMode.Name, cfg.ConnCtx.DebugMode, cliflags.CliDebugMode.Description)
	f.BoolVar(&cfg.ConnCtx.Echo, cliflags.EchoSQL.Name, cfg.ConnCtx.Echo, cliflags.EchoSQL.Description)

	errCode := exit.Success()
	if err := sqlCmd.Execute(); err != nil {
		__antithesis_instrumentation__.Notify(38147)
		clierror.OutputError(os.Stderr, err, true, false)

		errCode = exit.UnspecifiedError()
		var cliErr *clierror.Error
		if errors.As(err, &cliErr) {
			__antithesis_instrumentation__.Notify(38148)
			errCode = cliErr.GetExitCode()
		} else {
			__antithesis_instrumentation__.Notify(38149)
		}
	} else {
		__antithesis_instrumentation__.Notify(38150)
	}
	__antithesis_instrumentation__.Notify(38146)
	exit.WithCode(errCode)
}

var sqlCmd = &cobra.Command{
	Use:   "sql [[--url] <url>] [-f <inputfile>] [-e <stmt>]",
	Short: "start the SQL shell",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSQL,

	SilenceUsage: true,

	SilenceErrors: true,
}

func runSQL(_ *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(38151)
	if url == "" && func() bool {
		__antithesis_instrumentation__.Notify(38158)
		return len(args) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(38159)
		url = args[0]
	} else {
		__antithesis_instrumentation__.Notify(38160)
	}
	__antithesis_instrumentation__.Notify(38152)
	if url == "" {
		__antithesis_instrumentation__.Notify(38161)
		return errors.New("no connection URL specified")
	} else {
		__antithesis_instrumentation__.Notify(38162)
	}
	__antithesis_instrumentation__.Notify(38153)

	closeFn, err := cfg.Open(os.Stdin)
	if err != nil {
		__antithesis_instrumentation__.Notify(38163)
		return err
	} else {
		__antithesis_instrumentation__.Notify(38164)
	}
	__antithesis_instrumentation__.Notify(38154)
	defer closeFn()

	if cfg.CliCtx.IsInteractive {
		__antithesis_instrumentation__.Notify(38165)
		const welcomeMessage = `#
# Welcome to the CockroachDB SQL shell.
# All statements must be terminated by a semicolon.
# To exit, type: \q.
#
`
		fmt.Print(welcomeMessage)
	} else {
		__antithesis_instrumentation__.Notify(38166)
	}
	__antithesis_instrumentation__.Notify(38155)

	conn, err := cfg.MakeConn(url)
	if err != nil {
		__antithesis_instrumentation__.Notify(38167)
		return err
	} else {
		__antithesis_instrumentation__.Notify(38168)
	}
	__antithesis_instrumentation__.Notify(38156)
	defer func() {
		__antithesis_instrumentation__.Notify(38169)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(38157)

	return cfg.Run(conn)
}
