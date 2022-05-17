package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type TenantID struct {
	roachpb.TenantID

	Specified bool
}

var _ NodeFormatter = (*TenantID)(nil)

func (t *TenantID) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614517)
	if ctx.flags.HasFlags(FmtHideConstants) || func() bool {
		__antithesis_instrumentation__.Notify(614518)
		return !t.IsSet() == true
	}() == true {
		__antithesis_instrumentation__.Notify(614519)

		ctx.WriteByte('_')
	} else {
		__antithesis_instrumentation__.Notify(614520)
		ctx.WriteString(strconv.FormatUint(t.ToUint64(), 10))
	}
}
