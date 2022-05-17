// Package timer defines an Analyzer that detects correct use of
// timeutil.Timer.
package timer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for correct use of timeutil.Timer`

var Analyzer = &analysis.Analyzer{
	Name:     "timer",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	__antithesis_instrumentation__.Notify(645353)
	selectorIsTimer := func(s *ast.SelectorExpr) bool {
		__antithesis_instrumentation__.Notify(645356)
		tv, ok := pass.TypesInfo.Types[s.X]
		if !ok {
			__antithesis_instrumentation__.Notify(645361)
			return false
		} else {
			__antithesis_instrumentation__.Notify(645362)
		}
		__antithesis_instrumentation__.Notify(645357)
		typ := tv.Type.Underlying()
		for {
			__antithesis_instrumentation__.Notify(645363)
			ptr, pok := typ.(*types.Pointer)
			if !pok {
				__antithesis_instrumentation__.Notify(645365)
				break
			} else {
				__antithesis_instrumentation__.Notify(645366)
			}
			__antithesis_instrumentation__.Notify(645364)
			typ = ptr.Elem()
		}
		__antithesis_instrumentation__.Notify(645358)
		named, ok := typ.(*types.Named)
		if !ok {
			__antithesis_instrumentation__.Notify(645367)
			return false
		} else {
			__antithesis_instrumentation__.Notify(645368)
		}
		__antithesis_instrumentation__.Notify(645359)
		if named.Obj().Type().String() != "github.com/cockroachdb/cockroach/pkg/util/timeutil.Timer" {
			__antithesis_instrumentation__.Notify(645369)
			return false
		} else {
			__antithesis_instrumentation__.Notify(645370)
		}
		__antithesis_instrumentation__.Notify(645360)
		return true
	}
	__antithesis_instrumentation__.Notify(645354)

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.ForStmt)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		__antithesis_instrumentation__.Notify(645371)
		fr, ok := n.(*ast.ForStmt)
		if !ok {
			__antithesis_instrumentation__.Notify(645373)
			return
		} else {
			__antithesis_instrumentation__.Notify(645374)
		}
		__antithesis_instrumentation__.Notify(645372)
		walkStmts(fr.Body.List, func(s ast.Stmt) bool {
			__antithesis_instrumentation__.Notify(645375)
			return walkSelectStmts(s, func(s ast.Stmt) bool {
				__antithesis_instrumentation__.Notify(645376)
				comm, ok := s.(*ast.CommClause)
				if !ok || func() bool {
					__antithesis_instrumentation__.Notify(645385)
					return comm.Comm == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(645386)
					return true
				} else {
					__antithesis_instrumentation__.Notify(645387)
				}
				__antithesis_instrumentation__.Notify(645377)

				var unary ast.Expr
				switch v := comm.Comm.(type) {
				case *ast.AssignStmt:
					__antithesis_instrumentation__.Notify(645388)

					unary = v.Rhs[0]
				case *ast.ExprStmt:
					__antithesis_instrumentation__.Notify(645389)

					unary = v.X
				default:
					__antithesis_instrumentation__.Notify(645390)
					return true
				}
				__antithesis_instrumentation__.Notify(645378)
				chanRead, ok := unary.(*ast.UnaryExpr)
				if !ok || func() bool {
					__antithesis_instrumentation__.Notify(645391)
					return chanRead.Op != token.ARROW == true
				}() == true {
					__antithesis_instrumentation__.Notify(645392)
					return true
				} else {
					__antithesis_instrumentation__.Notify(645393)
				}
				__antithesis_instrumentation__.Notify(645379)
				selector, ok := chanRead.X.(*ast.SelectorExpr)
				if !ok {
					__antithesis_instrumentation__.Notify(645394)
					return true
				} else {
					__antithesis_instrumentation__.Notify(645395)
				}
				__antithesis_instrumentation__.Notify(645380)
				if !selectorIsTimer(selector) {
					__antithesis_instrumentation__.Notify(645396)
					return true
				} else {
					__antithesis_instrumentation__.Notify(645397)
				}
				__antithesis_instrumentation__.Notify(645381)
				selectorName := fmt.Sprint(selector.X)
				if selector.Sel.String() != timerChanName {
					__antithesis_instrumentation__.Notify(645398)
					return true
				} else {
					__antithesis_instrumentation__.Notify(645399)
				}
				__antithesis_instrumentation__.Notify(645382)

				noRead := walkStmts(comm.Body, func(s ast.Stmt) bool {
					__antithesis_instrumentation__.Notify(645400)
					assign, ok := s.(*ast.AssignStmt)
					if !ok || func() bool {
						__antithesis_instrumentation__.Notify(645403)
						return assign.Tok != token.ASSIGN == true
					}() == true {
						__antithesis_instrumentation__.Notify(645404)
						return true
					} else {
						__antithesis_instrumentation__.Notify(645405)
					}
					__antithesis_instrumentation__.Notify(645401)
					for i := range assign.Lhs {
						__antithesis_instrumentation__.Notify(645406)
						l, r := assign.Lhs[i], assign.Rhs[i]

						assignSelector, ok := l.(*ast.SelectorExpr)
						if !ok {
							__antithesis_instrumentation__.Notify(645412)
							return true
						} else {
							__antithesis_instrumentation__.Notify(645413)
						}
						__antithesis_instrumentation__.Notify(645407)
						if !selectorIsTimer(assignSelector) {
							__antithesis_instrumentation__.Notify(645414)
							return true
						} else {
							__antithesis_instrumentation__.Notify(645415)
						}
						__antithesis_instrumentation__.Notify(645408)
						if fmt.Sprint(assignSelector.X) != selectorName {
							__antithesis_instrumentation__.Notify(645416)
							return true
						} else {
							__antithesis_instrumentation__.Notify(645417)
						}
						__antithesis_instrumentation__.Notify(645409)
						if assignSelector.Sel.String() != "Read" {
							__antithesis_instrumentation__.Notify(645418)
							return true
						} else {
							__antithesis_instrumentation__.Notify(645419)
						}
						__antithesis_instrumentation__.Notify(645410)

						val, ok := r.(*ast.Ident)
						if !ok {
							__antithesis_instrumentation__.Notify(645420)
							return true
						} else {
							__antithesis_instrumentation__.Notify(645421)
						}
						__antithesis_instrumentation__.Notify(645411)
						if val.String() == "true" {
							__antithesis_instrumentation__.Notify(645422)

							return false
						} else {
							__antithesis_instrumentation__.Notify(645423)
						}
					}
					__antithesis_instrumentation__.Notify(645402)
					return true
				})
				__antithesis_instrumentation__.Notify(645383)
				if noRead {
					__antithesis_instrumentation__.Notify(645424)
					pass.Reportf(comm.Pos(), "must set timer.Read = true after reading from timer.C (see timeutil/timer.go)")
				} else {
					__antithesis_instrumentation__.Notify(645425)
				}
				__antithesis_instrumentation__.Notify(645384)
				return true
			})
		})
	})
	__antithesis_instrumentation__.Notify(645355)

	return nil, nil
}

const timerChanName = "C"

func walkSelectStmts(n ast.Node, fn func(ast.Stmt) bool) bool {
	__antithesis_instrumentation__.Notify(645426)
	sel, ok := n.(*ast.SelectStmt)
	if !ok {
		__antithesis_instrumentation__.Notify(645428)
		return true
	} else {
		__antithesis_instrumentation__.Notify(645429)
	}
	__antithesis_instrumentation__.Notify(645427)
	return walkStmts(sel.Body.List, fn)
}

func walkStmts(stmts []ast.Stmt, fn func(ast.Stmt) bool) bool {
	__antithesis_instrumentation__.Notify(645430)
	for _, stmt := range stmts {
		__antithesis_instrumentation__.Notify(645432)
		if !fn(stmt) {
			__antithesis_instrumentation__.Notify(645433)
			return false
		} else {
			__antithesis_instrumentation__.Notify(645434)
		}
	}
	__antithesis_instrumentation__.Notify(645431)
	return true
}
