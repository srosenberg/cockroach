package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type FmtFlags int

func (f FmtFlags) HasFlags(subset FmtFlags) bool {
	__antithesis_instrumentation__.Notify(609737)
	return f&subset == subset
}

func (f FmtFlags) HasAnyFlags(subset FmtFlags) bool {
	__antithesis_instrumentation__.Notify(609738)
	return f&subset != 0
}

func (f FmtFlags) EncodeFlags() lexbase.EncodeFlags {
	__antithesis_instrumentation__.Notify(609739)
	return lexbase.EncodeFlags(f) & (lexbase.EncFirstFreeFlagBit - 1)
}

const (
	FmtSimple FmtFlags = 0

	FmtBareStrings FmtFlags = FmtFlags(lexbase.EncBareStrings)

	FmtBareIdentifiers FmtFlags = FmtFlags(lexbase.EncBareIdentifiers)

	FmtShowPasswords FmtFlags = FmtFlags(lexbase.EncFirstFreeFlagBit) << iota

	FmtShowTypes

	FmtHideConstants

	FmtAnonymize

	FmtAlwaysQualifyTableNames

	FmtAlwaysGroupExprs

	FmtShowTableAliases

	FmtSymbolicSubqueries

	fmtPgwireFormat

	fmtDisambiguateDatumTypes

	fmtSymbolicVars

	fmtRawStrings

	FmtParsableNumerics

	FmtPGCatalog

	fmtStaticallyFormatUserDefinedTypes

	fmtFormatByteLiterals

	FmtMarkRedactionNode

	FmtSummary

	FmtOmitNameRedaction
)

const PasswordSubstitution = "'*****'"

const ColumnLimit = 15

const TableLimit = 30

const (
	FmtPgwireText FmtFlags = fmtPgwireFormat | FmtFlags(lexbase.EncBareStrings)

	FmtParsable FmtFlags = fmtDisambiguateDatumTypes | FmtParsableNumerics

	FmtSerializable FmtFlags = FmtParsable | fmtStaticallyFormatUserDefinedTypes

	FmtCheckEquivalence FmtFlags = fmtSymbolicVars |
		fmtDisambiguateDatumTypes |
		FmtParsableNumerics |
		fmtStaticallyFormatUserDefinedTypes

	FmtArrayToString FmtFlags = FmtBareStrings | fmtRawStrings

	FmtExport FmtFlags = FmtBareStrings | fmtRawStrings
)

const flagsRequiringAnnotations FmtFlags = FmtAlwaysQualifyTableNames

type FmtCtx struct {
	_ util.NoCopy

	bytes.Buffer

	dataConversionConfig sessiondatapb.DataConversionConfig

	flags FmtFlags

	ann *Annotations

	indexedVarFormat func(ctx *FmtCtx, idx int)

	tableNameFormatter func(*FmtCtx, *TableName)

	placeholderFormat func(ctx *FmtCtx, p *Placeholder)

	indexedTypeFormatter func(*FmtCtx, *OIDTypeReference)
}

type FmtCtxOption func(*FmtCtx)

func FmtAnnotations(ann *Annotations) FmtCtxOption {
	__antithesis_instrumentation__.Notify(609740)
	return func(ctx *FmtCtx) {
		__antithesis_instrumentation__.Notify(609741)
		ctx.ann = ann
	}
}

func FmtIndexedVarFormat(fn func(ctx *FmtCtx, idx int)) FmtCtxOption {
	__antithesis_instrumentation__.Notify(609742)
	return func(ctx *FmtCtx) {
		__antithesis_instrumentation__.Notify(609743)
		ctx.indexedVarFormat = fn
	}
}

func FmtPlaceholderFormat(placeholderFn func(_ *FmtCtx, _ *Placeholder)) FmtCtxOption {
	__antithesis_instrumentation__.Notify(609744)
	return func(ctx *FmtCtx) {
		__antithesis_instrumentation__.Notify(609745)
		ctx.placeholderFormat = placeholderFn
	}
}

func FmtReformatTableNames(tableNameFmt func(*FmtCtx, *TableName)) FmtCtxOption {
	__antithesis_instrumentation__.Notify(609746)
	return func(ctx *FmtCtx) {
		__antithesis_instrumentation__.Notify(609747)
		ctx.tableNameFormatter = tableNameFmt
	}
}

