package scplan

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gojson "encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scgraph"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scgraphviz"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scstage"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/treeprinter"
	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v2"
)

func (p Plan) DecorateErrorWithPlanDetails(err error) error {
	__antithesis_instrumentation__.Notify(594819)
	if err == nil {
		__antithesis_instrumentation__.Notify(594823)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(594824)
	}
	__antithesis_instrumentation__.Notify(594820)

	if len(p.Stages) > 0 {
		__antithesis_instrumentation__.Notify(594825)
		explain, explainErr := p.ExplainVerbose()
		if explainErr != nil {
			__antithesis_instrumentation__.Notify(594827)
			explainErr = errors.Wrapf(explainErr, "error when generating EXPLAIN plan")
			err = errors.CombineErrors(err, explainErr)
		} else {
			__antithesis_instrumentation__.Notify(594828)
			err = errors.WithDetailf(err, "%s", explain)
		}
		__antithesis_instrumentation__.Notify(594826)

		stagesURL, stagesErr := p.StagesURL()
		if stagesErr != nil {
			__antithesis_instrumentation__.Notify(594829)
			stagesErr = errors.Wrapf(stagesErr, "error when generating EXPLAIN graphviz URL")
			err = errors.CombineErrors(err, stagesErr)
		} else {
			__antithesis_instrumentation__.Notify(594830)
			err = errors.WithDetailf(err, "stages graphviz: %s", stagesURL)
		}
	} else {
		__antithesis_instrumentation__.Notify(594831)
	}
	__antithesis_instrumentation__.Notify(594821)

	if p.Graph != nil {
		__antithesis_instrumentation__.Notify(594832)
		dependenciesURL, dependenciesErr := p.DependenciesURL()
		if dependenciesErr != nil {
			__antithesis_instrumentation__.Notify(594833)
			dependenciesErr = errors.Wrapf(dependenciesErr, "error when generating dependencies graphviz URL")
			err = errors.CombineErrors(err, dependenciesErr)
		} else {
			__antithesis_instrumentation__.Notify(594834)
			err = errors.WithDetailf(err, "dependencies graphviz: %s", dependenciesURL)
		}
	} else {
		__antithesis_instrumentation__.Notify(594835)
	}
	__antithesis_instrumentation__.Notify(594822)

	return errors.WithAssertionFailure(err)
}

func (p Plan) DependenciesURL() (string, error) {
	__antithesis_instrumentation__.Notify(594836)
	return scgraphviz.DependenciesURL(p.CurrentState, p.Graph)
}

func (p Plan) StagesURL() (string, error) {
	__antithesis_instrumentation__.Notify(594837)
	return scgraphviz.StagesURL(p.CurrentState, p.Graph, p.Stages)
}

func (p Plan) ExplainViz() (stagesURL, depsURL string, err error) {
	__antithesis_instrumentation__.Notify(594838)
	stagesURL, err = p.StagesURL()
	if err != nil {
		__antithesis_instrumentation__.Notify(594841)
		return "", "", err
	} else {
		__antithesis_instrumentation__.Notify(594842)
	}
	__antithesis_instrumentation__.Notify(594839)
	depsURL, err = p.DependenciesURL()
	if err != nil {
		__antithesis_instrumentation__.Notify(594843)
		return "", "", err
	} else {
		__antithesis_instrumentation__.Notify(594844)
	}
	__antithesis_instrumentation__.Notify(594840)
	return stagesURL, depsURL, nil
}

func (p Plan) ExplainCompact() (string, error) {
	__antithesis_instrumentation__.Notify(594845)
	return p.explain(treeprinter.DefaultStyle)
}

func (p Plan) ExplainVerbose() (string, error) {
	__antithesis_instrumentation__.Notify(594846)
	return p.explain(treeprinter.BulletStyle)
}

