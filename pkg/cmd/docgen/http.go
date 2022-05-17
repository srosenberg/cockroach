package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

func init() {
	var (
		protocPath  string
		protocFlags string
		genDocPath  string
		outPath     string
	)

	cmdHTTP := &cobra.Command{
		Use:   "http",
		Short: "Generate HTTP docs",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runHTTP(protocPath, genDocPath, protocFlags, outPath); err != nil {
				fmt.Fprintln(os.Stdout, err)
				os.Exit(1)
			}
		},
	}
	cmdHTTP.Flags().StringVar(&protocPath, "protoc", "", `Path to the protoc compiler.
If given, we will call into this executable to generate the code; otherwise, we will call
into "buf protoc".`)
	cmdHTTP.Flags().StringVar(&protocFlags, "protoc-flags", "",
		"Whitespace-separated list of flags to pass to {buf} protoc. This should include the list of input sources.")
	cmdHTTP.Flags().StringVar(&genDocPath, "gendoc", "protoc-gen-doc", "Path to protoc-gen-doc binary.")
	cmdHTTP.Flags().StringVar(&outPath, "out", "docs/generated/http", "File output path.")

	cmds = append(cmds, cmdHTTP)
}

var singleMethods = []string{
	"HotRanges",
	"Nodes",
	"Health",
}

