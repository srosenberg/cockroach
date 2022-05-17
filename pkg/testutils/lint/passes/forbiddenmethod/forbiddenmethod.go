// Package forbiddenmethod provides a general-purpose Analyzer
// factory to vet against calls to a forbidden method.
package forbiddenmethod

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/passesutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

type Options struct {
	PassName string

	Doc string

	Package, Type, Method string

	Hint string
}

func Analyzer(options Options) *analysis.Analyzer {
	__antithesis_instrumentation__.Notify(644695)
	methodRe := regexp.MustCompile(options.Method)
	return &analysis.Analyzer{
		Name:     options.PassName,
		Doc:      options.Doc,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			__antithesis_instrumentation__.Notify(644696)
			inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
			inspect.Preorder([]ast.Node{
				(*ast.CallExpr)(nil),
			}, func(n ast.Node) {
				__antithesis_instrumentation__.Notify(644698)
				call := n.(*ast.CallExpr)
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					__antithesis_instrumentation__.Notify(644706)
					return
				} else {
					__antithesis_instrumentation__.Notify(644707)
				}
				__antithesis_instrumentation__.Notify(644699)
				obj, ok := pass.TypesInfo.Uses[sel.Sel]
				if !ok {
					__antithesis_instrumentation__.Notify(644708)
					return
				} else {
					__antithesis_instrumentation__.Notify(644709)
				}
				__antithesis_instrumentation__.Notify(644700)
				f, ok := obj.(*types.Func)
				if !ok {
					__antithesis_instrumentation__.Notify(644710)
					return
				} else {
					__antithesis_instrumentation__.Notify(644711)
				}
				__antithesis_instrumentation__.Notify(644701)

				const debug = false
				if debug {
					__antithesis_instrumentation__.Notify(644712)
					pkgPath := ""
					if f.Pkg() != nil {
						__antithesis_instrumentation__.Notify(644714)
						pkgPath = f.Pkg().Path()
					} else {
						__antithesis_instrumentation__.Notify(644715)
					}
					__antithesis_instrumentation__.Notify(644713)
					pass.Report(analysis.Diagnostic{
						Pos:     n.Pos(),
						Message: fmt.Sprintf("method call: pkgpath=%s recvtype=%s method=%s", pkgPath, namedRecvTypeFor(f), f.Name()),
					})
				} else {
					__antithesis_instrumentation__.Notify(644716)
				}
				__antithesis_instrumentation__.Notify(644702)

				if f.Pkg() == nil || func() bool {
					__antithesis_instrumentation__.Notify(644717)
					return f.Pkg().Path() != options.Package == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(644718)
					return !methodRe.MatchString(f.Name()) == true
				}() == true {
					__antithesis_instrumentation__.Notify(644719)
					return
				} else {
					__antithesis_instrumentation__.Notify(644720)
				}
				__antithesis_instrumentation__.Notify(644703)
				if !isMethodForNamedType(f, options.Type) {
					__antithesis_instrumentation__.Notify(644721)
					return
				} else {
					__antithesis_instrumentation__.Notify(644722)
				}
				__antithesis_instrumentation__.Notify(644704)

				if passesutil.HasNolintComment(pass, sel, options.PassName) {
					__antithesis_instrumentation__.Notify(644723)
					return
				} else {
					__antithesis_instrumentation__.Notify(644724)
				}
				__antithesis_instrumentation__.Notify(644705)
				pass.Report(analysis.Diagnostic{
					Pos:     n.Pos(),
					Message: fmt.Sprintf("Illegal call to %s.%s(), %s", options.Type, f.Name(), options.Hint),
				})
			})
			__antithesis_instrumentation__.Notify(644697)
			return nil, nil
		},
	}
}

func namedRecvTypeFor(f *types.Func) string {
	__antithesis_instrumentation__.Notify(644725)
	sig := f.Type().(*types.Signature)
	recv := sig.Recv()
	if recv == nil {
		__antithesis_instrumentation__.Notify(644728)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(644729)
	}
	__antithesis_instrumentation__.Notify(644726)
	switch recv := recv.Type().(type) {
	case *types.Named:
		__antithesis_instrumentation__.Notify(644730)
		return recv.Obj().Name()
	case *types.Pointer:
		__antithesis_instrumentation__.Notify(644731)
		named, ok := recv.Elem().(*types.Named)
		if !ok {
			__antithesis_instrumentation__.Notify(644733)
			return ""
		} else {
			__antithesis_instrumentation__.Notify(644734)
		}
		__antithesis_instrumentation__.Notify(644732)
		return named.Obj().Name()
	}
	__antithesis_instrumentation__.Notify(644727)
	return ""
}

func isMethodForNamedType(f *types.Func, name string) bool {
	__antithesis_instrumentation__.Notify(644735)
	return namedRecvTypeFor(f) == name
}
