package scgraphviz

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scgraph"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scstage"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/emicklei/dot"
	"github.com/gogo/protobuf/jsonpb"
)

func StagesURL(cs scpb.CurrentState, g *scgraph.Graph, stages []scstage.Stage) (string, error) {
	__antithesis_instrumentation__.Notify(594372)
	gv, err := DrawStages(cs, g, stages)
	if err != nil {
		__antithesis_instrumentation__.Notify(594374)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(594375)
	}
	__antithesis_instrumentation__.Notify(594373)
	return buildURL(gv)
}

func DependenciesURL(cs scpb.CurrentState, g *scgraph.Graph) (string, error) {
	__antithesis_instrumentation__.Notify(594376)
	gv, err := DrawDependencies(cs, g)
	if err != nil {
		__antithesis_instrumentation__.Notify(594378)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(594379)
	}
	__antithesis_instrumentation__.Notify(594377)
	return buildURL(gv)
}

func buildURL(gv string) (string, error) {
	__antithesis_instrumentation__.Notify(594380)
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := io.WriteString(w, gv); err != nil {
		__antithesis_instrumentation__.Notify(594383)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(594384)
	}
	__antithesis_instrumentation__.Notify(594381)
	if err := w.Close(); err != nil {
		__antithesis_instrumentation__.Notify(594385)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(594386)
	}
	__antithesis_instrumentation__.Notify(594382)
	return (&url.URL{
		Scheme:   "https",
		Host:     "cockroachdb.github.io",
		Path:     "scplan/viz.html",
		Fragment: base64.StdEncoding.EncodeToString(buf.Bytes()),
	}).String(), nil
}

func DrawStages(cs scpb.CurrentState, g *scgraph.Graph, stages []scstage.Stage) (string, error) {
	__antithesis_instrumentation__.Notify(594387)
	if len(stages) == 0 {
		__antithesis_instrumentation__.Notify(594390)
		return "", errors.Errorf("missing stages in plan")
	} else {
		__antithesis_instrumentation__.Notify(594391)
	}
	__antithesis_instrumentation__.Notify(594388)
	gv, err := drawStages(cs, g, stages)
	if err != nil {
		__antithesis_instrumentation__.Notify(594392)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(594393)
	}
	__antithesis_instrumentation__.Notify(594389)
	return gv.String(), nil
}

func DrawDependencies(cs scpb.CurrentState, g *scgraph.Graph) (string, error) {
	__antithesis_instrumentation__.Notify(594394)
	if g == nil {
		__antithesis_instrumentation__.Notify(594397)
		return "", errors.Errorf("missing graph in plan")
	} else {
		__antithesis_instrumentation__.Notify(594398)
	}
	__antithesis_instrumentation__.Notify(594395)
	gv, err := drawDeps(cs, g)
	if err != nil {
		__antithesis_instrumentation__.Notify(594399)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(594400)
	}
	__antithesis_instrumentation__.Notify(594396)
	return gv.String(), nil
}

