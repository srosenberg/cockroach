package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treewindow"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/pretty"
	"github.com/cockroachdb/errors"
)

type PrettyCfg struct {
	LineWidth int

	TabWidth int

	Align PrettyAlignMode

	UseTabs bool

	Simplify bool

	Case func(string) string

	JSONFmt bool
}

func DefaultPrettyCfg() PrettyCfg {
	__antithesis_instrumentation__.Notify(611853)
	return PrettyCfg{
		LineWidth: DefaultLineWidth,
		Simplify:  true,
		TabWidth:  4,
		UseTabs:   true,
		Align:     PrettyNoAlign,
	}
}

type PrettyAlignMode int

const (
	PrettyNoAlign PrettyAlignMode = 0

	PrettyAlignOnly = 1

	PrettyAlignAndDeindent = 2

	PrettyAlignAndExtraIndent = 3
)

type CaseMode int

const (
	LowerCase CaseMode = 0

	UpperCase CaseMode = 1
)

type LineWidthMode int

const (
	DefaultLineWidth = 60

	ConsoleLineWidth = 108
)

func (p *PrettyCfg) keywordWithText(left, keyword, right string) pretty.Doc {
	__antithesis_instrumentation__.Notify(611854)
	doc := pretty.Keyword(keyword)
	if left != "" {
		__antithesis_instrumentation__.Notify(611857)
		doc = pretty.Concat(pretty.Text(left), doc)
	} else {
		__antithesis_instrumentation__.Notify(611858)
	}
	__antithesis_instrumentation__.Notify(611855)
	if right != "" {
		__antithesis_instrumentation__.Notify(611859)
		doc = pretty.Concat(doc, pretty.Text(right))
	} else {
		__antithesis_instrumentation__.Notify(611860)
	}
	__antithesis_instrumentation__.Notify(611856)
	return doc
}

func (p *PrettyCfg) bracket(l string, d pretty.Doc, r string) pretty.Doc {
	__antithesis_instrumentation__.Notify(611861)
	return p.bracketDoc(pretty.Text(l), d, pretty.Text(r))
}

func (p *PrettyCfg) bracketDoc(l, d, r pretty.Doc) pretty.Doc {
	__antithesis_instrumentation__.Notify(611862)
	return pretty.BracketDoc(l, d, r)
}

func (p *PrettyCfg) bracketKeyword(
	leftKeyword, leftParen string, inner pretty.Doc, rightParen, rightKeyword string,
) pretty.Doc {
	__antithesis_instrumentation__.Notify(611863)
	var left, right pretty.Doc
	if leftKeyword != "" {
		__antithesis_instrumentation__.Notify(611866)
		left = p.keywordWithText("", leftKeyword, leftParen)
	} else {
		__antithesis_instrumentation__.Notify(611867)
		left = pretty.Text(leftParen)
	}
	__antithesis_instrumentation__.Notify(611864)
	if rightKeyword != "" {
		__antithesis_instrumentation__.Notify(611868)
		right = p.keywordWithText(rightParen, rightKeyword, "")
	} else {
		__antithesis_instrumentation__.Notify(611869)
		right = pretty.Text(rightParen)
	}
	__antithesis_instrumentation__.Notify(611865)
	return p.bracketDoc(left, inner, right)
}

func Pretty(stmt NodeFormatter) string {
	__antithesis_instrumentation__.Notify(611870)
	cfg := DefaultPrettyCfg()
	return cfg.Pretty(stmt)
}

func (p *PrettyCfg) Pretty(stmt NodeFormatter) string {
	__antithesis_instrumentation__.Notify(611871)
	doc := p.Doc(stmt)
	return pretty.Pretty(doc, p.LineWidth, p.UseTabs, p.TabWidth, p.Case)
}

func (p *PrettyCfg) Doc(f NodeFormatter) pretty.Doc {
	__antithesis_instrumentation__.Notify(611872)
	if f, ok := f.(docer); ok {
		__antithesis_instrumentation__.Notify(611874)
		doc := f.doc(p)
		return doc
	} else {
		__antithesis_instrumentation__.Notify(611875)
	}
	__antithesis_instrumentation__.Notify(611873)
	return p.docAsString(f)
}

func (p *PrettyCfg) docAsString(f NodeFormatter) pretty.Doc {
	__antithesis_instrumentation__.Notify(611876)
	const prettyFlags = FmtShowPasswords | FmtParsable
	txt := AsStringWithFlags(f, prettyFlags)
	return pretty.Text(strings.TrimSpace(txt))
}

func (p *PrettyCfg) nestUnder(a, b pretty.Doc) pretty.Doc {
	__antithesis_instrumentation__.Notify(611877)
	if p.Align != PrettyNoAlign {
		__antithesis_instrumentation__.Notify(611879)
		return pretty.AlignUnder(a, b)
	} else {
		__antithesis_instrumentation__.Notify(611880)
	}
	__antithesis_instrumentation__.Notify(611878)
	return pretty.NestUnder(a, b)
}

func (p *PrettyCfg) rlTable(rows ...pretty.TableRow) pretty.Doc {
	__antithesis_instrumentation__.Notify(611881)
	alignment := pretty.TableNoAlign
	if p.Align != PrettyNoAlign {
		__antithesis_instrumentation__.Notify(611883)
		alignment = pretty.TableRightAlignFirstColumn
	} else {
		__antithesis_instrumentation__.Notify(611884)
	}
	__antithesis_instrumentation__.Notify(611882)
	return pretty.Table(alignment, pretty.Keyword, rows...)
}

func (p *PrettyCfg) llTable(docFn func(string) pretty.Doc, rows ...pretty.TableRow) pretty.Doc {
	__antithesis_instrumentation__.Notify(611885)
	alignment := pretty.TableNoAlign
	if p.Align != PrettyNoAlign {
		__antithesis_instrumentation__.Notify(611887)
		alignment = pretty.TableLeftAlignFirstColumn
	} else {
		__antithesis_instrumentation__.Notify(611888)
	}
	__antithesis_instrumentation__.Notify(611886)
	return pretty.Table(alignment, docFn, rows...)
}

func (p *PrettyCfg) row(lbl string, d pretty.Doc) pretty.TableRow {
	__antithesis_instrumentation__.Notify(611889)
	return pretty.TableRow{Label: lbl, Doc: d}
}

var emptyRow = pretty.TableRow{}

func (p *PrettyCfg) unrow(r pretty.TableRow) pretty.Doc {
	__antithesis_instrumentation__.Notify(611890)
	if r.Doc == nil {
		__antithesis_instrumentation__.Notify(611893)
		return pretty.Nil
	} else {
		__antithesis_instrumentation__.Notify(611894)
	}
	__antithesis_instrumentation__.Notify(611891)
	if r.Label == "" {
		__antithesis_instrumentation__.Notify(611895)
		return r.Doc
	} else {
		__antithesis_instrumentation__.Notify(611896)
	}
	__antithesis_instrumentation__.Notify(611892)
	return p.nestUnder(pretty.Text(r.Label), r.Doc)
}

func (p *PrettyCfg) commaSeparated(d ...pretty.Doc) pretty.Doc {
	__antithesis_instrumentation__.Notify(611897)
	return pretty.Join(",", d...)
}

func (p *PrettyCfg) joinNestedOuter(lbl string, d ...pretty.Doc) pretty.Doc {
	__antithesis_instrumentation__.Notify(611898)
	if len(d) == 0 {
		__antithesis_instrumentation__.Notify(611900)
		return pretty.Nil
	} else {
		__antithesis_instrumentation__.Notify(611901)
	}
	__antithesis_instrumentation__.Notify(611899)
	switch p.Align {
	case PrettyAlignAndDeindent:
		__antithesis_instrumentation__.Notify(611902)
		return pretty.JoinNestedOuter(lbl, pretty.Keyword, d...)
	case PrettyAlignAndExtraIndent:
		__antithesis_instrumentation__.Notify(611903)
		items := make([]pretty.TableRow, len(d))
		for i, dd := range d {
			__antithesis_instrumentation__.Notify(611906)
			if i > 0 {
				__antithesis_instrumentation__.Notify(611908)
				items[i].Label = lbl
			} else {
				__antithesis_instrumentation__.Notify(611909)
			}
			__antithesis_instrumentation__.Notify(611907)
			items[i].Doc = dd
		}
		__antithesis_instrumentation__.Notify(611904)
		return pretty.Table(pretty.TableRightAlignFirstColumn, pretty.Keyword, items...)
	default:
		__antithesis_instrumentation__.Notify(611905)
		return pretty.JoinNestedRight(pretty.Keyword(lbl), d...)
	}
}

type docer interface {
	doc(*PrettyCfg) pretty.Doc
}

type tableDocer interface {
	docTable(*PrettyCfg) []pretty.TableRow
}

func (node SelectExprs) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(611910)
	d := make([]pretty.Doc, len(node))
	for i, e := range node {
		__antithesis_instrumentation__.Notify(611912)
		d[i] = e.doc(p)
	}
	__antithesis_instrumentation__.Notify(611911)
	return p.commaSeparated(d...)
}

func (node SelectExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(611913)
	e := node.Expr
	if p.Simplify {
		__antithesis_instrumentation__.Notify(611916)
		e = StripParens(e)
	} else {
		__antithesis_instrumentation__.Notify(611917)
	}
	__antithesis_instrumentation__.Notify(611914)
	d := p.Doc(e)
	if node.As != "" {
		__antithesis_instrumentation__.Notify(611918)
		d = p.nestUnder(
			d,
			pretty.Concat(p.keywordWithText("", "AS", " "), p.Doc(&node.As)),
		)
	} else {
		__antithesis_instrumentation__.Notify(611919)
	}
	__antithesis_instrumentation__.Notify(611915)
	return d
}

func (node TableExprs) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(611920)
	if len(node) == 0 {
		__antithesis_instrumentation__.Notify(611923)
		return pretty.Nil
	} else {
		__antithesis_instrumentation__.Notify(611924)
	}
	__antithesis_instrumentation__.Notify(611921)
	d := make([]pretty.Doc, len(node))
	for i, e := range node {
		__antithesis_instrumentation__.Notify(611925)
		if p.Simplify {
			__antithesis_instrumentation__.Notify(611927)
			e = StripTableParens(e)
		} else {
			__antithesis_instrumentation__.Notify(611928)
		}
		__antithesis_instrumentation__.Notify(611926)
		d[i] = p.Doc(e)
	}
	__antithesis_instrumentation__.Notify(611922)
	return p.commaSeparated(d...)
}

func (node *Where) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(611929)
	return p.unrow(node.docRow(p))
}

func (node *Where) docRow(p *PrettyCfg) pretty.TableRow {
	__antithesis_instrumentation__.Notify(611930)
	if node == nil {
		__antithesis_instrumentation__.Notify(611933)
		return emptyRow
	} else {
		__antithesis_instrumentation__.Notify(611934)
	}
	__antithesis_instrumentation__.Notify(611931)
	e := node.Expr
	if p.Simplify {
		__antithesis_instrumentation__.Notify(611935)
		e = StripParens(e)
	} else {
		__antithesis_instrumentation__.Notify(611936)
	}
	__antithesis_instrumentation__.Notify(611932)
	return p.row(node.Type, p.Doc(e))
}

func (node *GroupBy) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(611937)
	return p.unrow(node.docRow(p))
}

func (node *GroupBy) docRow(p *PrettyCfg) pretty.TableRow {
	__antithesis_instrumentation__.Notify(611938)
	if len(*node) == 0 {
		__antithesis_instrumentation__.Notify(611941)
		return emptyRow
	} else {
		__antithesis_instrumentation__.Notify(611942)
	}
	__antithesis_instrumentation__.Notify(611939)
	d := make([]pretty.Doc, len(*node))
	for i, e := range *node {
		__antithesis_instrumentation__.Notify(611943)

		d[i] = p.Doc(e)
	}
	__antithesis_instrumentation__.Notify(611940)
	return p.row("GROUP BY", p.commaSeparated(d...))
}

func (p *PrettyCfg) flattenOp(
	e Expr,
	pred func(e Expr, recurse func(e Expr)) bool,
	formatOperand func(e Expr) pretty.Doc,
	in []pretty.Doc,
) []pretty.Doc {
	__antithesis_instrumentation__.Notify(611944)
	if ok := pred(e, func(sub Expr) {
		__antithesis_instrumentation__.Notify(611946)
		in = p.flattenOp(sub, pred, formatOperand, in)
	}); ok {
		__antithesis_instrumentation__.Notify(611947)
		return in
	} else {
		__antithesis_instrumentation__.Notify(611948)
	}
	__antithesis_instrumentation__.Notify(611945)
	return append(in, formatOperand(e))
}

func (p *PrettyCfg) peelAndOrOperand(e Expr) Expr {
	__antithesis_instrumentation__.Notify(611949)
	if !p.Simplify {
		__antithesis_instrumentation__.Notify(611952)
		return e
	} else {
		__antithesis_instrumentation__.Notify(611953)
	}
	__antithesis_instrumentation__.Notify(611950)
	stripped := StripParens(e)
	switch stripped.(type) {
	case *BinaryExpr, *ComparisonExpr, *RangeCond, *FuncExpr, *IndirectionExpr,
		*UnaryExpr, *AnnotateTypeExpr, *CastExpr, *ColumnItem, *UnresolvedName:
		__antithesis_instrumentation__.Notify(611954)

		return stripped
	}
	__antithesis_instrumentation__.Notify(611951)

	return e
}

func (node *AndExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(611955)
	pred := func(e Expr, recurse func(e Expr)) bool {
		__antithesis_instrumentation__.Notify(611958)
		if a, ok := e.(*AndExpr); ok {
			__antithesis_instrumentation__.Notify(611960)
			recurse(a.Left)
			recurse(a.Right)
			return true
		} else {
			__antithesis_instrumentation__.Notify(611961)
		}
		__antithesis_instrumentation__.Notify(611959)
		return false
	}
	__antithesis_instrumentation__.Notify(611956)
	formatOperand := func(e Expr) pretty.Doc {
		__antithesis_instrumentation__.Notify(611962)
		return p.Doc(p.peelAndOrOperand(e))
	}
	__antithesis_instrumentation__.Notify(611957)
	operands := p.flattenOp(node.Left, pred, formatOperand, nil)
	operands = p.flattenOp(node.Right, pred, formatOperand, operands)
	return p.joinNestedOuter("AND", operands...)
}

