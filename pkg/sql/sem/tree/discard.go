package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Discard struct {
	Mode DiscardMode
}

var _ Statement = &Discard{}

type DiscardMode int

const (
	DiscardModeAll DiscardMode = iota
)

func (node *Discard) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607614)
	switch node.Mode {
	case DiscardModeAll:
		__antithesis_instrumentation__.Notify(607615)
		ctx.WriteString("DISCARD ALL")
	default:
		__antithesis_instrumentation__.Notify(607616)
	}
}

func (node *Discard) String() string {
	__antithesis_instrumentation__.Notify(607617)
	return AsString(node)
}
