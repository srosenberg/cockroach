package sqlutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func VerifyStatementPrettyRoundtrip(t *testing.T, sql string) {
	__antithesis_instrumentation__.Notify(646180)
	t.Helper()

	stmts, err := parser.Parse(sql)
	if err != nil {
		__antithesis_instrumentation__.Notify(646182)
		t.Fatalf("%s: %s", err, sql)
	} else {
		__antithesis_instrumentation__.Notify(646183)
	}
	__antithesis_instrumentation__.Notify(646181)
	cfg := tree.DefaultPrettyCfg()

	cfg.Simplify = false
	for i := range stmts {
		__antithesis_instrumentation__.Notify(646184)

		origStmt := stmts[i].AST

		prettyStmt := cfg.Pretty(origStmt)
		parsedPretty, err := parser.ParseOne(prettyStmt)
		if err != nil {
			__antithesis_instrumentation__.Notify(646186)
			t.Fatalf("%s: %s", err, prettyStmt)
		} else {
			__antithesis_instrumentation__.Notify(646187)
		}
		__antithesis_instrumentation__.Notify(646185)
		prettyFormatted := tree.AsStringWithFlags(parsedPretty.AST, tree.FmtSimple)
		origFormatted := tree.AsStringWithFlags(origStmt, tree.FmtParsable)
		if prettyFormatted != origFormatted {
			__antithesis_instrumentation__.Notify(646188)

			reparsedStmt, err := parser.ParseOne(origFormatted)
			if err != nil {
				__antithesis_instrumentation__.Notify(646190)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(646191)
			}
			__antithesis_instrumentation__.Notify(646189)
			origFormatted = tree.AsStringWithFlags(reparsedStmt.AST, tree.FmtParsable)
			if prettyFormatted != origFormatted {
				__antithesis_instrumentation__.Notify(646192)
				t.Fatalf("orig formatted != pretty formatted\norig SQL: %q\norig formatted: %q\npretty printed: %s\npretty formatted: %q",
					sql,
					origFormatted,
					prettyStmt,
					prettyFormatted,
				)
			} else {
				__antithesis_instrumentation__.Notify(646193)
			}
		} else {
			__antithesis_instrumentation__.Notify(646194)
		}
	}
}