func (node *OrExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(611963)
	pred := func(e Expr, recurse func(e Expr)) bool {
		__antithesis_instrumentation__.Notify(611966)
		if a, ok := e.(*OrExpr); ok {
			__antithesis_instrumentation__.Notify(611968)
			recurse(a.Left)
			recurse(a.Right)
			return true
		} else {
			__antithesis_instrumentation__.Notify(611969)
		}
		__antithesis_instrumentation__.Notify(611967)
		return false
	}
	__antithesis_instrumentation__.Notify(611964)
	formatOperand := func(e Expr) pretty.Doc {
		__antithesis_instrumentation__.Notify(611970)
		return p.Doc(p.peelAndOrOperand(e))
	}
	__antithesis_instrumentation__.Notify(611965)
	operands := p.flattenOp(node.Left, pred, formatOperand, nil)
	operands = p.flattenOp(node.Right, pred, formatOperand, operands)
	return p.joinNestedOuter("OR", operands...)
}

func (node *Exprs) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(611971)
	if node == nil || func() bool {
		__antithesis_instrumentation__.Notify(611974)
		return len(*node) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(611975)
		return pretty.Nil
	} else {
		__antithesis_instrumentation__.Notify(611976)
	}
	__antithesis_instrumentation__.Notify(611972)
	d := make([]pretty.Doc, len(*node))
	for i, e := range *node {
		__antithesis_instrumentation__.Notify(611977)
		if p.Simplify {
			__antithesis_instrumentation__.Notify(611979)
			e = StripParens(e)
		} else {
			__antithesis_instrumentation__.Notify(611980)
		}
		__antithesis_instrumentation__.Notify(611978)
		d[i] = p.Doc(e)
	}
	__antithesis_instrumentation__.Notify(611973)
	return p.commaSeparated(d...)
}

func (p *PrettyCfg) peelBinaryOperand(e Expr, sameLevel bool, parenPrio int) Expr {
	__antithesis_instrumentation__.Notify(611981)
	if !p.Simplify {
		__antithesis_instrumentation__.Notify(611984)
		return e
	} else {
		__antithesis_instrumentation__.Notify(611985)
	}
	__antithesis_instrumentation__.Notify(611982)
	stripped := StripParens(e)
	switch te := stripped.(type) {
	case *BinaryExpr:
		__antithesis_instrumentation__.Notify(611986)

		if te.Operator.IsExplicitOperator {
			__antithesis_instrumentation__.Notify(611989)
			return e
		} else {
			__antithesis_instrumentation__.Notify(611990)
		}
		__antithesis_instrumentation__.Notify(611987)
		childPrio := binaryOpPrio[te.Operator.Symbol]
		if childPrio < parenPrio || func() bool {
			__antithesis_instrumentation__.Notify(611991)
			return (sameLevel && func() bool {
				__antithesis_instrumentation__.Notify(611992)
				return childPrio == parenPrio == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(611993)
			return stripped
		} else {
			__antithesis_instrumentation__.Notify(611994)
		}
	case *FuncExpr, *UnaryExpr, *AnnotateTypeExpr, *IndirectionExpr,
		*CastExpr, *ColumnItem, *UnresolvedName:
		__antithesis_instrumentation__.Notify(611988)

		return stripped
	}
	__antithesis_instrumentation__.Notify(611983)

	return e
}

func (node *BinaryExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(611995)

	parenPrio := binaryOpPrio[node.Operator.Symbol]
	leftOperand := p.peelBinaryOperand(node.Left, true, parenPrio)

	opFullyAssoc := binaryOpFullyAssoc[node.Operator.Symbol]
	rightOperand := p.peelBinaryOperand(node.Right, opFullyAssoc, parenPrio)

	opDoc := pretty.Text(node.Operator.String())
	var res pretty.Doc
	if !node.Operator.Symbol.IsPadded() {
		__antithesis_instrumentation__.Notify(611997)
		res = pretty.JoinDoc(opDoc, p.Doc(leftOperand), p.Doc(rightOperand))
	} else {
		__antithesis_instrumentation__.Notify(611998)
		pred := func(e Expr, recurse func(e Expr)) bool {
			__antithesis_instrumentation__.Notify(612001)
			if b, ok := e.(*BinaryExpr); ok && func() bool {
				__antithesis_instrumentation__.Notify(612003)
				return b.Operator == node.Operator == true
			}() == true {
				__antithesis_instrumentation__.Notify(612004)
				leftSubOperand := p.peelBinaryOperand(b.Left, true, parenPrio)
				rightSubOperand := p.peelBinaryOperand(b.Right, opFullyAssoc, parenPrio)
				recurse(leftSubOperand)
				recurse(rightSubOperand)
				return true
			} else {
				__antithesis_instrumentation__.Notify(612005)
			}
			__antithesis_instrumentation__.Notify(612002)
			return false
		}
		__antithesis_instrumentation__.Notify(611999)
		formatOperand := func(e Expr) pretty.Doc {
			__antithesis_instrumentation__.Notify(612006)
			return p.Doc(e)
		}
		__antithesis_instrumentation__.Notify(612000)
		operands := p.flattenOp(leftOperand, pred, formatOperand, nil)
		operands = p.flattenOp(rightOperand, pred, formatOperand, operands)
		res = pretty.JoinNestedRight(
			opDoc, operands...)
	}
	__antithesis_instrumentation__.Notify(611996)
	return pretty.Group(res)
}

func (node *ParenExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612007)
	return p.bracket("(", p.Doc(node.Expr), ")")
}

func (node *ParenSelect) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612008)
	return p.bracket("(", p.Doc(node.Select), ")")
}

func (node *ParenTableExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612009)
	return p.bracket("(", p.Doc(node.Expr), ")")
}

func (node *Limit) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612010)
	res := pretty.Nil
	for i, r := range node.docTable(p) {
		__antithesis_instrumentation__.Notify(612012)
		if r.Doc != nil {
			__antithesis_instrumentation__.Notify(612013)
			if i > 0 {
				__antithesis_instrumentation__.Notify(612015)
				res = pretty.Concat(res, pretty.Line)
			} else {
				__antithesis_instrumentation__.Notify(612016)
			}
			__antithesis_instrumentation__.Notify(612014)
			res = pretty.Concat(res, p.nestUnder(pretty.Text(r.Label), r.Doc))
		} else {
			__antithesis_instrumentation__.Notify(612017)
		}
	}
	__antithesis_instrumentation__.Notify(612011)
	return res
}

func (node *Limit) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612018)
	if node == nil {
		__antithesis_instrumentation__.Notify(612022)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(612023)
	}
	__antithesis_instrumentation__.Notify(612019)
	res := make([]pretty.TableRow, 0, 2)
	if node.Count != nil {
		__antithesis_instrumentation__.Notify(612024)
		e := node.Count
		if p.Simplify {
			__antithesis_instrumentation__.Notify(612026)
			e = StripParens(e)
		} else {
			__antithesis_instrumentation__.Notify(612027)
		}
		__antithesis_instrumentation__.Notify(612025)
		res = append(res, p.row("LIMIT", p.Doc(e)))
	} else {
		__antithesis_instrumentation__.Notify(612028)
		if node.LimitAll {
			__antithesis_instrumentation__.Notify(612029)
			res = append(res, p.row("LIMIT", pretty.Keyword("ALL")))
		} else {
			__antithesis_instrumentation__.Notify(612030)
		}
	}
	__antithesis_instrumentation__.Notify(612020)
	if node.Offset != nil {
		__antithesis_instrumentation__.Notify(612031)
		e := node.Offset
		if p.Simplify {
			__antithesis_instrumentation__.Notify(612033)
			e = StripParens(e)
		} else {
			__antithesis_instrumentation__.Notify(612034)
		}
		__antithesis_instrumentation__.Notify(612032)
		res = append(res, p.row("OFFSET", p.Doc(e)))
	} else {
		__antithesis_instrumentation__.Notify(612035)
	}
	__antithesis_instrumentation__.Notify(612021)
	return res
}

func (node *OrderBy) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612036)
	return p.unrow(node.docRow(p))
}

func (node *OrderBy) docRow(p *PrettyCfg) pretty.TableRow {
	__antithesis_instrumentation__.Notify(612037)
	if node == nil || func() bool {
		__antithesis_instrumentation__.Notify(612040)
		return len(*node) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(612041)
		return emptyRow
	} else {
		__antithesis_instrumentation__.Notify(612042)
	}
	__antithesis_instrumentation__.Notify(612038)
	d := make([]pretty.Doc, len(*node))
	for i, e := range *node {
		__antithesis_instrumentation__.Notify(612043)

		d[i] = p.Doc(e)
	}
	__antithesis_instrumentation__.Notify(612039)
	return p.row("ORDER BY", p.commaSeparated(d...))
}

func (node *Select) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612044)
	return p.rlTable(node.docTable(p)...)
}

func (node *Select) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612045)
	items := make([]pretty.TableRow, 0, 9)
	items = append(items, node.With.docRow(p))
	if s, ok := node.Select.(tableDocer); ok {
		__antithesis_instrumentation__.Notify(612047)
		items = append(items, s.docTable(p)...)
	} else {
		__antithesis_instrumentation__.Notify(612048)
		items = append(items, p.row("", p.Doc(node.Select)))
	}
	__antithesis_instrumentation__.Notify(612046)
	items = append(items, node.OrderBy.docRow(p))
	items = append(items, node.Limit.docTable(p)...)
	items = append(items, node.Locking.docTable(p)...)
	return items
}

func (node *SelectClause) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612049)
	return p.rlTable(node.docTable(p)...)
}

func (node *SelectClause) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612050)
	if node.TableSelect {
		__antithesis_instrumentation__.Notify(612053)
		return []pretty.TableRow{p.row("TABLE", p.Doc(node.From.Tables[0]))}
	} else {
		__antithesis_instrumentation__.Notify(612054)
	}
	__antithesis_instrumentation__.Notify(612051)
	exprs := node.Exprs.doc(p)
	if node.Distinct {
		__antithesis_instrumentation__.Notify(612055)
		if node.DistinctOn != nil {
			__antithesis_instrumentation__.Notify(612056)
			exprs = pretty.ConcatLine(p.Doc(&node.DistinctOn), exprs)
		} else {
			__antithesis_instrumentation__.Notify(612057)
			exprs = pretty.ConcatLine(pretty.Keyword("DISTINCT"), exprs)
		}
	} else {
		__antithesis_instrumentation__.Notify(612058)
	}
	__antithesis_instrumentation__.Notify(612052)
	return []pretty.TableRow{
		p.row("SELECT", exprs),
		node.From.docRow(p),
		node.Where.docRow(p),
		node.GroupBy.docRow(p),
		node.Having.docRow(p),
		node.Window.docRow(p),
	}
}

func (node *From) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612059)
	return p.unrow(node.docRow(p))
}

func (node *From) docRow(p *PrettyCfg) pretty.TableRow {
	__antithesis_instrumentation__.Notify(612060)
	if node == nil || func() bool {
		__antithesis_instrumentation__.Notify(612063)
		return len(node.Tables) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(612064)
		return emptyRow
	} else {
		__antithesis_instrumentation__.Notify(612065)
	}
	__antithesis_instrumentation__.Notify(612061)
	d := node.Tables.doc(p)
	if node.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(612066)
		d = p.nestUnder(
			d,
			p.Doc(&node.AsOf),
		)
	} else {
		__antithesis_instrumentation__.Notify(612067)
	}
	__antithesis_instrumentation__.Notify(612062)
	return p.row("FROM", d)
}

func (node *Window) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612068)
	return p.unrow(node.docRow(p))
}

func (node *Window) docRow(p *PrettyCfg) pretty.TableRow {
	__antithesis_instrumentation__.Notify(612069)
	if node == nil || func() bool {
		__antithesis_instrumentation__.Notify(612072)
		return len(*node) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(612073)
		return emptyRow
	} else {
		__antithesis_instrumentation__.Notify(612074)
	}
	__antithesis_instrumentation__.Notify(612070)
	d := make([]pretty.Doc, len(*node))
	for i, e := range *node {
		__antithesis_instrumentation__.Notify(612075)
		d[i] = pretty.Fold(pretty.Concat,
			pretty.Text(e.Name.String()),
			p.keywordWithText(" ", "AS", " "),
			p.Doc(e),
		)
	}
	__antithesis_instrumentation__.Notify(612071)
	return p.row("WINDOW", p.commaSeparated(d...))
}

func (node *With) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612076)
	return p.unrow(node.docRow(p))
}

func (node *With) docRow(p *PrettyCfg) pretty.TableRow {
	__antithesis_instrumentation__.Notify(612077)
	if node == nil {
		__antithesis_instrumentation__.Notify(612081)
		return emptyRow
	} else {
		__antithesis_instrumentation__.Notify(612082)
	}
	__antithesis_instrumentation__.Notify(612078)
	d := make([]pretty.Doc, len(node.CTEList))
	for i, cte := range node.CTEList {
		__antithesis_instrumentation__.Notify(612083)
		asString := "AS"
		if cte.Mtr.Set {
			__antithesis_instrumentation__.Notify(612085)
			if !cte.Mtr.Materialize {
				__antithesis_instrumentation__.Notify(612087)
				asString += " NOT"
			} else {
				__antithesis_instrumentation__.Notify(612088)
			}
			__antithesis_instrumentation__.Notify(612086)
			asString += " MATERIALIZED"
		} else {
			__antithesis_instrumentation__.Notify(612089)
		}
		__antithesis_instrumentation__.Notify(612084)
		d[i] = p.nestUnder(
			p.Doc(&cte.Name),
			p.bracketKeyword(asString, " (", p.Doc(cte.Stmt), ")", ""),
		)
	}
	__antithesis_instrumentation__.Notify(612079)
	kw := "WITH"
	if node.Recursive {
		__antithesis_instrumentation__.Notify(612090)
		kw = "WITH RECURSIVE"
	} else {
		__antithesis_instrumentation__.Notify(612091)
	}
	__antithesis_instrumentation__.Notify(612080)
	return p.row(kw, p.commaSeparated(d...))
}

func (node *Subquery) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612092)
	d := pretty.Text("<unknown>")
	if node.Select != nil {
		__antithesis_instrumentation__.Notify(612095)
		d = p.Doc(node.Select)
	} else {
		__antithesis_instrumentation__.Notify(612096)
	}
	__antithesis_instrumentation__.Notify(612093)
	if node.Exists {
		__antithesis_instrumentation__.Notify(612097)
		d = pretty.Concat(
			pretty.Keyword("EXISTS"),
			d,
		)
	} else {
		__antithesis_instrumentation__.Notify(612098)
	}
	__antithesis_instrumentation__.Notify(612094)
	return d
}

