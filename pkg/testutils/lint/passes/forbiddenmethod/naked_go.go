package forbiddenmethod

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/testutils/lint/passes/passesutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const nakedGoPassName = "nakedgo"

var NakedGoAnalyzer = &analysis.Analyzer{
	Name:     nakedGoPassName,
	Doc:      "Prevents direct use of the 'go' keyword. Goroutines should be launched through Stopper.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (interface{}, error) {
		__antithesis_instrumentation__.Notify(644736)
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		inspect.Preorder([]ast.Node{
			(*ast.GoStmt)(nil),
		}, func(n ast.Node) {
			__antithesis_instrumentation__.Notify(644738)
			node := n.(*ast.GoStmt)

			const debug = false

			f := passesutil.FindContainingFile(pass, n)
			cm := ast.NewCommentMap(pass.Fset, node, f.Comments)
			var cc *ast.Comment
			for _, cg := range cm[n] {
				__antithesis_instrumentation__.Notify(644741)
				for _, c := range cg.List {
					__antithesis_instrumentation__.Notify(644742)
					if c.Pos() < node.Go {
						__antithesis_instrumentation__.Notify(644744)

						continue
					} else {
						__antithesis_instrumentation__.Notify(644745)
					}
					__antithesis_instrumentation__.Notify(644743)
					if cc == nil || func() bool {
						__antithesis_instrumentation__.Notify(644746)
						return cc.Pos() > node.Go == true
					}() == true {
						__antithesis_instrumentation__.Notify(644747)

						cc = c
						if debug {
							__antithesis_instrumentation__.Notify(644748)
							fmt.Printf("closest comment now %d-%d: %s\n", cc.Pos(), cc.End(), cc.Text)
						} else {
							__antithesis_instrumentation__.Notify(644749)
						}
					} else {
						__antithesis_instrumentation__.Notify(644750)
					}
				}
			}
			__antithesis_instrumentation__.Notify(644739)
			if cc != nil && func() bool {
				__antithesis_instrumentation__.Notify(644751)
				return strings.Contains(cc.Text, "nolint:"+nakedGoPassName) == true
			}() == true {
				__antithesis_instrumentation__.Notify(644752)
				if debug {
					__antithesis_instrumentation__.Notify(644754)
					fmt.Printf("GoStmt at: %d-%d\n", n.Pos(), n.End())
					fmt.Printf("GoStmt.Go at: %d\n", node.Go)
					fmt.Printf("GoStmt.Call at: %d-%d\n", node.Call.Pos(), node.Call.End())
				} else {
					__antithesis_instrumentation__.Notify(644755)
				}
				__antithesis_instrumentation__.Notify(644753)

				goPos := pass.Fset.Position(node.End())
				cmPos := pass.Fset.Position(cc.Pos())

				if goPos.Line == cmPos.Line {
					__antithesis_instrumentation__.Notify(644756)
					if debug {
						__antithesis_instrumentation__.Notify(644758)
						fmt.Printf("suppressing lint because of %d-%d: %s\n", cc.Pos(), cc.End(), cc.Text)
					} else {
						__antithesis_instrumentation__.Notify(644759)
					}
					__antithesis_instrumentation__.Notify(644757)
					return
				} else {
					__antithesis_instrumentation__.Notify(644760)
				}
			} else {
				__antithesis_instrumentation__.Notify(644761)
			}
			__antithesis_instrumentation__.Notify(644740)

			pass.Report(analysis.Diagnostic{
				Pos:     n.Pos(),
				Message: "Use of go keyword not allowed, use a Stopper instead",
			})
		})
		__antithesis_instrumentation__.Notify(644737)
		return nil, nil
	},
}
