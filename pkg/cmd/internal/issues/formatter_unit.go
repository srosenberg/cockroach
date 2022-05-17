package issues

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"sort"
)

var UnitTestFormatter = IssueFormatter{
	Title: func(data TemplateData) string {
		__antithesis_instrumentation__.Notify(41085)
		return fmt.Sprintf("%s: %s failed", data.PackageNameShort, data.TestName)
	},
	Body: func(r *Renderer, data TemplateData) error {
		__antithesis_instrumentation__.Notify(41086)
		r.Escaped(fmt.Sprintf("%s.%s ", data.PackageNameShort, data.TestName))
		r.A(
			"failed",
			data.URL,
		)
		if data.ArtifactsURL != "" {
			__antithesis_instrumentation__.Notify(41094)
			r.Escaped(" with ")
			r.A(
				"artifacts",
				data.ArtifactsURL,
			)
		} else {
			__antithesis_instrumentation__.Notify(41095)
		}
		__antithesis_instrumentation__.Notify(41087)
		r.Escaped(" on " + data.Branch + " @ ")
		r.A(
			data.Commit,
			data.CommitURL,
		)
		r.Escaped(`:

`)
		if fop, ok := data.CondensedMessage.FatalOrPanic(50); ok {
			__antithesis_instrumentation__.Notify(41096)
			if fop.Error != "" {
				__antithesis_instrumentation__.Notify(41099)
				r.Escaped("Fatal error:")
				r.CodeBlock("", fop.Error)
			} else {
				__antithesis_instrumentation__.Notify(41100)
			}
			__antithesis_instrumentation__.Notify(41097)
			if fop.FirstStack != "" {
				__antithesis_instrumentation__.Notify(41101)
				r.Escaped("Stack: ")
				r.CodeBlock("", fop.FirstStack)
			} else {
				__antithesis_instrumentation__.Notify(41102)
			}
			__antithesis_instrumentation__.Notify(41098)

			r.Collapsed("Log preceding fatal error", func() {
				__antithesis_instrumentation__.Notify(41103)
				r.CodeBlock("", fop.LastLines)
			})
		} else {
			__antithesis_instrumentation__.Notify(41104)
			if rsgCrash, ok := data.CondensedMessage.RSGCrash(100); ok {
				__antithesis_instrumentation__.Notify(41105)
				r.Escaped("Random syntax error:")
				r.CodeBlock("", rsgCrash.Error)
				r.Escaped("Query:")
				r.CodeBlock("", rsgCrash.Query)
				if rsgCrash.Schema != "" {
					__antithesis_instrumentation__.Notify(41106)
					r.Escaped("Schema:")
					r.CodeBlock("", rsgCrash.Schema)
				} else {
					__antithesis_instrumentation__.Notify(41107)
				}
			} else {
				__antithesis_instrumentation__.Notify(41108)
				r.CodeBlock("", data.CondensedMessage.Digest(50))
			}
		}
		__antithesis_instrumentation__.Notify(41088)

		r.Collapsed("Help", func() {
			__antithesis_instrumentation__.Notify(41109)
			if data.HelpCommand != nil {
				__antithesis_instrumentation__.Notify(41111)
				data.HelpCommand(r)
			} else {
				__antithesis_instrumentation__.Notify(41112)
			}
			__antithesis_instrumentation__.Notify(41110)

			if len(data.Parameters) != 0 {
				__antithesis_instrumentation__.Notify(41113)
				r.Escaped("Parameters in this failure:\n")
				for _, p := range data.Parameters {
					__antithesis_instrumentation__.Notify(41114)
					r.Escaped("\n- ")
					r.Escaped(p)
					r.Escaped("\n")
				}
			} else {
				__antithesis_instrumentation__.Notify(41115)
			}
		})
		__antithesis_instrumentation__.Notify(41089)

		if len(data.RelatedIssues) > 0 {
			__antithesis_instrumentation__.Notify(41116)
			r.Collapsed("Same failure on other branches", func() {
				__antithesis_instrumentation__.Notify(41117)
				for _, iss := range data.RelatedIssues {
					__antithesis_instrumentation__.Notify(41119)
					var ls []string
					for _, l := range iss.Labels {
						__antithesis_instrumentation__.Notify(41121)
						ls = append(ls, l.GetName())
					}
					__antithesis_instrumentation__.Notify(41120)
					sort.Strings(ls)
					r.Escaped("\n- ")
					r.Escaped(fmt.Sprintf("#%d %s %v", iss.GetNumber(), iss.GetTitle(), ls))
				}
				__antithesis_instrumentation__.Notify(41118)
				r.Escaped("\n")
			})
		} else {
			__antithesis_instrumentation__.Notify(41122)
		}
		__antithesis_instrumentation__.Notify(41090)

		if data.InternalLog != "" {
			__antithesis_instrumentation__.Notify(41123)
			r.Collapsed("Internal log", func() {
				__antithesis_instrumentation__.Notify(41125)
				r.CodeBlock("", data.InternalLog)
			})
			__antithesis_instrumentation__.Notify(41124)
			r.Escaped("\n")
		} else {
			__antithesis_instrumentation__.Notify(41126)
		}
		__antithesis_instrumentation__.Notify(41091)

		if len(data.MentionOnCreate) > 0 {
			__antithesis_instrumentation__.Notify(41127)
			r.Escaped("/cc")
			for _, handle := range data.MentionOnCreate {
				__antithesis_instrumentation__.Notify(41129)
				r.Escaped(" ")
				r.Escaped(handle)
			}
			__antithesis_instrumentation__.Notify(41128)
			r.Escaped("\n")
		} else {
			__antithesis_instrumentation__.Notify(41130)
		}
		__antithesis_instrumentation__.Notify(41092)

		r.HTML("sub", func() {
			__antithesis_instrumentation__.Notify(41131)
			r.Escaped("\n\n")
			r.A(
				"This test on roachdash",
				"https://roachdash.crdb.dev/?filter=status:open%20t:.*"+
					data.TestName+
					".*&sort=title+created&display=lastcommented+project",
			)
			r.Escaped(" | ")
			r.A("Improve this report!",
				"https://github.com/cockroachdb/cockroach/tree/master/pkg/cmd/internal/issues",
			)
			r.Escaped("\n")
		})
		__antithesis_instrumentation__.Notify(41093)
		return nil
	},
}

func UnitTestHelpCommand(repro string) func(r *Renderer) {
	__antithesis_instrumentation__.Notify(41132)
	return func(r *Renderer) {
		__antithesis_instrumentation__.Notify(41133)
		ReproductionCommandFromString(repro)
		r.Escaped("\n")
		r.Escaped("See also: ")
		r.A("How To Investigate a Go Test Failure (internal)", "https://cockroachlabs.atlassian.net/l/c/HgfXfJgM")
		r.Escaped("\n")
	}
}
