package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"text/template"

	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/spf13/cobra"
)

func init() {
	cmds = append(cmds, &cobra.Command{
		Use:   "logformats",
		Short: "Generate the markdown documentation for logging formats.",
		Args:  cobra.MaximumNArgs(1),
		Run:   runLogFormats,
	})
}

func runLogFormats(_ *cobra.Command, args []string) {
	__antithesis_instrumentation__.Notify(39972)
	if err := runLogFormatsInternal(args); err != nil {
		__antithesis_instrumentation__.Notify(39973)
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		exit.WithCode(exit.UnspecifiedError())
	} else {
		__antithesis_instrumentation__.Notify(39974)
	}
}

func runLogFormatsInternal(args []string) error {
	__antithesis_instrumentation__.Notify(39975)

	tmpl, err := template.New("format docs").Parse(fmtDocTemplate)
	if err != nil {
		__antithesis_instrumentation__.Notify(39982)
		return err
	} else {
		__antithesis_instrumentation__.Notify(39983)
	}
	__antithesis_instrumentation__.Notify(39976)

	m := log.GetFormatterDocs()

	fNames := make([]string, 0, len(m))
	for k := range m {
		__antithesis_instrumentation__.Notify(39984)
		fNames = append(fNames, k)
	}
	__antithesis_instrumentation__.Notify(39977)
	sort.Strings(fNames)

	type info struct {
		Name string
		Doc  string
	}
	var infos []info
	for _, k := range fNames {
		__antithesis_instrumentation__.Notify(39985)
		infos = append(infos, info{Name: k, Doc: m[k]})
	}
	__antithesis_instrumentation__.Notify(39978)

	var src bytes.Buffer
	if err := tmpl.Execute(&src, struct {
		Formats []info
	}{infos}); err != nil {
		__antithesis_instrumentation__.Notify(39986)
		return err
	} else {
		__antithesis_instrumentation__.Notify(39987)
	}
	__antithesis_instrumentation__.Notify(39979)

	w := os.Stdout
	if len(args) > 0 {
		__antithesis_instrumentation__.Notify(39988)
		f, err := os.OpenFile(args[0], os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		if err != nil {
			__antithesis_instrumentation__.Notify(39991)
			return err
		} else {
			__antithesis_instrumentation__.Notify(39992)
		}
		__antithesis_instrumentation__.Notify(39989)
		defer func() { __antithesis_instrumentation__.Notify(39993); _ = f.Close() }()
		__antithesis_instrumentation__.Notify(39990)
		w = f
	} else {
		__antithesis_instrumentation__.Notify(39994)
	}
	__antithesis_instrumentation__.Notify(39980)
	if _, err := w.Write(src.Bytes()); err != nil {
		__antithesis_instrumentation__.Notify(39995)
		return err
	} else {
		__antithesis_instrumentation__.Notify(39996)
	}
	__antithesis_instrumentation__.Notify(39981)

	return nil
}

const fmtDocTemplate = `
The supported log output formats are documented below.

{{range .Formats}}
- [` + "`{{.Name}}`" + `](#format-{{.Name}})
{{end}}

{{range .Formats}}
## Format ` + "`{{.Name}}`" + `

{{.Doc}}
{{end}}
`
