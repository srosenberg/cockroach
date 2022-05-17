package fmtsafe

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/cockroachdb/errors"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

var Doc = `checks that log and error functions don't leak PII.

This linter checks the following:

- that the format string in Infof(), Errorf() and similar calls is a
  constant string.

  This check is essential for correctness because format strings
  are assumed to be PII-free and always safe for reporting in
  telemetry or PII-free logs.

- that the message strings in errors.New() and similar calls that
  construct error objects is a constant string.

  This check is provided to encourage hygiene: errors
  constructed using non-constant strings are better constructed using
  a formatting function instead, which makes the construction more
  readable and encourage the provision of PII-free reportable details.

It is possible for a call site *in a test file* to opt the format/message
string out of the linter using /* nolint:fmtsafe */ after the format
argument. This escape hatch is not available in non-test code.
`

var Analyzer = &analysis.Analyzer{
	Name:     "fmtsafe",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	__antithesis_instrumentation__.Notify(644592)
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.CallExpr)(nil),
	}

	var fmtOrMsgStr *types.Var
	var enclosingFnName string

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		__antithesis_instrumentation__.Notify(644594)

		defer func() {
			__antithesis_instrumentation__.Notify(644597)
			if r := recover(); r != nil {
				__antithesis_instrumentation__.Notify(644598)
				if err, ok := r.(error); ok {
					__antithesis_instrumentation__.Notify(644600)
					pass.Reportf(n.Pos(), "internal linter error: %v", err)
					return
				} else {
					__antithesis_instrumentation__.Notify(644601)
				}
				__antithesis_instrumentation__.Notify(644599)
				panic(r)
			} else {
				__antithesis_instrumentation__.Notify(644602)
			}
		}()
		__antithesis_instrumentation__.Notify(644595)

		if fd, ok := n.(*ast.FuncDecl); ok {
			__antithesis_instrumentation__.Notify(644603)

			enclosingFnName, fmtOrMsgStr = maybeGetConstStr(pass, fd)
			return
		} else {
			__antithesis_instrumentation__.Notify(644604)
		}
		__antithesis_instrumentation__.Notify(644596)

		call := n.(*ast.CallExpr)
		checkCallExpr(pass, enclosingFnName, call, fmtOrMsgStr)
	})
	__antithesis_instrumentation__.Notify(644593)
	return nil, nil
}

func maybeGetConstStr(
	pass *analysis.Pass, fd *ast.FuncDecl,
) (enclosingFnName string, res *types.Var) {
	__antithesis_instrumentation__.Notify(644605)
	if fd.Body == nil {
		__antithesis_instrumentation__.Notify(644612)

		return "", nil
	} else {
		__antithesis_instrumentation__.Notify(644613)
	}
	__antithesis_instrumentation__.Notify(644606)

	fn := pass.TypesInfo.Defs[fd.Name].(*types.Func)
	if fn == nil {
		__antithesis_instrumentation__.Notify(644614)
		return "", nil
	} else {
		__antithesis_instrumentation__.Notify(644615)
	}
	__antithesis_instrumentation__.Notify(644607)
	fnName := stripVendor(fn.FullName())

	var wantVariadic bool
	var argIdx int

	if requireConstFmt[fnName] {
		__antithesis_instrumentation__.Notify(644616)

		wantVariadic = true
		argIdx = -2
	} else {
		__antithesis_instrumentation__.Notify(644617)
		if requireConstMsg[fnName] {
			__antithesis_instrumentation__.Notify(644618)

			wantVariadic = false
			argIdx = -1
		} else {
			__antithesis_instrumentation__.Notify(644619)

			return fn.Name(), nil
		}
	}
	__antithesis_instrumentation__.Notify(644608)

	sig := fn.Type().(*types.Signature)
	if sig.Variadic() != wantVariadic {
		__antithesis_instrumentation__.Notify(644620)
		panic(errors.Newf("expected variadic %v, got %v", wantVariadic, sig.Variadic()))
	} else {
		__antithesis_instrumentation__.Notify(644621)
	}
	__antithesis_instrumentation__.Notify(644609)

	params := sig.Params()
	nparams := params.Len()

	if nparams+argIdx < 0 {
		__antithesis_instrumentation__.Notify(644622)
		panic(errors.New("not enough arguments"))
	} else {
		__antithesis_instrumentation__.Notify(644623)
	}
	__antithesis_instrumentation__.Notify(644610)
	if p := params.At(nparams + argIdx); p.Type() == types.Typ[types.String] {
		__antithesis_instrumentation__.Notify(644624)

		return fn.Name(), p
	} else {
		__antithesis_instrumentation__.Notify(644625)
	}
	__antithesis_instrumentation__.Notify(644611)
	return fn.Name(), nil
}

