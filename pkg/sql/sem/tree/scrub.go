package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "fmt"

type ScrubType int

const (
	ScrubTable = iota

	ScrubDatabase = iota
)

type Scrub struct {
	Typ     ScrubType
	Options ScrubOptions

	Table *UnresolvedObjectName

	Database Name
	AsOf     AsOfClause
}

func (n *Scrub) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613039)
	ctx.WriteString("EXPERIMENTAL SCRUB ")
	switch n.Typ {
	case ScrubTable:
		__antithesis_instrumentation__.Notify(613042)
		ctx.WriteString("TABLE ")
		ctx.FormatNode(n.Table)
	case ScrubDatabase:
		__antithesis_instrumentation__.Notify(613043)
		ctx.WriteString("DATABASE ")
		ctx.FormatNode(&n.Database)
	default:
		__antithesis_instrumentation__.Notify(613044)
		panic("Unhandled ScrubType")
	}
	__antithesis_instrumentation__.Notify(613040)

	if n.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(613045)
		ctx.WriteByte(' ')
		ctx.FormatNode(&n.AsOf)
	} else {
		__antithesis_instrumentation__.Notify(613046)
	}
	__antithesis_instrumentation__.Notify(613041)

	if len(n.Options) > 0 {
		__antithesis_instrumentation__.Notify(613047)
		ctx.WriteString(" WITH OPTIONS ")
		ctx.FormatNode(&n.Options)
	} else {
		__antithesis_instrumentation__.Notify(613048)
	}
}

type ScrubOptions []ScrubOption

func (n *ScrubOptions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613049)
	for i, option := range *n {
		__antithesis_instrumentation__.Notify(613050)
		if i > 0 {
			__antithesis_instrumentation__.Notify(613052)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(613053)
		}
		__antithesis_instrumentation__.Notify(613051)
		ctx.FormatNode(option)
	}
}

func (n *ScrubOptions) String() string {
	__antithesis_instrumentation__.Notify(613054)
	return AsString(n)
}

type ScrubOption interface {
	fmt.Stringer
	NodeFormatter

	scrubOptionType()
}

func (*ScrubOptionIndex) scrubOptionType()      { __antithesis_instrumentation__.Notify(613055) }
func (*ScrubOptionPhysical) scrubOptionType()   { __antithesis_instrumentation__.Notify(613056) }
func (*ScrubOptionConstraint) scrubOptionType() { __antithesis_instrumentation__.Notify(613057) }

func (n *ScrubOptionIndex) String() string {
	__antithesis_instrumentation__.Notify(613058)
	return AsString(n)
}
func (n *ScrubOptionPhysical) String() string {
	__antithesis_instrumentation__.Notify(613059)
	return AsString(n)
}
func (n *ScrubOptionConstraint) String() string {
	__antithesis_instrumentation__.Notify(613060)
	return AsString(n)
}

type ScrubOptionIndex struct {
	IndexNames NameList
}

func (n *ScrubOptionIndex) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613061)
	ctx.WriteString("INDEX ")
	if n.IndexNames != nil {
		__antithesis_instrumentation__.Notify(613062)
		ctx.WriteByte('(')
		ctx.FormatNode(&n.IndexNames)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(613063)
		ctx.WriteString("ALL")
	}
}

type ScrubOptionPhysical struct{}

func (n *ScrubOptionPhysical) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613064)
	ctx.WriteString("PHYSICAL")
}

type ScrubOptionConstraint struct {
	ConstraintNames NameList
}

func (n *ScrubOptionConstraint) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613065)
	ctx.WriteString("CONSTRAINT ")
	if n.ConstraintNames != nil {
		__antithesis_instrumentation__.Notify(613066)
		ctx.WriteByte('(')
		ctx.FormatNode(&n.ConstraintNames)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(613067)
		ctx.WriteString("ALL")
	}
}