func (node *AliasedTableExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612099)
	d := p.Doc(node.Expr)
	if node.Lateral {
		__antithesis_instrumentation__.Notify(612104)
		d = pretty.Concat(
			p.keywordWithText("", "LATERAL", " "),
			d,
		)
	} else {
		__antithesis_instrumentation__.Notify(612105)
	}
	__antithesis_instrumentation__.Notify(612100)
	if node.IndexFlags != nil {
		__antithesis_instrumentation__.Notify(612106)
		d = pretty.Concat(
			d,
			p.Doc(node.IndexFlags),
		)
	} else {
		__antithesis_instrumentation__.Notify(612107)
	}
	__antithesis_instrumentation__.Notify(612101)
	if node.Ordinality {
		__antithesis_instrumentation__.Notify(612108)
		d = pretty.Concat(
			d,
			p.keywordWithText(" ", "WITH ORDINALITY", ""),
		)
	} else {
		__antithesis_instrumentation__.Notify(612109)
	}
	__antithesis_instrumentation__.Notify(612102)
	if node.As.Alias != "" {
		__antithesis_instrumentation__.Notify(612110)
		d = p.nestUnder(
			d,
			pretty.Concat(
				p.keywordWithText("", "AS", " "),
				p.Doc(&node.As),
			),
		)
	} else {
		__antithesis_instrumentation__.Notify(612111)
	}
	__antithesis_instrumentation__.Notify(612103)
	return d
}

func (node *FuncExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612112)
	d := p.Doc(&node.Func)

	if len(node.Exprs) > 0 {
		__antithesis_instrumentation__.Notify(612117)
		args := node.Exprs.doc(p)
		if node.Type != 0 {
			__antithesis_instrumentation__.Notify(612120)
			args = pretty.ConcatLine(
				pretty.Text(funcTypeName[node.Type]),
				args,
			)
		} else {
			__antithesis_instrumentation__.Notify(612121)
		}
		__antithesis_instrumentation__.Notify(612118)

		if node.AggType == GeneralAgg && func() bool {
			__antithesis_instrumentation__.Notify(612122)
			return len(node.OrderBy) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(612123)
			args = pretty.ConcatSpace(args, node.OrderBy.doc(p))
		} else {
			__antithesis_instrumentation__.Notify(612124)
		}
		__antithesis_instrumentation__.Notify(612119)
		d = pretty.Concat(d, p.bracket("(", args, ")"))
	} else {
		__antithesis_instrumentation__.Notify(612125)
		d = pretty.Concat(d, pretty.Text("()"))
	}
	__antithesis_instrumentation__.Notify(612113)
	if node.AggType == OrderedSetAgg && func() bool {
		__antithesis_instrumentation__.Notify(612126)
		return len(node.OrderBy) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(612127)
		args := node.OrderBy.doc(p)
		d = pretty.Concat(d, p.bracket("WITHIN GROUP (", args, ")"))
	} else {
		__antithesis_instrumentation__.Notify(612128)
	}
	__antithesis_instrumentation__.Notify(612114)
	if node.Filter != nil {
		__antithesis_instrumentation__.Notify(612129)
		d = pretty.Fold(pretty.ConcatSpace,
			d,
			pretty.Keyword("FILTER"),
			p.bracket("(",
				p.nestUnder(pretty.Keyword("WHERE"), p.Doc(node.Filter)),
				")"))
	} else {
		__antithesis_instrumentation__.Notify(612130)
	}
	__antithesis_instrumentation__.Notify(612115)
	if window := node.WindowDef; window != nil {
		__antithesis_instrumentation__.Notify(612131)
		var over pretty.Doc
		if window.Name != "" {
			__antithesis_instrumentation__.Notify(612133)
			over = p.Doc(&window.Name)
		} else {
			__antithesis_instrumentation__.Notify(612134)
			over = p.Doc(window)
		}
		__antithesis_instrumentation__.Notify(612132)
		d = pretty.Fold(pretty.ConcatSpace,
			d,
			pretty.Keyword("OVER"),
			over,
		)
	} else {
		__antithesis_instrumentation__.Notify(612135)
	}
	__antithesis_instrumentation__.Notify(612116)
	return d
}

func (node *WindowDef) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612136)
	rows := make([]pretty.TableRow, 0, 4)
	if node.RefName != "" {
		__antithesis_instrumentation__.Notify(612142)
		rows = append(rows, p.row("", p.Doc(&node.RefName)))
	} else {
		__antithesis_instrumentation__.Notify(612143)
	}
	__antithesis_instrumentation__.Notify(612137)
	if len(node.Partitions) > 0 {
		__antithesis_instrumentation__.Notify(612144)
		rows = append(rows, p.row("PARTITION BY", p.Doc(&node.Partitions)))
	} else {
		__antithesis_instrumentation__.Notify(612145)
	}
	__antithesis_instrumentation__.Notify(612138)
	if len(node.OrderBy) > 0 {
		__antithesis_instrumentation__.Notify(612146)
		rows = append(rows, node.OrderBy.docRow(p))
	} else {
		__antithesis_instrumentation__.Notify(612147)
	}
	__antithesis_instrumentation__.Notify(612139)
	if node.Frame != nil {
		__antithesis_instrumentation__.Notify(612148)
		rows = append(rows, node.Frame.docRow(p))
	} else {
		__antithesis_instrumentation__.Notify(612149)
	}
	__antithesis_instrumentation__.Notify(612140)
	if len(rows) == 0 {
		__antithesis_instrumentation__.Notify(612150)
		return pretty.Text("()")
	} else {
		__antithesis_instrumentation__.Notify(612151)
	}
	__antithesis_instrumentation__.Notify(612141)
	return p.bracket("(", p.rlTable(rows...), ")")
}

func (wf *WindowFrame) docRow(p *PrettyCfg) pretty.TableRow {
	__antithesis_instrumentation__.Notify(612152)
	kw := "RANGE"
	if wf.Mode == treewindow.ROWS {
		__antithesis_instrumentation__.Notify(612156)
		kw = "ROWS"
	} else {
		__antithesis_instrumentation__.Notify(612157)
		if wf.Mode == treewindow.GROUPS {
			__antithesis_instrumentation__.Notify(612158)
			kw = "GROUPS"
		} else {
			__antithesis_instrumentation__.Notify(612159)
		}
	}
	__antithesis_instrumentation__.Notify(612153)
	d := p.Doc(wf.Bounds.StartBound)
	if wf.Bounds.EndBound != nil {
		__antithesis_instrumentation__.Notify(612160)
		d = p.rlTable(
			p.row("BETWEEN", d),
			p.row("AND", p.Doc(wf.Bounds.EndBound)),
		)
	} else {
		__antithesis_instrumentation__.Notify(612161)
	}
	__antithesis_instrumentation__.Notify(612154)
	if wf.Exclusion != treewindow.NoExclusion {
		__antithesis_instrumentation__.Notify(612162)
		d = pretty.Stack(d, pretty.Keyword(wf.Exclusion.String()))
	} else {
		__antithesis_instrumentation__.Notify(612163)
	}
	__antithesis_instrumentation__.Notify(612155)
	return p.row(kw, d)
}

func (node *WindowFrameBound) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612164)
	switch node.BoundType {
	case treewindow.UnboundedPreceding:
		__antithesis_instrumentation__.Notify(612165)
		return pretty.Keyword("UNBOUNDED PRECEDING")
	case treewindow.OffsetPreceding:
		__antithesis_instrumentation__.Notify(612166)
		return pretty.ConcatSpace(p.Doc(node.OffsetExpr), pretty.Keyword("PRECEDING"))
	case treewindow.CurrentRow:
		__antithesis_instrumentation__.Notify(612167)
		return pretty.Keyword("CURRENT ROW")
	case treewindow.OffsetFollowing:
		__antithesis_instrumentation__.Notify(612168)
		return pretty.ConcatSpace(p.Doc(node.OffsetExpr), pretty.Keyword("FOLLOWING"))
	case treewindow.UnboundedFollowing:
		__antithesis_instrumentation__.Notify(612169)
		return pretty.Keyword("UNBOUNDED FOLLOWING")
	default:
		__antithesis_instrumentation__.Notify(612170)
		panic(errors.AssertionFailedf("unexpected type %d", errors.Safe(node.BoundType)))
	}
}

func (node *LockingClause) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612171)
	return p.rlTable(node.docTable(p)...)
}

func (node *LockingClause) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612172)
	items := make([]pretty.TableRow, len(*node))
	for i, n := range *node {
		__antithesis_instrumentation__.Notify(612174)
		items[i] = p.row("", p.Doc(n))
	}
	__antithesis_instrumentation__.Notify(612173)
	return items
}

func (node *LockingItem) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612175)
	return p.rlTable(node.docTable(p)...)
}

func (node *LockingItem) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612176)
	if node.Strength == ForNone {
		__antithesis_instrumentation__.Notify(612179)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(612180)
	}
	__antithesis_instrumentation__.Notify(612177)
	items := make([]pretty.TableRow, 0, 3)
	items = append(items, node.Strength.docTable(p)...)
	if len(node.Targets) > 0 {
		__antithesis_instrumentation__.Notify(612181)
		items = append(items, p.row("OF", p.Doc(&node.Targets)))
	} else {
		__antithesis_instrumentation__.Notify(612182)
	}
	__antithesis_instrumentation__.Notify(612178)
	items = append(items, node.WaitPolicy.docTable(p)...)
	return items
}

func (node LockingStrength) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612183)
	return p.rlTable(node.docTable(p)...)
}

func (node LockingStrength) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612184)
	str := node.String()
	if str == "" {
		__antithesis_instrumentation__.Notify(612186)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(612187)
	}
	__antithesis_instrumentation__.Notify(612185)
	return []pretty.TableRow{p.row("", pretty.Keyword(str))}
}

func (node LockingWaitPolicy) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612188)
	return p.rlTable(node.docTable(p)...)
}

func (node LockingWaitPolicy) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612189)
	str := node.String()
	if str == "" {
		__antithesis_instrumentation__.Notify(612191)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(612192)
	}
	__antithesis_instrumentation__.Notify(612190)
	return []pretty.TableRow{p.row("", pretty.Keyword(str))}
}

func (p *PrettyCfg) peelCompOperand(e Expr) Expr {
	__antithesis_instrumentation__.Notify(612193)
	if !p.Simplify {
		__antithesis_instrumentation__.Notify(612196)
		return e
	} else {
		__antithesis_instrumentation__.Notify(612197)
	}
	__antithesis_instrumentation__.Notify(612194)
	stripped := StripParens(e)
	switch stripped.(type) {
	case *FuncExpr, *IndirectionExpr, *UnaryExpr,
		*AnnotateTypeExpr, *CastExpr, *ColumnItem, *UnresolvedName:
		__antithesis_instrumentation__.Notify(612198)
		return stripped
	}
	__antithesis_instrumentation__.Notify(612195)
	return e
}

func (node *ComparisonExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612199)
	opStr := node.Operator.String()

	if node.Operator.Symbol == treecmp.IsDistinctFrom && func() bool {
		__antithesis_instrumentation__.Notify(612202)
		return (node.Right == DBoolTrue || func() bool {
			__antithesis_instrumentation__.Notify(612203)
			return node.Right == DBoolFalse == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(612204)
		opStr = "IS NOT"
	} else {
		__antithesis_instrumentation__.Notify(612205)
		if node.Operator.Symbol == treecmp.IsNotDistinctFrom && func() bool {
			__antithesis_instrumentation__.Notify(612206)
			return (node.Right == DBoolTrue || func() bool {
				__antithesis_instrumentation__.Notify(612207)
				return node.Right == DBoolFalse == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(612208)
			opStr = "IS"
		} else {
			__antithesis_instrumentation__.Notify(612209)
		}
	}
	__antithesis_instrumentation__.Notify(612200)
	opDoc := pretty.Keyword(opStr)
	if node.Operator.Symbol.HasSubOperator() {
		__antithesis_instrumentation__.Notify(612210)
		opDoc = pretty.ConcatSpace(pretty.Text(node.SubOperator.String()), opDoc)
	} else {
		__antithesis_instrumentation__.Notify(612211)
	}
	__antithesis_instrumentation__.Notify(612201)
	return pretty.Group(
		pretty.JoinNestedRight(
			opDoc,
			p.Doc(p.peelCompOperand(node.Left)),
			p.Doc(p.peelCompOperand(node.Right))))
}

func (node *AliasClause) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612212)
	d := pretty.Text(node.Alias.String())
	if len(node.Cols) != 0 {
		__antithesis_instrumentation__.Notify(612214)
		d = p.nestUnder(d, p.bracket("(", p.Doc(&node.Cols), ")"))
	} else {
		__antithesis_instrumentation__.Notify(612215)
	}
	__antithesis_instrumentation__.Notify(612213)
	return d
}

func (node *JoinTableExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612216)

	var buf bytes.Buffer
	cond := pretty.Nil
	if _, isNatural := node.Cond.(NaturalJoinCond); isNatural {
		__antithesis_instrumentation__.Notify(612219)

		buf.WriteString("NATURAL ")
	} else {
		__antithesis_instrumentation__.Notify(612220)

		if node.Cond != nil {
			__antithesis_instrumentation__.Notify(612221)
			cond = p.Doc(node.Cond)
		} else {
			__antithesis_instrumentation__.Notify(612222)
		}
	}
	__antithesis_instrumentation__.Notify(612217)

	if node.JoinType != "" {
		__antithesis_instrumentation__.Notify(612223)
		buf.WriteString(node.JoinType)
		buf.WriteByte(' ')
		if node.Hint != "" {
			__antithesis_instrumentation__.Notify(612224)
			buf.WriteString(node.Hint)
			buf.WriteByte(' ')
		} else {
			__antithesis_instrumentation__.Notify(612225)
		}
	} else {
		__antithesis_instrumentation__.Notify(612226)
	}
	__antithesis_instrumentation__.Notify(612218)
	buf.WriteString("JOIN")

	return p.joinNestedOuter(
		buf.String(),
		p.Doc(node.Left),
		pretty.ConcatSpace(p.Doc(node.Right), cond))
}

func (node *OnJoinCond) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612227)
	e := node.Expr
	if p.Simplify {
		__antithesis_instrumentation__.Notify(612229)
		e = StripParens(e)
	} else {
		__antithesis_instrumentation__.Notify(612230)
	}
	__antithesis_instrumentation__.Notify(612228)
	return p.nestUnder(pretty.Keyword("ON"), p.Doc(e))
}

