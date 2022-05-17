package clisqlshell

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"text/tabwriter"

	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/errors"
)

func (c *cliState) handleStatementDiag(
	args []string, loopState, errState cliStateEnum,
) (resState cliStateEnum) {
	__antithesis_instrumentation__.Notify(30194)
	var cmd string
	if len(args) > 0 {
		__antithesis_instrumentation__.Notify(30198)
		cmd = args[0]
		args = args[1:]
	} else {
		__antithesis_instrumentation__.Notify(30199)
	}
	__antithesis_instrumentation__.Notify(30195)

	var cmdErr error
	switch cmd {
	case "list":
		__antithesis_instrumentation__.Notify(30200)
		if len(args) > 0 {
			__antithesis_instrumentation__.Notify(30207)
			return c.invalidSyntax(errState)
		} else {
			__antithesis_instrumentation__.Notify(30208)
		}
		__antithesis_instrumentation__.Notify(30201)
		cmdErr = c.statementDiagList()

	case "download":
		__antithesis_instrumentation__.Notify(30202)
		if len(args) < 1 || func() bool {
			__antithesis_instrumentation__.Notify(30209)
			return len(args) > 2 == true
		}() == true {
			__antithesis_instrumentation__.Notify(30210)
			return c.invalidSyntax(errState)
		} else {
			__antithesis_instrumentation__.Notify(30211)
		}
		__antithesis_instrumentation__.Notify(30203)
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			__antithesis_instrumentation__.Notify(30212)
			return c.invalidSyntaxf(
				errState, "%s", errors.Wrapf(err, "%q is not a valid bundle ID", args[1]),
			)
		} else {
			__antithesis_instrumentation__.Notify(30213)
		}
		__antithesis_instrumentation__.Notify(30204)
		var filename string
		if len(args) > 1 {
			__antithesis_instrumentation__.Notify(30214)
			filename = args[1]
		} else {
			__antithesis_instrumentation__.Notify(30215)
			filename = fmt.Sprintf("stmt-bundle-%d.zip", id)
		}
		__antithesis_instrumentation__.Notify(30205)
		cmdErr = clisqlclient.StmtDiagDownloadBundle(
			context.Background(), c.conn, id, filename)
		if cmdErr == nil {
			__antithesis_instrumentation__.Notify(30216)
			fmt.Fprintf(c.iCtx.stdout, "Bundle saved to %q\n", filename)
		} else {
			__antithesis_instrumentation__.Notify(30217)
		}

	default:
		__antithesis_instrumentation__.Notify(30206)
		return c.invalidSyntax(errState)
	}
	__antithesis_instrumentation__.Notify(30196)

	if cmdErr != nil {
		__antithesis_instrumentation__.Notify(30218)
		fmt.Fprintln(c.iCtx.stderr, cmdErr)
		c.exitErr = cmdErr
		return errState
	} else {
		__antithesis_instrumentation__.Notify(30219)
	}
	__antithesis_instrumentation__.Notify(30197)
	return loopState
}

func (c *cliState) statementDiagList() error {
	__antithesis_instrumentation__.Notify(30220)
	const timeFmt = "2006-01-02 15:04:05 MST"

	bundles, err := clisqlclient.StmtDiagListBundles(context.Background(), c.conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(30223)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30224)
	}
	__antithesis_instrumentation__.Notify(30221)

	if len(bundles) == 0 {
		__antithesis_instrumentation__.Notify(30225)
		fmt.Fprintf(c.iCtx.stdout, "No statement diagnostics bundles available.\n")
	} else {
		__antithesis_instrumentation__.Notify(30226)
		var buf bytes.Buffer
		fmt.Fprintf(c.iCtx.stdout, "Statement diagnostics bundles:\n")
		w := tabwriter.NewWriter(&buf, 4, 0, 2, ' ', 0)
		fmt.Fprint(w, "  ID\tCollection time\tStatement\n")
		for _, b := range bundles {
			__antithesis_instrumentation__.Notify(30228)
			fmt.Fprintf(w, "  %d\t%s\t%s\n", b.ID, b.CollectedAt.UTC().Format(timeFmt), b.Statement)
		}
		__antithesis_instrumentation__.Notify(30227)
		_ = w.Flush()
		_, _ = buf.WriteTo(c.iCtx.stdout)
	}
	__antithesis_instrumentation__.Notify(30222)
	fmt.Fprintln(c.iCtx.stdout)

	return nil
}
