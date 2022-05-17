package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/lib/pq/oid"
)

func ResolveBlankPaddedChar(s string, t *types.T) string {
	__antithesis_instrumentation__.Notify(611667)
	if t.Oid() == oid.T_bpchar && func() bool {
		__antithesis_instrumentation__.Notify(611669)
		return len(s) < int(t.Width()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(611670)

		return fmt.Sprintf("%-*v", t.Width(), s)
	} else {
		__antithesis_instrumentation__.Notify(611671)
	}
	__antithesis_instrumentation__.Notify(611668)
	return s
}

func (d *DTuple) pgwireFormat(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(611672)

	ctx.WriteByte('(')
	comma := ""
	for i, v := range d.D {
		__antithesis_instrumentation__.Notify(611674)
		ctx.WriteString(comma)
		t := d.ResolvedType().TupleContents()[i]
		switch dv := UnwrapDatum(nil, v).(type) {
		case dNull:
			__antithesis_instrumentation__.Notify(611676)
		case *DString:
			__antithesis_instrumentation__.Notify(611677)
			s := ResolveBlankPaddedChar(string(*dv), t)
			pgwireFormatStringInTuple(&ctx.Buffer, s)
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(611678)
			s := ResolveBlankPaddedChar(dv.Contents, t)
			pgwireFormatStringInTuple(&ctx.Buffer, s)

		case *DBytes:
			__antithesis_instrumentation__.Notify(611679)
			ctx.FormatNode(dv)
		case *DJSON:
			__antithesis_instrumentation__.Notify(611680)
			var buf bytes.Buffer
			dv.JSON.Format(&buf)
			pgwireFormatStringInTuple(&ctx.Buffer, buf.String())
		default:
			__antithesis_instrumentation__.Notify(611681)
			s := AsStringWithFlags(v, ctx.flags, FmtDataConversionConfig(ctx.dataConversionConfig))
			pgwireFormatStringInTuple(&ctx.Buffer, s)
		}
		__antithesis_instrumentation__.Notify(611675)
		comma = ","
	}
	__antithesis_instrumentation__.Notify(611673)
	ctx.WriteByte(')')
}

func pgwireFormatStringInTuple(buf *bytes.Buffer, in string) {
	__antithesis_instrumentation__.Notify(611682)
	quote := pgwireQuoteStringInTuple(in)
	if quote {
		__antithesis_instrumentation__.Notify(611685)
		buf.WriteByte('"')
	} else {
		__antithesis_instrumentation__.Notify(611686)
	}
	__antithesis_instrumentation__.Notify(611683)

	for _, r := range in {
		__antithesis_instrumentation__.Notify(611687)
		if r == '"' || func() bool {
			__antithesis_instrumentation__.Notify(611688)
			return r == '\\' == true
		}() == true {
			__antithesis_instrumentation__.Notify(611689)

			buf.WriteByte(byte(r))
			buf.WriteByte(byte(r))
		} else {
			__antithesis_instrumentation__.Notify(611690)
			buf.WriteRune(r)
		}
	}
	__antithesis_instrumentation__.Notify(611684)
	if quote {
		__antithesis_instrumentation__.Notify(611691)
		buf.WriteByte('"')
	} else {
		__antithesis_instrumentation__.Notify(611692)
	}
}

func (d *DArray) pgwireFormat(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(611693)

	switch d.ResolvedType().Oid() {
	case oid.T_int2vector, oid.T_oidvector:
		__antithesis_instrumentation__.Notify(611697)

		sep := ""

		for _, d := range d.Array {
			__antithesis_instrumentation__.Notify(611700)
			ctx.WriteString(sep)
			ctx.FormatNode(d)
			sep = " "
		}
		__antithesis_instrumentation__.Notify(611698)
		return
	default:
		__antithesis_instrumentation__.Notify(611699)
	}
	__antithesis_instrumentation__.Notify(611694)

	if ctx.HasFlags(FmtPGCatalog) {
		__antithesis_instrumentation__.Notify(611701)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(611702)
	}
	__antithesis_instrumentation__.Notify(611695)
	ctx.WriteByte('{')
	comma := ""
	for _, v := range d.Array {
		__antithesis_instrumentation__.Notify(611703)
		ctx.WriteString(comma)
		switch dv := UnwrapDatum(nil, v).(type) {
		case dNull:
			__antithesis_instrumentation__.Notify(611705)
			ctx.WriteString("NULL")
		case *DString:
			__antithesis_instrumentation__.Notify(611706)
			pgwireFormatStringInArray(ctx, string(*dv))
		case *DCollatedString:
			__antithesis_instrumentation__.Notify(611707)
			pgwireFormatStringInArray(ctx, dv.Contents)

		case *DBytes:
			__antithesis_instrumentation__.Notify(611708)
			ctx.FormatNode(dv)
		default:
			__antithesis_instrumentation__.Notify(611709)
			s := AsStringWithFlags(v, ctx.flags, FmtDataConversionConfig(ctx.dataConversionConfig))
			pgwireFormatStringInArray(ctx, s)
		}
		__antithesis_instrumentation__.Notify(611704)
		comma = ","
	}
	__antithesis_instrumentation__.Notify(611696)
	ctx.WriteByte('}')
	if ctx.HasFlags(FmtPGCatalog) {
		__antithesis_instrumentation__.Notify(611710)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(611711)
	}
}

var tupleQuoteSet, arrayQuoteSet asciiSet

func init() {
	var ok bool
	tupleQuoteSet, ok = makeASCIISet(" \t\v\f\r\n(),\"\\")
	if !ok {
		panic("tuple asciiset")
	}
	arrayQuoteSet, ok = makeASCIISet(" \t\v\f\r\n{},\"\\")
	if !ok {
		panic("array asciiset")
	}
}

func pgwireQuoteStringInTuple(in string) bool {
	__antithesis_instrumentation__.Notify(611712)
	return in == "" || func() bool {
		__antithesis_instrumentation__.Notify(611713)
		return tupleQuoteSet.in(in) == true
	}() == true
}

func pgwireQuoteStringInArray(in string) bool {
	__antithesis_instrumentation__.Notify(611714)
	if in == "" || func() bool {
		__antithesis_instrumentation__.Notify(611717)
		return arrayQuoteSet.in(in) == true
	}() == true {
		__antithesis_instrumentation__.Notify(611718)
		return true
	} else {
		__antithesis_instrumentation__.Notify(611719)
	}
	__antithesis_instrumentation__.Notify(611715)
	if len(in) == 4 && func() bool {
		__antithesis_instrumentation__.Notify(611720)
		return (in[0] == 'n' || func() bool {
			__antithesis_instrumentation__.Notify(611721)
			return in[0] == 'N' == true
		}() == true) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(611722)
		return (in[1] == 'u' || func() bool {
			__antithesis_instrumentation__.Notify(611723)
			return in[1] == 'U' == true
		}() == true) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(611724)
		return (in[2] == 'l' || func() bool {
			__antithesis_instrumentation__.Notify(611725)
			return in[2] == 'L' == true
		}() == true) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(611726)
		return (in[3] == 'l' || func() bool {
			__antithesis_instrumentation__.Notify(611727)
			return in[3] == 'L' == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(611728)
		return true
	} else {
		__antithesis_instrumentation__.Notify(611729)
	}
	__antithesis_instrumentation__.Notify(611716)
	return false
}

func pgwireFormatStringInArray(ctx *FmtCtx, in string) {
	__antithesis_instrumentation__.Notify(611730)
	buf := &ctx.Buffer
	quote := pgwireQuoteStringInArray(in)
	if quote {
		__antithesis_instrumentation__.Notify(611733)
		buf.WriteByte('"')
	} else {
		__antithesis_instrumentation__.Notify(611734)
	}
	__antithesis_instrumentation__.Notify(611731)

	for _, r := range in {
		__antithesis_instrumentation__.Notify(611735)
		if r == '"' || func() bool {
			__antithesis_instrumentation__.Notify(611736)
			return r == '\\' == true
		}() == true {
			__antithesis_instrumentation__.Notify(611737)

			buf.WriteByte('\\')
			buf.WriteByte(byte(r))
		} else {
			__antithesis_instrumentation__.Notify(611738)
			if ctx.HasFlags(FmtPGCatalog) && func() bool {
				__antithesis_instrumentation__.Notify(611739)
				return r == '\'' == true
			}() == true {
				__antithesis_instrumentation__.Notify(611740)
				buf.WriteByte('\'')
				buf.WriteByte('\'')
			} else {
				__antithesis_instrumentation__.Notify(611741)
				buf.WriteRune(r)
			}
		}
	}
	__antithesis_instrumentation__.Notify(611732)
	if quote {
		__antithesis_instrumentation__.Notify(611742)
		buf.WriteByte('"')
	} else {
		__antithesis_instrumentation__.Notify(611743)
	}
}

type asciiSet [8]uint32

func makeASCIISet(chars string) (as asciiSet, ok bool) {
	__antithesis_instrumentation__.Notify(611744)
	for i := 0; i < len(chars); i++ {
		__antithesis_instrumentation__.Notify(611746)
		c := chars[i]
		if c >= utf8.RuneSelf {
			__antithesis_instrumentation__.Notify(611748)
			return as, false
		} else {
			__antithesis_instrumentation__.Notify(611749)
		}
		__antithesis_instrumentation__.Notify(611747)
		as[c>>5] |= 1 << uint(c&31)
	}
	__antithesis_instrumentation__.Notify(611745)
	return as, true
}

func (as *asciiSet) contains(c byte) bool {
	__antithesis_instrumentation__.Notify(611750)
	return (as[c>>5] & (1 << uint(c&31))) != 0
}

func (as *asciiSet) in(s string) bool {
	__antithesis_instrumentation__.Notify(611751)
	for i := 0; i < len(s); i++ {
		__antithesis_instrumentation__.Notify(611753)
		if as.contains(s[i]) {
			__antithesis_instrumentation__.Notify(611754)
			return true
		} else {
			__antithesis_instrumentation__.Notify(611755)
		}
	}
	__antithesis_instrumentation__.Notify(611752)
	return false
}
