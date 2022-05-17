package reltest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type ComparisonTests struct {
	Entities []string
	Tests    []ComparisonTest
}

type ComparisonTest struct {
	Attrs []rel.Attr
	Order [][]string
}

func (ct ComparisonTests) run(t *testing.T, s Suite) {
	__antithesis_instrumentation__.Notify(578975)
	for _, tc := range ct.Tests {
		__antithesis_instrumentation__.Notify(578976)
		strs := make([]string, len(tc.Attrs))
		for i, a := range tc.Attrs {
			__antithesis_instrumentation__.Notify(578978)
			strs[i] = a.String()
		}
		__antithesis_instrumentation__.Notify(578977)
		t.Run(strings.Join(strs, ","), func(t *testing.T) {
			__antithesis_instrumentation__.Notify(578979)
			tc.run(t, s, ct.Entities)
		})
	}
}

func (ct ComparisonTest) run(t *testing.T, s Suite, entityNames []string) {
	__antithesis_instrumentation__.Notify(578980)
	entities := entitiesForComparison{
		schema:    s.Schema,
		attrs:     ct.Attrs,
		names:     entityNames,
		entities:  make([]interface{}, len(entityNames)),
		nameOrder: make(map[string]int, len(entityNames)),
	}
	for i, name := range entityNames {
		__antithesis_instrumentation__.Notify(578983)
		entities.entities[i] = s.Registry.MustGetByName(t, name)
		entities.nameOrder[name] = i
	}
	__antithesis_instrumentation__.Notify(578981)
	sort.Sort(entities)
	groups := [][]string{
		{entities.names[0]},
	}
	for i := 1; i < len(entities.entities); i++ {
		__antithesis_instrumentation__.Notify(578984)
		if s.Schema.EqualOn(ct.Attrs, entities.entities[i-1], entities.entities[i]) {
			__antithesis_instrumentation__.Notify(578985)
			last := len(groups) - 1
			groups[last] = append(groups[last], entities.names[i])
		} else {
			__antithesis_instrumentation__.Notify(578986)
			groups = append(groups, []string{entities.names[i]})
		}
	}
	__antithesis_instrumentation__.Notify(578982)
	require.Equal(t, ct.Order, groups)
}

type entitiesForComparison struct {
	schema    *rel.Schema
	attrs     []rel.Attr
	names     []string
	entities  []interface{}
	nameOrder map[string]int
}

func (efc entitiesForComparison) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(578987)
	less, eq := efc.schema.CompareOn(efc.attrs, efc.entities[i], efc.entities[j])
	if !eq {
		__antithesis_instrumentation__.Notify(578989)
		return less
	} else {
		__antithesis_instrumentation__.Notify(578990)
	}
	__antithesis_instrumentation__.Notify(578988)
	return efc.nameOrder[efc.names[i]] < efc.nameOrder[efc.names[j]]
}

func (efc entitiesForComparison) Len() int {
	__antithesis_instrumentation__.Notify(578991)
	return len(efc.entities)
}
func (efc entitiesForComparison) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(578992)
	efc.entities[i], efc.entities[j] = efc.entities[j], efc.entities[i]
	efc.names[i], efc.names[j] = efc.names[j], efc.names[i]
}

func (ct ComparisonTests) encode(t *testing.T) *yaml.Node {
	__antithesis_instrumentation__.Notify(578993)
	var entities yaml.Node
	require.NoError(t, entities.Encode(ct.Entities))
	entities.Style = yaml.FlowStyle
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			scalarYAML("entities"),
			&entities,
			scalarYAML("tests"),
			ct.encodeTests(t),
		},
	}
}

func (ct ComparisonTests) encodeTests(t *testing.T) *yaml.Node {
	__antithesis_instrumentation__.Notify(578994)
	var tests yaml.Node
	tests.Kind = yaml.SequenceNode
	for _, subtest := range ct.Tests {
		__antithesis_instrumentation__.Notify(578996)
		tests.Content = append(tests.Content, subtest.encode(t))
	}
	__antithesis_instrumentation__.Notify(578995)
	return &tests
}

func (ct ComparisonTest) encode(t *testing.T) *yaml.Node {
	__antithesis_instrumentation__.Notify(578997)
	var exp yaml.Node
	require.NoError(t, exp.Encode(ct.Order))
	exp.Style = yaml.FlowStyle
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			scalarYAML("attrs"),
			encodeAttrs(ct.Attrs),
			scalarYAML("order"),
			&exp,
		},
	}
}
