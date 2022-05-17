// Package errcmp defines an Analyzer which checks
// for usage of errors.Is instead of direct ==/!= comparisons.
package errcmp

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for comparison of error objects`

var errorType = types.Universe.Lookup("error").Type()

var Analyzer = &analysis.Analyzer{
	Name:     "errcmp",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	__antithesis_instrumentation__.Notify(644451)
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BinaryExpr)(nil),
		(*ast.TypeAssertExpr)(nil),
		(*ast.SwitchStmt)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		__antithesis_instrumentation__.Notify(644453)

		defer func() {
			__antithesis_instrumentation__.Notify(644457)
			if r := recover(); r != nil {
				__antithesis_instrumentation__.Notify(644458)
				if err, ok := r.(error); ok {
					__antithesis_instrumentation__.Notify(644460)
					pass.Reportf(n.Pos(), "internal linter error: %v", err)
					return
				} else {
					__antithesis_instrumentation__.Notify(644461)
				}
				__antithesis_instrumentation__.Notify(644459)
				panic(r)
			} else {
				__antithesis_instrumentation__.Notify(644462)
			}
		}()
		__antithesis_instrumentation__.Notify(644454)

		if cmp, ok := n.(*ast.BinaryExpr); ok {
			__antithesis_instrumentation__.Notify(644463)
			checkErrCmp(pass, cmp)
			return
		} else {
			__antithesis_instrumentation__.Notify(644464)
		}
		__antithesis_instrumentation__.Notify(644455)
		if cmp, ok := n.(*ast.TypeAssertExpr); ok {
			__antithesis_instrumentation__.Notify(644465)
			checkErrCast(pass, cmp)
			return
		} else {
			__antithesis_instrumentation__.Notify(644466)
		}
		__antithesis_instrumentation__.Notify(644456)
		if cmp, ok := n.(*ast.SwitchStmt); ok {
			__antithesis_instrumentation__.Notify(644467)
			checkErrSwitch(pass, cmp)
			return
		} else {
			__antithesis_instrumentation__.Notify(644468)
		}
	})
	__antithesis_instrumentation__.Notify(644452)

	return nil, nil
}

func checkErrSwitch(pass *analysis.Pass, s *ast.SwitchStmt) {
	__antithesis_instrumentation__.Notify(644469)
	if pass.TypesInfo.Types[s.Tag].Type == errorType {
		__antithesis_instrumentation__.Notify(644470)
		pass.Reportf(s.Switch, escNl(`invalid direct comparison of error object
Tip:
   switch err { case errRef:...
-> switch { case errors.Is(err, errRef): ...
`))
	} else {
		__antithesis_instrumentation__.Notify(644471)
	}
}

func checkErrCast(pass *analysis.Pass, texpr *ast.TypeAssertExpr) {
	__antithesis_instrumentation__.Notify(644472)
	if pass.TypesInfo.Types[texpr.X].Type == errorType {
		__antithesis_instrumentation__.Notify(644473)
		pass.Reportf(texpr.Lparen, escNl(`invalid direct cast on error object
Alternatives:
   if _, ok := err.(*T); ok        ->   if errors.HasType(err, (*T)(nil))
   if _, ok := err.(I); ok         ->   if errors.HasInterface(err, (*I)(nil))
   if myErr, ok := err.(*T); ok    ->   if myErr := (*T)(nil); errors.As(err, &myErr)
   if myErr, ok := err.(I); ok     ->   if myErr := (I)(nil); errors.As(err, &myErr)
   switch err.(type) { case *T:... ->   switch { case errors.HasType(err, (*T)(nil)): ...
`))
	} else {
		__antithesis_instrumentation__.Notify(644474)
	}
}

func isEOFError(e ast.Expr) bool {
	__antithesis_instrumentation__.Notify(644475)
	if s, ok := e.(*ast.SelectorExpr); ok {
		__antithesis_instrumentation__.Notify(644477)
		if io, ok := s.X.(*ast.Ident); ok && func() bool {
			__antithesis_instrumentation__.Notify(644478)
			return io.Name == "io" == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(644479)
			return io.Obj == (*ast.Object)(nil) == true
		}() == true {
			__antithesis_instrumentation__.Notify(644480)
			if s.Sel.Name == "EOF" || func() bool {
				__antithesis_instrumentation__.Notify(644481)
				return s.Sel.Name == "ErrUnexpectedEOF" == true
			}() == true {
				__antithesis_instrumentation__.Notify(644482)
				return true
			} else {
				__antithesis_instrumentation__.Notify(644483)
			}
		} else {
			__antithesis_instrumentation__.Notify(644484)
		}
	} else {
		__antithesis_instrumentation__.Notify(644485)
	}
	__antithesis_instrumentation__.Notify(644476)
	return false
}

func checkErrCmp(pass *analysis.Pass, binaryExpr *ast.BinaryExpr) {
	__antithesis_instrumentation__.Notify(644486)
	switch binaryExpr.Op {
	case token.NEQ, token.EQL:
		__antithesis_instrumentation__.Notify(644487)
		if pass.TypesInfo.Types[binaryExpr.X].Type == errorType && func() bool {
			__antithesis_instrumentation__.Notify(644489)
			return !pass.TypesInfo.Types[binaryExpr.Y].IsNil() == true
		}() == true {
			__antithesis_instrumentation__.Notify(644490)

			if isEOFError(binaryExpr.Y) {
				__antithesis_instrumentation__.Notify(644492)
				return
			} else {
				__antithesis_instrumentation__.Notify(644493)
			}
			__antithesis_instrumentation__.Notify(644491)

			pass.Reportf(binaryExpr.OpPos, escNl(`use errors.Is instead of a direct comparison
For example:
   if errors.Is(err, errMyOwnErrReference) {
     ...
   }
`))
		} else {
			__antithesis_instrumentation__.Notify(644494)
		}
	default:
		__antithesis_instrumentation__.Notify(644488)
	}
}

func escNl(msg string) string {
	__antithesis_instrumentation__.Notify(644495)
	return strings.ReplaceAll(msg, "\n", "\\n++")
}