func (node *Insert) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612231)
	items := make([]pretty.TableRow, 0, 8)
	items = append(items, node.With.docRow(p))
	kw := "INSERT"
	if node.OnConflict.IsUpsertAlias() {
		__antithesis_instrumentation__.Notify(612236)
		kw = "UPSERT"
	} else {
		__antithesis_instrumentation__.Notify(612237)
	}
	__antithesis_instrumentation__.Notify(612232)
	items = append(items, p.row(kw, pretty.Nil))

	into := p.Doc(node.Table)
	if node.Columns != nil {
		__antithesis_instrumentation__.Notify(612238)
		into = p.nestUnder(into, p.bracket("(", p.Doc(&node.Columns), ")"))
	} else {
		__antithesis_instrumentation__.Notify(612239)
	}
	__antithesis_instrumentation__.Notify(612233)
	items = append(items, p.row("INTO", into))

	if node.DefaultValues() {
		__antithesis_instrumentation__.Notify(612240)
		items = append(items, p.row("", pretty.Keyword("DEFAULT VALUES")))
	} else {
		__antithesis_instrumentation__.Notify(612241)
		items = append(items, node.Rows.docTable(p)...)
	}
	__antithesis_instrumentation__.Notify(612234)

	if node.OnConflict != nil && func() bool {
		__antithesis_instrumentation__.Notify(612242)
		return !node.OnConflict.IsUpsertAlias() == true
	}() == true {
		__antithesis_instrumentation__.Notify(612243)
		cond := pretty.Nil
		if len(node.OnConflict.Constraint) > 0 {
			__antithesis_instrumentation__.Notify(612247)
			cond = p.nestUnder(pretty.Text("ON CONSTRAINT"), p.Doc(&node.OnConflict.Constraint))
		} else {
			__antithesis_instrumentation__.Notify(612248)
		}
		__antithesis_instrumentation__.Notify(612244)
		if len(node.OnConflict.Columns) > 0 {
			__antithesis_instrumentation__.Notify(612249)
			cond = p.bracket("(", p.Doc(&node.OnConflict.Columns), ")")
		} else {
			__antithesis_instrumentation__.Notify(612250)
		}
		__antithesis_instrumentation__.Notify(612245)
		items = append(items, p.row("ON CONFLICT", cond))
		if node.OnConflict.ArbiterPredicate != nil {
			__antithesis_instrumentation__.Notify(612251)
			items = append(items, p.row("WHERE", p.Doc(node.OnConflict.ArbiterPredicate)))
		} else {
			__antithesis_instrumentation__.Notify(612252)
		}
		__antithesis_instrumentation__.Notify(612246)

		if node.OnConflict.DoNothing {
			__antithesis_instrumentation__.Notify(612253)
			items = append(items, p.row("DO", pretty.Keyword("NOTHING")))
		} else {
			__antithesis_instrumentation__.Notify(612254)
			items = append(items, p.row("DO",
				p.nestUnder(pretty.Keyword("UPDATE SET"), p.Doc(&node.OnConflict.Exprs))))
			if node.OnConflict.Where != nil {
				__antithesis_instrumentation__.Notify(612255)
				items = append(items, node.OnConflict.Where.docRow(p))
			} else {
				__antithesis_instrumentation__.Notify(612256)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(612257)
	}
	__antithesis_instrumentation__.Notify(612235)

	items = append(items, p.docReturning(node.Returning))
	return p.rlTable(items...)
}

func (node *NameList) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612258)
	d := make([]pretty.Doc, len(*node))
	for i, n := range *node {
		__antithesis_instrumentation__.Notify(612260)
		d[i] = p.Doc(&n)
	}
	__antithesis_instrumentation__.Notify(612259)
	return p.commaSeparated(d...)
}

func (node *CastExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612261)
	typ := pretty.Text(node.Type.SQLString())

	switch node.SyntaxMode {
	case CastPrepend:
		__antithesis_instrumentation__.Notify(612262)

		if _, ok := node.Expr.(*StrVal); ok {
			__antithesis_instrumentation__.Notify(612269)
			return pretty.Fold(pretty.Concat,
				typ,
				pretty.Text(" "),
				p.Doc(node.Expr),
			)
		} else {
			__antithesis_instrumentation__.Notify(612270)
		}
		__antithesis_instrumentation__.Notify(612263)
		fallthrough
	case CastShort:
		__antithesis_instrumentation__.Notify(612264)
		if typ, ok := GetStaticallyKnownType(node.Type); ok {
			__antithesis_instrumentation__.Notify(612271)
			switch typ.Family() {
			case types.JsonFamily:
				__antithesis_instrumentation__.Notify(612272)
				if sv, ok := node.Expr.(*StrVal); ok && func() bool {
					__antithesis_instrumentation__.Notify(612274)
					return p.JSONFmt == true
				}() == true {
					__antithesis_instrumentation__.Notify(612275)
					return p.jsonCast(sv, "::", typ)
				} else {
					__antithesis_instrumentation__.Notify(612276)
				}
			default:
				__antithesis_instrumentation__.Notify(612273)
			}
		} else {
			__antithesis_instrumentation__.Notify(612277)
		}
		__antithesis_instrumentation__.Notify(612265)
		return pretty.Fold(pretty.Concat,
			p.exprDocWithParen(node.Expr),
			pretty.Text("::"),
			typ,
		)
	default:
		__antithesis_instrumentation__.Notify(612266)
		if nTyp, ok := GetStaticallyKnownType(node.Type); ok && func() bool {
			__antithesis_instrumentation__.Notify(612278)
			return nTyp.Family() == types.CollatedStringFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(612279)

			strTyp := types.MakeScalar(
				types.StringFamily,
				nTyp.Oid(),
				nTyp.Precision(),
				nTyp.Width(),
				"",
			)
			typ = pretty.Text(strTyp.SQLString())
		} else {
			__antithesis_instrumentation__.Notify(612280)
		}
		__antithesis_instrumentation__.Notify(612267)

		ret := pretty.Fold(pretty.Concat,
			pretty.Keyword("CAST"),
			p.bracket(
				"(",
				p.nestUnder(
					p.Doc(node.Expr),
					pretty.Concat(
						p.keywordWithText("", "AS", " "),
						typ,
					),
				),
				")",
			),
		)

		if nTyp, ok := GetStaticallyKnownType(node.Type); ok && func() bool {
			__antithesis_instrumentation__.Notify(612281)
			return nTyp.Family() == types.CollatedStringFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(612282)
			ret = pretty.Fold(pretty.ConcatSpace,
				ret,
				pretty.Keyword("COLLATE"),
				pretty.Text(nTyp.Locale()))
		} else {
			__antithesis_instrumentation__.Notify(612283)
		}
		__antithesis_instrumentation__.Notify(612268)
		return ret
	}
}

func (node *ValuesClause) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612284)
	return p.rlTable(node.docTable(p)...)
}

func (node *ValuesClause) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612285)
	d := make([]pretty.Doc, len(node.Rows))
	for i := range node.Rows {
		__antithesis_instrumentation__.Notify(612287)
		d[i] = p.bracket("(", p.Doc(&node.Rows[i]), ")")
	}
	__antithesis_instrumentation__.Notify(612286)
	return []pretty.TableRow{p.row("VALUES", p.commaSeparated(d...))}
}

func (node *StatementSource) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612288)
	return p.bracket("[", p.Doc(node.Statement), "]")
}

func (node *RowsFromExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612289)
	if p.Simplify && func() bool {
		__antithesis_instrumentation__.Notify(612291)
		return len(node.Items) == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(612292)
		return p.Doc(node.Items[0])
	} else {
		__antithesis_instrumentation__.Notify(612293)
	}
	__antithesis_instrumentation__.Notify(612290)
	return p.bracketKeyword("ROWS FROM", " (", p.Doc(&node.Items), ")", "")
}

func (node *Array) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612294)
	return p.bracketKeyword("ARRAY", "[", p.Doc(&node.Exprs), "]", "")
}

func (node *Tuple) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612295)
	exprDoc := p.Doc(&node.Exprs)
	if len(node.Exprs) == 1 {
		__antithesis_instrumentation__.Notify(612298)
		exprDoc = pretty.Concat(exprDoc, pretty.Text(","))
	} else {
		__antithesis_instrumentation__.Notify(612299)
	}
	__antithesis_instrumentation__.Notify(612296)
	d := p.bracket("(", exprDoc, ")")
	if len(node.Labels) > 0 {
		__antithesis_instrumentation__.Notify(612300)
		labels := make([]pretty.Doc, len(node.Labels))
		for i, n := range node.Labels {
			__antithesis_instrumentation__.Notify(612302)
			labels[i] = p.Doc((*Name)(&n))
		}
		__antithesis_instrumentation__.Notify(612301)
		d = p.bracket("(", pretty.Stack(
			d,
			p.nestUnder(pretty.Keyword("AS"), p.commaSeparated(labels...)),
		), ")")
	} else {
		__antithesis_instrumentation__.Notify(612303)
	}
	__antithesis_instrumentation__.Notify(612297)
	return d
}

func (node *UpdateExprs) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612304)
	d := make([]pretty.Doc, len(*node))
	for i, n := range *node {
		__antithesis_instrumentation__.Notify(612306)
		d[i] = p.Doc(n)
	}
	__antithesis_instrumentation__.Notify(612305)
	return p.commaSeparated(d...)
}

func (p *PrettyCfg) exprDocWithParen(e Expr) pretty.Doc {
	__antithesis_instrumentation__.Notify(612307)
	if _, ok := e.(operatorExpr); ok {
		__antithesis_instrumentation__.Notify(612309)
		return p.bracket("(", p.Doc(e), ")")
	} else {
		__antithesis_instrumentation__.Notify(612310)
	}
	__antithesis_instrumentation__.Notify(612308)
	return p.Doc(e)
}

func (node *Update) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612311)
	items := make([]pretty.TableRow, 8)
	items = append(items,
		node.With.docRow(p),
		p.row("UPDATE", p.Doc(node.Table)),
		p.row("SET", p.Doc(&node.Exprs)))
	if len(node.From) > 0 {
		__antithesis_instrumentation__.Notify(612313)
		items = append(items,
			p.row("FROM", p.Doc(&node.From)))
	} else {
		__antithesis_instrumentation__.Notify(612314)
	}
	__antithesis_instrumentation__.Notify(612312)
	items = append(items,
		node.Where.docRow(p),
		node.OrderBy.docRow(p))
	items = append(items, node.Limit.docTable(p)...)
	items = append(items, p.docReturning(node.Returning))
	return p.rlTable(items...)
}

func (node *Delete) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612315)
	items := make([]pretty.TableRow, 6)
	items = append(items,
		node.With.docRow(p),
		p.row("DELETE FROM", p.Doc(node.Table)),
		node.Where.docRow(p),
		node.OrderBy.docRow(p))
	items = append(items, node.Limit.docTable(p)...)
	items = append(items, p.docReturning(node.Returning))
	return p.rlTable(items...)
}

func (p *PrettyCfg) docReturning(node ReturningClause) pretty.TableRow {
	__antithesis_instrumentation__.Notify(612316)
	switch r := node.(type) {
	case *NoReturningClause:
		__antithesis_instrumentation__.Notify(612317)
		return p.row("", nil)
	case *ReturningNothing:
		__antithesis_instrumentation__.Notify(612318)
		return p.row("RETURNING", pretty.Keyword("NOTHING"))
	case *ReturningExprs:
		__antithesis_instrumentation__.Notify(612319)
		return p.row("RETURNING", p.Doc((*SelectExprs)(r)))
	default:
		__antithesis_instrumentation__.Notify(612320)
		panic(errors.AssertionFailedf("unhandled case: %T", node))
	}
}

func (node *Order) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612321)
	var d pretty.Doc
	if node.OrderType == OrderByColumn {
		__antithesis_instrumentation__.Notify(612325)
		d = p.Doc(node.Expr)
	} else {
		__antithesis_instrumentation__.Notify(612326)
		if node.Index == "" {
			__antithesis_instrumentation__.Notify(612327)
			d = pretty.ConcatSpace(
				pretty.Keyword("PRIMARY KEY"),
				p.Doc(&node.Table),
			)
		} else {
			__antithesis_instrumentation__.Notify(612328)
			d = pretty.ConcatSpace(
				pretty.Keyword("INDEX"),
				pretty.Fold(pretty.Concat,
					p.Doc(&node.Table),
					pretty.Text("@"),
					p.Doc(&node.Index),
				),
			)
		}
	}
	__antithesis_instrumentation__.Notify(612322)
	if node.Direction != DefaultDirection {
		__antithesis_instrumentation__.Notify(612329)
		d = p.nestUnder(d, pretty.Text(node.Direction.String()))
	} else {
		__antithesis_instrumentation__.Notify(612330)
	}
	__antithesis_instrumentation__.Notify(612323)
	if node.NullsOrder != DefaultNullsOrder {
		__antithesis_instrumentation__.Notify(612331)
		d = p.nestUnder(d, pretty.Text(node.NullsOrder.String()))
	} else {
		__antithesis_instrumentation__.Notify(612332)
	}
	__antithesis_instrumentation__.Notify(612324)
	return d
}

func (node *UpdateExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612333)
	d := p.Doc(&node.Names)
	if node.Tuple {
		__antithesis_instrumentation__.Notify(612336)
		d = p.bracket("(", d, ")")
	} else {
		__antithesis_instrumentation__.Notify(612337)
	}
	__antithesis_instrumentation__.Notify(612334)
	e := node.Expr
	if p.Simplify {
		__antithesis_instrumentation__.Notify(612338)
		e = StripParens(e)
	} else {
		__antithesis_instrumentation__.Notify(612339)
	}
	__antithesis_instrumentation__.Notify(612335)
	return p.nestUnder(d, pretty.ConcatSpace(pretty.Text("="), p.Doc(e)))
}