func checkCallExpr(pass *analysis.Pass, enclosingFnName string, call *ast.CallExpr, fv *types.Var) {
	__antithesis_instrumentation__.Notify(644626)

	cfn := typeutil.Callee(pass.TypesInfo, call)
	if cfn == nil {
		__antithesis_instrumentation__.Notify(644637)

		return
	} else {
		__antithesis_instrumentation__.Notify(644638)
	}
	__antithesis_instrumentation__.Notify(644627)
	fn, ok := cfn.(*types.Func)
	if !ok {
		__antithesis_instrumentation__.Notify(644639)

		return
	} else {
		__antithesis_instrumentation__.Notify(644640)
	}
	__antithesis_instrumentation__.Notify(644628)

	fnName := stripVendor(fn.FullName())

	var wantVariadic bool
	var argIdx int
	var argType string

	if requireConstFmt[fnName] {
		__antithesis_instrumentation__.Notify(644641)

		wantVariadic = true
		argIdx = -2
		argType = "format"
	} else {
		__antithesis_instrumentation__.Notify(644642)
		if requireConstMsg[fnName] {
			__antithesis_instrumentation__.Notify(644643)

			wantVariadic = false
			argIdx = -1
			argType = "message"
		} else {
			__antithesis_instrumentation__.Notify(644644)

			return
		}
	}
	__antithesis_instrumentation__.Notify(644629)

	typ := pass.TypesInfo.Types[call.Fun].Type
	if typ == nil {
		__antithesis_instrumentation__.Notify(644645)
		panic(errors.New("can't find function type"))
	} else {
		__antithesis_instrumentation__.Notify(644646)
	}
	__antithesis_instrumentation__.Notify(644630)

	sig, ok := typ.(*types.Signature)
	if !ok {
		__antithesis_instrumentation__.Notify(644647)
		panic(errors.New("can't derive signature"))
	} else {
		__antithesis_instrumentation__.Notify(644648)
	}
	__antithesis_instrumentation__.Notify(644631)
	if sig.Variadic() != wantVariadic {
		__antithesis_instrumentation__.Notify(644649)
		panic(errors.Newf("expected variadic %v, got %v", wantVariadic, sig.Variadic()))
	} else {
		__antithesis_instrumentation__.Notify(644650)
	}
	__antithesis_instrumentation__.Notify(644632)

	idx := sig.Params().Len() + argIdx
	if idx < 0 {
		__antithesis_instrumentation__.Notify(644651)
		panic(errors.New("not enough parameters"))
	} else {
		__antithesis_instrumentation__.Notify(644652)
	}
	__antithesis_instrumentation__.Notify(644633)

	lit := pass.TypesInfo.Types[call.Args[idx]].Value
	if lit != nil {
		__antithesis_instrumentation__.Notify(644653)

		return
	} else {
		__antithesis_instrumentation__.Notify(644654)
	}
	__antithesis_instrumentation__.Notify(644634)

	if fv != nil {
		__antithesis_instrumentation__.Notify(644655)
		if id, ok := call.Args[idx].(*ast.Ident); ok {
			__antithesis_instrumentation__.Notify(644656)
			if pass.TypesInfo.ObjectOf(id) == fv {
				__antithesis_instrumentation__.Notify(644657)

				return
			} else {
				__antithesis_instrumentation__.Notify(644658)
			}
		} else {
			__antithesis_instrumentation__.Notify(644659)
		}
	} else {
		__antithesis_instrumentation__.Notify(644660)
	}
	__antithesis_instrumentation__.Notify(644635)

	if hasNoLintComment(pass, call, idx) {
		__antithesis_instrumentation__.Notify(644661)
		return
	} else {
		__antithesis_instrumentation__.Notify(644662)
	}
	__antithesis_instrumentation__.Notify(644636)

	pass.Reportf(call.Lparen, escNl("%s(): %s argument is not a constant expression"+Tip),
		enclosingFnName, argType)
}