func drawStages(
	cs scpb.CurrentState, g *scgraph.Graph, stages []scstage.Stage,
) (*dot.Graph, error) {
	__antithesis_instrumentation__.Notify(594401)
	dg := dot.NewGraph()
	stagesSubgraph := dg.Subgraph("stages", dot.ClusterOption{})
	targetsSubgraph := stagesSubgraph.Subgraph("targets", dot.ClusterOption{})
	statementsSubgraph := stagesSubgraph.Subgraph("statements", dot.ClusterOption{})
	targetNodes := make([]dot.Node, len(cs.TargetState.Targets))

	for idx, stmt := range cs.TargetState.Statements {
		__antithesis_instrumentation__.Notify(594406)
		stmtNode := statementsSubgraph.Node(itoa(idx, len(cs.TargetState.Statements)))
		stmtNode.Attr("label", htmlLabel(stmt))
		stmtNode.Attr("fontsize", "9")
		stmtNode.Attr("shape", "none")
	}
	__antithesis_instrumentation__.Notify(594402)
	for idx, t := range cs.TargetState.Targets {
		__antithesis_instrumentation__.Notify(594407)
		tn := targetsSubgraph.Node(itoa(idx, len(cs.TargetState.Targets)))
		tn.Attr("label", htmlLabel(t.Element()))
		tn.Attr("fontsize", "9")
		tn.Attr("shape", "none")
		targetNodes[idx] = tn
	}
	__antithesis_instrumentation__.Notify(594403)

	curNodes := make([]dot.Node, len(cs.Current))
	cur := cs.Current
	curDummy := targetsSubgraph.Node("dummy")
	curDummy.Attr("shape", "point")
	curDummy.Attr("style", "invis")
	for i, status := range cs.Current {
		__antithesis_instrumentation__.Notify(594408)
		label := targetStatusID(i, status)
		tsn := stagesSubgraph.Node(fmt.Sprintf("initial %d", i))
		tsn.Attr("label", label)
		tn := targetNodes[i]
		e := tn.Edge(tsn)
		e.Dashed()
		e.Label(fmt.Sprintf("to %s", cs.TargetState.Targets[i].TargetStatus.String()))
		curNodes[i] = tsn
	}
	__antithesis_instrumentation__.Notify(594404)
	for _, st := range stages {
		__antithesis_instrumentation__.Notify(594409)
		stage := st.String()
		sg := stagesSubgraph.Subgraph(stage, dot.ClusterOption{})
		next := st.After
		nextNodes := make([]dot.Node, len(curNodes))
		m := make(map[scpb.Element][]scop.Op, len(curNodes))
		for _, op := range st.EdgeOps {
			__antithesis_instrumentation__.Notify(594413)
			if oe := g.GetOpEdgeFromOp(op); oe != nil {
				__antithesis_instrumentation__.Notify(594414)
				e := oe.To().Element()
				m[e] = append(m[e], op)
			} else {
				__antithesis_instrumentation__.Notify(594415)
			}
		}
		__antithesis_instrumentation__.Notify(594410)

		for i, status := range next {
			__antithesis_instrumentation__.Notify(594416)
			cst := sg.Node(fmt.Sprintf("%s: %d", stage, i))
			cst.Attr("label", targetStatusID(i, status))
			ge := curNodes[i].Edge(cst)
			if status != cur[i] {
				__antithesis_instrumentation__.Notify(594418)
				if ops := m[cs.TargetState.Targets[i].Element()]; len(ops) > 0 {
					__antithesis_instrumentation__.Notify(594419)
					ge.Attr("label", htmlLabel(ops))
					ge.Attr("fontsize", "9")
				} else {
					__antithesis_instrumentation__.Notify(594420)
				}
			} else {
				__antithesis_instrumentation__.Notify(594421)
				ge.Dotted()
			}
			__antithesis_instrumentation__.Notify(594417)
			nextNodes[i] = cst
		}
		__antithesis_instrumentation__.Notify(594411)
		nextDummy := sg.Node(fmt.Sprintf("%s: dummy", stage))
		nextDummy.Attr("shape", "point")
		nextDummy.Attr("style", "invis")
		if len(st.ExtraOps) > 0 {
			__antithesis_instrumentation__.Notify(594422)
			ge := curDummy.Edge(nextDummy)
			ge.Attr("label", htmlLabel(st.ExtraOps))
			ge.Attr("fontsize", "9")
		} else {
			__antithesis_instrumentation__.Notify(594423)
		}
		__antithesis_instrumentation__.Notify(594412)
		cur, curNodes, curDummy = next, nextNodes, nextDummy
	}
	__antithesis_instrumentation__.Notify(594405)

	return dg, nil
}

