package reltest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type AttributeTestCases []AttributeTestCase

type AttributeTestCase struct {
	Entity   string
	Expected map[rel.Attr]interface{}
}

func (c AttributeTestCase) run(t *testing.T, s Suite) {
	__antithesis_instrumentation__.Notify(578961)
	got := make(map[rel.Attr]interface{})
	e := s.Registry.MustGetByName(t, c.Entity)
	require.NoError(t, s.Schema.IterateAttributes(e, func(
		attribute rel.Attr, value interface{},
	) error {
		__antithesis_instrumentation__.Notify(578963)
		got[attribute] = value
		return nil
	}))
	__antithesis_instrumentation__.Notify(578962)
	require.Equal(t, c.Expected, got)
	for a, v := range got {
		__antithesis_instrumentation__.Notify(578964)
		gotVal, err := s.Schema.GetAttribute(a, e)
		require.NoError(t, err)
		require.Equalf(t, v, gotVal, "%T[%v]", v, a.String())
	}
}
func (c AttributeTestCase) encode(t *testing.T, s Suite) *yaml.Node {
	__antithesis_instrumentation__.Notify(578965)
	type kv struct {
		k string
		v *yaml.Node
	}
	var kvs []kv
	for attr, value := range c.Expected {
		__antithesis_instrumentation__.Notify(578969)
		var v *yaml.Node
		if name, ok := s.Registry.GetName(value); ok {
			__antithesis_instrumentation__.Notify(578971)
			v = scalarYAML(name)
		} else {
			__antithesis_instrumentation__.Notify(578972)
			v = s.Registry.EncodeToYAML(t, value)
		}
		__antithesis_instrumentation__.Notify(578970)
		kvs = append(kvs, kv{attr.String(), v})
	}
	__antithesis_instrumentation__.Notify(578966)
	sort.Slice(kvs, func(i, j int) bool { __antithesis_instrumentation__.Notify(578973); return kvs[i].k < kvs[j].k })
	__antithesis_instrumentation__.Notify(578967)
	var n yaml.Node
	n.Kind = yaml.MappingNode
	n.Style = yaml.FlowStyle
	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(578974)
		n.Content = append(n.Content,
			scalarYAML(kv.k),
			kv.v,
		)
	}
	__antithesis_instrumentation__.Notify(578968)
	return &n
}
