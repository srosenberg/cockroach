package entitynodetest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel/reltest"
	"gopkg.in/yaml.v3"
)

type entity struct {
	I8  int8
	PI8 *int8
	I16 int16
}

type node struct {
	Value       *entity
	Left, Right *node
}

func (n *node) EncodeToYAML(t *testing.T, r *reltest.Registry) interface{} {
	__antithesis_instrumentation__.Notify(578567)
	yn := yaml.Node{Kind: yaml.MappingNode, Style: yaml.FlowStyle}
	for _, f := range []struct {
		name  string
		field interface{}
		ok    bool
	}{
		{"value", n.Value, n.Value != nil},
		{"left", n.Left, n.Left != nil},
		{"right", n.Right, n.Right != nil},
	} {
		__antithesis_instrumentation__.Notify(578569)
		if !f.ok {
			__antithesis_instrumentation__.Notify(578571)
			continue
		} else {
			__antithesis_instrumentation__.Notify(578572)
		}
		__antithesis_instrumentation__.Notify(578570)
		yn.Content = append(yn.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: f.name},
			&yaml.Node{Kind: yaml.ScalarNode, Value: r.MustGetName(t, f.field)},
		)
	}
	__antithesis_instrumentation__.Notify(578568)
	return &yn
}

var _ reltest.RegistryYAMLEncoder = (*node)(nil)

type testAttr int8

var _ rel.Attr = testAttr(0)

const (
	i8 testAttr = iota
	pi8
	i16
	value
	left
	right
)

var schema = rel.MustSchema("testschema",
	rel.EntityMapping(reflect.TypeOf((*entity)(nil)),
		rel.EntityAttr(i8, "I8"),
		rel.EntityAttr(pi8, "PI8"),
		rel.EntityAttr(i16, "I16"),
	),
	rel.EntityMapping(reflect.TypeOf((*node)(nil)),
		rel.EntityAttr(value, "Value"),
		rel.EntityAttr(left, "Left"),
		rel.EntityAttr(right, "Right"),
	),
)
