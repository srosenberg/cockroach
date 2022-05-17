package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func crdbInternalSendNotice(
	ctx *tree.EvalContext, severity string, msg string,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(601656)
	if ctx.ClientNoticeSender == nil {
		__antithesis_instrumentation__.Notify(601658)
		return nil, errors.AssertionFailedf("notice sender not set")
	} else {
		__antithesis_instrumentation__.Notify(601659)
	}
	__antithesis_instrumentation__.Notify(601657)
	ctx.ClientNoticeSender.BufferClientNotice(
		ctx.Context,
		pgnotice.NewWithSeverityf(strings.ToUpper(severity), "%s", msg),
	)
	return tree.NewDInt(0), nil
}
