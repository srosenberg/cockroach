// Package returnerrcheck defines an suite of Analyzers that
// detects conditionals which check for a non-nil error and then
// proceed to return a nil error.
package returnerrcheck

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/passesutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for return of nil in a conditional which check for non-nil errors`

var errorType = types.Universe.Lookup("error").Type()

const name = "returnerrcheck"

var Analyzer = &analysis.Analyzer{
	Name:     name,
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (interface{}, error) {
		__antithesis_instrumentation__.Notify(645144)
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		inspect.Preorder([]ast.Node{
			(*ast.IfStmt)(nil),
		}, func(n ast.Node) {
			__antithesis_instrumentation__.Notify(645146)
			ifStmt := n.(*ast.IfStmt)

			containing := findContainingFunc(pass, n)
			if !funcReturnsErrorLast(containing) {
				__antithesis_instrumentation__.Notify(645149)
				return
			} else {
				__antithesis_instrumentation__.Notify(645150)
			}
			__antithesis_instrumentation__.Notify(645147)

			errObj, ok := isNonNilErrCheck(pass, ifStmt.Cond)
			if !ok {
				__antithesis_instrumentation__.Notify(645151)
				return
			} else {
				__antithesis_instrumentation__.Notify(645152)
			}
			__antithesis_instrumentation__.Notify(645148)
			for i, n := range ifStmt.Body.List {
				__antithesis_instrumentation__.Notify(645153)
				returnStmt, ok := n.(*ast.ReturnStmt)
				if !ok {
					__antithesis_instrumentation__.Notify(645159)
					continue
				} else {
					__antithesis_instrumentation__.Notify(645160)
				}
				__antithesis_instrumentation__.Notify(645154)
				if len(returnStmt.Results) == 0 {
					__antithesis_instrumentation__.Notify(645161)
					continue
				} else {
					__antithesis_instrumentation__.Notify(645162)
				}
				__antithesis_instrumentation__.Notify(645155)
				lastRes := returnStmt.Results[len(returnStmt.Results)-1]
				if !pass.TypesInfo.Types[lastRes].IsNil() {
					__antithesis_instrumentation__.Notify(645163)
					continue
				} else {
					__antithesis_instrumentation__.Notify(645164)
				}
				__antithesis_instrumentation__.Notify(645156)
				if hasAcceptableAction(pass, errObj, ifStmt.Body.List[:i+1]) {
					__antithesis_instrumentation__.Notify(645165)
					continue
				} else {
					__antithesis_instrumentation__.Notify(645166)
				}
				__antithesis_instrumentation__.Notify(645157)
				if passesutil.HasNolintComment(pass, returnStmt, name) {
					__antithesis_instrumentation__.Notify(645167)
					continue
				} else {
					__antithesis_instrumentation__.Notify(645168)
				}
				__antithesis_instrumentation__.Notify(645158)
				pass.Report(analysis.Diagnostic{
					Pos: n.Pos(),
					Message: fmt.Sprintf("unexpected nil error return after checking for a non-nil error" +
						"; if this is not a mistake, add a \"//nolint:returnerrcheck\" comment"),
				})
			}
		})
		__antithesis_instrumentation__.Notify(645145)
		return nil, nil
	},
}

func hasAcceptableAction(pass *analysis.Pass, errObj types.Object, statements []ast.Stmt) bool {
	__antithesis_instrumentation__.Notify(645169)
	var seen bool
	isObj := func(n ast.Node) bool {
		__antithesis_instrumentation__.Notify(645173)
		if id, ok := n.(*ast.Ident); ok {
			__antithesis_instrumentation__.Notify(645175)
			if seen = pass.TypesInfo.Uses[id] == errObj; seen {
				__antithesis_instrumentation__.Notify(645176)
				return true
			} else {
				__antithesis_instrumentation__.Notify(645177)
			}
		} else {
			__antithesis_instrumentation__.Notify(645178)
		}
		__antithesis_instrumentation__.Notify(645174)
		return false
	}
	__antithesis_instrumentation__.Notify(645170)
	inspect := func(n ast.Node) (wantMore bool) {
		__antithesis_instrumentation__.Notify(645179)
		switch n := n.(type) {
		case *ast.CallExpr:
			__antithesis_instrumentation__.Notify(645181)

			for _, arg := range n.Args {
				__antithesis_instrumentation__.Notify(645186)
				if isObj(arg) {
					__antithesis_instrumentation__.Notify(645187)
					return false
				} else {
					__antithesis_instrumentation__.Notify(645188)
				}
			}
		case *ast.AssignStmt:
			__antithesis_instrumentation__.Notify(645182)

			for _, rhs := range n.Rhs {
				__antithesis_instrumentation__.Notify(645189)
				if isObj(rhs) {
					__antithesis_instrumentation__.Notify(645190)
					return false
				} else {
					__antithesis_instrumentation__.Notify(645191)
				}
			}
		case *ast.KeyValueExpr:
			__antithesis_instrumentation__.Notify(645183)

			if isObj(n.Value) {
				__antithesis_instrumentation__.Notify(645192)
				return false
			} else {
				__antithesis_instrumentation__.Notify(645193)
			}

		case *ast.SelectorExpr:
			__antithesis_instrumentation__.Notify(645184)

			if isObj(n.X) {
				__antithesis_instrumentation__.Notify(645194)
				return false
			} else {
				__antithesis_instrumentation__.Notify(645195)
			}
		case *ast.ReturnStmt:
			__antithesis_instrumentation__.Notify(645185)

			if numRes := len(n.Results); numRes > 0 && func() bool {
				__antithesis_instrumentation__.Notify(645196)
				return isObj(n.Results[numRes-1]) == true
			}() == true {
				__antithesis_instrumentation__.Notify(645197)
				return false
			} else {
				__antithesis_instrumentation__.Notify(645198)
			}
		}
		__antithesis_instrumentation__.Notify(645180)
		return true
	}
	__antithesis_instrumentation__.Notify(645171)
	for _, stmt := range statements {
		__antithesis_instrumentation__.Notify(645199)
		if ast.Inspect(stmt, inspect); seen {
			__antithesis_instrumentation__.Notify(645200)
			return true
		} else {
			__antithesis_instrumentation__.Notify(645201)
		}
	}
	__antithesis_instrumentation__.Notify(645172)
	return false
}

func isNonNilErrCheck(pass *analysis.Pass, expr ast.Expr) (errObj types.Object, ok bool) {
	__antithesis_instrumentation__.Notify(645202)

	binaryExpr, ok := expr.(*ast.BinaryExpr)
	if !ok {
		__antithesis_instrumentation__.Notify(645205)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(645206)
	}
	__antithesis_instrumentation__.Notify(645203)

	switch binaryExpr.Op {
	case token.NEQ:
		__antithesis_instrumentation__.Notify(645207)
		var id *ast.Ident
		if pass.TypesInfo.Types[binaryExpr.X].Type == errorType && func() bool {
			__antithesis_instrumentation__.Notify(645212)
			return pass.TypesInfo.Types[binaryExpr.Y].IsNil() == true
		}() == true {
			__antithesis_instrumentation__.Notify(645213)
			id, ok = binaryExpr.X.(*ast.Ident)
		} else {
			__antithesis_instrumentation__.Notify(645214)
			if pass.TypesInfo.Types[binaryExpr.Y].Type == errorType && func() bool {
				__antithesis_instrumentation__.Notify(645215)
				return pass.TypesInfo.Types[binaryExpr.X].IsNil() == true
			}() == true {
				__antithesis_instrumentation__.Notify(645216)
				id, ok = binaryExpr.Y.(*ast.Ident)
			} else {
				__antithesis_instrumentation__.Notify(645217)
			}
		}
		__antithesis_instrumentation__.Notify(645208)
		if ok {
			__antithesis_instrumentation__.Notify(645218)
			errObj, ok := pass.TypesInfo.Uses[id]
			return errObj, ok
		} else {
			__antithesis_instrumentation__.Notify(645219)
		}
	case token.LAND, token.LOR:
		__antithesis_instrumentation__.Notify(645209)
		if errObj, ok := isNonNilErrCheck(pass, binaryExpr.X); ok {
			__antithesis_instrumentation__.Notify(645220)
			return errObj, ok
		} else {
			__antithesis_instrumentation__.Notify(645221)
		}
		__antithesis_instrumentation__.Notify(645210)
		return isNonNilErrCheck(pass, binaryExpr.Y)
	default:
		__antithesis_instrumentation__.Notify(645211)
	}
	__antithesis_instrumentation__.Notify(645204)
	return nil, false
}

func funcReturnsErrorLast(f *types.Signature) bool {
	__antithesis_instrumentation__.Notify(645222)
	results := f.Results()
	return results.Len() > 0 && func() bool {
		__antithesis_instrumentation__.Notify(645223)
		return results.At(results.Len()-1).Type() == errorType == true
	}() == true
}

func findContainingFunc(pass *analysis.Pass, n ast.Node) *types.Signature {
	__antithesis_instrumentation__.Notify(645224)
	stack, _ := astutil.PathEnclosingInterval(passesutil.FindContainingFile(pass, n), n.Pos(), n.End())
	for _, n := range stack {
		__antithesis_instrumentation__.Notify(645226)
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			__antithesis_instrumentation__.Notify(645228)
			return pass.TypesInfo.ObjectOf(funcDecl.Name).(*types.Func).Type().(*types.Signature)
		} else {
			__antithesis_instrumentation__.Notify(645229)
		}
		__antithesis_instrumentation__.Notify(645227)
		if funcLit, ok := n.(*ast.FuncLit); ok {
			__antithesis_instrumentation__.Notify(645230)
			return pass.TypesInfo.Types[funcLit].Type.(*types.Signature)
		} else {
			__antithesis_instrumentation__.Notify(645231)
		}
	}
	__antithesis_instrumentation__.Notify(645225)
	return nil
}
