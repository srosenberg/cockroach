package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/errors"

type RelocateRange struct {
	Rows            *Select
	ToStoreID       Expr
	FromStoreID     Expr
	SubjectReplicas RelocateSubject
}

type RelocateSubject int8

const (
	RelocateLease RelocateSubject = iota

	RelocateVoters

	RelocateNonVoters
)

func (n *RelocateSubject) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603025)
	ctx.WriteString(n.String())
}

func (n RelocateSubject) String() string {
	__antithesis_instrumentation__.Notify(603026)
	switch n {
	case RelocateLease:
		__antithesis_instrumentation__.Notify(603027)
		return "LEASE"
	case RelocateVoters:
		__antithesis_instrumentation__.Notify(603028)
		return "VOTERS"
	case RelocateNonVoters:
		__antithesis_instrumentation__.Notify(603029)
		return "NONVOTERS"
	default:
		__antithesis_instrumentation__.Notify(603030)
		panic(errors.AssertionFailedf("programming error: unhandled case %d", int(n)))
	}
}

func (n *RelocateRange) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603031)
	ctx.WriteString("ALTER RANGE RELOCATE ")
	ctx.FormatNode(&n.SubjectReplicas)

	if n.SubjectReplicas != RelocateLease {
		__antithesis_instrumentation__.Notify(603033)
		ctx.WriteString(" FROM ")
		ctx.FormatNode(n.FromStoreID)
	} else {
		__antithesis_instrumentation__.Notify(603034)
	}
	__antithesis_instrumentation__.Notify(603032)
	ctx.WriteString(" TO ")
	ctx.FormatNode(n.ToStoreID)
	ctx.WriteString(" FOR ")
	ctx.FormatNode(n.Rows)
}
