package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var sqlfmtCmd = &cobra.Command{
	Use:   "sqlfmt",
	Short: "format SQL statements",
	Long:  "Formats SQL statements from stdin to line length n.",
	RunE:  runSQLFmt,
}

func runSQLFmt(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(33964)
	if sqlfmtCtx.len < 1 {
		__antithesis_instrumentation__.Notify(33970)
		return errors.Errorf("line length must be > 0: %d", sqlfmtCtx.len)
	} else {
		__antithesis_instrumentation__.Notify(33971)
	}
	__antithesis_instrumentation__.Notify(33965)
	if sqlfmtCtx.tabWidth < 1 {
		__antithesis_instrumentation__.Notify(33972)
		return errors.Errorf("tab width must be > 0: %d", sqlfmtCtx.tabWidth)
	} else {
		__antithesis_instrumentation__.Notify(33973)
	}
	__antithesis_instrumentation__.Notify(33966)

	var sl parser.Statements
	if len(sqlfmtCtx.execStmts) != 0 {
		__antithesis_instrumentation__.Notify(33974)
		for _, exec := range sqlfmtCtx.execStmts {
			__antithesis_instrumentation__.Notify(33975)
			stmts, err := parser.Parse(exec)
			if err != nil {
				__antithesis_instrumentation__.Notify(33977)
				return err
			} else {
				__antithesis_instrumentation__.Notify(33978)
			}
			__antithesis_instrumentation__.Notify(33976)
			sl = append(sl, stmts...)
		}
	} else {
		__antithesis_instrumentation__.Notify(33979)
		in, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			__antithesis_instrumentation__.Notify(33981)
			return err
		} else {
			__antithesis_instrumentation__.Notify(33982)
		}
		__antithesis_instrumentation__.Notify(33980)
		sl, err = parser.Parse(string(in))
		if err != nil {
			__antithesis_instrumentation__.Notify(33983)
			return err
		} else {
			__antithesis_instrumentation__.Notify(33984)
		}
	}
	__antithesis_instrumentation__.Notify(33967)

	cfg := tree.DefaultPrettyCfg()
	cfg.UseTabs = !sqlfmtCtx.useSpaces
	cfg.LineWidth = sqlfmtCtx.len
	cfg.TabWidth = sqlfmtCtx.tabWidth
	cfg.Simplify = !sqlfmtCtx.noSimplify
	cfg.Align = tree.PrettyNoAlign
	cfg.JSONFmt = true
	if sqlfmtCtx.align {
		__antithesis_instrumentation__.Notify(33985)
		cfg.Align = tree.PrettyAlignAndDeindent
	} else {
		__antithesis_instrumentation__.Notify(33986)
	}
	__antithesis_instrumentation__.Notify(33968)

	for i := range sl {
		__antithesis_instrumentation__.Notify(33987)
		fmt.Print(cfg.Pretty(sl[i].AST))
		if len(sl) > 1 {
			__antithesis_instrumentation__.Notify(33989)
			fmt.Print(";")
		} else {
			__antithesis_instrumentation__.Notify(33990)
		}
		__antithesis_instrumentation__.Notify(33988)
		fmt.Println()
	}
	__antithesis_instrumentation__.Notify(33969)
	return nil
}