func (p Plan) explain(style treeprinter.Style) (string, error) {
	__antithesis_instrumentation__.Notify(594847)

	tp := treeprinter.NewWithStyle(style)
	var sb strings.Builder
	{
		__antithesis_instrumentation__.Notify(594850)
		sb.WriteString("Schema change plan for ")
		if p.InRollback {
			__antithesis_instrumentation__.Notify(594852)
			sb.WriteString("rolling back ")
		} else {
			__antithesis_instrumentation__.Notify(594853)
		}
		__antithesis_instrumentation__.Notify(594851)
		for _, stmt := range p.Statements {
			__antithesis_instrumentation__.Notify(594854)
			sb.WriteString(strings.TrimSuffix(stmt.RedactedStatement, ";"))
			sb.WriteString("; ")
		}
	}
	__antithesis_instrumentation__.Notify(594848)
	root := tp.Child(sb.String())
	var pn treeprinter.Node
	for i, s := range p.Stages {
		__antithesis_instrumentation__.Notify(594855)

		if i == 0 || func() bool {
			__antithesis_instrumentation__.Notify(594858)
			return s.Phase != p.Stages[i-1].Phase == true
		}() == true {
			__antithesis_instrumentation__.Notify(594859)
			pn = root.Childf("%s", s.Phase)
		} else {
			__antithesis_instrumentation__.Notify(594860)
		}
		__antithesis_instrumentation__.Notify(594856)
		sn := pn.Childf("Stage %d of %d in %s", s.Ordinal, s.StagesInPhase, s.Phase)

		if err := p.explainTargets(s, sn, style); err != nil {
			__antithesis_instrumentation__.Notify(594861)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(594862)
		}
		__antithesis_instrumentation__.Notify(594857)

		if err := p.explainOps(s, sn, style); err != nil {
			__antithesis_instrumentation__.Notify(594863)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(594864)
		}
	}
	__antithesis_instrumentation__.Notify(594849)
	return tp.String(), nil
}

