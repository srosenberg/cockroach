package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
	"github.com/golang-commonmark/markdown"
	"github.com/spf13/cobra"
)

func init() {
	cmds = append(cmds, &cobra.Command{
		Use:   "functions <output-dir>",
		Short: "generate markdown documentation of functions and operators",
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir := filepath.Join("docs", "generated", "sql")
			if len(args) > 0 {
				outDir = args[0]
			}

			if stat, err := os.Stat(outDir); err != nil {
				return err
			} else if !stat.IsDir() {
				return errors.Errorf("%q is not a directory", outDir)
			}

			if err := ioutil.WriteFile(
				filepath.Join(outDir, "functions.md"), generateFunctions(builtins.AllBuiltinNames, true), 0644,
			); err != nil {
				return err
			}
			if err := ioutil.WriteFile(
				filepath.Join(outDir, "aggregates.md"), generateFunctions(builtins.AllAggregateBuiltinNames, false), 0644,
			); err != nil {
				return err
			}
			if err := ioutil.WriteFile(
				filepath.Join(outDir, "window_functions.md"), generateFunctions(builtins.AllWindowBuiltinNames, false), 0644,
			); err != nil {
				return err
			}
			return ioutil.WriteFile(
				filepath.Join(outDir, "operators.md"), generateOperators(), 0644,
			)
		},
	})
}

type operation struct {
	left  string
	right string
	ret   string
	op    string
}

func (o operation) String() string {
	__antithesis_instrumentation__.Notify(39790)
	if o.right == "" {
		__antithesis_instrumentation__.Notify(39792)
		return fmt.Sprintf("<code>%s</code>%s", o.op, linkTypeName(o.left))
	} else {
		__antithesis_instrumentation__.Notify(39793)
	}
	__antithesis_instrumentation__.Notify(39791)
	return fmt.Sprintf("%s <code>%s</code> %s", linkTypeName(o.left), o.op, linkTypeName(o.right))
}

type operations []operation

func (p operations) Len() int { __antithesis_instrumentation__.Notify(39794); return len(p) }
func (p operations) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(39795)
	p[i], p[j] = p[j], p[i]
}
func (p operations) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(39796)
	if p[i].right != "" && func() bool {
		__antithesis_instrumentation__.Notify(39801)
		return p[j].right == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(39802)
		return false
	} else {
		__antithesis_instrumentation__.Notify(39803)
	}
	__antithesis_instrumentation__.Notify(39797)
	if p[i].right == "" && func() bool {
		__antithesis_instrumentation__.Notify(39804)
		return p[j].right != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(39805)
		return true
	} else {
		__antithesis_instrumentation__.Notify(39806)
	}
	__antithesis_instrumentation__.Notify(39798)
	if p[i].left != p[j].left {
		__antithesis_instrumentation__.Notify(39807)
		return p[i].left < p[j].left
	} else {
		__antithesis_instrumentation__.Notify(39808)
	}
	__antithesis_instrumentation__.Notify(39799)
	if p[i].right != p[j].right {
		__antithesis_instrumentation__.Notify(39809)
		return p[i].right < p[j].right
	} else {
		__antithesis_instrumentation__.Notify(39810)
	}
	__antithesis_instrumentation__.Notify(39800)
	return p[i].ret < p[j].ret
}

