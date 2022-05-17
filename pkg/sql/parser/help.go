package parser

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
)

type HelpMessage struct {
	Command string

	Function string

	HelpMessageBody
}

const helpHintPrefix = "help:"

func (h *HelpMessage) String() string {
	__antithesis_instrumentation__.Notify(552299)
	var buf bytes.Buffer
	buf.WriteString(helpHintPrefix + "\n")
	h.Format(&buf)
	return buf.String()
}

func (h *HelpMessage) Format(w io.Writer) {
	__antithesis_instrumentation__.Notify(552300)
	if h.Command != "" {
		__antithesis_instrumentation__.Notify(552306)
		fmt.Fprintf(w, "Command:     %s\n", h.Command)
	} else {
		__antithesis_instrumentation__.Notify(552307)
	}
	__antithesis_instrumentation__.Notify(552301)
	if h.Function != "" {
		__antithesis_instrumentation__.Notify(552308)
		fmt.Fprintf(w, "Function:    %s\n", h.Function)
	} else {
		__antithesis_instrumentation__.Notify(552309)
	}
	__antithesis_instrumentation__.Notify(552302)
	if h.ShortDescription != "" {
		__antithesis_instrumentation__.Notify(552310)
		fmt.Fprintf(w, "Description: %s\n", h.ShortDescription)
	} else {
		__antithesis_instrumentation__.Notify(552311)
	}
	__antithesis_instrumentation__.Notify(552303)
	if h.Category != "" {
		__antithesis_instrumentation__.Notify(552312)
		fmt.Fprintf(w, "Category:    %s\n", h.Category)
	} else {
		__antithesis_instrumentation__.Notify(552313)
	}
	__antithesis_instrumentation__.Notify(552304)
	if h.Command != "" {
		__antithesis_instrumentation__.Notify(552314)
		fmt.Fprintln(w, "Syntax:")
	} else {
		__antithesis_instrumentation__.Notify(552315)
	}
	__antithesis_instrumentation__.Notify(552305)
	fmt.Fprintln(w, strings.TrimSpace(h.Text))
	if h.SeeAlso != "" {
		__antithesis_instrumentation__.Notify(552316)
		fmt.Fprintf(w, "\nSee also:\n  %s\n", h.SeeAlso)
	} else {
		__antithesis_instrumentation__.Notify(552317)
	}
}

func helpWith(sqllex sqlLexer, helpText string) int {
	__antithesis_instrumentation__.Notify(552318)
	scan := sqllex.(*lexer)
	if helpText == "" {
		__antithesis_instrumentation__.Notify(552320)
		scan.lastError = pgerror.WithCandidateCode(errors.New("help upon syntax error"), pgcode.Syntax)
		scan.populateHelpMsg("help:\n" + AllHelp)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(552321)
	}
	__antithesis_instrumentation__.Notify(552319)
	msg := HelpMessage{Command: helpText, HelpMessageBody: HelpMessages[helpText]}
	scan.SetHelp(msg)

	return 1
}

func helpWithFunction(sqllex sqlLexer, f tree.ResolvableFunctionReference) int {
	__antithesis_instrumentation__.Notify(552322)
	d, err := f.Resolve(sessiondata.SearchPath{})
	if err != nil {
		__antithesis_instrumentation__.Notify(552325)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(552326)
	}
	__antithesis_instrumentation__.Notify(552323)

	msg := HelpMessage{
		Function: f.String(),
		HelpMessageBody: HelpMessageBody{
			Category: d.Category,
			SeeAlso:  docs.URL("functions-and-operators.html"),
		},
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)

	lastInfo := ""
	for i, overload := range d.Definition {
		__antithesis_instrumentation__.Notify(552327)
		b := overload.(*tree.Overload)
		if b.Info != "" && func() bool {
			__antithesis_instrumentation__.Notify(552329)
			return b.Info != lastInfo == true
		}() == true {
			__antithesis_instrumentation__.Notify(552330)
			if i > 0 {
				__antithesis_instrumentation__.Notify(552332)
				fmt.Fprintln(w, "---")
			} else {
				__antithesis_instrumentation__.Notify(552333)
			}
			__antithesis_instrumentation__.Notify(552331)
			fmt.Fprintf(w, "\n%s\n\n", b.Info)
			fmt.Fprintln(w, "Signature")
		} else {
			__antithesis_instrumentation__.Notify(552334)
		}
		__antithesis_instrumentation__.Notify(552328)
		lastInfo = b.Info

		simplifyRet := d.Class == tree.GeneratorClass
		fmt.Fprintf(w, "%s%s\n", d.Name, b.Signature(simplifyRet))
	}
	__antithesis_instrumentation__.Notify(552324)
	_ = w.Flush()
	msg.Text = buf.String()

	sqllex.(*lexer).SetHelp(msg)
	return 1
}

func helpWithFunctionByName(sqllex sqlLexer, s string) int {
	__antithesis_instrumentation__.Notify(552335)
	un := &tree.UnresolvedName{NumParts: 1, Parts: tree.NameParts{s}}
	return helpWithFunction(sqllex, tree.ResolvableFunctionReference{FunctionReference: un})
}

