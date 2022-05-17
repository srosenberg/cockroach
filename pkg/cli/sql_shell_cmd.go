package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlshell"
	"github.com/cockroachdb/cockroach/pkg/server/pgurl"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var sqlShellCmd = &cobra.Command{
	Use:   "sql [options]",
	Short: "open a sql shell",
	Long: `
Open a sql shell running against a cockroach database.
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runTerm),
}

func runTerm(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(33947)
	closeFn, err := sqlCtx.Open(os.Stdin)
	if err != nil {
		__antithesis_instrumentation__.Notify(33952)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33953)
	}
	__antithesis_instrumentation__.Notify(33948)
	defer closeFn()

	if cliCtx.IsInteractive {
		__antithesis_instrumentation__.Notify(33954)

		const welcomeMessage = `#
# Welcome to the CockroachDB SQL shell.
# All statements must be terminated by a semicolon.
# To exit, type: \q.
#
`
		fmt.Print(welcomeMessage)
	} else {
		__antithesis_instrumentation__.Notify(33955)
	}
	__antithesis_instrumentation__.Notify(33949)

	conn, err := makeSQLClient(catconstants.InternalSQLAppName, useDefaultDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(33956)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33957)
	}
	__antithesis_instrumentation__.Notify(33950)
	defer func() {
		__antithesis_instrumentation__.Notify(33958)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(33951)

	sqlCtx.ShellCtx.ParseURL = makeURLParser(cmd)
	return sqlCtx.Run(conn)
}

func makeURLParser(cmd *cobra.Command) clisqlshell.URLParser {
	__antithesis_instrumentation__.Notify(33959)
	return func(url string) (*pgurl.URL, error) {
		__antithesis_instrumentation__.Notify(33960)

		up := urlParser{cmd: cmd, cliCtx: &cliCtx}
		if err := up.setInternal(url, false); err != nil {
			__antithesis_instrumentation__.Notify(33962)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(33963)
		}
		__antithesis_instrumentation__.Notify(33961)
		return cliCtx.sqlConnURL, nil
	}
}
