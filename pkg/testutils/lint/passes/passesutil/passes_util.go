// Package passesutil provides useful functionality for implementing passes.
package passesutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

func FindContainingFile(pass *analysis.Pass, n ast.Node) *ast.File {
	__antithesis_instrumentation__.Notify(645080)
	fPos := pass.Fset.File(n.Pos())
	for _, f := range pass.Files {
		__antithesis_instrumentation__.Notify(645082)
		if pass.Fset.File(f.Pos()) == fPos {
			__antithesis_instrumentation__.Notify(645083)
			return f
		} else {
			__antithesis_instrumentation__.Notify(645084)
		}
	}
	__antithesis_instrumentation__.Notify(645081)
	panic(fmt.Errorf("cannot file file for %v", n))
}

func HasNolintComment(pass *analysis.Pass, n ast.Node, nolintName string) bool {
	__antithesis_instrumentation__.Notify(645085)
	f := FindContainingFile(pass, n)
	relevant, containing := findNodesInBlock(f, n)
	cm := ast.NewCommentMap(pass.Fset, containing, f.Comments)

	nolintComment := "nolint:" + nolintName
	for _, cn := range relevant {
		__antithesis_instrumentation__.Notify(645087)

		_, isIdent := cn.(*ast.Ident)
		for _, cg := range cm[cn] {
			__antithesis_instrumentation__.Notify(645088)
			for _, c := range cg.List {
				__antithesis_instrumentation__.Notify(645089)
				if !strings.Contains(c.Text, nolintComment) {
					__antithesis_instrumentation__.Notify(645092)
					continue
				} else {
					__antithesis_instrumentation__.Notify(645093)
				}
				__antithesis_instrumentation__.Notify(645090)
				outermost := relevant[len(relevant)-1]
				if isIdent && func() bool {
					__antithesis_instrumentation__.Notify(645094)
					return (cg.Pos() < outermost.Pos() || func() bool {
						__antithesis_instrumentation__.Notify(645095)
						return cg.End() > outermost.End() == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(645096)
					continue
				} else {
					__antithesis_instrumentation__.Notify(645097)
				}
				__antithesis_instrumentation__.Notify(645091)
				return true
			}
		}
	}
	__antithesis_instrumentation__.Notify(645086)
	return false
}

func findNodesInBlock(f *ast.File, n ast.Node) (relevant []ast.Node, containing ast.Node) {
	__antithesis_instrumentation__.Notify(645098)
	stack, _ := astutil.PathEnclosingInterval(f, n.Pos(), n.End())

	ast.Walk(funcVisitor(func(node ast.Node) {
		__antithesis_instrumentation__.Notify(645101)
		relevant = append(relevant, node)
	}), n)
	__antithesis_instrumentation__.Notify(645099)

	reverseNodes(relevant)

	containing = f
	for _, n := range stack {
		__antithesis_instrumentation__.Notify(645102)
		switch n.(type) {
		case *ast.GenDecl, *ast.BlockStmt:
			__antithesis_instrumentation__.Notify(645103)
			containing = n
			return relevant, containing
		default:
			__antithesis_instrumentation__.Notify(645104)

			relevant = append(relevant, n)
		}
	}
	__antithesis_instrumentation__.Notify(645100)
	return relevant, containing
}

func reverseNodes(n []ast.Node) {
	__antithesis_instrumentation__.Notify(645105)
	for i := 0; i < len(n)/2; i++ {
		__antithesis_instrumentation__.Notify(645106)
		n[i], n[len(n)-i-1] = n[len(n)-i-1], n[i]
	}
}

type funcVisitor func(node ast.Node)

var _ ast.Visitor = (funcVisitor)(nil)

func (f funcVisitor) Visit(node ast.Node) (w ast.Visitor) {
	__antithesis_instrumentation__.Notify(645107)
	if node != nil {
		__antithesis_instrumentation__.Notify(645109)
		f(node)
	} else {
		__antithesis_instrumentation__.Notify(645110)
	}
	__antithesis_instrumentation__.Notify(645108)
	return f
}
