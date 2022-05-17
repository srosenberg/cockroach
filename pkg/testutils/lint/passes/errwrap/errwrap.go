package errwrap

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/passesutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `checks for unwrapped errors.

This linter checks that:

- err.Error() is not passed as an argument to an error-creating
  function.

- the '%s', '%v', and '%+v' format verbs are not used to format
  errors when creating a new error.

In both cases, an error-wrapping function can be used to correctly
preserve the chain of errors so that user-directed hints, links to
documentation issues, and telemetry data are all propagated.

It is possible for a call site to opt the format/message string
out of the linter using /* nolint:errwrap */ on or before the line
that creates the error.`

var errorType = types.Universe.Lookup("error").Type().String()

var Analyzer = &analysis.Analyzer{
	Name:     "errwrap",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	__antithesis_instrumentation__.Notify(644496)
	inspctr := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspctr.Preorder(nodeFilter, func(n ast.Node) {
		__antithesis_instrumentation__.Notify(644498)

		defer func() {
			__antithesis_instrumentation__.Notify(644510)
			if r := recover(); r != nil {
				__antithesis_instrumentation__.Notify(644511)
				if err, ok := r.(error); ok {
					__antithesis_instrumentation__.Notify(644513)
					pass.Reportf(n.Pos(), "internal linter error: %v", err)
					return
				} else {
					__antithesis_instrumentation__.Notify(644514)
				}
				__antithesis_instrumentation__.Notify(644512)
				panic(r)
			} else {
				__antithesis_instrumentation__.Notify(644515)
			}
		}()
		__antithesis_instrumentation__.Notify(644499)

		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(644516)
			return
		} else {
			__antithesis_instrumentation__.Notify(644517)
		}
		__antithesis_instrumentation__.Notify(644500)
		if pass.TypesInfo.TypeOf(callExpr).String() != errorType {
			__antithesis_instrumentation__.Notify(644518)
			return
		} else {
			__antithesis_instrumentation__.Notify(644519)
		}
		__antithesis_instrumentation__.Notify(644501)
		sel, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(644520)
			return
		} else {
			__antithesis_instrumentation__.Notify(644521)
		}
		__antithesis_instrumentation__.Notify(644502)
		obj, ok := pass.TypesInfo.Uses[sel.Sel]
		if !ok {
			__antithesis_instrumentation__.Notify(644522)
			return
		} else {
			__antithesis_instrumentation__.Notify(644523)
		}
		__antithesis_instrumentation__.Notify(644503)
		fn, ok := obj.(*types.Func)
		if !ok {
			__antithesis_instrumentation__.Notify(644524)
			return
		} else {
			__antithesis_instrumentation__.Notify(644525)
		}
		__antithesis_instrumentation__.Notify(644504)
		pkg := obj.Pkg()
		if pkg == nil {
			__antithesis_instrumentation__.Notify(644526)
			return
		} else {
			__antithesis_instrumentation__.Notify(644527)
		}
		__antithesis_instrumentation__.Notify(644505)

		file := pass.Fset.File(n.Pos())
		if strings.HasSuffix(file.Name(), "/embedded.go") {
			__antithesis_instrumentation__.Notify(644528)
			return
		} else {
			__antithesis_instrumentation__.Notify(644529)
		}
		__antithesis_instrumentation__.Notify(644506)
		fnName := stripVendor(fn.FullName())

		if _, found := ErrorFnFormatStringIndex[fnName]; found {
			__antithesis_instrumentation__.Notify(644530)
			for i := range callExpr.Args {
				__antithesis_instrumentation__.Notify(644531)
				if isErrorStringCall(pass, callExpr.Args[i]) {
					__antithesis_instrumentation__.Notify(644532)

					if passesutil.HasNolintComment(pass, sel, "errwrap") {
						__antithesis_instrumentation__.Notify(644534)
						continue
					} else {
						__antithesis_instrumentation__.Notify(644535)
					}
					__antithesis_instrumentation__.Notify(644533)

					pass.Report(analysis.Diagnostic{
						Pos: n.Pos(),
						Message: fmt.Sprintf(
							"err.Error() is passed to %s.%s; use pgerror.Wrap/errors.Wrap/errors.CombineErrors/"+
								"errors.WithSecondaryError/errors.NewAssertionErrorWithWrappedErrf instead",
							pkg.Name(), fn.Name()),
					})
				} else {
					__antithesis_instrumentation__.Notify(644536)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(644537)
		}
		__antithesis_instrumentation__.Notify(644507)

		formatStringIdx, ok := ErrorFnFormatStringIndex[fnName]
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(644538)
			return formatStringIdx < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(644539)

			return
		} else {
			__antithesis_instrumentation__.Notify(644540)
		}
		__antithesis_instrumentation__.Notify(644508)

		formatVerbs, ok := getFormatStringVerbs(pass, callExpr, formatStringIdx)
		if !ok {
			__antithesis_instrumentation__.Notify(644541)
			return
		} else {
			__antithesis_instrumentation__.Notify(644542)
		}
		__antithesis_instrumentation__.Notify(644509)

		args := callExpr.Args[formatStringIdx+1:]
		for i := 0; i < len(args) && func() bool {
			__antithesis_instrumentation__.Notify(644543)
			return i < len(formatVerbs) == true
		}() == true; i++ {
			__antithesis_instrumentation__.Notify(644544)
			if pass.TypesInfo.TypeOf(args[i]).String() != errorType {
				__antithesis_instrumentation__.Notify(644546)
				continue
			} else {
				__antithesis_instrumentation__.Notify(644547)
			}
			__antithesis_instrumentation__.Notify(644545)

			if formatVerbs[i] == "%v" || func() bool {
				__antithesis_instrumentation__.Notify(644548)
				return formatVerbs[i] == "%+v" == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(644549)
				return formatVerbs[i] == "%s" == true
			}() == true {
				__antithesis_instrumentation__.Notify(644550)

				if passesutil.HasNolintComment(pass, sel, "errwrap") {
					__antithesis_instrumentation__.Notify(644552)
					continue
				} else {
					__antithesis_instrumentation__.Notify(644553)
				}
				__antithesis_instrumentation__.Notify(644551)

				pass.Report(analysis.Diagnostic{
					Pos: n.Pos(),
					Message: fmt.Sprintf(
						"non-wrapped error is passed to %s.%s; use pgerror.Wrap/errors.Wrap/errors.CombineErrors/"+
							"errors.WithSecondaryError/errors.NewAssertionErrorWithWrappedErrf instead",
						pkg.Name(), fn.Name(),
					),
				})
			} else {
				__antithesis_instrumentation__.Notify(644554)
			}
		}
	})
	__antithesis_instrumentation__.Notify(644497)

	return nil, nil
}

