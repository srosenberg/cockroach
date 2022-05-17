package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
)

type CopyFrom struct {
	Table   TableName
	Columns NameList
	Stdin   bool
	Options CopyOptions
}

type CopyOptions struct {
	Destination Expr
	CopyFormat  CopyFormat
	Delimiter   Expr
	Null        Expr
	Escape      *StrVal
}

var _ NodeFormatter = &CopyOptions{}

func (node *CopyFrom) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604644)
	ctx.WriteString("COPY ")
	ctx.FormatNode(&node.Table)
	if len(node.Columns) > 0 {
		__antithesis_instrumentation__.Notify(604647)
		ctx.WriteString(" (")
		ctx.FormatNode(&node.Columns)
		ctx.WriteString(")")
	} else {
		__antithesis_instrumentation__.Notify(604648)
	}
	__antithesis_instrumentation__.Notify(604645)
	ctx.WriteString(" FROM ")
	if node.Stdin {
		__antithesis_instrumentation__.Notify(604649)
		ctx.WriteString("STDIN")
	} else {
		__antithesis_instrumentation__.Notify(604650)
	}
	__antithesis_instrumentation__.Notify(604646)
	if !node.Options.IsDefault() {
		__antithesis_instrumentation__.Notify(604651)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(&node.Options)
	} else {
		__antithesis_instrumentation__.Notify(604652)
	}
}

func (o *CopyOptions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(604653)
	var addSep bool
	maybeAddSep := func() {
		__antithesis_instrumentation__.Notify(604659)
		if addSep {
			__antithesis_instrumentation__.Notify(604661)
			ctx.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(604662)
		}
		__antithesis_instrumentation__.Notify(604660)
		addSep = true
	}
	__antithesis_instrumentation__.Notify(604654)
	if o.CopyFormat != CopyFormatText {
		__antithesis_instrumentation__.Notify(604663)
		maybeAddSep()
		switch o.CopyFormat {
		case CopyFormatBinary:
			__antithesis_instrumentation__.Notify(604664)
			ctx.WriteString("BINARY")
		case CopyFormatCSV:
			__antithesis_instrumentation__.Notify(604665)
			ctx.WriteString("CSV")
		default:
			__antithesis_instrumentation__.Notify(604666)
		}
	} else {
		__antithesis_instrumentation__.Notify(604667)
	}
	__antithesis_instrumentation__.Notify(604655)
	if o.Delimiter != nil {
		__antithesis_instrumentation__.Notify(604668)
		maybeAddSep()
		ctx.WriteString("DELIMITER ")
		ctx.FormatNode(o.Delimiter)
		addSep = true
	} else {
		__antithesis_instrumentation__.Notify(604669)
	}
	__antithesis_instrumentation__.Notify(604656)
	if o.Null != nil {
		__antithesis_instrumentation__.Notify(604670)
		maybeAddSep()
		ctx.WriteString("NULL ")
		ctx.FormatNode(o.Null)
		addSep = true
	} else {
		__antithesis_instrumentation__.Notify(604671)
	}
	__antithesis_instrumentation__.Notify(604657)
	if o.Destination != nil {
		__antithesis_instrumentation__.Notify(604672)
		maybeAddSep()

		ctx.WriteString("destination = ")
		ctx.FormatNode(o.Destination)
		addSep = true
	} else {
		__antithesis_instrumentation__.Notify(604673)
	}
	__antithesis_instrumentation__.Notify(604658)
	if o.Escape != nil {
		__antithesis_instrumentation__.Notify(604674)
		maybeAddSep()
		ctx.WriteString("ESCAPE ")
		ctx.FormatNode(o.Escape)
	} else {
		__antithesis_instrumentation__.Notify(604675)
	}
}

func (o CopyOptions) IsDefault() bool {
	__antithesis_instrumentation__.Notify(604676)
	return o == CopyOptions{}
}

func (o *CopyOptions) CombineWith(other *CopyOptions) error {
	__antithesis_instrumentation__.Notify(604677)
	if other.Destination != nil {
		__antithesis_instrumentation__.Notify(604683)
		if o.Destination != nil {
			__antithesis_instrumentation__.Notify(604685)
			return pgerror.Newf(pgcode.Syntax, "destination option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(604686)
		}
		__antithesis_instrumentation__.Notify(604684)
		o.Destination = other.Destination
	} else {
		__antithesis_instrumentation__.Notify(604687)
	}
	__antithesis_instrumentation__.Notify(604678)
	if other.CopyFormat != CopyFormatText {
		__antithesis_instrumentation__.Notify(604688)
		if o.CopyFormat != CopyFormatText {
			__antithesis_instrumentation__.Notify(604690)
			return pgerror.Newf(pgcode.Syntax, "format option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(604691)
		}
		__antithesis_instrumentation__.Notify(604689)
		o.CopyFormat = other.CopyFormat
	} else {
		__antithesis_instrumentation__.Notify(604692)
	}
	__antithesis_instrumentation__.Notify(604679)
	if other.Delimiter != nil {
		__antithesis_instrumentation__.Notify(604693)
		if o.Delimiter != nil {
			__antithesis_instrumentation__.Notify(604695)
			return pgerror.Newf(pgcode.Syntax, "delimiter option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(604696)
		}
		__antithesis_instrumentation__.Notify(604694)
		o.Delimiter = other.Delimiter
	} else {
		__antithesis_instrumentation__.Notify(604697)
	}
	__antithesis_instrumentation__.Notify(604680)
	if other.Null != nil {
		__antithesis_instrumentation__.Notify(604698)
		if o.Null != nil {
			__antithesis_instrumentation__.Notify(604700)
			return pgerror.Newf(pgcode.Syntax, "null option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(604701)
		}
		__antithesis_instrumentation__.Notify(604699)
		o.Null = other.Null
	} else {
		__antithesis_instrumentation__.Notify(604702)
	}
	__antithesis_instrumentation__.Notify(604681)
	if other.Escape != nil {
		__antithesis_instrumentation__.Notify(604703)
		if o.Escape != nil {
			__antithesis_instrumentation__.Notify(604705)
			return pgerror.Newf(pgcode.Syntax, "escape option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(604706)
		}
		__antithesis_instrumentation__.Notify(604704)
		o.Escape = other.Escape
	} else {
		__antithesis_instrumentation__.Notify(604707)
	}
	__antithesis_instrumentation__.Notify(604682)
	return nil
}

type CopyFormat int

const (
	CopyFormatText CopyFormat = iota
	CopyFormatBinary
	CopyFormatCSV
)