func (p Plan) explainTargets(s scstage.Stage, sn treeprinter.Node, style treeprinter.Style) error {
	__antithesis_instrumentation__.Notify(594865)
	var targetTypeMap util.FastIntMap
	depEdgeByElement := make(map[scpb.Element][]*scgraph.DepEdge)
	noOpByElement := make(map[scpb.Element][]*scgraph.OpEdge)
	var beforeMaxLen, afterMaxLen int

	for j, before := range s.Before {
		__antithesis_instrumentation__.Notify(594868)
		t := &p.TargetState.Targets[j]
		after := s.After[j]
		if before == after {
			__antithesis_instrumentation__.Notify(594873)
			continue
		} else {
			__antithesis_instrumentation__.Notify(594874)
		}
		__antithesis_instrumentation__.Notify(594869)

		if l := utf8.RuneCountInString(before.String()); l > beforeMaxLen {
			__antithesis_instrumentation__.Notify(594875)
			beforeMaxLen = l
		} else {
			__antithesis_instrumentation__.Notify(594876)
		}
		__antithesis_instrumentation__.Notify(594870)
		if l := utf8.RuneCountInString(after.String()); l > afterMaxLen {
			__antithesis_instrumentation__.Notify(594877)
			afterMaxLen = l
		} else {
			__antithesis_instrumentation__.Notify(594878)
		}
		__antithesis_instrumentation__.Notify(594871)

		k := int(scpb.AsTargetStatus(t.TargetStatus))
		numTransitions, found := targetTypeMap.Get(k)
		if !found {
			__antithesis_instrumentation__.Notify(594879)
			targetTypeMap.Set(k, 1)
		} else {
			__antithesis_instrumentation__.Notify(594880)
			targetTypeMap.Set(k, numTransitions+1)
		}
		__antithesis_instrumentation__.Notify(594872)

		if style == treeprinter.BulletStyle {
			__antithesis_instrumentation__.Notify(594881)
			n, nodeFound := p.Graph.GetNode(t, before)
			if !nodeFound {
				__antithesis_instrumentation__.Notify(594883)
				return errors.Errorf("could not find node [[%s, %s], %s] in graph",
					screl.ElementString(t.Element()), t.TargetStatus, before)
			} else {
				__antithesis_instrumentation__.Notify(594884)
			}
			__antithesis_instrumentation__.Notify(594882)
			for n.CurrentStatus != after {
				__antithesis_instrumentation__.Notify(594885)
				oe, edgeFound := p.Graph.GetOpEdgeFrom(n)
				if !edgeFound {
					__antithesis_instrumentation__.Notify(594888)
					return errors.Errorf("could not find op edge from %s in graph", screl.NodeString(n))
				} else {
					__antithesis_instrumentation__.Notify(594889)
				}
				__antithesis_instrumentation__.Notify(594886)
				n = oe.To()
				if p.Graph.IsNoOp(oe) {
					__antithesis_instrumentation__.Notify(594890)
					noOpByElement[t.Element()] = append(noOpByElement[t.Element()], oe)
				} else {
					__antithesis_instrumentation__.Notify(594891)
				}
				__antithesis_instrumentation__.Notify(594887)
				if err := p.Graph.ForEachDepEdgeTo(n, func(de *scgraph.DepEdge) error {
					__antithesis_instrumentation__.Notify(594892)
					depEdgeByElement[t.Element()] = append(depEdgeByElement[t.Element()], de)
					return nil
				}); err != nil {
					__antithesis_instrumentation__.Notify(594893)
					return err
				} else {
					__antithesis_instrumentation__.Notify(594894)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(594895)
		}
	}
	__antithesis_instrumentation__.Notify(594866)

	fmtCompactTransition := fmt.Sprintf("%%-%ds → %%-%ds %%s", beforeMaxLen, afterMaxLen)

	targetTypeMap.ForEach(func(key, numTransitions int) {
		__antithesis_instrumentation__.Notify(594896)
		ts := scpb.TargetStatus(key)
		plural := "s"
		if numTransitions == 1 {
			__antithesis_instrumentation__.Notify(594898)
			plural = ""
		} else {
			__antithesis_instrumentation__.Notify(594899)
		}
		__antithesis_instrumentation__.Notify(594897)
		tn := sn.Childf("%d element%s transitioning toward %s",
			numTransitions, plural, ts.Status())
		for j, before := range s.Before {
			__antithesis_instrumentation__.Notify(594900)
			t := &p.TargetState.Targets[j]
			after := s.After[j]
			if t.TargetStatus != ts.Status() || func() bool {
				__antithesis_instrumentation__.Notify(594904)
				return before == after == true
			}() == true {
				__antithesis_instrumentation__.Notify(594905)
				continue
			} else {
				__antithesis_instrumentation__.Notify(594906)
			}
			__antithesis_instrumentation__.Notify(594901)

			var en treeprinter.Node
			if style == treeprinter.BulletStyle {
				__antithesis_instrumentation__.Notify(594907)
				en = tn.Child(screl.ElementString(t.Element()))
				en.AddLine(fmt.Sprintf("%s → %s", before, after))
			} else {
				__antithesis_instrumentation__.Notify(594908)
				en = tn.Childf(fmtCompactTransition, before, after, screl.ElementString(t.Element()))
			}
			__antithesis_instrumentation__.Notify(594902)
			depEdges := depEdgeByElement[t.Element()]
			for _, de := range depEdges {
				__antithesis_instrumentation__.Notify(594909)
				rn := en.Childf("%s dependency from %s %s",
					de.Kind(), de.From().CurrentStatus, screl.ElementString(de.From().Element()))
				rn.AddLine(fmt.Sprintf("rule: %q", de.Name()))
			}
			__antithesis_instrumentation__.Notify(594903)
			noOpEdges := noOpByElement[t.Element()]
			for _, oe := range noOpEdges {
				__antithesis_instrumentation__.Notify(594910)
				noOpRules := p.Graph.NoOpRules(oe)
				if len(noOpRules) == 0 {
					__antithesis_instrumentation__.Notify(594912)
					continue
				} else {
					__antithesis_instrumentation__.Notify(594913)
				}
				__antithesis_instrumentation__.Notify(594911)
				nn := en.Childf("skip %s → %s operations",
					oe.From().CurrentStatus, oe.To().CurrentStatus)
				for _, rule := range noOpRules {
					__antithesis_instrumentation__.Notify(594914)
					nn.AddLine(fmt.Sprintf("rule: %q", rule))
				}
			}
		}
	})
	__antithesis_instrumentation__.Notify(594867)
	return nil
}

func (p Plan) explainOps(s scstage.Stage, sn treeprinter.Node, style treeprinter.Style) error {
	__antithesis_instrumentation__.Notify(594915)
	ops := s.Ops()
	if len(ops) == 0 {
		__antithesis_instrumentation__.Notify(594919)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(594920)
	}
	__antithesis_instrumentation__.Notify(594916)
	plural := "s"
	if len(ops) == 1 {
		__antithesis_instrumentation__.Notify(594921)
		plural = ""
	} else {
		__antithesis_instrumentation__.Notify(594922)
	}
	__antithesis_instrumentation__.Notify(594917)
	on := sn.Childf("%d %s operation%s", len(ops), strings.TrimSuffix(s.Type().String(), "Type"), plural)
	for _, op := range ops {
		__antithesis_instrumentation__.Notify(594923)
		if setJobStateOp, ok := op.(*scop.SetJobStateOnDescriptor); ok {
			__antithesis_instrumentation__.Notify(594925)
			clone := *setJobStateOp
			clone.State = scpb.DescriptorState{}
			op = &clone
		} else {
			__antithesis_instrumentation__.Notify(594926)
		}
		__antithesis_instrumentation__.Notify(594924)
		opName := strings.TrimPrefix(fmt.Sprintf("%T", op), "*scop.")
		if style == treeprinter.BulletStyle {
			__antithesis_instrumentation__.Notify(594927)
			n := on.Child(opName)
			opBody, err := explainOpBodyVerbose(op)
			if err != nil {
				__antithesis_instrumentation__.Notify(594929)
				return err
			} else {
				__antithesis_instrumentation__.Notify(594930)
			}
			__antithesis_instrumentation__.Notify(594928)
			for _, line := range strings.Split(opBody, "\n") {
				__antithesis_instrumentation__.Notify(594931)
				n.AddLine(line)
			}
		} else {
			__antithesis_instrumentation__.Notify(594932)
			opBody, err := explainOpBodyCompact(op)
			if err != nil {
				__antithesis_instrumentation__.Notify(594934)
				return err
			} else {
				__antithesis_instrumentation__.Notify(594935)
			}
			__antithesis_instrumentation__.Notify(594933)
			if len(opBody) == 0 {
				__antithesis_instrumentation__.Notify(594936)
				on.Child(opName)
			} else {
				__antithesis_instrumentation__.Notify(594937)
				on.Childf("%s %s", opName, opBody)
			}
		}
	}
	__antithesis_instrumentation__.Notify(594918)
	return nil
}

func explainOpBodyVerbose(op scop.Op) (string, error) {
	__antithesis_instrumentation__.Notify(594938)
	opMap, err := scgraphviz.ToMap(op)
	if err != nil {
		__antithesis_instrumentation__.Notify(594941)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(594942)
	}
	__antithesis_instrumentation__.Notify(594939)
	yml, err := yaml.Marshal(opMap)
	if err != nil {
		__antithesis_instrumentation__.Notify(594943)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(594944)
	}
	__antithesis_instrumentation__.Notify(594940)
	return strings.TrimSuffix(string(yml), "\n"), nil
}

func explainOpBodyCompact(op scop.Op) (string, error) {
	__antithesis_instrumentation__.Notify(594945)
	m := map[string]interface{}{}

	{
		__antithesis_instrumentation__.Notify(594948)
		opMap, err := scgraphviz.ToMap(op)
		if err != nil {
			__antithesis_instrumentation__.Notify(594951)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(594952)
		}
		__antithesis_instrumentation__.Notify(594949)
		jb, err := gojson.Marshal(opMap)
		if err != nil {
			__antithesis_instrumentation__.Notify(594953)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(594954)
		}
		__antithesis_instrumentation__.Notify(594950)
		if err := gojson.Unmarshal(jb, &m); err != nil {
			__antithesis_instrumentation__.Notify(594955)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(594956)
		}
	}
	__antithesis_instrumentation__.Notify(594946)

	tm := opMapTrim(m)
	if len(tm) == 0 {
		__antithesis_instrumentation__.Notify(594957)

		for k, v := range m {
			__antithesis_instrumentation__.Notify(594959)
			vm, ok := v.(map[string]interface{})
			if !ok {
				__antithesis_instrumentation__.Notify(594962)
				continue
			} else {
				__antithesis_instrumentation__.Notify(594963)
			}
			__antithesis_instrumentation__.Notify(594960)
			tvm := opMapTrim(vm)
			if len(tvm) == 0 {
				__antithesis_instrumentation__.Notify(594964)
				continue
			} else {
				__antithesis_instrumentation__.Notify(594965)
			}
			__antithesis_instrumentation__.Notify(594961)
			tm[k] = tvm
		}
		__antithesis_instrumentation__.Notify(594958)
		if len(tm) == 0 {
			__antithesis_instrumentation__.Notify(594966)
			return "", nil
		} else {
			__antithesis_instrumentation__.Notify(594967)
		}
	} else {
		__antithesis_instrumentation__.Notify(594968)
	}
	__antithesis_instrumentation__.Notify(594947)
	jb, err := gojson.Marshal(tm)
	return string(jb), err
}

func opMapTrim(in map[string]interface{}) map[string]interface{} {
	__antithesis_instrumentation__.Notify(594969)
	const jobIDKey = "JobID"
	out := make(map[string]interface{})
	for k, v := range in {
		__antithesis_instrumentation__.Notify(594971)
		vv := reflect.ValueOf(v)
		switch vv.Type().Kind() {
		case reflect.Bool:
			__antithesis_instrumentation__.Notify(594972)
			out[k] = vv.Bool()
		case reflect.Float64:
			__antithesis_instrumentation__.Notify(594973)
			if k != jobIDKey || func() bool {
				__antithesis_instrumentation__.Notify(594976)
				return vv.Float() != 1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(594977)
				out[k] = vv.Float()
			} else {
				__antithesis_instrumentation__.Notify(594978)
			}
		case reflect.String:
			__antithesis_instrumentation__.Notify(594974)
			if str := vv.String(); utf8.RuneCountInString(str) > 16 {
				__antithesis_instrumentation__.Notify(594979)
				out[k] = fmt.Sprintf("%.16s...", str)
			} else {
				__antithesis_instrumentation__.Notify(594980)
				out[k] = str
			}
		default:
			__antithesis_instrumentation__.Notify(594975)
		}
	}
	__antithesis_instrumentation__.Notify(594970)
	return out
}