func (node *CreateTable) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612340)

	title := pretty.Keyword("CREATE")
	switch node.Persistence {
	case PersistenceTemporary:
		__antithesis_instrumentation__.Notify(612349)
		title = pretty.ConcatSpace(title, pretty.Keyword("TEMPORARY"))
	case PersistenceUnlogged:
		__antithesis_instrumentation__.Notify(612350)
		title = pretty.ConcatSpace(title, pretty.Keyword("UNLOGGED"))
	default:
		__antithesis_instrumentation__.Notify(612351)
	}
	__antithesis_instrumentation__.Notify(612341)
	title = pretty.ConcatSpace(title, pretty.Keyword("TABLE"))
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(612352)
		title = pretty.ConcatSpace(title, pretty.Keyword("IF NOT EXISTS"))
	} else {
		__antithesis_instrumentation__.Notify(612353)
	}
	__antithesis_instrumentation__.Notify(612342)
	title = pretty.ConcatSpace(title, p.Doc(&node.Table))

	if node.As() {
		__antithesis_instrumentation__.Notify(612354)
		if len(node.Defs) > 0 {
			__antithesis_instrumentation__.Notify(612356)
			title = pretty.ConcatSpace(title,
				p.bracket("(", p.Doc(&node.Defs), ")"))
		} else {
			__antithesis_instrumentation__.Notify(612357)
		}
		__antithesis_instrumentation__.Notify(612355)
		title = pretty.ConcatSpace(title, pretty.Keyword("AS"))
	} else {
		__antithesis_instrumentation__.Notify(612358)
		title = pretty.ConcatSpace(title,
			p.bracket("(", p.Doc(&node.Defs), ")"),
		)
	}
	__antithesis_instrumentation__.Notify(612343)

	clauses := make([]pretty.Doc, 0, 4)
	if node.As() {
		__antithesis_instrumentation__.Notify(612359)
		clauses = append(clauses, p.Doc(node.AsSource))
	} else {
		__antithesis_instrumentation__.Notify(612360)
	}
	__antithesis_instrumentation__.Notify(612344)
	if node.PartitionByTable != nil {
		__antithesis_instrumentation__.Notify(612361)
		clauses = append(clauses, p.Doc(node.PartitionByTable))
	} else {
		__antithesis_instrumentation__.Notify(612362)
	}
	__antithesis_instrumentation__.Notify(612345)
	if node.StorageParams != nil {
		__antithesis_instrumentation__.Notify(612363)
		clauses = append(
			clauses,
			pretty.ConcatSpace(
				pretty.Keyword(`WITH`),
				p.bracket(`(`, p.Doc(&node.StorageParams), `)`),
			),
		)
	} else {
		__antithesis_instrumentation__.Notify(612364)
	}
	__antithesis_instrumentation__.Notify(612346)
	if node.Locality != nil {
		__antithesis_instrumentation__.Notify(612365)
		clauses = append(clauses, p.Doc(node.Locality))
	} else {
		__antithesis_instrumentation__.Notify(612366)
	}
	__antithesis_instrumentation__.Notify(612347)
	if len(clauses) == 0 {
		__antithesis_instrumentation__.Notify(612367)
		return title
	} else {
		__antithesis_instrumentation__.Notify(612368)
	}
	__antithesis_instrumentation__.Notify(612348)
	return p.nestUnder(title, pretty.Group(pretty.Stack(clauses...)))
}

func (node *CreateView) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612369)

	title := pretty.Keyword("CREATE")
	if node.Replace {
		__antithesis_instrumentation__.Notify(612375)
		title = pretty.ConcatSpace(title, pretty.Keyword("OR REPLACE"))
	} else {
		__antithesis_instrumentation__.Notify(612376)
	}
	__antithesis_instrumentation__.Notify(612370)
	if node.Persistence == PersistenceTemporary {
		__antithesis_instrumentation__.Notify(612377)
		title = pretty.ConcatSpace(title, pretty.Keyword("TEMPORARY"))
	} else {
		__antithesis_instrumentation__.Notify(612378)
	}
	__antithesis_instrumentation__.Notify(612371)
	if node.Materialized {
		__antithesis_instrumentation__.Notify(612379)
		title = pretty.ConcatSpace(title, pretty.Keyword("MATERIALIZED"))
	} else {
		__antithesis_instrumentation__.Notify(612380)
	}
	__antithesis_instrumentation__.Notify(612372)
	title = pretty.ConcatSpace(title, pretty.Keyword("VIEW"))
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(612381)
		title = pretty.ConcatSpace(title, pretty.Keyword("IF NOT EXISTS"))
	} else {
		__antithesis_instrumentation__.Notify(612382)
	}
	__antithesis_instrumentation__.Notify(612373)
	d := pretty.ConcatSpace(
		title,
		p.Doc(&node.Name),
	)
	if len(node.ColumnNames) > 0 {
		__antithesis_instrumentation__.Notify(612383)
		d = pretty.ConcatSpace(
			d,
			p.bracket("(", p.Doc(&node.ColumnNames), ")"),
		)
	} else {
		__antithesis_instrumentation__.Notify(612384)
	}
	__antithesis_instrumentation__.Notify(612374)
	return p.nestUnder(
		pretty.ConcatSpace(d, pretty.Keyword("AS")),
		p.Doc(node.AsSource),
	)
}

func (node *TableDefs) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612385)

	defs := *node
	colDefRows := make([]pretty.TableRow, 0, len(defs))
	items := make([]pretty.Doc, 0, len(defs))

	for i := 0; i < len(defs); i++ {
		__antithesis_instrumentation__.Notify(612387)
		if _, ok := defs[i].(*ColumnTableDef); ok {
			__antithesis_instrumentation__.Notify(612388)

			j := i
			colDefRows = colDefRows[:0]
			for ; j < len(defs); j++ {
				__antithesis_instrumentation__.Notify(612391)
				cdef, ok := defs[j].(*ColumnTableDef)
				if !ok {
					__antithesis_instrumentation__.Notify(612393)
					break
				} else {
					__antithesis_instrumentation__.Notify(612394)
				}
				__antithesis_instrumentation__.Notify(612392)
				colDefRows = append(colDefRows, cdef.docRow(p))
			}
			__antithesis_instrumentation__.Notify(612389)

			i = j - 1

			for j = 0; j < len(colDefRows)-1; j++ {
				__antithesis_instrumentation__.Notify(612395)
				colDefRows[j].Doc = pretty.Concat(colDefRows[j].Doc, pretty.Text(","))
			}
			__antithesis_instrumentation__.Notify(612390)
			items = append(items, p.llTable(pretty.Text, colDefRows...))
		} else {
			__antithesis_instrumentation__.Notify(612396)

			items = append(items, p.Doc(defs[i]))
		}
	}
	__antithesis_instrumentation__.Notify(612386)

	return p.commaSeparated(items...)
}

func (node *CaseExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612397)
	d := make([]pretty.Doc, 0, len(node.Whens)+3)
	c := pretty.Keyword("CASE")
	if node.Expr != nil {
		__antithesis_instrumentation__.Notify(612401)
		c = pretty.Group(pretty.ConcatSpace(c, p.Doc(node.Expr)))
	} else {
		__antithesis_instrumentation__.Notify(612402)
	}
	__antithesis_instrumentation__.Notify(612398)
	d = append(d, c)
	for _, when := range node.Whens {
		__antithesis_instrumentation__.Notify(612403)
		d = append(d, p.Doc(when))
	}
	__antithesis_instrumentation__.Notify(612399)
	if node.Else != nil {
		__antithesis_instrumentation__.Notify(612404)
		d = append(d, pretty.Group(pretty.ConcatSpace(
			pretty.Keyword("ELSE"),
			p.Doc(node.Else),
		)))
	} else {
		__antithesis_instrumentation__.Notify(612405)
	}
	__antithesis_instrumentation__.Notify(612400)
	d = append(d, pretty.Keyword("END"))
	return pretty.Stack(d...)
}

func (node *When) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612406)
	return pretty.Group(pretty.ConcatLine(
		pretty.Group(pretty.ConcatSpace(
			pretty.Keyword("WHEN"),
			p.Doc(node.Cond),
		)),
		pretty.Group(pretty.ConcatSpace(
			pretty.Keyword("THEN"),
			p.Doc(node.Val),
		)),
	))
}

func (node *UnionClause) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612407)
	op := node.Type.String()
	if node.All {
		__antithesis_instrumentation__.Notify(612409)
		op += " ALL"
	} else {
		__antithesis_instrumentation__.Notify(612410)
	}
	__antithesis_instrumentation__.Notify(612408)
	return pretty.Stack(p.Doc(node.Left), p.nestUnder(pretty.Keyword(op), p.Doc(node.Right)))
}

func (node *IfErrExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612411)
	var s string
	if node.Else != nil {
		__antithesis_instrumentation__.Notify(612415)
		s = "IFERROR"
	} else {
		__antithesis_instrumentation__.Notify(612416)
		s = "ISERROR"
	}
	__antithesis_instrumentation__.Notify(612412)
	d := []pretty.Doc{p.Doc(node.Cond)}
	if node.Else != nil {
		__antithesis_instrumentation__.Notify(612417)
		d = append(d, p.Doc(node.Else))
	} else {
		__antithesis_instrumentation__.Notify(612418)
	}
	__antithesis_instrumentation__.Notify(612413)
	if node.ErrCode != nil {
		__antithesis_instrumentation__.Notify(612419)
		d = append(d, p.Doc(node.ErrCode))
	} else {
		__antithesis_instrumentation__.Notify(612420)
	}
	__antithesis_instrumentation__.Notify(612414)
	return p.bracketKeyword(s, "(", p.commaSeparated(d...), ")", "")
}

func (node *IfExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612421)
	return p.bracketKeyword("IF", "(",
		p.commaSeparated(
			p.Doc(node.Cond),
			p.Doc(node.True),
			p.Doc(node.Else),
		), ")", "")
}

func (node *NullIfExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612422)
	return p.bracketKeyword("NULLIF", "(",
		p.commaSeparated(
			p.Doc(node.Expr1),
			p.Doc(node.Expr2),
		), ")", "")
}

func (node *PartitionByTable) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612423)

	var kw string
	kw = `PARTITION `
	if node.All {
		__antithesis_instrumentation__.Notify(612425)
		kw += `ALL `
	} else {
		__antithesis_instrumentation__.Notify(612426)
	}
	__antithesis_instrumentation__.Notify(612424)
	return node.PartitionBy.docInner(p, kw+`BY `)
}

func (node *PartitionBy) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612427)

	return node.docInner(p, `PARTITION BY `)
}

func (node *PartitionBy) docInner(p *PrettyCfg, kw string) pretty.Doc {
	__antithesis_instrumentation__.Notify(612428)
	if node == nil {
		__antithesis_instrumentation__.Notify(612433)
		return pretty.Keyword(kw + `NOTHING`)
	} else {
		__antithesis_instrumentation__.Notify(612434)
	}
	__antithesis_instrumentation__.Notify(612429)
	if len(node.List) > 0 {
		__antithesis_instrumentation__.Notify(612435)
		kw += `LIST`
	} else {
		__antithesis_instrumentation__.Notify(612436)
		if len(node.Range) > 0 {
			__antithesis_instrumentation__.Notify(612437)
			kw += `RANGE`
		} else {
			__antithesis_instrumentation__.Notify(612438)
		}
	}
	__antithesis_instrumentation__.Notify(612430)
	title := pretty.ConcatSpace(pretty.Keyword(kw),
		p.bracket("(", p.Doc(&node.Fields), ")"))

	inner := make([]pretty.Doc, 0, len(node.List)+len(node.Range))
	for _, v := range node.List {
		__antithesis_instrumentation__.Notify(612439)
		inner = append(inner, p.Doc(&v))
	}
	__antithesis_instrumentation__.Notify(612431)
	for _, v := range node.Range {
		__antithesis_instrumentation__.Notify(612440)
		inner = append(inner, p.Doc(&v))
	}
	__antithesis_instrumentation__.Notify(612432)
	return p.nestUnder(title,
		p.bracket("(", p.commaSeparated(inner...), ")"),
	)
}

func (node *Locality) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612441)

	localityKW := pretty.Keyword("LOCALITY")
	switch node.LocalityLevel {
	case LocalityLevelGlobal:
		__antithesis_instrumentation__.Notify(612443)
		return pretty.ConcatSpace(localityKW, pretty.Keyword("GLOBAL"))
	case LocalityLevelRow:
		__antithesis_instrumentation__.Notify(612444)
		ret := pretty.ConcatSpace(localityKW, pretty.Keyword("REGIONAL BY ROW"))
		if node.RegionalByRowColumn != "" {
			__antithesis_instrumentation__.Notify(612449)
			return pretty.ConcatSpace(
				ret,
				pretty.ConcatSpace(
					pretty.Keyword("AS"),
					p.Doc(&node.RegionalByRowColumn),
				),
			)
		} else {
			__antithesis_instrumentation__.Notify(612450)
		}
		__antithesis_instrumentation__.Notify(612445)
		return ret
	case LocalityLevelTable:
		__antithesis_instrumentation__.Notify(612446)
		byTable := pretty.ConcatSpace(localityKW, pretty.Keyword("REGIONAL BY TABLE IN"))
		if node.TableRegion == "" {
			__antithesis_instrumentation__.Notify(612451)
			return pretty.ConcatSpace(
				byTable,
				pretty.Keyword("PRIMARY REGION"),
			)
		} else {
			__antithesis_instrumentation__.Notify(612452)
		}
		__antithesis_instrumentation__.Notify(612447)
		return pretty.ConcatSpace(
			byTable,
			p.Doc(&node.TableRegion),
		)
	default:
		__antithesis_instrumentation__.Notify(612448)
	}
	__antithesis_instrumentation__.Notify(612442)
	panic(fmt.Sprintf("unknown locality: %v", *node))
}

func (node *ListPartition) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612453)

	title := pretty.ConcatSpace(pretty.Keyword("PARTITION"), p.Doc(&node.Name))

	clauses := make([]pretty.Doc, 1, 2)
	clauses[0] = pretty.ConcatSpace(
		pretty.Keyword("VALUES IN"),
		p.bracket("(", p.Doc(&node.Exprs), ")"),
	)
	if node.Subpartition != nil {
		__antithesis_instrumentation__.Notify(612455)
		clauses = append(clauses, p.Doc(node.Subpartition))
	} else {
		__antithesis_instrumentation__.Notify(612456)
	}
	__antithesis_instrumentation__.Notify(612454)
	return p.nestUnder(title, pretty.Group(pretty.Stack(clauses...)))
}

func (node *RangePartition) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612457)

	title := pretty.ConcatSpace(
		pretty.Keyword("PARTITION"),
		p.Doc(&node.Name),
	)

	clauses := make([]pretty.Doc, 2, 3)
	clauses[0] = pretty.ConcatSpace(
		pretty.Keyword("VALUES FROM"),
		p.bracket("(", p.Doc(&node.From), ")"))
	clauses[1] = pretty.ConcatSpace(
		pretty.Keyword("TO"),
		p.bracket("(", p.Doc(&node.To), ")"))

	if node.Subpartition != nil {
		__antithesis_instrumentation__.Notify(612459)
		clauses = append(clauses, p.Doc(node.Subpartition))
	} else {
		__antithesis_instrumentation__.Notify(612460)
	}
	__antithesis_instrumentation__.Notify(612458)

	return p.nestUnder(title, pretty.Group(pretty.Stack(clauses...)))
}

func (node *ShardedIndexDef) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612461)

	if _, ok := node.ShardBuckets.(DefaultVal); ok {
		__antithesis_instrumentation__.Notify(612463)
		return pretty.Keyword("USING HASH")
	} else {
		__antithesis_instrumentation__.Notify(612464)
	}
	__antithesis_instrumentation__.Notify(612462)
	parts := []pretty.Doc{
		pretty.Keyword("USING HASH WITH BUCKET_COUNT = "),
		p.Doc(node.ShardBuckets),
	}
	return pretty.Fold(pretty.ConcatSpace, parts...)
}