func FmtIndexedTypeFormat(fn func(*FmtCtx, *OIDTypeReference)) FmtCtxOption {
	__antithesis_instrumentation__.Notify(609748)
	return func(ctx *FmtCtx) {
		__antithesis_instrumentation__.Notify(609749)
		ctx.indexedTypeFormatter = fn
	}
}

func FmtDataConversionConfig(dcc sessiondatapb.DataConversionConfig) FmtCtxOption {
	__antithesis_instrumentation__.Notify(609750)
	return func(ctx *FmtCtx) {
		__antithesis_instrumentation__.Notify(609751)
		ctx.dataConversionConfig = dcc
	}
}

func NewFmtCtx(f FmtFlags, opts ...FmtCtxOption) *FmtCtx {
	__antithesis_instrumentation__.Notify(609752)
	ctx := fmtCtxPool.Get().(*FmtCtx)
	ctx.flags = f
	for _, opts := range opts {
		__antithesis_instrumentation__.Notify(609755)
		opts(ctx)
	}
	__antithesis_instrumentation__.Notify(609753)
	if ctx.ann == nil && func() bool {
		__antithesis_instrumentation__.Notify(609756)
		return f&flagsRequiringAnnotations != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(609757)
		panic(errors.AssertionFailedf("no Annotations provided"))
	} else {
		__antithesis_instrumentation__.Notify(609758)
	}
	__antithesis_instrumentation__.Notify(609754)
	return ctx
}

func (ctx *FmtCtx) SetDataConversionConfig(
	dcc sessiondatapb.DataConversionConfig,
) sessiondatapb.DataConversionConfig {
	__antithesis_instrumentation__.Notify(609759)
	old := ctx.dataConversionConfig
	ctx.dataConversionConfig = dcc
	return old
}

func (ctx *FmtCtx) WithReformatTableNames(tableNameFmt func(*FmtCtx, *TableName), fn func()) {
	__antithesis_instrumentation__.Notify(609760)
	old := ctx.tableNameFormatter
	ctx.tableNameFormatter = tableNameFmt
	defer func() { __antithesis_instrumentation__.Notify(609762); ctx.tableNameFormatter = old }()
	__antithesis_instrumentation__.Notify(609761)

	fn()
}

func (ctx *FmtCtx) WithFlags(flags FmtFlags, fn func()) {
	__antithesis_instrumentation__.Notify(609763)
	if ctx.ann == nil && func() bool {
		__antithesis_instrumentation__.Notify(609766)
		return flags&flagsRequiringAnnotations != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(609767)
		panic(errors.AssertionFailedf("no Annotations provided"))
	} else {
		__antithesis_instrumentation__.Notify(609768)
	}
	__antithesis_instrumentation__.Notify(609764)
	oldFlags := ctx.flags
	ctx.flags = flags
	defer func() { __antithesis_instrumentation__.Notify(609769); ctx.flags = oldFlags }()
	__antithesis_instrumentation__.Notify(609765)

	fn()
}

func (ctx *FmtCtx) HasFlags(f FmtFlags) bool {
	__antithesis_instrumentation__.Notify(609770)
	return ctx.flags.HasFlags(f)
}

func (ctx *FmtCtx) Printf(f string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(609771)
	fmt.Fprintf(&ctx.Buffer, f, args...)
}

type NodeFormatter interface {
	Format(ctx *FmtCtx)
}

func (ctx *FmtCtx) FormatName(s string) {
	__antithesis_instrumentation__.Notify(609772)
	ctx.FormatNode((*Name)(&s))
}

func (ctx *FmtCtx) FormatNameP(s *string) {
	__antithesis_instrumentation__.Notify(609773)
	ctx.FormatNode((*Name)(s))
}