var Tip = `
Tip: use YourFuncf("descriptive prefix %%s", ...) or list new formatting wrappers in pkg/testutils/lint/passes/fmtsafe/functions.go.`

func hasNoLintComment(pass *analysis.Pass, call *ast.CallExpr, idx int) bool {
	__antithesis_instrumentation__.Notify(644663)
	fPos, f := findContainingFile(pass, call)

	if !strings.HasSuffix(fPos.Name(), "_test.go") {
		__antithesis_instrumentation__.Notify(644667)

		return false
	} else {
		__antithesis_instrumentation__.Notify(644668)
	}
	__antithesis_instrumentation__.Notify(644664)

	startPos := call.Args[idx].End()
	endPos := call.Rparen
	if idx < len(call.Args)-1 {
		__antithesis_instrumentation__.Notify(644669)
		endPos = call.Args[idx+1].Pos()
	} else {
		__antithesis_instrumentation__.Notify(644670)
	}
	__antithesis_instrumentation__.Notify(644665)
	for _, cg := range f.Comments {
		__antithesis_instrumentation__.Notify(644671)
		if cg.Pos() > endPos || func() bool {
			__antithesis_instrumentation__.Notify(644673)
			return cg.End() < startPos == true
		}() == true {
			__antithesis_instrumentation__.Notify(644674)
			continue
		} else {
			__antithesis_instrumentation__.Notify(644675)
		}
		__antithesis_instrumentation__.Notify(644672)
		for _, c := range cg.List {
			__antithesis_instrumentation__.Notify(644676)
			if strings.Contains(c.Text, "nolint:fmtsafe") {
				__antithesis_instrumentation__.Notify(644677)
				return true
			} else {
				__antithesis_instrumentation__.Notify(644678)
			}
		}
	}
	__antithesis_instrumentation__.Notify(644666)
	return false
}

func findContainingFile(pass *analysis.Pass, n ast.Node) (*token.File, *ast.File) {
	__antithesis_instrumentation__.Notify(644679)
	fPos := pass.Fset.File(n.Pos())
	for _, f := range pass.Files {
		__antithesis_instrumentation__.Notify(644681)
		if pass.Fset.File(f.Pos()) == fPos {
			__antithesis_instrumentation__.Notify(644682)
			return fPos, f
		} else {
			__antithesis_instrumentation__.Notify(644683)
		}
	}
	__antithesis_instrumentation__.Notify(644680)
	panic(fmt.Errorf("cannot file file for %v", n))
}

func stripVendor(s string) string {
	__antithesis_instrumentation__.Notify(644684)
	if i := strings.Index(s, "/vendor/"); i != -1 {
		__antithesis_instrumentation__.Notify(644686)
		s = s[i+len("/vendor/"):]
	} else {
		__antithesis_instrumentation__.Notify(644687)
	}
	__antithesis_instrumentation__.Notify(644685)
	return s
}

func escNl(msg string) string {
	__antithesis_instrumentation__.Notify(644688)
	return strings.ReplaceAll(msg, "\n", "\\n++")
}
