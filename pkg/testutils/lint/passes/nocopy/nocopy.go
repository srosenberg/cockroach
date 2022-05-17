// Package nocopy defines an Analyzer that detects invalid uses of util.NoCopy.
package nocopy

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for invalid uses of util.NoCopy`

var Analyzer = &analysis.Analyzer{
	Name:     "nocopy",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

const noCopyType = "github.com/cockroachdb/cockroach/pkg/util.NoCopy"

func run(pass *analysis.Pass) (interface{}, error) {
	__antithesis_instrumentation__.Notify(645062)
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{
		(*ast.StructType)(nil),
	}, func(n ast.Node) {
		__antithesis_instrumentation__.Notify(645064)
		str := n.(*ast.StructType)
		if str.Fields == nil {
			__antithesis_instrumentation__.Notify(645066)
			return
		} else {
			__antithesis_instrumentation__.Notify(645067)
		}
		__antithesis_instrumentation__.Notify(645065)
		for i, f := range str.Fields.List {
			__antithesis_instrumentation__.Notify(645068)
			tv, ok := pass.TypesInfo.Types[f.Type]
			if !ok {
				__antithesis_instrumentation__.Notify(645071)
				continue
			} else {
				__antithesis_instrumentation__.Notify(645072)
			}
			__antithesis_instrumentation__.Notify(645069)
			if tv.Type.String() != noCopyType {
				__antithesis_instrumentation__.Notify(645073)
				continue
			} else {
				__antithesis_instrumentation__.Notify(645074)
			}
			__antithesis_instrumentation__.Notify(645070)
			switch {
			case i != 0:
				__antithesis_instrumentation__.Notify(645075)
				pass.Reportf(f.Pos(), "Illegal use of util.NoCopy - must be first field in struct")
			case len(f.Names) == 0:
				__antithesis_instrumentation__.Notify(645076)
				pass.Reportf(f.Pos(), "Illegal use of util.NoCopy - should not be embedded")
			case len(f.Names) > 1:
				__antithesis_instrumentation__.Notify(645077)
				pass.Reportf(f.Pos(), "Illegal use of util.NoCopy - should be included only once")
			case f.Names[0].Name != "_":
				__antithesis_instrumentation__.Notify(645078)
				pass.Reportf(f.Pos(), "Illegal use of util.NoCopy - should be unnamed")
			default:
				__antithesis_instrumentation__.Notify(645079)

			}
		}
	})
	__antithesis_instrumentation__.Notify(645063)
	return nil, nil
}