func isErrorStringCall(pass *analysis.Pass, expr ast.Expr) bool {
	__antithesis_instrumentation__.Notify(644555)
	if call, ok := expr.(*ast.CallExpr); ok {
		__antithesis_instrumentation__.Notify(644557)
		if pass.TypesInfo.TypeOf(call).String() == "string" {
			__antithesis_instrumentation__.Notify(644558)
			if callSel, ok := call.Fun.(*ast.SelectorExpr); ok {
				__antithesis_instrumentation__.Notify(644559)
				fun := pass.TypesInfo.Uses[callSel.Sel].(*types.Func)
				return fun.Type().String() == "func() string" && func() bool {
					__antithesis_instrumentation__.Notify(644560)
					return fun.Name() == "Error" == true
				}() == true
			} else {
				__antithesis_instrumentation__.Notify(644561)
			}
		} else {
			__antithesis_instrumentation__.Notify(644562)
		}
	} else {
		__antithesis_instrumentation__.Notify(644563)
	}
	__antithesis_instrumentation__.Notify(644556)
	return false
}

var formatVerbRegexp = regexp.MustCompile(`%([^%+]|\+v)`)

func getFormatStringVerbs(
	pass *analysis.Pass, call *ast.CallExpr, formatStringIdx int,
) ([]string, bool) {
	__antithesis_instrumentation__.Notify(644564)
	if len(call.Args) <= formatStringIdx {
		__antithesis_instrumentation__.Notify(644567)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(644568)
	}
	__antithesis_instrumentation__.Notify(644565)
	strLit, ok := call.Args[formatStringIdx].(*ast.BasicLit)
	if !ok {
		__antithesis_instrumentation__.Notify(644569)

		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(644570)
	}
	__antithesis_instrumentation__.Notify(644566)
	formatString := constant.StringVal(pass.TypesInfo.Types[strLit].Value)

	return formatVerbRegexp.FindAllString(formatString, -1), true
}

func stripVendor(s string) string {
	__antithesis_instrumentation__.Notify(644571)
	if i := strings.Index(s, "/vendor/"); i != -1 {
		__antithesis_instrumentation__.Notify(644573)
		s = s[i+len("/vendor/"):]
	} else {
		__antithesis_instrumentation__.Notify(644574)
	}
	__antithesis_instrumentation__.Notify(644572)
	return s
}