const (
	hGroup        = ""
	hDDL          = "schema manipulation"
	hDML          = "data manipulation"
	hTxn          = "transaction control"
	hPriv         = "privileges and security"
	hMisc         = "miscellaneous"
	hCfg          = "configuration"
	hExperimental = "experimental"
	hCCL          = "enterprise features"
)

type HelpMessageBody struct {
	Category         string
	ShortDescription string
	Text             string
	SeeAlso          string
}

var HelpMessages = func(h map[string]HelpMessageBody) map[string]HelpMessageBody {
	__antithesis_instrumentation__.Notify(552336)
	appendSeeAlso := func(newItem, prevItems string) string {
		__antithesis_instrumentation__.Notify(552340)

		if prevItems != "" {
			__antithesis_instrumentation__.Notify(552342)
			return newItem + "\n  " + prevItems
		} else {
			__antithesis_instrumentation__.Notify(552343)
		}
		__antithesis_instrumentation__.Notify(552341)
		return newItem
	}
	__antithesis_instrumentation__.Notify(552337)
	reformatSeeAlso := func(seeAlso string) string {
		__antithesis_instrumentation__.Notify(552344)
		return strings.Replace(
			strings.Replace(seeAlso, ", ", "\n  ", -1),
			"WEBDOCS", docs.URLBase, -1)
	}
	__antithesis_instrumentation__.Notify(552338)
	srcMsg := h["<SOURCE>"]
	srcMsg.SeeAlso = reformatSeeAlso(strings.TrimSpace(srcMsg.SeeAlso))
	selectMsg := h["<SELECTCLAUSE>"]
	selectMsg.SeeAlso = reformatSeeAlso(strings.TrimSpace(selectMsg.SeeAlso))
	for k, m := range h {
		__antithesis_instrumentation__.Notify(552345)
		m = h[k]
		m.ShortDescription = strings.TrimSpace(m.ShortDescription)
		m.Text = strings.TrimSpace(m.Text)
		m.SeeAlso = strings.TrimSpace(m.SeeAlso)

		if strings.Contains(m.Text, "<source>") && func() bool {
			__antithesis_instrumentation__.Notify(552349)
			return k != "<SOURCE>" == true
		}() == true {
			__antithesis_instrumentation__.Notify(552350)
			m.Text = strings.TrimSpace(m.Text) + "\n\n" + strings.TrimSpace(srcMsg.Text)
			m.SeeAlso = appendSeeAlso(srcMsg.SeeAlso, m.SeeAlso)
		} else {
			__antithesis_instrumentation__.Notify(552351)
		}
		__antithesis_instrumentation__.Notify(552346)

		if strings.Contains(m.Text, "<selectclause>") && func() bool {
			__antithesis_instrumentation__.Notify(552352)
			return k != "<SELECTCLAUSE>" == true
		}() == true {
			__antithesis_instrumentation__.Notify(552353)
			m.Text = strings.TrimSpace(m.Text) + "\n\n" + strings.TrimSpace(selectMsg.Text)
			m.SeeAlso = appendSeeAlso(selectMsg.SeeAlso, m.SeeAlso)
		} else {
			__antithesis_instrumentation__.Notify(552354)
		}
		__antithesis_instrumentation__.Notify(552347)

		if strings.Contains(m.Text, "<tablename>") {
			__antithesis_instrumentation__.Notify(552355)
			m.SeeAlso = appendSeeAlso("SHOW TABLES", m.SeeAlso)
		} else {
			__antithesis_instrumentation__.Notify(552356)
		}
		__antithesis_instrumentation__.Notify(552348)
		m.SeeAlso = reformatSeeAlso(m.SeeAlso)
		h[k] = m
	}
	__antithesis_instrumentation__.Notify(552339)
	return h
}(helpMessages)

var AllHelp = func(h map[string]HelpMessageBody) string {
	__antithesis_instrumentation__.Notify(552357)

	cmds := make(map[string][]string)
	for c, details := range h {
		__antithesis_instrumentation__.Notify(552361)
		if details.Category == "" {
			__antithesis_instrumentation__.Notify(552363)
			continue
		} else {
			__antithesis_instrumentation__.Notify(552364)
		}
		__antithesis_instrumentation__.Notify(552362)
		cmds[details.Category] = append(cmds[details.Category], c)
	}
	__antithesis_instrumentation__.Notify(552358)

	var categories []string
	for c, l := range cmds {
		__antithesis_instrumentation__.Notify(552365)
		categories = append(categories, c)
		sort.Strings(l)
	}
	__antithesis_instrumentation__.Notify(552359)
	sort.Strings(categories)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
	for _, cat := range categories {
		__antithesis_instrumentation__.Notify(552366)
		fmt.Fprintf(w, "%s:\n", strings.Title(cat))
		for _, item := range cmds[cat] {
			__antithesis_instrumentation__.Notify(552368)
			fmt.Fprintf(w, "\t\t%s\t%s\n", item, h[item].ShortDescription)
		}
		__antithesis_instrumentation__.Notify(552367)
		fmt.Fprintln(w)
	}
	__antithesis_instrumentation__.Notify(552360)
	_ = w.Flush()
	return buf.String()
}(helpMessages)