func (node *CreateIndex) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612465)

	title := make([]pretty.Doc, 0, 6)
	title = append(title, pretty.Keyword("CREATE"))
	if node.Unique {
		__antithesis_instrumentation__.Notify(612476)
		title = append(title, pretty.Keyword("UNIQUE"))
	} else {
		__antithesis_instrumentation__.Notify(612477)
	}
	__antithesis_instrumentation__.Notify(612466)
	if node.Inverted {
		__antithesis_instrumentation__.Notify(612478)
		title = append(title, pretty.Keyword("INVERTED"))
	} else {
		__antithesis_instrumentation__.Notify(612479)
	}
	__antithesis_instrumentation__.Notify(612467)
	title = append(title, pretty.Keyword("INDEX"))
	if node.Concurrently {
		__antithesis_instrumentation__.Notify(612480)
		title = append(title, pretty.Keyword("CONCURRENTLY"))
	} else {
		__antithesis_instrumentation__.Notify(612481)
	}
	__antithesis_instrumentation__.Notify(612468)
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(612482)
		title = append(title, pretty.Keyword("IF NOT EXISTS"))
	} else {
		__antithesis_instrumentation__.Notify(612483)
	}
	__antithesis_instrumentation__.Notify(612469)
	if node.Name != "" {
		__antithesis_instrumentation__.Notify(612484)
		title = append(title, p.Doc(&node.Name))
	} else {
		__antithesis_instrumentation__.Notify(612485)
	}
	__antithesis_instrumentation__.Notify(612470)

	clauses := make([]pretty.Doc, 0, 5)
	clauses = append(clauses, pretty.Fold(pretty.ConcatSpace,
		pretty.Keyword("ON"),
		p.Doc(&node.Table),
		p.bracket("(", p.Doc(&node.Columns), ")")))

	if node.Sharded != nil {
		__antithesis_instrumentation__.Notify(612486)
		clauses = append(clauses, p.Doc(node.Sharded))
	} else {
		__antithesis_instrumentation__.Notify(612487)
	}
	__antithesis_instrumentation__.Notify(612471)
	if len(node.Storing) > 0 {
		__antithesis_instrumentation__.Notify(612488)
		clauses = append(clauses, p.bracketKeyword(
			"STORING", " (",
			p.Doc(&node.Storing),
			")", "",
		))
	} else {
		__antithesis_instrumentation__.Notify(612489)
	}
	__antithesis_instrumentation__.Notify(612472)
	if node.PartitionByIndex != nil {
		__antithesis_instrumentation__.Notify(612490)
		clauses = append(clauses, p.Doc(node.PartitionByIndex))
	} else {
		__antithesis_instrumentation__.Notify(612491)
	}
	__antithesis_instrumentation__.Notify(612473)
	if node.StorageParams != nil {
		__antithesis_instrumentation__.Notify(612492)
		clauses = append(clauses, p.bracketKeyword(
			"WITH", " (",
			p.Doc(&node.StorageParams),
			")", "",
		))
	} else {
		__antithesis_instrumentation__.Notify(612493)
	}
	__antithesis_instrumentation__.Notify(612474)
	if node.Predicate != nil {
		__antithesis_instrumentation__.Notify(612494)
		clauses = append(clauses, p.nestUnder(pretty.Keyword("WHERE"), p.Doc(node.Predicate)))
	} else {
		__antithesis_instrumentation__.Notify(612495)
	}
	__antithesis_instrumentation__.Notify(612475)
	return p.nestUnder(
		pretty.Fold(pretty.ConcatSpace, title...),
		pretty.Group(pretty.Stack(clauses...)))
}

func (node *FamilyTableDef) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612496)

	d := pretty.Keyword("FAMILY")
	if node.Name != "" {
		__antithesis_instrumentation__.Notify(612498)
		d = pretty.ConcatSpace(d, p.Doc(&node.Name))
	} else {
		__antithesis_instrumentation__.Notify(612499)
	}
	__antithesis_instrumentation__.Notify(612497)
	return pretty.ConcatSpace(d, p.bracket("(", p.Doc(&node.Columns), ")"))
}

func (node *LikeTableDef) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612500)
	d := pretty.Keyword("LIKE")
	d = pretty.ConcatSpace(d, p.Doc(&node.Name))
	for _, opt := range node.Options {
		__antithesis_instrumentation__.Notify(612502)
		word := "INCLUDING"
		if opt.Excluded {
			__antithesis_instrumentation__.Notify(612504)
			word = "EXCLUDING"
		} else {
			__antithesis_instrumentation__.Notify(612505)
		}
		__antithesis_instrumentation__.Notify(612503)
		d = pretty.ConcatSpace(d, pretty.Keyword(word))
		d = pretty.ConcatSpace(d, pretty.Keyword(opt.Opt.String()))
	}
	__antithesis_instrumentation__.Notify(612501)
	return d
}

func (node *IndexTableDef) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612506)

	title := pretty.Keyword("INDEX")
	if node.Name != "" {
		__antithesis_instrumentation__.Notify(612515)
		title = pretty.ConcatSpace(title, p.Doc(&node.Name))
	} else {
		__antithesis_instrumentation__.Notify(612516)
	}
	__antithesis_instrumentation__.Notify(612507)
	if node.Inverted {
		__antithesis_instrumentation__.Notify(612517)
		title = pretty.ConcatSpace(pretty.Keyword("INVERTED"), title)
	} else {
		__antithesis_instrumentation__.Notify(612518)
	}
	__antithesis_instrumentation__.Notify(612508)
	title = pretty.ConcatSpace(title, p.bracket("(", p.Doc(&node.Columns), ")"))

	clauses := make([]pretty.Doc, 0, 4)
	if node.Sharded != nil {
		__antithesis_instrumentation__.Notify(612519)
		clauses = append(clauses, p.Doc(node.Sharded))
	} else {
		__antithesis_instrumentation__.Notify(612520)
	}
	__antithesis_instrumentation__.Notify(612509)
	if node.Storing != nil {
		__antithesis_instrumentation__.Notify(612521)
		clauses = append(clauses, p.bracketKeyword(
			"STORING", "(",
			p.Doc(&node.Storing),
			")", ""))
	} else {
		__antithesis_instrumentation__.Notify(612522)
	}
	__antithesis_instrumentation__.Notify(612510)
	if node.PartitionByIndex != nil {
		__antithesis_instrumentation__.Notify(612523)
		clauses = append(clauses, p.Doc(node.PartitionByIndex))
	} else {
		__antithesis_instrumentation__.Notify(612524)
	}
	__antithesis_instrumentation__.Notify(612511)
	if node.StorageParams != nil {
		__antithesis_instrumentation__.Notify(612525)
		clauses = append(
			clauses,
			p.bracketKeyword("WITH", "(", p.Doc(&node.StorageParams), ")", ""),
		)
	} else {
		__antithesis_instrumentation__.Notify(612526)
	}
	__antithesis_instrumentation__.Notify(612512)
	if node.Predicate != nil {
		__antithesis_instrumentation__.Notify(612527)
		clauses = append(clauses, p.nestUnder(pretty.Keyword("WHERE"), p.Doc(node.Predicate)))
	} else {
		__antithesis_instrumentation__.Notify(612528)
	}
	__antithesis_instrumentation__.Notify(612513)

	if len(clauses) == 0 {
		__antithesis_instrumentation__.Notify(612529)
		return title
	} else {
		__antithesis_instrumentation__.Notify(612530)
	}
	__antithesis_instrumentation__.Notify(612514)
	return p.nestUnder(title, pretty.Group(pretty.Stack(clauses...)))
}

func (node *UniqueConstraintTableDef) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612531)

	clauses := make([]pretty.Doc, 0, 5)
	var title pretty.Doc
	if node.PrimaryKey {
		__antithesis_instrumentation__.Notify(612539)
		title = pretty.Keyword("PRIMARY KEY")
	} else {
		__antithesis_instrumentation__.Notify(612540)
		title = pretty.Keyword("UNIQUE")
		if node.WithoutIndex {
			__antithesis_instrumentation__.Notify(612541)
			title = pretty.ConcatSpace(title, pretty.Keyword("WITHOUT INDEX"))
		} else {
			__antithesis_instrumentation__.Notify(612542)
		}
	}
	__antithesis_instrumentation__.Notify(612532)
	title = pretty.ConcatSpace(title, p.bracket("(", p.Doc(&node.Columns), ")"))
	if node.Name != "" {
		__antithesis_instrumentation__.Notify(612543)
		clauses = append(clauses, title)
		title = pretty.ConcatSpace(pretty.Keyword("CONSTRAINT"), p.Doc(&node.Name))
	} else {
		__antithesis_instrumentation__.Notify(612544)
	}
	__antithesis_instrumentation__.Notify(612533)
	if node.Sharded != nil {
		__antithesis_instrumentation__.Notify(612545)
		clauses = append(clauses, p.Doc(node.Sharded))
	} else {
		__antithesis_instrumentation__.Notify(612546)
	}
	__antithesis_instrumentation__.Notify(612534)
	if node.Storing != nil {
		__antithesis_instrumentation__.Notify(612547)
		clauses = append(clauses, p.bracketKeyword(
			"STORING", "(",
			p.Doc(&node.Storing),
			")", ""))
	} else {
		__antithesis_instrumentation__.Notify(612548)
	}
	__antithesis_instrumentation__.Notify(612535)

	if node.PartitionByIndex != nil {
		__antithesis_instrumentation__.Notify(612549)
		clauses = append(clauses, p.Doc(node.PartitionByIndex))
	} else {
		__antithesis_instrumentation__.Notify(612550)
	}
	__antithesis_instrumentation__.Notify(612536)
	if node.Predicate != nil {
		__antithesis_instrumentation__.Notify(612551)
		clauses = append(clauses, p.nestUnder(pretty.Keyword("WHERE"), p.Doc(node.Predicate)))
	} else {
		__antithesis_instrumentation__.Notify(612552)
	}
	__antithesis_instrumentation__.Notify(612537)

	if len(clauses) == 0 {
		__antithesis_instrumentation__.Notify(612553)
		return title
	} else {
		__antithesis_instrumentation__.Notify(612554)
	}
	__antithesis_instrumentation__.Notify(612538)
	return p.nestUnder(title, pretty.Group(pretty.Stack(clauses...)))
}

func (node *ForeignKeyConstraintTableDef) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612555)

	clauses := make([]pretty.Doc, 0, 4)
	title := pretty.ConcatSpace(
		pretty.Keyword("FOREIGN KEY"),
		p.bracket("(", p.Doc(&node.FromCols), ")"))

	if node.Name != "" {
		__antithesis_instrumentation__.Notify(612560)
		clauses = append(clauses, title)
		title = pretty.ConcatSpace(pretty.Keyword("CONSTRAINT"), p.Doc(&node.Name))
	} else {
		__antithesis_instrumentation__.Notify(612561)
	}
	__antithesis_instrumentation__.Notify(612556)

	ref := pretty.ConcatSpace(
		pretty.Keyword("REFERENCES"), p.Doc(&node.Table))
	if len(node.ToCols) > 0 {
		__antithesis_instrumentation__.Notify(612562)
		ref = pretty.ConcatSpace(ref, p.bracket("(", p.Doc(&node.ToCols), ")"))
	} else {
		__antithesis_instrumentation__.Notify(612563)
	}
	__antithesis_instrumentation__.Notify(612557)
	clauses = append(clauses, ref)

	if node.Match != MatchSimple {
		__antithesis_instrumentation__.Notify(612564)
		clauses = append(clauses, pretty.Keyword(node.Match.String()))
	} else {
		__antithesis_instrumentation__.Notify(612565)
	}
	__antithesis_instrumentation__.Notify(612558)

	if actions := p.Doc(&node.Actions); ref != pretty.Nil {
		__antithesis_instrumentation__.Notify(612566)
		clauses = append(clauses, actions)
	} else {
		__antithesis_instrumentation__.Notify(612567)
	}
	__antithesis_instrumentation__.Notify(612559)

	return p.nestUnder(title, pretty.Group(pretty.Stack(clauses...)))
}

func (p *PrettyCfg) maybePrependConstraintName(constraintName *Name, d pretty.Doc) pretty.Doc {
	__antithesis_instrumentation__.Notify(612568)
	if *constraintName != "" {
		__antithesis_instrumentation__.Notify(612570)
		return pretty.Fold(pretty.ConcatSpace,
			pretty.Keyword("CONSTRAINT"),
			p.Doc(constraintName),
			d)
	} else {
		__antithesis_instrumentation__.Notify(612571)
	}
	__antithesis_instrumentation__.Notify(612569)
	return d
}

func (node *ColumnTableDef) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612572)
	return p.unrow(node.docRow(p))
}

