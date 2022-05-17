// Package unconvert defines an Analyzer that detects unnecessary type
// conversions.
package unconvert

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/passesutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for unnecessary type conversions`

const name = "unconvert"

var Analyzer = &analysis.Analyzer{
	Name:     name,
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	__antithesis_instrumentation__.Notify(645441)
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		__antithesis_instrumentation__.Notify(645443)
		call, ok := n.(*ast.CallExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(645453)
			return
		} else {
			__antithesis_instrumentation__.Notify(645454)
		}
		__antithesis_instrumentation__.Notify(645444)
		if len(call.Args) != 1 || func() bool {
			__antithesis_instrumentation__.Notify(645455)
			return call.Ellipsis != token.NoPos == true
		}() == true {
			__antithesis_instrumentation__.Notify(645456)
			return
		} else {
			__antithesis_instrumentation__.Notify(645457)
		}
		__antithesis_instrumentation__.Notify(645445)
		ft, ok := pass.TypesInfo.Types[call.Fun]
		if !ok {
			__antithesis_instrumentation__.Notify(645458)
			pass.Reportf(call.Pos(), "missing type")
			return
		} else {
			__antithesis_instrumentation__.Notify(645459)
		}
		__antithesis_instrumentation__.Notify(645446)
		if !ft.IsType() {
			__antithesis_instrumentation__.Notify(645460)

			return
		} else {
			__antithesis_instrumentation__.Notify(645461)
		}
		__antithesis_instrumentation__.Notify(645447)
		at, ok := pass.TypesInfo.Types[call.Args[0]]
		if !ok {
			__antithesis_instrumentation__.Notify(645462)
			pass.Reportf(call.Pos(), "missing type")
			return
		} else {
			__antithesis_instrumentation__.Notify(645463)
		}
		__antithesis_instrumentation__.Notify(645448)
		if !types.Identical(ft.Type, at.Type) {
			__antithesis_instrumentation__.Notify(645464)

			return
		} else {
			__antithesis_instrumentation__.Notify(645465)
		}
		__antithesis_instrumentation__.Notify(645449)
		if isUntypedValue(call.Args[0], pass.TypesInfo) {
			__antithesis_instrumentation__.Notify(645466)

			return
		} else {
			__antithesis_instrumentation__.Notify(645467)
		}
		__antithesis_instrumentation__.Notify(645450)

		if ident, ok := call.Fun.(*ast.Ident); ok {
			__antithesis_instrumentation__.Notify(645468)
			if ident.Name == "_cgoCheckPointer" {
				__antithesis_instrumentation__.Notify(645469)
				return
			} else {
				__antithesis_instrumentation__.Notify(645470)
			}
		} else {
			__antithesis_instrumentation__.Notify(645471)
		}
		__antithesis_instrumentation__.Notify(645451)
		if passesutil.HasNolintComment(pass, call, name) {
			__antithesis_instrumentation__.Notify(645472)
			return
		} else {
			__antithesis_instrumentation__.Notify(645473)
		}
		__antithesis_instrumentation__.Notify(645452)
		pass.Reportf(call.Pos(), "unnecessary conversion")
	})
	__antithesis_instrumentation__.Notify(645442)

	return nil, nil
}

func isUntypedValue(n ast.Expr, info *types.Info) bool {
	__antithesis_instrumentation__.Notify(645474)
	switch n := n.(type) {
	case *ast.BinaryExpr:
		__antithesis_instrumentation__.Notify(645476)
		switch n.Op {
		case token.SHL, token.SHR:
			__antithesis_instrumentation__.Notify(645483)

			return isUntypedValue(n.X, info)
		case token.EQL, token.NEQ, token.LSS, token.GTR, token.LEQ, token.GEQ:
			__antithesis_instrumentation__.Notify(645484)

			return true
		case token.ADD, token.SUB, token.MUL, token.QUO, token.REM,
			token.AND, token.OR, token.XOR, token.AND_NOT,
			token.LAND, token.LOR:
			__antithesis_instrumentation__.Notify(645485)
			return isUntypedValue(n.X, info) && func() bool {
				__antithesis_instrumentation__.Notify(645487)
				return isUntypedValue(n.Y, info) == true
			}() == true
		default:
			__antithesis_instrumentation__.Notify(645486)
		}
	case *ast.UnaryExpr:
		__antithesis_instrumentation__.Notify(645477)
		switch n.Op {
		case token.ADD, token.SUB, token.NOT, token.XOR:
			__antithesis_instrumentation__.Notify(645488)
			return isUntypedValue(n.X, info)
		default:
			__antithesis_instrumentation__.Notify(645489)
		}
	case *ast.BasicLit:
		__antithesis_instrumentation__.Notify(645478)

		return true
	case *ast.ParenExpr:
		__antithesis_instrumentation__.Notify(645479)
		return isUntypedValue(n.X, info)
	case *ast.SelectorExpr:
		__antithesis_instrumentation__.Notify(645480)
		return isUntypedValue(n.Sel, info)
	case *ast.Ident:
		__antithesis_instrumentation__.Notify(645481)
		if obj, ok := info.Uses[n]; ok {
			__antithesis_instrumentation__.Notify(645490)
			if obj.Pkg() == nil && func() bool {
				__antithesis_instrumentation__.Notify(645492)
				return obj.Name() == "nil" == true
			}() == true {
				__antithesis_instrumentation__.Notify(645493)

				return true
			} else {
				__antithesis_instrumentation__.Notify(645494)
			}
			__antithesis_instrumentation__.Notify(645491)
			if b, ok := obj.Type().(*types.Basic); ok && func() bool {
				__antithesis_instrumentation__.Notify(645495)
				return b.Info()&types.IsUntyped != 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(645496)

				return true
			} else {
				__antithesis_instrumentation__.Notify(645497)
			}
		} else {
			__antithesis_instrumentation__.Notify(645498)
		}
	case *ast.CallExpr:
		__antithesis_instrumentation__.Notify(645482)
		if b, ok := asBuiltin(n.Fun, info); ok {
			__antithesis_instrumentation__.Notify(645499)
			switch b.Name() {
			case "real", "imag":
				__antithesis_instrumentation__.Notify(645500)
				return isUntypedValue(n.Args[0], info)
			case "complex":
				__antithesis_instrumentation__.Notify(645501)
				return isUntypedValue(n.Args[0], info) && func() bool {
					__antithesis_instrumentation__.Notify(645503)
					return isUntypedValue(n.Args[1], info) == true
				}() == true
			default:
				__antithesis_instrumentation__.Notify(645502)
			}
		} else {
			__antithesis_instrumentation__.Notify(645504)
		}
	}
	__antithesis_instrumentation__.Notify(645475)

	return false
}

func asBuiltin(n ast.Expr, info *types.Info) (*types.Builtin, bool) {
	__antithesis_instrumentation__.Notify(645505)
	for {
		__antithesis_instrumentation__.Notify(645509)
		paren, ok := n.(*ast.ParenExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(645511)
			break
		} else {
			__antithesis_instrumentation__.Notify(645512)
		}
		__antithesis_instrumentation__.Notify(645510)
		n = paren.X
	}
	__antithesis_instrumentation__.Notify(645506)

	ident, ok := n.(*ast.Ident)
	if !ok {
		__antithesis_instrumentation__.Notify(645513)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(645514)
	}
	__antithesis_instrumentation__.Notify(645507)

	obj, ok := info.Uses[ident]
	if !ok {
		__antithesis_instrumentation__.Notify(645515)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(645516)
	}
	__antithesis_instrumentation__.Notify(645508)

	b, ok := obj.(*types.Builtin)
	return b, ok
}
