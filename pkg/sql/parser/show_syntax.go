package parser

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func RunShowSyntax(
	ctx context.Context,
	stmt string,
	report func(ctx context.Context, field, msg string),
	reportErr func(ctx context.Context, err error),
) {
	__antithesis_instrumentation__.Notify(552619)
	if strings.HasSuffix(stmt, "??") {
		__antithesis_instrumentation__.Notify(552621)

		prefix := strings.ToUpper(strings.TrimSpace(stmt[:len(stmt)-2]))
		if h, ok := HelpMessages[prefix]; ok {
			__antithesis_instrumentation__.Notify(552622)
			msg := HelpMessage{Command: prefix, HelpMessageBody: h}
			msgs := msg.String()
			err := errors.WithHint(pgerror.WithCandidateCode(errors.New(specialHelpErrorPrefix), pgcode.Syntax), msgs)
			doErr(ctx, report, reportErr, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(552623)
		}
	} else {
		__antithesis_instrumentation__.Notify(552624)
	}
	__antithesis_instrumentation__.Notify(552620)

	stmts, err := Parse(stmt)
	if err != nil {
		__antithesis_instrumentation__.Notify(552625)
		doErr(ctx, report, reportErr, err)
	} else {
		__antithesis_instrumentation__.Notify(552626)
		for i := range stmts {
			__antithesis_instrumentation__.Notify(552627)
			report(ctx, "sql", tree.AsStringWithFlags(stmts[i].AST, tree.FmtParsable))
		}
	}
}

func doErr(
	ctx context.Context,
	report func(ctx context.Context, field, msg string),
	reportErr func(ctx context.Context, err error),
	err error,
) {
	__antithesis_instrumentation__.Notify(552628)
	if reportErr != nil {
		__antithesis_instrumentation__.Notify(552632)
		reportErr(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(552633)
	}
	__antithesis_instrumentation__.Notify(552629)

	pqErr := pgerror.Flatten(err)
	report(ctx, "error", pqErr.Message)
	report(ctx, "code", pqErr.Code)
	if pqErr.Source != nil {
		__antithesis_instrumentation__.Notify(552634)
		if pqErr.Source.File != "" {
			__antithesis_instrumentation__.Notify(552637)
			report(ctx, "file", pqErr.Source.File)
		} else {
			__antithesis_instrumentation__.Notify(552638)
		}
		__antithesis_instrumentation__.Notify(552635)
		if pqErr.Source.Line > 0 {
			__antithesis_instrumentation__.Notify(552639)
			report(ctx, "line", fmt.Sprintf("%d", pqErr.Source.Line))
		} else {
			__antithesis_instrumentation__.Notify(552640)
		}
		__antithesis_instrumentation__.Notify(552636)
		if pqErr.Source.Function != "" {
			__antithesis_instrumentation__.Notify(552641)
			report(ctx, "function", pqErr.Source.Function)
		} else {
			__antithesis_instrumentation__.Notify(552642)
		}
	} else {
		__antithesis_instrumentation__.Notify(552643)
	}
	__antithesis_instrumentation__.Notify(552630)
	if pqErr.Detail != "" {
		__antithesis_instrumentation__.Notify(552644)
		report(ctx, "detail", pqErr.Detail)
	} else {
		__antithesis_instrumentation__.Notify(552645)
	}
	__antithesis_instrumentation__.Notify(552631)
	if pqErr.Hint != "" {
		__antithesis_instrumentation__.Notify(552646)
		report(ctx, "hint", pqErr.Hint)
	} else {
		__antithesis_instrumentation__.Notify(552647)
	}
}