func (node *ColumnTableDef) docRow(p *PrettyCfg) pretty.TableRow {
	__antithesis_instrumentation__.Notify(612573)

	clauses := make([]pretty.Doc, 0, 7)

	if node.Type != nil {
		__antithesis_instrumentation__.Notify(612590)
		clauses = append(clauses, pretty.Text(node.columnTypeString()))
	} else {
		__antithesis_instrumentation__.Notify(612591)
	}
	__antithesis_instrumentation__.Notify(612574)

	if node.IsComputed() {
		__antithesis_instrumentation__.Notify(612592)
		var typ string
		if node.Computed.Virtual {
			__antithesis_instrumentation__.Notify(612594)
			typ = "VIRTUAL"
		} else {
			__antithesis_instrumentation__.Notify(612595)
			typ = "STORED"
		}
		__antithesis_instrumentation__.Notify(612593)

		clauses = append(clauses, pretty.ConcatSpace(
			pretty.Keyword("AS"),
			pretty.ConcatSpace(
				p.bracket("(", p.Doc(node.Computed.Expr), ")"),
				pretty.Keyword(typ),
			),
		))
	} else {
		__antithesis_instrumentation__.Notify(612596)
	}
	__antithesis_instrumentation__.Notify(612575)

	if node.GeneratedIdentity.IsGeneratedAsIdentity {
		__antithesis_instrumentation__.Notify(612597)
		var generatedConstraint pretty.Doc
		switch node.GeneratedIdentity.GeneratedAsIdentityType {
		case GeneratedAlways:
			__antithesis_instrumentation__.Notify(612599)
			generatedConstraint = pretty.Keyword("GENERATED ALWAYS AS IDENTITY")
		case GeneratedByDefault:
			__antithesis_instrumentation__.Notify(612600)
			generatedConstraint = pretty.Keyword("GENERATED BY DEFAULT AS IDENTITY")
		default:
			__antithesis_instrumentation__.Notify(612601)
		}
		__antithesis_instrumentation__.Notify(612598)
		clauses = append(clauses, generatedConstraint)
		if node.GeneratedIdentity.SeqOptions != nil {
			__antithesis_instrumentation__.Notify(612602)
			const prettyFlags = FmtShowPasswords | FmtParsable
			curGenSeqOpts := node.GeneratedIdentity.SeqOptions
			txt := AsStringWithFlags(&curGenSeqOpts, prettyFlags)
			bracketedTxt := p.bracket("(", pretty.Text(strings.TrimSpace(txt)), ")")
			clauses = append(clauses, bracketedTxt)
		} else {
			__antithesis_instrumentation__.Notify(612603)
		}
	} else {
		__antithesis_instrumentation__.Notify(612604)
	}
	__antithesis_instrumentation__.Notify(612576)

	if node.HasColumnFamily() {
		__antithesis_instrumentation__.Notify(612605)
		d := pretty.Keyword("FAMILY")
		if node.Family.Name != "" {
			__antithesis_instrumentation__.Notify(612608)
			d = pretty.ConcatSpace(d, p.Doc(&node.Family.Name))
		} else {
			__antithesis_instrumentation__.Notify(612609)
		}
		__antithesis_instrumentation__.Notify(612606)
		if node.Family.Create {
			__antithesis_instrumentation__.Notify(612610)
			c := pretty.Keyword("CREATE")
			if node.Family.IfNotExists {
				__antithesis_instrumentation__.Notify(612612)
				c = pretty.ConcatSpace(c, pretty.Keyword("IF NOT EXISTS"))
			} else {
				__antithesis_instrumentation__.Notify(612613)
			}
			__antithesis_instrumentation__.Notify(612611)
			d = pretty.ConcatSpace(c, d)
		} else {
			__antithesis_instrumentation__.Notify(612614)
		}
		__antithesis_instrumentation__.Notify(612607)
		clauses = append(clauses, d)
	} else {
		__antithesis_instrumentation__.Notify(612615)
	}
	__antithesis_instrumentation__.Notify(612577)

	if node.HasDefaultExpr() {
		__antithesis_instrumentation__.Notify(612616)
		clauses = append(clauses, p.maybePrependConstraintName(&node.DefaultExpr.ConstraintName,
			pretty.ConcatSpace(pretty.Keyword("DEFAULT"), p.Doc(node.DefaultExpr.Expr))))
	} else {
		__antithesis_instrumentation__.Notify(612617)
	}
	__antithesis_instrumentation__.Notify(612578)

	if node.HasOnUpdateExpr() {
		__antithesis_instrumentation__.Notify(612618)
		clauses = append(clauses, p.maybePrependConstraintName(&node.OnUpdateExpr.ConstraintName,
			pretty.ConcatSpace(pretty.Keyword("ON UPDATE"), p.Doc(node.OnUpdateExpr.Expr))))
	} else {
		__antithesis_instrumentation__.Notify(612619)
	}
	__antithesis_instrumentation__.Notify(612579)

	if node.Hidden {
		__antithesis_instrumentation__.Notify(612620)
		hiddenConstraint := pretty.Keyword("NOT VISIBLE")
		clauses = append(clauses, p.maybePrependConstraintName(&node.Nullable.ConstraintName, hiddenConstraint))
	} else {
		__antithesis_instrumentation__.Notify(612621)
	}
	__antithesis_instrumentation__.Notify(612580)

	nConstraint := pretty.Nil
	switch node.Nullable.Nullability {
	case Null:
		__antithesis_instrumentation__.Notify(612622)
		nConstraint = pretty.Keyword("NULL")
	case NotNull:
		__antithesis_instrumentation__.Notify(612623)
		nConstraint = pretty.Keyword("NOT NULL")
	default:
		__antithesis_instrumentation__.Notify(612624)
	}
	__antithesis_instrumentation__.Notify(612581)
	if nConstraint != pretty.Nil {
		__antithesis_instrumentation__.Notify(612625)
		clauses = append(clauses, p.maybePrependConstraintName(&node.Nullable.ConstraintName, nConstraint))
	} else {
		__antithesis_instrumentation__.Notify(612626)
	}
	__antithesis_instrumentation__.Notify(612582)

	pkConstraint := pretty.Nil
	if node.PrimaryKey.IsPrimaryKey {
		__antithesis_instrumentation__.Notify(612627)
		pkConstraint = pretty.Keyword("PRIMARY KEY")
	} else {
		__antithesis_instrumentation__.Notify(612628)
		if node.Unique.IsUnique {
			__antithesis_instrumentation__.Notify(612629)
			pkConstraint = pretty.Keyword("UNIQUE")
			if node.Unique.WithoutIndex {
				__antithesis_instrumentation__.Notify(612630)
				pkConstraint = pretty.ConcatSpace(pkConstraint, pretty.Keyword("WITHOUT INDEX"))
			} else {
				__antithesis_instrumentation__.Notify(612631)
			}
		} else {
			__antithesis_instrumentation__.Notify(612632)
		}
	}
	__antithesis_instrumentation__.Notify(612583)
	if pkConstraint != pretty.Nil {
		__antithesis_instrumentation__.Notify(612633)
		clauses = append(clauses, p.maybePrependConstraintName(&node.Unique.ConstraintName, pkConstraint))
	} else {
		__antithesis_instrumentation__.Notify(612634)
	}
	__antithesis_instrumentation__.Notify(612584)

	pkStorageParams := node.PrimaryKey.StorageParams
	if node.PrimaryKey.Sharded {
		__antithesis_instrumentation__.Notify(612635)
		clauses = append(clauses, pretty.Keyword("USING HASH"))
		bcStorageParam := node.PrimaryKey.StorageParams.GetVal(`bucket_count`)
		if _, ok := node.PrimaryKey.ShardBuckets.(DefaultVal); !ok && func() bool {
			__antithesis_instrumentation__.Notify(612636)
			return bcStorageParam == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(612637)
			pkStorageParams = append(
				pkStorageParams, StorageParam{
					Key:   `bucket_count`,
					Value: node.PrimaryKey.ShardBuckets,
				},
			)
		} else {
			__antithesis_instrumentation__.Notify(612638)
		}
	} else {
		__antithesis_instrumentation__.Notify(612639)
	}
	__antithesis_instrumentation__.Notify(612585)
	if len(pkStorageParams) > 0 {
		__antithesis_instrumentation__.Notify(612640)
		clauses = append(clauses, p.bracketKeyword(
			"WITH", " (",
			p.Doc(&pkStorageParams),
			")", "",
		))
	} else {
		__antithesis_instrumentation__.Notify(612641)
	}
	__antithesis_instrumentation__.Notify(612586)

	for _, checkExpr := range node.CheckExprs {
		__antithesis_instrumentation__.Notify(612642)
		clauses = append(clauses, p.maybePrependConstraintName(&checkExpr.ConstraintName,
			pretty.ConcatSpace(pretty.Keyword("CHECK"), p.bracket("(", p.Doc(checkExpr.Expr), ")"))))
	}
	__antithesis_instrumentation__.Notify(612587)

	if node.HasFKConstraint() {
		__antithesis_instrumentation__.Notify(612643)
		fkHead := pretty.ConcatSpace(pretty.Keyword("REFERENCES"), p.Doc(node.References.Table))
		if node.References.Col != "" {
			__antithesis_instrumentation__.Notify(612648)
			fkHead = pretty.ConcatSpace(fkHead, p.bracket("(", p.Doc(&node.References.Col), ")"))
		} else {
			__antithesis_instrumentation__.Notify(612649)
		}
		__antithesis_instrumentation__.Notify(612644)
		fkDetails := make([]pretty.Doc, 0, 2)

		if node.References.Match != MatchSimple {
			__antithesis_instrumentation__.Notify(612650)
			fkDetails = append(fkDetails, pretty.Keyword(node.References.Match.String()))
		} else {
			__antithesis_instrumentation__.Notify(612651)
		}
		__antithesis_instrumentation__.Notify(612645)
		if ref := p.Doc(&node.References.Actions); ref != pretty.Nil {
			__antithesis_instrumentation__.Notify(612652)
			fkDetails = append(fkDetails, ref)
		} else {
			__antithesis_instrumentation__.Notify(612653)
		}
		__antithesis_instrumentation__.Notify(612646)
		fk := fkHead
		if len(fkDetails) > 0 {
			__antithesis_instrumentation__.Notify(612654)
			fk = p.nestUnder(fk, pretty.Group(pretty.Stack(fkDetails...)))
		} else {
			__antithesis_instrumentation__.Notify(612655)
		}
		__antithesis_instrumentation__.Notify(612647)
		clauses = append(clauses, p.maybePrependConstraintName(&node.References.ConstraintName, fk))
	} else {
		__antithesis_instrumentation__.Notify(612656)
	}
	__antithesis_instrumentation__.Notify(612588)

	var tblRow pretty.TableRow
	if node.Type == nil {
		__antithesis_instrumentation__.Notify(612657)
		tblRow = pretty.TableRow{
			Label: node.Name.String(),
			Doc:   pretty.Stack(clauses...),
		}
	} else {
		__antithesis_instrumentation__.Notify(612658)
		tblRow = pretty.TableRow{
			Label: node.Name.String(),
			Doc:   pretty.Group(pretty.Stack(clauses...)),
		}
	}
	__antithesis_instrumentation__.Notify(612589)

	return tblRow
}

func (node *CheckConstraintTableDef) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612659)

	d := pretty.ConcatSpace(pretty.Keyword("CHECK"),
		p.bracket("(", p.Doc(node.Expr), ")"))

	if node.Name != "" {
		__antithesis_instrumentation__.Notify(612661)
		d = p.nestUnder(
			pretty.ConcatSpace(
				pretty.Keyword("CONSTRAINT"),
				p.Doc(&node.Name),
			),
			d,
		)
	} else {
		__antithesis_instrumentation__.Notify(612662)
	}
	__antithesis_instrumentation__.Notify(612660)
	return d
}

func (node *ReferenceActions) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612663)
	var docs []pretty.Doc
	if node.Delete != NoAction {
		__antithesis_instrumentation__.Notify(612666)
		docs = append(docs,
			pretty.Keyword("ON DELETE"),
			pretty.Keyword(node.Delete.String()),
		)
	} else {
		__antithesis_instrumentation__.Notify(612667)
	}
	__antithesis_instrumentation__.Notify(612664)
	if node.Update != NoAction {
		__antithesis_instrumentation__.Notify(612668)
		docs = append(docs,
			pretty.Keyword("ON UPDATE"),
			pretty.Keyword(node.Update.String()),
		)
	} else {
		__antithesis_instrumentation__.Notify(612669)
	}
	__antithesis_instrumentation__.Notify(612665)
	return pretty.Fold(pretty.ConcatSpace, docs...)
}

func (node *Backup) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612670)
	items := make([]pretty.TableRow, 0, 6)

	items = append(items, p.row("BACKUP", pretty.Nil))
	if node.Targets != nil {
		__antithesis_instrumentation__.Notify(612676)
		items = append(items, node.Targets.docRow(p))
	} else {
		__antithesis_instrumentation__.Notify(612677)
	}
	__antithesis_instrumentation__.Notify(612671)
	if node.Nested {
		__antithesis_instrumentation__.Notify(612678)
		if node.Subdir != nil {
			__antithesis_instrumentation__.Notify(612679)
			items = append(items, p.row("INTO ", p.Doc(node.Subdir)))
			items = append(items, p.row(" IN ", p.Doc(&node.To)))
		} else {
			__antithesis_instrumentation__.Notify(612680)
			if node.AppendToLatest {
				__antithesis_instrumentation__.Notify(612681)
				items = append(items, p.row("INTO LATEST IN", p.Doc(&node.To)))
			} else {
				__antithesis_instrumentation__.Notify(612682)
				items = append(items, p.row("INTO", p.Doc(&node.To)))
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(612683)
		items = append(items, p.row("TO", p.Doc(&node.To)))
	}
	__antithesis_instrumentation__.Notify(612672)

	if node.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(612684)
		items = append(items, node.AsOf.docRow(p))
	} else {
		__antithesis_instrumentation__.Notify(612685)
	}
	__antithesis_instrumentation__.Notify(612673)
	if node.IncrementalFrom != nil {
		__antithesis_instrumentation__.Notify(612686)
		items = append(items, p.row("INCREMENTAL FROM", p.Doc(&node.IncrementalFrom)))
	} else {
		__antithesis_instrumentation__.Notify(612687)
	}
	__antithesis_instrumentation__.Notify(612674)
	if !node.Options.IsDefault() {
		__antithesis_instrumentation__.Notify(612688)
		items = append(items, p.row("WITH", p.Doc(&node.Options)))
	} else {
		__antithesis_instrumentation__.Notify(612689)
	}
	__antithesis_instrumentation__.Notify(612675)
	return p.rlTable(items...)
}

func (node *Restore) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612690)
	items := make([]pretty.TableRow, 0, 5)

	items = append(items, p.row("RESTORE", pretty.Nil))
	if node.DescriptorCoverage == RequestedDescriptors {
		__antithesis_instrumentation__.Notify(612696)
		items = append(items, node.Targets.docRow(p))
	} else {
		__antithesis_instrumentation__.Notify(612697)
	}
	__antithesis_instrumentation__.Notify(612691)
	from := make([]pretty.Doc, len(node.From))
	for i := range node.From {
		__antithesis_instrumentation__.Notify(612698)
		from[i] = p.Doc(&node.From[i])
	}
	__antithesis_instrumentation__.Notify(612692)
	if node.Subdir != nil {
		__antithesis_instrumentation__.Notify(612699)
		items = append(items, p.row("FROM", p.Doc(node.Subdir)))
		items = append(items, p.row("IN", p.commaSeparated(from...)))
	} else {
		__antithesis_instrumentation__.Notify(612700)
		items = append(items, p.row("FROM", p.commaSeparated(from...)))
	}
	__antithesis_instrumentation__.Notify(612693)

	if node.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(612701)
		items = append(items, node.AsOf.docRow(p))
	} else {
		__antithesis_instrumentation__.Notify(612702)
	}
	__antithesis_instrumentation__.Notify(612694)
	if !node.Options.IsDefault() {
		__antithesis_instrumentation__.Notify(612703)
		items = append(items, p.row("WITH", p.Doc(&node.Options)))
	} else {
		__antithesis_instrumentation__.Notify(612704)
	}
	__antithesis_instrumentation__.Notify(612695)
	return p.rlTable(items...)
}

func (node *TargetList) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612705)
	return p.unrow(node.docRow(p))
}