func drawDeps(cs scpb.CurrentState, g *scgraph.Graph) (*dot.Graph, error) {
	__antithesis_instrumentation__.Notify(594424)
	dg := dot.NewGraph()

	depsSubgraph := dg.Subgraph("deps", dot.ClusterOption{})
	targetsSubgraph := depsSubgraph.Subgraph("targets", dot.ClusterOption{})
	statementsSubgraph := depsSubgraph.Subgraph("statements", dot.ClusterOption{})
	targetNodes := make([]dot.Node, len(cs.Current))
	targetIdxMap := make(map[*scpb.Target]int)

	for idx, stmt := range cs.TargetState.Statements {
		__antithesis_instrumentation__.Notify(594430)
		stmtNode := statementsSubgraph.Node(itoa(idx, len(cs.TargetState.Statements)))
		stmtNode.Attr("label", htmlLabel(stmt))
		stmtNode.Attr("fontsize", "9")
		stmtNode.Attr("shape", "none")
	}
	__antithesis_instrumentation__.Notify(594425)
	targetStatusNodes := make([]map[scpb.Status]dot.Node, len(cs.Current))
	for idx, status := range cs.Current {
		__antithesis_instrumentation__.Notify(594431)
		t := &cs.TargetState.Targets[idx]
		tn := targetsSubgraph.Node(itoa(idx, len(cs.Current)))
		tn.Attr("label", htmlLabel(t.Element()))
		tn.Attr("fontsize", "9")
		tn.Attr("shape", "none")
		targetNodes[idx] = tn
		targetIdxMap[t] = idx
		targetStatusNodes[idx] = map[scpb.Status]dot.Node{status: tn}
	}
	__antithesis_instrumentation__.Notify(594426)
	_ = g.ForEachNode(func(n *screl.Node) error {
		__antithesis_instrumentation__.Notify(594432)
		tn := depsSubgraph.Node(targetStatusID(targetIdxMap[n.Target], n.CurrentStatus))
		targetStatusNodes[targetIdxMap[n.Target]][n.CurrentStatus] = tn
		return nil
	})
	__antithesis_instrumentation__.Notify(594427)
	for idx, status := range cs.Current {
		__antithesis_instrumentation__.Notify(594433)
		nn := targetStatusNodes[idx][status]
		tn := targetNodes[idx]
		e := tn.Edge(nn)
		e.Label(fmt.Sprintf("to %s", cs.TargetState.Targets[idx].TargetStatus.String()))
		e.Dashed()
	}
	__antithesis_instrumentation__.Notify(594428)

	_ = g.ForEachEdge(func(e scgraph.Edge) error {
		__antithesis_instrumentation__.Notify(594434)
		from := targetStatusNodes[targetIdxMap[e.From().Target]][e.From().CurrentStatus]
		to := targetStatusNodes[targetIdxMap[e.To().Target]][e.To().CurrentStatus]
		ge := from.Edge(to)
		switch e := e.(type) {
		case *scgraph.OpEdge:
			__antithesis_instrumentation__.Notify(594436)
			ge.Attr("label", htmlLabel(e.Op()))
			ge.Attr("fontsize", "9")
		case *scgraph.DepEdge:
			__antithesis_instrumentation__.Notify(594437)
			ge.Attr("color", "red")
			ge.Attr("label", e.Name())
			if e.Kind() == scgraph.SameStagePrecedence {
				__antithesis_instrumentation__.Notify(594438)
				ge.Attr("arrowhead", "diamond")
			} else {
				__antithesis_instrumentation__.Notify(594439)
			}
		}
		__antithesis_instrumentation__.Notify(594435)
		return nil
	})
	__antithesis_instrumentation__.Notify(594429)
	return dg, nil
}

func targetStatusID(targetID int, status scpb.Status) string {
	__antithesis_instrumentation__.Notify(594440)
	return fmt.Sprintf("%d:%s", targetID, status)
}

func htmlLabel(o interface{}) dot.HTML {
	__antithesis_instrumentation__.Notify(594441)
	var buf strings.Builder
	if err := objectTemplate.Execute(&buf, o); err != nil {
		__antithesis_instrumentation__.Notify(594443)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(594444)
	}
	__antithesis_instrumentation__.Notify(594442)
	return dot.HTML(buf.String())
}

func itoa(i, ub int) string {
	__antithesis_instrumentation__.Notify(594445)
	return fmt.Sprintf(fmt.Sprintf("%%0%dd", len(strconv.Itoa(ub))), i)
}

func ToMap(v interface{}) (interface{}, error) {
	__antithesis_instrumentation__.Notify(594446)
	if v == nil {
		__antithesis_instrumentation__.Notify(594451)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(594452)
	}
	__antithesis_instrumentation__.Notify(594447)
	if msg, ok := v.(protoutil.Message); ok {
		__antithesis_instrumentation__.Notify(594453)
		var buf bytes.Buffer
		jsonEncoder := jsonpb.Marshaler{EmitDefaults: false}
		if err := jsonEncoder.Marshal(&buf, msg); err != nil {
			__antithesis_instrumentation__.Notify(594456)
			return nil, errors.Wrapf(err, "%T %v", v, v)
		} else {
			__antithesis_instrumentation__.Notify(594457)
		}
		__antithesis_instrumentation__.Notify(594454)
		var m map[string]interface{}
		if err := json.NewDecoder(&buf).Decode(&m); err != nil {
			__antithesis_instrumentation__.Notify(594458)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(594459)
		}
		__antithesis_instrumentation__.Notify(594455)
		return m, nil
	} else {
		__antithesis_instrumentation__.Notify(594460)
	}
	__antithesis_instrumentation__.Notify(594448)
	vv := reflect.ValueOf(v)
	vt := vv.Type()
	switch vt.Kind() {
	case reflect.Struct:
		__antithesis_instrumentation__.Notify(594461)
	case reflect.Ptr:
		__antithesis_instrumentation__.Notify(594462)
		if vt.Elem().Kind() != reflect.Struct {
			__antithesis_instrumentation__.Notify(594465)
			return v, nil
		} else {
			__antithesis_instrumentation__.Notify(594466)
		}
		__antithesis_instrumentation__.Notify(594463)
		vv = vv.Elem()
		vt = vt.Elem()
	default:
		__antithesis_instrumentation__.Notify(594464)
		return v, nil
	}
	__antithesis_instrumentation__.Notify(594449)

	m := make(map[string]interface{}, vt.NumField())
	for i := 0; i < vt.NumField(); i++ {
		__antithesis_instrumentation__.Notify(594467)
		vvf := vv.Field(i)
		if !vvf.CanInterface() || func() bool {
			__antithesis_instrumentation__.Notify(594469)
			return vvf.IsZero() == true
		}() == true {
			__antithesis_instrumentation__.Notify(594470)
			continue
		} else {
			__antithesis_instrumentation__.Notify(594471)
		}
		__antithesis_instrumentation__.Notify(594468)
		var err error
		if m[vt.Field(i).Name], err = ToMap(vvf.Interface()); err != nil {
			__antithesis_instrumentation__.Notify(594472)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(594473)
		}
	}
	__antithesis_instrumentation__.Notify(594450)
	return m, nil
}