func generateOperators() []byte {
	__antithesis_instrumentation__.Notify(39811)
	ops := make(map[string]operations)
	for optyp, overloads := range tree.UnaryOps {
		__antithesis_instrumentation__.Notify(39817)
		op := optyp.String()
		for _, untyped := range overloads {
			__antithesis_instrumentation__.Notify(39818)
			v := untyped.(*tree.UnaryOp)
			ops[op] = append(ops[op], operation{
				left: v.Typ.String(),
				ret:  v.ReturnType.String(),
				op:   op,
			})
		}
	}
	__antithesis_instrumentation__.Notify(39812)
	for optyp, overloads := range tree.BinOps {
		__antithesis_instrumentation__.Notify(39819)
		op := optyp.String()
		for _, untyped := range overloads {
			__antithesis_instrumentation__.Notify(39820)
			v := untyped.(*tree.BinOp)
			left := v.LeftType.String()
			right := v.RightType.String()
			ops[op] = append(ops[op], operation{
				left:  left,
				right: right,
				ret:   v.ReturnType.String(),
				op:    op,
			})
		}
	}
	__antithesis_instrumentation__.Notify(39813)
	for optyp, overloads := range tree.CmpOps {
		__antithesis_instrumentation__.Notify(39821)
		op := optyp.String()
		for _, untyped := range overloads {
			__antithesis_instrumentation__.Notify(39822)
			v := untyped.(*tree.CmpOp)
			left := v.LeftType.String()
			right := v.RightType.String()
			ops[op] = append(ops[op], operation{
				left:  left,
				right: right,
				ret:   "bool",
				op:    op,
			})
		}
	}
	__antithesis_instrumentation__.Notify(39814)
	var opstrs []string
	for k, v := range ops {
		__antithesis_instrumentation__.Notify(39823)
		sort.Sort(v)
		opstrs = append(opstrs, k)
	}
	__antithesis_instrumentation__.Notify(39815)
	sort.Strings(opstrs)
	b := new(bytes.Buffer)
	seen := map[string]bool{}
	for _, op := range opstrs {
		__antithesis_instrumentation__.Notify(39824)
		fmt.Fprintf(b, "<table><thead>\n")
		fmt.Fprintf(b, "<tr><td><code>%s</code></td><td>Return</td></tr>\n", op)
		fmt.Fprintf(b, "</thead><tbody>\n")
		for _, v := range ops[op] {
			__antithesis_instrumentation__.Notify(39826)
			s := fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>\n", v.String(), linkTypeName(v.ret))
			if seen[s] {
				__antithesis_instrumentation__.Notify(39828)
				continue
			} else {
				__antithesis_instrumentation__.Notify(39829)
			}
			__antithesis_instrumentation__.Notify(39827)
			seen[s] = true
			b.WriteString(s)
		}
		__antithesis_instrumentation__.Notify(39825)
		fmt.Fprintf(b, "</tbody></table>")
		fmt.Fprintln(b)
	}
	__antithesis_instrumentation__.Notify(39816)
	return b.Bytes()
}

const notUsableInfo = "Not usable; exposed only for compatibility with PostgreSQL."