func (node *TargetList) docRow(p *PrettyCfg) pretty.TableRow {
	__antithesis_instrumentation__.Notify(612706)
	if node.Databases != nil {
		__antithesis_instrumentation__.Notify(612709)
		return p.row("DATABASE", p.Doc(&node.Databases))
	} else {
		__antithesis_instrumentation__.Notify(612710)
	}
	__antithesis_instrumentation__.Notify(612707)
	if node.TenantID.Specified {
		__antithesis_instrumentation__.Notify(612711)
		return p.row("TENANT", p.Doc(&node.TenantID))
	} else {
		__antithesis_instrumentation__.Notify(612712)
	}
	__antithesis_instrumentation__.Notify(612708)
	return p.row("TABLE", p.Doc(&node.Tables))
}

func (node *AsOfClause) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612713)
	return p.unrow(node.docRow(p))
}

func (node *AsOfClause) docRow(p *PrettyCfg) pretty.TableRow {
	__antithesis_instrumentation__.Notify(612714)
	return p.row("AS OF SYSTEM TIME", p.Doc(node.Expr))
}

func (node *KVOptions) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612715)
	var opts []pretty.Doc
	for _, opt := range *node {
		__antithesis_instrumentation__.Notify(612717)
		d := p.Doc(&opt.Key)
		if opt.Value != nil {
			__antithesis_instrumentation__.Notify(612719)
			d = pretty.Fold(pretty.ConcatSpace,
				d,
				pretty.Text("="),
				p.Doc(opt.Value),
			)
		} else {
			__antithesis_instrumentation__.Notify(612720)
		}
		__antithesis_instrumentation__.Notify(612718)
		opts = append(opts, d)
	}
	__antithesis_instrumentation__.Notify(612716)
	return p.commaSeparated(opts...)
}

func (node *Import) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612721)
	items := make([]pretty.TableRow, 0, 5)
	items = append(items, p.row("IMPORT", pretty.Nil))

	if node.Bundle {
		__antithesis_instrumentation__.Notify(612724)
		if node.Table != nil {
			__antithesis_instrumentation__.Notify(612726)
			items = append(items, p.row("TABLE", p.Doc(node.Table)))
			items = append(items, p.row("FROM", pretty.Nil))
		} else {
			__antithesis_instrumentation__.Notify(612727)
		}
		__antithesis_instrumentation__.Notify(612725)
		items = append(items, p.row(node.FileFormat, p.Doc(&node.Files)))
	} else {
		__antithesis_instrumentation__.Notify(612728)
		if node.Into {
			__antithesis_instrumentation__.Notify(612729)
			into := p.Doc(node.Table)
			if node.IntoCols != nil {
				__antithesis_instrumentation__.Notify(612731)
				into = p.nestUnder(into, p.bracket("(", p.Doc(&node.IntoCols), ")"))
			} else {
				__antithesis_instrumentation__.Notify(612732)
			}
			__antithesis_instrumentation__.Notify(612730)
			items = append(items, p.row("INTO", into))
			data := p.bracketKeyword(
				"DATA", " (",
				p.Doc(&node.Files),
				")", "",
			)
			items = append(items, p.row(node.FileFormat, data))
		} else {
			__antithesis_instrumentation__.Notify(612733)
		}
	}
	__antithesis_instrumentation__.Notify(612722)

	if node.Options != nil {
		__antithesis_instrumentation__.Notify(612734)
		items = append(items, p.row("WITH", p.Doc(&node.Options)))
	} else {
		__antithesis_instrumentation__.Notify(612735)
	}
	__antithesis_instrumentation__.Notify(612723)
	return p.rlTable(items...)
}

func (node *Export) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612736)
	items := make([]pretty.TableRow, 0, 5)
	items = append(items, p.row("EXPORT", pretty.Nil))
	items = append(items, p.row("INTO "+node.FileFormat, p.Doc(node.File)))
	if node.Options != nil {
		__antithesis_instrumentation__.Notify(612738)
		items = append(items, p.row("WITH", p.Doc(&node.Options)))
	} else {
		__antithesis_instrumentation__.Notify(612739)
	}
	__antithesis_instrumentation__.Notify(612737)
	items = append(items, p.row("FROM", p.Doc(node.Query)))
	return p.rlTable(items...)
}

func (node *NotExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612740)
	return p.nestUnder(
		pretty.Keyword("NOT"),
		p.exprDocWithParen(node.Expr),
	)
}

func (node *IsNullExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612741)
	return pretty.ConcatSpace(
		p.exprDocWithParen(node.Expr),
		pretty.Keyword("IS NULL"),
	)
}

func (node *IsNotNullExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612742)
	return pretty.ConcatSpace(
		p.exprDocWithParen(node.Expr),
		pretty.Keyword("IS NOT NULL"),
	)
}

func (node *CoalesceExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612743)
	return p.bracketKeyword(
		node.Name, "(",
		p.Doc(&node.Exprs),
		")", "",
	)
}

func (node *AlterTable) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612744)
	title := pretty.Keyword("ALTER TABLE")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(612746)
		title = pretty.ConcatSpace(title, pretty.Keyword("IF EXISTS"))
	} else {
		__antithesis_instrumentation__.Notify(612747)
	}
	__antithesis_instrumentation__.Notify(612745)
	title = pretty.ConcatSpace(title, p.Doc(node.Table))
	return p.nestUnder(
		title,
		p.Doc(&node.Cmds),
	)
}

func (node *AlterTableCmds) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612748)
	cmds := make([]pretty.Doc, len(*node))
	for i, c := range *node {
		__antithesis_instrumentation__.Notify(612750)
		cmds[i] = p.Doc(c)
	}
	__antithesis_instrumentation__.Notify(612749)
	return p.commaSeparated(cmds...)
}

func (node *AlterTableAddColumn) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612751)
	title := pretty.Keyword("ADD COLUMN")
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(612753)
		title = pretty.ConcatSpace(title, pretty.Keyword("IF NOT EXISTS"))
	} else {
		__antithesis_instrumentation__.Notify(612754)
	}
	__antithesis_instrumentation__.Notify(612752)
	return p.nestUnder(
		title,
		p.Doc(node.ColumnDef),
	)
}

func (node *Prepare) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612755)
	return p.rlTable(node.docTable(p)...)
}

func (node *Prepare) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612756)
	name := p.Doc(&node.Name)
	if len(node.Types) > 0 {
		__antithesis_instrumentation__.Notify(612758)
		typs := make([]pretty.Doc, len(node.Types))
		for i, t := range node.Types {
			__antithesis_instrumentation__.Notify(612760)
			typs[i] = pretty.Text(t.SQLString())
		}
		__antithesis_instrumentation__.Notify(612759)
		name = pretty.ConcatSpace(name,
			p.bracket("(", p.commaSeparated(typs...), ")"),
		)
	} else {
		__antithesis_instrumentation__.Notify(612761)
	}
	__antithesis_instrumentation__.Notify(612757)
	return []pretty.TableRow{
		p.row("PREPARE", name),
		p.row("AS", p.Doc(node.Statement)),
	}
}

func (node *Execute) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612762)
	return p.rlTable(node.docTable(p)...)
}

func (node *Execute) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612763)
	name := p.Doc(&node.Name)
	if len(node.Params) > 0 {
		__antithesis_instrumentation__.Notify(612766)
		name = pretty.ConcatSpace(
			name,
			p.bracket("(", p.Doc(&node.Params), ")"),
		)
	} else {
		__antithesis_instrumentation__.Notify(612767)
	}
	__antithesis_instrumentation__.Notify(612764)
	rows := []pretty.TableRow{p.row("EXECUTE", name)}
	if node.DiscardRows {
		__antithesis_instrumentation__.Notify(612768)
		rows = append(rows, p.row("", pretty.Keyword("DISCARD ROWS")))
	} else {
		__antithesis_instrumentation__.Notify(612769)
	}
	__antithesis_instrumentation__.Notify(612765)
	return rows
}

func (node *AnnotateTypeExpr) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612770)
	if node.SyntaxMode == AnnotateShort {
		__antithesis_instrumentation__.Notify(612772)
		if typ, ok := GetStaticallyKnownType(node.Type); ok {
			__antithesis_instrumentation__.Notify(612773)
			switch typ.Family() {
			case types.JsonFamily:
				__antithesis_instrumentation__.Notify(612774)
				if sv, ok := node.Expr.(*StrVal); ok && func() bool {
					__antithesis_instrumentation__.Notify(612776)
					return p.JSONFmt == true
				}() == true {
					__antithesis_instrumentation__.Notify(612777)
					return p.jsonCast(sv, ":::", typ)
				} else {
					__antithesis_instrumentation__.Notify(612778)
				}
			default:
				__antithesis_instrumentation__.Notify(612775)
			}
		} else {
			__antithesis_instrumentation__.Notify(612779)
		}
	} else {
		__antithesis_instrumentation__.Notify(612780)
	}
	__antithesis_instrumentation__.Notify(612771)
	return p.docAsString(node)
}

func (node *DeclareCursor) docTable(p *PrettyCfg) []pretty.TableRow {
	__antithesis_instrumentation__.Notify(612781)
	optionsRow := pretty.Nil
	if node.Binary {
		__antithesis_instrumentation__.Notify(612786)
		optionsRow = pretty.ConcatSpace(optionsRow, pretty.Keyword("BINARY"))
	} else {
		__antithesis_instrumentation__.Notify(612787)
	}
	__antithesis_instrumentation__.Notify(612782)
	if node.Sensitivity != UnspecifiedSensitivity {
		__antithesis_instrumentation__.Notify(612788)
		optionsRow = pretty.ConcatSpace(optionsRow, pretty.Keyword(node.Sensitivity.String()))
	} else {
		__antithesis_instrumentation__.Notify(612789)
	}
	__antithesis_instrumentation__.Notify(612783)
	if node.Scroll != UnspecifiedScroll {
		__antithesis_instrumentation__.Notify(612790)
		optionsRow = pretty.ConcatSpace(optionsRow, pretty.Keyword(node.Scroll.String()))
	} else {
		__antithesis_instrumentation__.Notify(612791)
	}
	__antithesis_instrumentation__.Notify(612784)
	cursorRow := pretty.Nil
	if node.Hold {
		__antithesis_instrumentation__.Notify(612792)
		cursorRow = pretty.ConcatSpace(cursorRow, pretty.Keyword("WITH HOLD"))
	} else {
		__antithesis_instrumentation__.Notify(612793)
	}
	__antithesis_instrumentation__.Notify(612785)
	return []pretty.TableRow{
		p.row("DECLARE", pretty.ConcatLine(p.Doc(&node.Name), optionsRow)),
		p.row("CURSOR", cursorRow),
		p.row("FOR", node.Select.doc(p)),
	}
}

func (node *DeclareCursor) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612794)
	return p.rlTable(node.docTable(p)...)
}

func (node *CursorStmt) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612795)
	ret := pretty.Nil
	fetchType := node.FetchType.String()
	if fetchType != "" {
		__antithesis_instrumentation__.Notify(612798)
		ret = pretty.ConcatSpace(ret, pretty.Keyword(fetchType))
	} else {
		__antithesis_instrumentation__.Notify(612799)
	}
	__antithesis_instrumentation__.Notify(612796)
	if node.FetchType.HasCount() {
		__antithesis_instrumentation__.Notify(612800)
		ret = pretty.ConcatSpace(ret, pretty.Text(strconv.Itoa(int(node.Count))))
	} else {
		__antithesis_instrumentation__.Notify(612801)
	}
	__antithesis_instrumentation__.Notify(612797)
	return pretty.Fold(pretty.ConcatSpace,
		ret,
		pretty.Keyword("FROM"),
		p.Doc(&node.Name),
	)
}

func (node *FetchCursor) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612802)
	return pretty.ConcatSpace(pretty.Keyword("FETCH"), node.CursorStmt.doc(p))
}

func (node *MoveCursor) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612803)
	return pretty.ConcatSpace(pretty.Keyword("MOVE"), node.CursorStmt.doc(p))
}

func (node *CloseCursor) doc(p *PrettyCfg) pretty.Doc {
	__antithesis_instrumentation__.Notify(612804)
	close := pretty.Keyword("CLOSE")
	if node.All {
		__antithesis_instrumentation__.Notify(612806)
		return pretty.ConcatSpace(close, pretty.Keyword("ALL"))
	} else {
		__antithesis_instrumentation__.Notify(612807)
	}
	__antithesis_instrumentation__.Notify(612805)
	return pretty.ConcatSpace(close, p.Doc(&node.Name))
}

func (p *PrettyCfg) jsonCast(sv *StrVal, op string, typ *types.T) pretty.Doc {
	__antithesis_instrumentation__.Notify(612808)
	return pretty.Fold(pretty.Concat,
		p.jsonString(sv.RawString()),
		pretty.Text(op),
		pretty.Text(typ.SQLString()),
	)
}

func (p *PrettyCfg) jsonString(s string) pretty.Doc {
	__antithesis_instrumentation__.Notify(612809)
	j, err := json.ParseJSON(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(612811)
		return pretty.Text(s)
	} else {
		__antithesis_instrumentation__.Notify(612812)
	}
	__antithesis_instrumentation__.Notify(612810)
	return p.bracket(`'`, p.jsonNode(j), `'`)
}

func (p *PrettyCfg) jsonNode(j json.JSON) pretty.Doc {
	__antithesis_instrumentation__.Notify(612813)

	if it, _ := j.ObjectIter(); it != nil {
		__antithesis_instrumentation__.Notify(612815)

		elems := make([]pretty.Doc, 0, j.Len())
		for it.Next() {
			__antithesis_instrumentation__.Notify(612817)
			elems = append(elems, p.nestUnder(
				pretty.Concat(
					pretty.Text(json.FromString(it.Key()).String()),
					pretty.Text(`:`),
				),
				p.jsonNode(it.Value()),
			))
		}
		__antithesis_instrumentation__.Notify(612816)
		return p.bracket("{", p.commaSeparated(elems...), "}")
	} else {
		__antithesis_instrumentation__.Notify(612818)
		if n := j.Len(); n > 0 {
			__antithesis_instrumentation__.Notify(612819)

			elems := make([]pretty.Doc, n)
			for i := 0; i < n; i++ {
				__antithesis_instrumentation__.Notify(612821)
				elem, err := j.FetchValIdx(i)
				if err != nil {
					__antithesis_instrumentation__.Notify(612823)
					return pretty.Text(j.String())
				} else {
					__antithesis_instrumentation__.Notify(612824)
				}
				__antithesis_instrumentation__.Notify(612822)
				elems[i] = p.jsonNode(elem)
			}
			__antithesis_instrumentation__.Notify(612820)
			return p.bracket("[", p.commaSeparated(elems...), "]")
		} else {
			__antithesis_instrumentation__.Notify(612825)
		}
	}
	__antithesis_instrumentation__.Notify(612814)

	return pretty.Text(j.String())
}