var objectTemplate = template.Must(template.New("obj").Funcs(template.FuncMap{
	"typeOf": func(v interface{}) string {
		__antithesis_instrumentation__.Notify(594474)
		return fmt.Sprintf("%T", v)
	},
	"isMap": func(v interface{}) bool {
		__antithesis_instrumentation__.Notify(594475)
		_, ok := v.(map[string]interface{})
		return ok
	},
	"isSlice": func(v interface{}) bool {
		__antithesis_instrumentation__.Notify(594476)
		vv := reflect.ValueOf(v)
		if !vv.IsValid() {
			__antithesis_instrumentation__.Notify(594478)
			return false
		} else {
			__antithesis_instrumentation__.Notify(594479)
		}
		__antithesis_instrumentation__.Notify(594477)
		return vv.Kind() == reflect.Slice
	},
	"emptyMap": func(v interface{}) bool {
		__antithesis_instrumentation__.Notify(594480)
		m, ok := v.(map[string]interface{})
		return ok && func() bool {
			__antithesis_instrumentation__.Notify(594481)
			return len(m) == 0 == true
		}() == true
	},
	"emptySlice": func(v interface{}) bool {
		__antithesis_instrumentation__.Notify(594482)
		m, ok := v.([]interface{})
		return ok && func() bool {
			__antithesis_instrumentation__.Notify(594483)
			return len(m) == 0 == true
		}() == true
	},
	"isStruct": func(v interface{}) bool {
		__antithesis_instrumentation__.Notify(594484)
		return reflect.Indirect(reflect.ValueOf(v)).Kind() == reflect.Struct
	},
	"toMap": ToMap,
}).Parse(`
{{- define "key" -}}
<td>
{{- . -}}
</td>
{{- end -}}

{{- define "val" -}}
<td>
{{- if (isMap .) -}}
{{- template "mapVal" . -}}
{{- else if (isSlice .) -}}
{{- template "sliceVal" . -}}
{{- else if (isStruct .) -}}
<td>
{{- typeOf . -}}
</td>
<td>
{{- template "mapVal" (toMap .) -}}
</td>
{{- else -}}
{{- . -}}
{{- end -}}
</td>
{{- end -}}

{{- define "sliceVal" -}}
{{- if not (emptySlice .) -}}
<table class="val"><tr>
{{- range . -}}
{{- template "val" . -}}
{{- end -}}
</tr></table>
{{- end -}}
{{- end -}}

{{- define "mapVal" -}}
<table class="table">
{{- range $k, $v :=  . -}}
{{- if not (emptyMap $v) -}}
<tr>
{{- template "key" $k -}}
{{- if (isStruct $v) -}}
<td>
{{- typeOf . -}}
</td>
<td>
{{- template "mapVal" (toMap $v) -}}
</td>
{{- else -}}
{{- template "val" $v -}}
{{- end -}}
</tr>
{{- end -}}
{{- end -}}
</table>
{{- end -}}

{{- define "header" -}}
<tr><td class="header">
{{- typeOf . -}}
</td></tr>
{{- end -}}

<table class="outer">
{{- template "header" . -}}
<tr><td>
{{- template "mapVal" (toMap .) -}}
</td></tr>
</table>
`))