func runHTTP(protocPath, genDocPath, protocFlags, outPath string) error {
	__antithesis_instrumentation__.Notify(39880)

	if err := os.MkdirAll(outPath, 0777); err != nil {
		__antithesis_instrumentation__.Notify(39896)
		return err
	} else {
		__antithesis_instrumentation__.Notify(39897)
	}
	__antithesis_instrumentation__.Notify(39881)
	tmpJSON, err := ioutil.TempDir("", "docgen-*")
	if err != nil {
		__antithesis_instrumentation__.Notify(39898)
		return err
	} else {
		__antithesis_instrumentation__.Notify(39899)
	}
	__antithesis_instrumentation__.Notify(39882)

	jsonTmpl := filepath.Join(tmpJSON, "json.tmpl")
	if err := ioutil.WriteFile(jsonTmpl, []byte(tmplJSON), 0666); err != nil {
		__antithesis_instrumentation__.Notify(39900)
		return err
	} else {
		__antithesis_instrumentation__.Notify(39901)
	}
	__antithesis_instrumentation__.Notify(39883)
	defer func() {
		__antithesis_instrumentation__.Notify(39902)
		_ = os.RemoveAll(tmpJSON)
	}()
	__antithesis_instrumentation__.Notify(39884)
	var args []string
	if protocPath == "" {
		__antithesis_instrumentation__.Notify(39903)
		args = append(args, "protoc")
	} else {
		__antithesis_instrumentation__.Notify(39904)
	}
	__antithesis_instrumentation__.Notify(39885)
	args = append(args,
		fmt.Sprintf("--doc_out=%s", tmpJSON),
		fmt.Sprintf("--doc_opt=%s,http.json", jsonTmpl),
		fmt.Sprintf("--plugin=protoc-gen-doc=%s", genDocPath))
	args = append(args, strings.Fields(protocFlags)...)

	executable := protocPath
	if protocPath == "" {
		__antithesis_instrumentation__.Notify(39905)
		executable = "buf"
	} else {
		__antithesis_instrumentation__.Notify(39906)
	}
	__antithesis_instrumentation__.Notify(39886)
	cmd := exec.Command(executable, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		__antithesis_instrumentation__.Notify(39907)
		fmt.Println(string(out))
		return err
	} else {
		__antithesis_instrumentation__.Notify(39908)
	}
	__antithesis_instrumentation__.Notify(39887)
	dataFile, err := ioutil.ReadFile(filepath.Join(tmpJSON, "http.json"))
	if err != nil {
		__antithesis_instrumentation__.Notify(39909)
		return err
	} else {
		__antithesis_instrumentation__.Notify(39910)
	}
	__antithesis_instrumentation__.Notify(39888)
	var data protoData
	if err := json.Unmarshal(dataFile, &data); err != nil {
		__antithesis_instrumentation__.Notify(39911)
		return fmt.Errorf("json unmarshal: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(39912)
	}
	__antithesis_instrumentation__.Notify(39889)

	for k := range data.Files {
		__antithesis_instrumentation__.Notify(39913)
		file := &data.Files[k]
		for i := range file.Messages {
			__antithesis_instrumentation__.Notify(39915)
			m := &file.Messages[i]
			if !(strings.HasSuffix(m.Name, "Entry") && func() bool {
				__antithesis_instrumentation__.Notify(39916)
				return len(m.Fields) == 2 == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(39917)
				return m.Fields[0].Name == "key" == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(39918)
				return m.Fields[1].Name == "value" == true
			}() == true) {
				__antithesis_instrumentation__.Notify(39919)

				annotateStatus(&m.Description, &m.SupportStatus, "payload")
				for j := range m.Fields {
					__antithesis_instrumentation__.Notify(39920)
					f := &m.Fields[j]
					annotateStatus(&f.Description, &f.SupportStatus, "field")
				}
			} else {
				__antithesis_instrumentation__.Notify(39921)
			}
		}
		__antithesis_instrumentation__.Notify(39914)
		for j := range file.Services {
			__antithesis_instrumentation__.Notify(39922)
			service := &file.Services[j]
			for i := range service.Methods {
				__antithesis_instrumentation__.Notify(39923)
				m := &service.Methods[i]
				annotateStatus(&m.Description, &m.SupportStatus, "endpoint")
				if len(m.Options.GoogleAPIHTTP.Rules) > 0 {
					__antithesis_instrumentation__.Notify(39924)

					m.Options.GoogleAPIHTTP.Rules = m.Options.GoogleAPIHTTP.Rules[len(m.Options.GoogleAPIHTTP.Rules)-1:]
				} else {
					__antithesis_instrumentation__.Notify(39925)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(39890)

	messages := make(map[string]*protoMessage)
	methods := make(map[string]*protoMethod)
	for f := range data.Files {
		__antithesis_instrumentation__.Notify(39926)
		file := &data.Files[f]
		for i := range file.Messages {
			__antithesis_instrumentation__.Notify(39928)
			messages[file.Messages[i].FullName] = &file.Messages[i]
		}
		__antithesis_instrumentation__.Notify(39927)

		for j := range file.Services {
			__antithesis_instrumentation__.Notify(39929)
			service := &file.Services[j]
			for i := range service.Methods {
				__antithesis_instrumentation__.Notify(39930)
				methods[service.Methods[i].Name] = &service.Methods[i]
			}
		}
	}
	__antithesis_instrumentation__.Notify(39891)

	extraMessages := func(name string) []string {
		__antithesis_instrumentation__.Notify(39931)
		seen := make(map[string]bool)
		var extraFn func(name string) []string
		extraFn = func(name string) []string {
			__antithesis_instrumentation__.Notify(39933)
			if seen[name] {
				__antithesis_instrumentation__.Notify(39937)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(39938)
			}
			__antithesis_instrumentation__.Notify(39934)
			seen[name] = true
			msg, ok := messages[name]
			if !ok {
				__antithesis_instrumentation__.Notify(39939)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(39940)
			}
			__antithesis_instrumentation__.Notify(39935)
			var other []string
			for _, field := range msg.Fields {
				__antithesis_instrumentation__.Notify(39941)
				if innerMsg, ok := messages[field.FullType]; ok {
					__antithesis_instrumentation__.Notify(39942)
					other = append(other, field.FullType)
					other = append(other, extraFn(innerMsg.FullName)...)
				} else {
					__antithesis_instrumentation__.Notify(39943)
				}
			}
			__antithesis_instrumentation__.Notify(39936)
			return other
		}
		__antithesis_instrumentation__.Notify(39932)
		return extraFn(name)
	}
	__antithesis_instrumentation__.Notify(39892)

	tmplFuncs := template.FuncMap{

		"tableCell": func(s string) string {
			__antithesis_instrumentation__.Notify(39944)
			s = strings.TrimSpace(s)
			if s == "" {
				__antithesis_instrumentation__.Notify(39946)
				return ""
			} else {
				__antithesis_instrumentation__.Notify(39947)
			}
			__antithesis_instrumentation__.Notify(39945)
			s = strings.ReplaceAll(s, "\r", "")

			s = strings.ReplaceAll(s, "\n\n", "<br><br>")

			s = strings.ReplaceAll(s, "\n", " ")
			return s
		},
		"getMessage": func(name string) *protoMessage {
			__antithesis_instrumentation__.Notify(39948)
			return messages[name]
		},
		"extraMessages": extraMessages,
	}
	__antithesis_instrumentation__.Notify(39893)
	tmplFull := template.Must(template.New("full").Funcs(tmplFuncs).Parse(fullTemplate))
	tmplMessages := template.Must(template.New("single").Funcs(tmplFuncs).Parse(messagesTemplate))

	if err := execHTTPTmpl(tmplFull, &data, filepath.Join(outPath, "full.md")); err != nil {
		__antithesis_instrumentation__.Notify(39949)
		return fmt.Errorf("execHTTPTmpl: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(39950)
	}
	__antithesis_instrumentation__.Notify(39894)

	for _, methodName := range singleMethods {
		__antithesis_instrumentation__.Notify(39951)
		method, ok := methods[methodName]
		if !ok {
			__antithesis_instrumentation__.Notify(39953)
			return fmt.Errorf("single method not found: %s", methodName)
		} else {
			__antithesis_instrumentation__.Notify(39954)
		}
		__antithesis_instrumentation__.Notify(39952)
		for name, messages := range map[string][]string{
			"request":  {method.RequestFullType},
			"response": {method.ResponseFullType},
			"other":    append(extraMessages(method.RequestFullType), extraMessages(method.ResponseFullType)...),
		} {
			__antithesis_instrumentation__.Notify(39955)
			path := filepath.Join(outPath, fmt.Sprintf("%s-%s.md", strings.ToLower(methodName), name))
			if err := execHTTPTmpl(tmplMessages, messages, path); err != nil {
				__antithesis_instrumentation__.Notify(39956)
				return err
			} else {
				__antithesis_instrumentation__.Notify(39957)
			}
		}
	}
	__antithesis_instrumentation__.Notify(39895)
	return nil
}

func annotateStatus(desc *string, status *string, kind string) {
	__antithesis_instrumentation__.Notify(39958)
	if !strings.Contains(*desc, "API: PUBLIC") {
		__antithesis_instrumentation__.Notify(39962)
		*status = `[reserved](#support-status)`
	} else {
		__antithesis_instrumentation__.Notify(39963)
	}
	__antithesis_instrumentation__.Notify(39959)
	if strings.Contains(*desc, "API: PUBLIC ALPHA") {
		__antithesis_instrumentation__.Notify(39964)
		*status = `[alpha](#support-status)`
	} else {
		__antithesis_instrumentation__.Notify(39965)
	}
	__antithesis_instrumentation__.Notify(39960)
	if *status == "" {
		__antithesis_instrumentation__.Notify(39966)
		*status = `[public](#support-status)`
	} else {
		__antithesis_instrumentation__.Notify(39967)
	}
	__antithesis_instrumentation__.Notify(39961)
	*desc = strings.Replace(*desc, "API: PUBLIC ALPHA", "", 1)
	*desc = strings.Replace(*desc, "API: PUBLIC", "", 1)
	*desc = strings.TrimSpace(*desc)
}

func execHTTPTmpl(tmpl *template.Template, data interface{}, path string) error {
	__antithesis_instrumentation__.Notify(39968)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		__antithesis_instrumentation__.Notify(39970)
		return err
	} else {
		__antithesis_instrumentation__.Notify(39971)
	}
	__antithesis_instrumentation__.Notify(39969)
	return ioutil.WriteFile(path, buf.Bytes(), 0666)
}

type protoData struct {
	Files []struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		Package       string `json:"package"`
		HasEnums      bool   `json:"hasEnums"`
		HasExtensions bool   `json:"hasExtensions"`
		HasMessages   bool   `json:"hasMessages"`
		HasServices   bool   `json:"hasServices"`
		Enums         []struct {
			Name        string `json:"name"`
			LongName    string `json:"longName"`
			FullName    string `json:"fullName"`
			Description string `json:"description"`
			Values      []struct {
				Name        string `json:"name"`
				Number      string `json:"number"`
				Description string `json:"description"`
			} `json:"values"`
		} `json:"enums"`
		Extensions []interface{}  `json:"extensions"`
		Messages   []protoMessage `json:"messages"`
		Services   []struct {
			Name        string        `json:"name"`
			LongName    string        `json:"longName"`
			FullName    string        `json:"fullName"`
			Description string        `json:"description"`
			Methods     []protoMethod `json:"methods"`
		} `json:"services"`
	} `json:"files"`
	ScalarValueTypes []struct {
		ProtoType  string `json:"protoType"`
		Notes      string `json:"notes"`
		CppType    string `json:"cppType"`
		CsType     string `json:"csType"`
		GoType     string `json:"goType"`
		JavaType   string `json:"javaType"`
		PhpType    string `json:"phpType"`
		PythonType string `json:"pythonType"`
		RubyType   string `json:"rubyType"`
	} `json:"scalarValueTypes"`
}

type protoMethod struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	SupportStatus     string `json:"supportStatus"`
	RequestType       string `json:"requestType"`
	RequestLongType   string `json:"requestLongType"`
	RequestFullType   string `json:"requestFullType"`
	RequestStreaming  bool   `json:"requestStreaming"`
	ResponseType      string `json:"responseType"`
	ResponseLongType  string `json:"responseLongType"`
	ResponseFullType  string `json:"responseFullType"`
	ResponseStreaming bool   `json:"responseStreaming"`
	Options           struct {
		GoogleAPIHTTP struct {
			Rules []struct {
				Method  string `json:"method"`
				Pattern string `json:"pattern"`
			} `json:"rules"`
		} `json:"google.api.http"`
	} `json:"options"`
}

type protoMessage struct {
	Name          string        `json:"name"`
	LongName      string        `json:"longName"`
	FullName      string        `json:"fullName"`
	Description   string        `json:"description"`
	SupportStatus string        `json:"supportStatus"`
	HasExtensions bool          `json:"hasExtensions"`
	HasFields     bool          `json:"hasFields"`
	Extensions    []interface{} `json:"extensions"`
	Fields        []struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		SupportStatus string `json:"supportStatus"`
		Label         string `json:"label"`
		Type          string `json:"type"`
		LongType      string `json:"longType"`
		FullType      string `json:"fullType"`
		Ismap         bool   `json:"ismap"`
		DefaultValue  string `json:"defaultValue"`
	} `json:"fields"`
}

const tmplJSON = `{{toPrettyJson .}}`

const fullTemplate = `
{{- define "FIELDS" -}}
{{with getMessage .}}
{{$message := .}}

{{with .Description}}{{.}}{{end}}

{{with .Fields}}
| Field | Type | Label | Description | Support status |
| ----- | ---- | ----- | ----------- | -------------- |
{{- range .}}
| {{.Name}} | [{{.LongType}}](#{{$message.FullName}}-{{.FullType}}) | {{.Label}} | {{.Description | tableCell}}{{if .DefaultValue}} Default: {{.DefaultValue}}{{end}} | {{.SupportStatus | tableCell}} |
{{- end}} {{- /* range */}}
{{end}} {{- /* with .Fields */}}

{{/* document extra messages */}}
{{range extraMessages .FullName}}
{{with getMessage .}}
{{if .Fields}}
<a name="{{$message.FullName}}-{{.FullName}}"></a>
#### {{.LongName}}

{{with .Description}}{{.}}{{end}}

| Field | Type | Label | Description | Support status |
| ----- | ---- | ----- | ----------- | -------------- |
{{- range .Fields}}
| {{.Name}} | [{{.LongType}}](#{{$message.FullName}}-{{.FullType}}) | {{.Label}} | {{.Description | tableCell}}{{if .DefaultValue}} Default: {{.DefaultValue}}{{end}} | {{.SupportStatus | tableCell}} |
{{- end}} {{- /* range */}}
{{end}} {{- /* if .Fields */}}
{{end}} {{- /* with getMessage */}}
{{end}} {{- /* range */}}

{{end}} {{- /* with getMessage */}}
{{- end}} {{- /* template */}}

{{- range .Files}}

{{- range .Services}}

{{- range .Methods -}}
## {{.Name}}

{{range .Options.GoogleAPIHTTP.Rules -}}
` + "`{{.Method}} {{.Pattern}}` " + `
{{- end}}

{{with .Description}}{{.}}{{end}}

{{with .SupportStatus}}Support status: {{.}}{{end}}

#### Request Parameters

{{template "FIELDS" .RequestFullType}}

#### Response Parameters

{{template "FIELDS" .ResponseFullType}}

{{end}} {{- /* methods */}}

{{- end -}} {{- /* services */ -}}
{{- end -}} {{- /* files */ -}}
`

var messagesTemplate = `
{{- range .}}
{{with getMessage .}}
<a name="{{.FullName}}"></a>
#### {{.LongName}}

{{with .Description}}{{.}}{{end}}

{{with .SupportStatus}}Support status: {{.}}{{end}}

{{with .Fields}}
| Field | Type | Label | Description | Support status |
| ----- | ---- | ----- | ----------- | -------------- |
{{- range .}}
| {{.Name}} | [{{.LongType}}](#{{.FullType}}) | {{.Label}} | {{.Description|tableCell}} | {{.SupportStatus | tableCell}} |
{{- end}} {{- /* range */}}
{{end}}{{- /* with .Fields */}}
{{end}}{{- /* with getMessage */}}
{{end}}{{- /* range */ -}}
`
