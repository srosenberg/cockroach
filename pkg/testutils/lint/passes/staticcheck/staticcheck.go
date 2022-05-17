// Package staticcheck provides utilities for consuming `staticcheck` checks in
// `nogo`.
package staticcheck

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/analysis/facts"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/analysis/report"
)

func MungeAnalyzer(analyzer *analysis.Analyzer) {
	__antithesis_instrumentation__.Notify(645333)

	analyzer.Requires = analyzer.Requires[0:len(analyzer.Requires):len(analyzer.Requires)]
	analyzer.Requires = append(analyzer.Requires, facts.Directives)
	oldRun := analyzer.Run
	analyzer.Run = func(p *analysis.Pass) (interface{}, error) {
		__antithesis_instrumentation__.Notify(645334)
		pass := *p
		oldReport := p.Report
		pass.Report = func(diag analysis.Diagnostic) {
			__antithesis_instrumentation__.Notify(645336)
			dirs := pass.ResultOf[facts.Directives].([]lint.Directive)
			for _, dir := range dirs {
				__antithesis_instrumentation__.Notify(645338)
				cmd := dir.Command
				args := dir.Arguments
				switch cmd {
				case "ignore":
					__antithesis_instrumentation__.Notify(645339)
					ignorePos := report.DisplayPosition(pass.Fset, dir.Node.Pos())
					nodePos := report.DisplayPosition(pass.Fset, diag.Pos)
					if ignorePos.Filename != nodePos.Filename || func() bool {
						__antithesis_instrumentation__.Notify(645343)
						return ignorePos.Line != nodePos.Line == true
					}() == true {
						__antithesis_instrumentation__.Notify(645344)
						continue
					} else {
						__antithesis_instrumentation__.Notify(645345)
					}
					__antithesis_instrumentation__.Notify(645340)
					for _, check := range strings.Split(args[0], ",") {
						__antithesis_instrumentation__.Notify(645346)
						if check == analyzer.Name {
							__antithesis_instrumentation__.Notify(645347)

							return
						} else {
							__antithesis_instrumentation__.Notify(645348)
						}
					}
				case "file-ignore":
					__antithesis_instrumentation__.Notify(645341)
					ignorePos := report.DisplayPosition(pass.Fset, dir.Node.Pos())
					nodePos := report.DisplayPosition(pass.Fset, diag.Pos)
					if ignorePos.Filename == nodePos.Filename {
						__antithesis_instrumentation__.Notify(645349)

						return
					} else {
						__antithesis_instrumentation__.Notify(645350)
					}
				default:
					__antithesis_instrumentation__.Notify(645342)

					continue
				}
			}
			__antithesis_instrumentation__.Notify(645337)
			oldReport(diag)
		}
		__antithesis_instrumentation__.Notify(645335)
		return oldRun(&pass)
	}
}
