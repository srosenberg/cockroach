package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

type DeclareCursor struct {
	Name        Name
	Select      *Select
	Binary      bool
	Scroll      CursorScrollOption
	Sensitivity CursorSensitivity
	Hold        bool
}

func (node *DeclareCursor) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605379)
	ctx.WriteString("DECLARE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" ")
	if node.Binary {
		__antithesis_instrumentation__.Notify(605384)
		ctx.WriteString("BINARY ")
	} else {
		__antithesis_instrumentation__.Notify(605385)
	}
	__antithesis_instrumentation__.Notify(605380)
	if node.Sensitivity != UnspecifiedSensitivity {
		__antithesis_instrumentation__.Notify(605386)
		ctx.WriteString(node.Sensitivity.String())
		ctx.WriteString(" ")
	} else {
		__antithesis_instrumentation__.Notify(605387)
	}
	__antithesis_instrumentation__.Notify(605381)
	if node.Scroll != UnspecifiedScroll {
		__antithesis_instrumentation__.Notify(605388)
		ctx.WriteString(node.Scroll.String())
		ctx.WriteString(" ")
	} else {
		__antithesis_instrumentation__.Notify(605389)
	}
	__antithesis_instrumentation__.Notify(605382)
	ctx.WriteString("CURSOR ")
	if node.Hold {
		__antithesis_instrumentation__.Notify(605390)
		ctx.WriteString("WITH HOLD ")
	} else {
		__antithesis_instrumentation__.Notify(605391)
	}
	__antithesis_instrumentation__.Notify(605383)
	ctx.WriteString("FOR ")
	ctx.FormatNode(node.Select)
}

type CursorScrollOption int8

const (
	UnspecifiedScroll CursorScrollOption = iota

	Scroll

	NoScroll
)

func (o CursorScrollOption) String() string {
	__antithesis_instrumentation__.Notify(605392)
	switch o {
	case Scroll:
		__antithesis_instrumentation__.Notify(605394)
		return "SCROLL"
	case NoScroll:
		__antithesis_instrumentation__.Notify(605395)
		return "NO SCROLL"
	default:
		__antithesis_instrumentation__.Notify(605396)
	}
	__antithesis_instrumentation__.Notify(605393)
	return ""
}

type CursorSensitivity int

const (
	UnspecifiedSensitivity CursorSensitivity = iota

	Insensitive

	Asensitive
)

func (o CursorSensitivity) String() string {
	__antithesis_instrumentation__.Notify(605397)
	switch o {
	case Insensitive:
		__antithesis_instrumentation__.Notify(605399)
		return "INSENSITIVE"
	case Asensitive:
		__antithesis_instrumentation__.Notify(605400)
		return "ASENSITIVE"
	default:
		__antithesis_instrumentation__.Notify(605401)
	}
	__antithesis_instrumentation__.Notify(605398)
	return ""
}

type CursorStmt struct {
	Name      Name
	FetchType FetchType
	Count     int64
}

type FetchCursor struct {
	CursorStmt
}

type MoveCursor struct {
	CursorStmt
}

type FetchType int

const (
	FetchNormal FetchType = iota

	FetchRelative

	FetchAbsolute

	FetchFirst

	FetchLast

	FetchAll

	FetchBackwardAll
)

func (o FetchType) String() string {
	__antithesis_instrumentation__.Notify(605402)
	switch o {
	case FetchNormal:
		__antithesis_instrumentation__.Notify(605404)
		return ""
	case FetchRelative:
		__antithesis_instrumentation__.Notify(605405)
		return "RELATIVE"
	case FetchAbsolute:
		__antithesis_instrumentation__.Notify(605406)
		return "ABSOLUTE"
	case FetchFirst:
		__antithesis_instrumentation__.Notify(605407)
		return "FIRST"
	case FetchLast:
		__antithesis_instrumentation__.Notify(605408)
		return "LAST"
	case FetchAll:
		__antithesis_instrumentation__.Notify(605409)
		return "ALL"
	case FetchBackwardAll:
		__antithesis_instrumentation__.Notify(605410)
		return "BACKWARD ALL"
	default:
		__antithesis_instrumentation__.Notify(605411)
	}
	__antithesis_instrumentation__.Notify(605403)
	return ""
}

func (o FetchType) HasCount() bool {
	__antithesis_instrumentation__.Notify(605412)
	switch o {
	case FetchNormal, FetchRelative, FetchAbsolute:
		__antithesis_instrumentation__.Notify(605414)
		return true
	default:
		__antithesis_instrumentation__.Notify(605415)
	}
	__antithesis_instrumentation__.Notify(605413)
	return false
}

func (c CursorStmt) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605416)
	fetchType := c.FetchType.String()
	if fetchType != "" {
		__antithesis_instrumentation__.Notify(605419)
		ctx.WriteString(fetchType)
		ctx.WriteString(" ")
	} else {
		__antithesis_instrumentation__.Notify(605420)
	}
	__antithesis_instrumentation__.Notify(605417)
	if c.FetchType.HasCount() {
		__antithesis_instrumentation__.Notify(605421)
		if ctx.HasFlags(FmtHideConstants) {
			__antithesis_instrumentation__.Notify(605423)
			ctx.WriteByte('0')
		} else {
			__antithesis_instrumentation__.Notify(605424)
			ctx.WriteString(strconv.Itoa(int(c.Count)))
		}
		__antithesis_instrumentation__.Notify(605422)
		ctx.WriteString(" ")
	} else {
		__antithesis_instrumentation__.Notify(605425)
	}
	__antithesis_instrumentation__.Notify(605418)
	ctx.FormatNode(&c.Name)
}

func (f FetchCursor) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605426)
	ctx.WriteString("FETCH ")
	f.CursorStmt.Format(ctx)
}

func (m MoveCursor) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605427)
	ctx.WriteString("MOVE ")
	m.CursorStmt.Format(ctx)
}

type CloseCursor struct {
	Name Name
	All  bool
}

func (c CloseCursor) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605428)
	ctx.WriteString("CLOSE ")
	if c.All {
		__antithesis_instrumentation__.Notify(605429)
		ctx.WriteString("ALL")
	} else {
		__antithesis_instrumentation__.Notify(605430)
		ctx.FormatNode(&c.Name)
	}
}