func generateFunctions(from []string, categorize bool) []byte {
	__antithesis_instrumentation__.Notify(39830)
	functions := make(map[string][]string)
	seen := make(map[string]struct{})
	md := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	for _, name := range from {
		__antithesis_instrumentation__.Notify(39835)

		name = strings.ToLower(name)
		if _, ok := seen[name]; ok {
			__antithesis_instrumentation__.Notify(39838)
			continue
		} else {
			__antithesis_instrumentation__.Notify(39839)
		}
		__antithesis_instrumentation__.Notify(39836)
		seen[name] = struct{}{}
		props, fns := builtins.GetBuiltinProperties(name)
		if !props.ShouldDocument() {
			__antithesis_instrumentation__.Notify(39840)
			continue
		} else {
			__antithesis_instrumentation__.Notify(39841)
		}
		__antithesis_instrumentation__.Notify(39837)
		for _, fn := range fns {
			__antithesis_instrumentation__.Notify(39842)
			if fn.Info == notUsableInfo {
				__antithesis_instrumentation__.Notify(39848)
				continue
			} else {
				__antithesis_instrumentation__.Notify(39849)
			}
			__antithesis_instrumentation__.Notify(39843)

			if categorize && func() bool {
				__antithesis_instrumentation__.Notify(39850)
				return (props.Class == tree.AggregateClass || func() bool {
					__antithesis_instrumentation__.Notify(39851)
					return props.Class == tree.WindowClass == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(39852)
				continue
			} else {
				__antithesis_instrumentation__.Notify(39853)
			}
			__antithesis_instrumentation__.Notify(39844)
			args := fn.Types.String()

			retType := fn.InferReturnTypeFromInputArgTypes(fn.Types.Types())
			ret := retType.String()

			cat := props.Category
			if cat == "" {
				__antithesis_instrumentation__.Notify(39854)
				cat = strings.ToUpper(ret)
			} else {
				__antithesis_instrumentation__.Notify(39855)
			}
			__antithesis_instrumentation__.Notify(39845)
			if !categorize {
				__antithesis_instrumentation__.Notify(39856)
				cat = ""
			} else {
				__antithesis_instrumentation__.Notify(39857)
			}
			__antithesis_instrumentation__.Notify(39846)
			extra := ""
			if fn.Info != "" {
				__antithesis_instrumentation__.Notify(39858)

				info := md.RenderToString([]byte(fn.Info))
				extra = fmt.Sprintf("<span class=\"funcdesc\">%s</span>", info)
			} else {
				__antithesis_instrumentation__.Notify(39859)
			}
			__antithesis_instrumentation__.Notify(39847)
			s := fmt.Sprintf("<tr><td><a name=\"%s\"></a><code>%s(%s) &rarr; %s</code></td><td>%s</td></tr>", name, name, linkArguments(args), linkArguments(ret), extra)
			functions[cat] = append(functions[cat], s)
		}
	}
	__antithesis_instrumentation__.Notify(39831)
	var cats []string
	for k, v := range functions {
		__antithesis_instrumentation__.Notify(39860)
		sort.Strings(v)
		cats = append(cats, k)
	}
	__antithesis_instrumentation__.Notify(39832)
	sort.Strings(cats)

	for i, cat := range cats {
		__antithesis_instrumentation__.Notify(39861)
		if cat == "Compatibility" {
			__antithesis_instrumentation__.Notify(39862)
			cats = append(append(cats[:i], cats[i+1:]...), "Compatibility")
			break
		} else {
			__antithesis_instrumentation__.Notify(39863)
		}
	}
	__antithesis_instrumentation__.Notify(39833)
	b := new(bytes.Buffer)
	for _, cat := range cats {
		__antithesis_instrumentation__.Notify(39864)
		if categorize {
			__antithesis_instrumentation__.Notify(39866)
			fmt.Fprintf(b, "### %s functions\n\n", cat)
		} else {
			__antithesis_instrumentation__.Notify(39867)
		}
		__antithesis_instrumentation__.Notify(39865)
		b.WriteString("<table>\n<thead><tr><th>Function &rarr; Returns</th><th>Description</th></tr></thead>\n")
		b.WriteString("<tbody>\n")
		b.WriteString(strings.Join(functions[cat], "\n"))
		b.WriteString("</tbody>\n</table>\n\n")
	}
	__antithesis_instrumentation__.Notify(39834)
	return b.Bytes()
}

var linkRE = regexp.MustCompile(`([a-z]+)([\.\[\]]*)$`)

func linkArguments(t string) string {
	__antithesis_instrumentation__.Notify(39868)
	sp := strings.Split(t, ", ")
	for i, s := range sp {
		__antithesis_instrumentation__.Notify(39870)
		sp[i] = linkRE.ReplaceAllStringFunc(s, func(s string) string {
			__antithesis_instrumentation__.Notify(39871)
			match := linkRE.FindStringSubmatch(s)
			s = linkTypeName(match[1])
			return s + match[2]
		})
	}
	__antithesis_instrumentation__.Notify(39869)
	return strings.Join(sp, ", ")
}

func linkTypeName(s string) string {
	__antithesis_instrumentation__.Notify(39872)
	s = strings.TrimSuffix(s, "{}")
	s = strings.TrimSuffix(s, "{*}")
	name := s
	switch s {
	case "timestamptz":
		__antithesis_instrumentation__.Notify(39875)
		s = "timestamp"
	case "collatedstring":
		__antithesis_instrumentation__.Notify(39876)
		s = "collate"
	default:
		__antithesis_instrumentation__.Notify(39877)
	}
	__antithesis_instrumentation__.Notify(39873)
	s = strings.TrimSuffix(s, "[]")
	s = strings.TrimSuffix(s, "*")
	switch s {
	case "int", "decimal", "float", "bool", "date", "timestamp", "interval", "string", "bytes",
		"inet", "uuid", "collate", "time":
		__antithesis_instrumentation__.Notify(39878)
		s = fmt.Sprintf("<a href=\"%s.html\">%s</a>", s, name)
	default:
		__antithesis_instrumentation__.Notify(39879)
	}
	__antithesis_instrumentation__.Notify(39874)
	return s
}