func (ctx *FmtCtx) FormatNode(n NodeFormatter) {
	__antithesis_instrumentation__.Notify(609774)
	f := ctx.flags
	if f.HasFlags(FmtShowTypes) {
		__antithesis_instrumentation__.Notify(609779)
		if te, ok := n.(TypedExpr); ok {
			__antithesis_instrumentation__.Notify(609780)
			ctx.WriteByte('(')

			if f.HasFlags(FmtMarkRedactionNode) {
				__antithesis_instrumentation__.Notify(609783)
				ctx.formatNodeMaybeMarkRedaction(n)
			} else {
				__antithesis_instrumentation__.Notify(609784)
				ctx.formatNodeOrHideConstants(n)
			}
			__antithesis_instrumentation__.Notify(609781)

			ctx.WriteString(")[")
			if rt := te.ResolvedType(); rt == nil {
				__antithesis_instrumentation__.Notify(609785)

				ctx.Printf("??? %v", te)
			} else {
				__antithesis_instrumentation__.Notify(609786)
				ctx.WriteString(rt.String())
			}
			__antithesis_instrumentation__.Notify(609782)
			ctx.WriteByte(']')
			return
		} else {
			__antithesis_instrumentation__.Notify(609787)
		}
	} else {
		__antithesis_instrumentation__.Notify(609788)
	}
	__antithesis_instrumentation__.Notify(609775)
	if f.HasFlags(FmtAlwaysGroupExprs) {
		__antithesis_instrumentation__.Notify(609789)
		if _, ok := n.(Expr); ok {
			__antithesis_instrumentation__.Notify(609790)
			ctx.WriteByte('(')
		} else {
			__antithesis_instrumentation__.Notify(609791)
		}
	} else {
		__antithesis_instrumentation__.Notify(609792)
	}
	__antithesis_instrumentation__.Notify(609776)

	if f.HasFlags(FmtMarkRedactionNode) {
		__antithesis_instrumentation__.Notify(609793)
		ctx.formatNodeMaybeMarkRedaction(n)
	} else {
		__antithesis_instrumentation__.Notify(609794)
		ctx.formatNodeOrHideConstants(n)
	}
	__antithesis_instrumentation__.Notify(609777)

	if f.HasFlags(FmtAlwaysGroupExprs) {
		__antithesis_instrumentation__.Notify(609795)
		if _, ok := n.(Expr); ok {
			__antithesis_instrumentation__.Notify(609796)
			ctx.WriteByte(')')
		} else {
			__antithesis_instrumentation__.Notify(609797)
		}
	} else {
		__antithesis_instrumentation__.Notify(609798)
	}
	__antithesis_instrumentation__.Notify(609778)
	if f.HasAnyFlags(fmtDisambiguateDatumTypes | FmtPGCatalog) {
		__antithesis_instrumentation__.Notify(609799)
		var typ *types.T
		if d, isDatum := n.(Datum); isDatum {
			__antithesis_instrumentation__.Notify(609801)
			if p, isPlaceholder := d.(*Placeholder); isPlaceholder {
				__antithesis_instrumentation__.Notify(609802)

				typ = p.typ
			} else {
				__antithesis_instrumentation__.Notify(609803)
				if d.AmbiguousFormat() {
					__antithesis_instrumentation__.Notify(609804)
					typ = d.ResolvedType()
				} else {
					__antithesis_instrumentation__.Notify(609805)
					if _, isArray := d.(*DArray); isArray && func() bool {
						__antithesis_instrumentation__.Notify(609806)
						return f.HasFlags(FmtPGCatalog) == true
					}() == true {
						__antithesis_instrumentation__.Notify(609807)
						typ = d.ResolvedType()
					} else {
						__antithesis_instrumentation__.Notify(609808)
					}
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(609809)
		}
		__antithesis_instrumentation__.Notify(609800)
		if typ != nil {
			__antithesis_instrumentation__.Notify(609810)
			if f.HasFlags(fmtDisambiguateDatumTypes) {
				__antithesis_instrumentation__.Notify(609811)
				ctx.WriteString(":::")
				ctx.FormatTypeReference(typ)
			} else {
				__antithesis_instrumentation__.Notify(609812)
				if f.HasFlags(FmtPGCatalog) && func() bool {
					__antithesis_instrumentation__.Notify(609813)
					return !typ.IsNumeric() == true
				}() == true {
					__antithesis_instrumentation__.Notify(609814)
					ctx.WriteString("::")
					ctx.FormatTypeReference(typ)
				} else {
					__antithesis_instrumentation__.Notify(609815)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(609816)
		}
	} else {
		__antithesis_instrumentation__.Notify(609817)
	}
}

func (ctx *FmtCtx) formatLimitLength(n NodeFormatter, maxLength int) {
	__antithesis_instrumentation__.Notify(609818)
	temp := NewFmtCtx(ctx.flags)
	temp.FormatNodeSummary(n)
	s := temp.CloseAndGetString()
	if len(s) > maxLength {
		__antithesis_instrumentation__.Notify(609820)
		truncated := s[:maxLength] + "..."

		if strings.Count(truncated, "(") > strings.Count(truncated, ")") {
			__antithesis_instrumentation__.Notify(609822)
			remaining := s[maxLength:]
			for i, c := range remaining {
				__antithesis_instrumentation__.Notify(609823)
				if c == ')' {
					__antithesis_instrumentation__.Notify(609824)
					truncated += ")"

					if i < len(remaining)-1 && func() bool {
						__antithesis_instrumentation__.Notify(609826)
						return string(remaining[i+1]) != ")" == true
					}() == true {
						__antithesis_instrumentation__.Notify(609827)
						truncated += "..."
					} else {
						__antithesis_instrumentation__.Notify(609828)
					}
					__antithesis_instrumentation__.Notify(609825)
					if strings.Count(truncated, "(") <= strings.Count(truncated, ")") {
						__antithesis_instrumentation__.Notify(609829)
						break
					} else {
						__antithesis_instrumentation__.Notify(609830)
					}
				} else {
					__antithesis_instrumentation__.Notify(609831)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(609832)
		}
		__antithesis_instrumentation__.Notify(609821)
		s = truncated
	} else {
		__antithesis_instrumentation__.Notify(609833)
	}
	__antithesis_instrumentation__.Notify(609819)
	ctx.WriteString(s)
}

func (ctx *FmtCtx) formatSummarySelect(node *Select) {
	__antithesis_instrumentation__.Notify(609834)
	if node.With == nil {
		__antithesis_instrumentation__.Notify(609835)
		s := node.Select
		if s, ok := s.(*SelectClause); ok {
			__antithesis_instrumentation__.Notify(609836)
			ctx.WriteString("SELECT ")
			ctx.formatLimitLength(&s.Exprs, ColumnLimit)
			if len(s.From.Tables) > 0 {
				__antithesis_instrumentation__.Notify(609837)
				ctx.WriteByte(' ')
				ctx.formatLimitLength(&s.From, TableLimit+len("FROM "))
			} else {
				__antithesis_instrumentation__.Notify(609838)
			}
		} else {
			__antithesis_instrumentation__.Notify(609839)
		}
	} else {
		__antithesis_instrumentation__.Notify(609840)
	}
}

func (ctx *FmtCtx) formatSummaryInsert(node *Insert) {
	__antithesis_instrumentation__.Notify(609841)
	if node.OnConflict.IsUpsertAlias() {
		__antithesis_instrumentation__.Notify(609843)
		ctx.WriteString("UPSERT")
	} else {
		__antithesis_instrumentation__.Notify(609844)
		ctx.WriteString("INSERT")
	}
	__antithesis_instrumentation__.Notify(609842)
	ctx.WriteString(" INTO ")
	ctx.formatLimitLength(node.Table, TableLimit)
	rows := node.Rows
	expr := rows.Select
	if _, ok := expr.(*SelectClause); ok {
		__antithesis_instrumentation__.Notify(609845)
		ctx.WriteByte(' ')
		ctx.FormatNodeSummary(rows)
	} else {
		__antithesis_instrumentation__.Notify(609846)
		if node.Columns != nil {
			__antithesis_instrumentation__.Notify(609847)
			ctx.WriteByte('(')
			ctx.formatLimitLength(&node.Columns, ColumnLimit)
			ctx.WriteByte(')')
		} else {
			__antithesis_instrumentation__.Notify(609848)
		}
	}
}

func (ctx *FmtCtx) formatSummaryUpdate(node *Update) {
	__antithesis_instrumentation__.Notify(609849)
	if node.With == nil {
		__antithesis_instrumentation__.Notify(609850)
		ctx.WriteString("UPDATE ")
		ctx.formatLimitLength(node.Table, TableLimit)
		ctx.WriteString(" SET ")
		ctx.formatLimitLength(&node.Exprs, ColumnLimit)
		if node.Where != nil {
			__antithesis_instrumentation__.Notify(609851)
			ctx.WriteByte(' ')
			ctx.formatLimitLength(node.Where, ColumnLimit+len("WHERE "))
		} else {
			__antithesis_instrumentation__.Notify(609852)
		}
	} else {
		__antithesis_instrumentation__.Notify(609853)
	}
}

func (ctx *FmtCtx) FormatNodeSummary(n NodeFormatter) {
	__antithesis_instrumentation__.Notify(609854)
	switch node := n.(type) {
	case *Insert:
		__antithesis_instrumentation__.Notify(609856)
		ctx.formatSummaryInsert(node)
		return
	case *Select:
		__antithesis_instrumentation__.Notify(609857)
		ctx.formatSummarySelect(node)
		return
	case *Update:
		__antithesis_instrumentation__.Notify(609858)
		ctx.formatSummaryUpdate(node)
		return
	}
	__antithesis_instrumentation__.Notify(609855)
	ctx.FormatNode(n)
}

func AsStringWithFlags(n NodeFormatter, fl FmtFlags, opts ...FmtCtxOption) string {
	__antithesis_instrumentation__.Notify(609859)
	ctx := NewFmtCtx(fl, opts...)
	if fl.HasFlags(FmtSummary) {
		__antithesis_instrumentation__.Notify(609861)
		ctx.FormatNodeSummary(n)
	} else {
		__antithesis_instrumentation__.Notify(609862)
		ctx.FormatNode(n)
	}
	__antithesis_instrumentation__.Notify(609860)
	return ctx.CloseAndGetString()
}

func AsStringWithFQNames(n NodeFormatter, ann *Annotations) string {
	__antithesis_instrumentation__.Notify(609863)
	ctx := NewFmtCtx(FmtAlwaysQualifyTableNames, FmtAnnotations(ann))
	ctx.FormatNode(n)
	return ctx.CloseAndGetString()
}

func AsString(n NodeFormatter) string {
	__antithesis_instrumentation__.Notify(609864)
	return AsStringWithFlags(n, FmtSimple)
}

func ErrString(n NodeFormatter) string {
	__antithesis_instrumentation__.Notify(609865)
	return AsStringWithFlags(n, FmtBareIdentifiers)
}

func Serialize(n NodeFormatter) string {
	__antithesis_instrumentation__.Notify(609866)
	return AsStringWithFlags(n, FmtSerializable)
}

func SerializeForDisplay(n NodeFormatter) string {
	__antithesis_instrumentation__.Notify(609867)
	return AsStringWithFlags(n, FmtParsable)
}

var fmtCtxPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(609868)
		return &FmtCtx{}
	},
}

func (ctx *FmtCtx) Close() {
	__antithesis_instrumentation__.Notify(609869)
	ctx.Buffer.Reset()
	*ctx = FmtCtx{
		Buffer: ctx.Buffer,
	}
	fmtCtxPool.Put(ctx)
}

func (ctx *FmtCtx) CloseAndGetString() string {
	__antithesis_instrumentation__.Notify(609870)
	s := ctx.String()
	ctx.Close()
	return s
}

func (ctx *FmtCtx) alwaysFormatTablePrefix() bool {
	__antithesis_instrumentation__.Notify(609871)
	return ctx.flags.HasFlags(FmtAlwaysQualifyTableNames) || func() bool {
		__antithesis_instrumentation__.Notify(609872)
		return ctx.tableNameFormatter != nil == true
	}() == true
}

func (ctx *FmtCtx) formatNodeMaybeMarkRedaction(n NodeFormatter) {
	__antithesis_instrumentation__.Notify(609873)
	if ctx.flags.HasFlags(FmtMarkRedactionNode) {
		__antithesis_instrumentation__.Notify(609874)
		switch v := n.(type) {
		case *Placeholder:
			__antithesis_instrumentation__.Notify(609876)

		case *Name, *UnrestrictedName:
			__antithesis_instrumentation__.Notify(609877)
			if ctx.flags.HasFlags(FmtOmitNameRedaction) {
				__antithesis_instrumentation__.Notify(609880)
				break
			} else {
				__antithesis_instrumentation__.Notify(609881)
			}
			__antithesis_instrumentation__.Notify(609878)
			ctx.WriteString(string(redact.StartMarker()))
			v.Format(ctx)
			ctx.WriteString(string(redact.EndMarker()))
			return
		case Datum, Constant:
			__antithesis_instrumentation__.Notify(609879)
			ctx.WriteString(string(redact.StartMarker()))
			v.Format(ctx)
			ctx.WriteString(string(redact.EndMarker()))
			return
		}
		__antithesis_instrumentation__.Notify(609875)
		n.Format(ctx)
	} else {
		__antithesis_instrumentation__.Notify(609882)
	}
}
