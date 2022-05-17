package cyclegraphtest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel/reltest"
	"gopkg.in/yaml.v3"
)

type testAttr int8

var _ rel.Attr = testAttr(0)

const (
	s testAttr = iota
	s1
	s2
	c
	name
)

type struct1 struct {
	Name string

	S1 *struct1

	S2 *struct2

	C *container
}

type struct2 struct1

type container struct {
	S1 *struct1
	S2 *struct2
}

type message interface{ message() }

func (s *struct1) message() { __antithesis_instrumentation__.Notify(578534) }
func (s *struct2) message() { __antithesis_instrumentation__.Notify(578535) }

var schema = rel.MustSchema(
	"testschema",
	rel.AttrType(
		s, reflect.TypeOf((*message)(nil)).Elem(),
	),
	rel.EntityMapping(
		reflect.TypeOf((*struct1)(nil)),
		rel.EntityAttr(c, "C"),
		rel.EntityAttr(s1, "S1"),
		rel.EntityAttr(s2, "S2"),
		rel.EntityAttr(name, "Name"),
	),
	rel.EntityMapping(
		reflect.TypeOf((*struct2)(nil)),
		rel.EntityAttr(c, "C"),
		rel.EntityAttr(s1, "S1"),
		rel.EntityAttr(s2, "S2"),
		rel.EntityAttr(name, "Name"),
	),
	rel.EntityMapping(
		reflect.TypeOf((*container)(nil)),
		rel.EntityAttr(s, "S1", "S2"),
	),
)

func (s *struct1) String() string {
	__antithesis_instrumentation__.Notify(578536)
	return fmt.Sprintf("struct1(%s)", s.Name)
}

func (s *struct2) String() string {
	__antithesis_instrumentation__.Notify(578537)
	return fmt.Sprintf("struct2(%s)", s.Name)
}

func (c *container) String() string {
	__antithesis_instrumentation__.Notify(578538)
	var name string
	if c.S1 != nil {
		__antithesis_instrumentation__.Notify(578540)
		name = c.S1.Name
	} else {
		__antithesis_instrumentation__.Notify(578541)
		name = c.S2.Name
	}
	__antithesis_instrumentation__.Notify(578539)
	return fmt.Sprintf("container(%s)", name)
}

func (s *struct1) EncodeToYAML(t *testing.T, r *reltest.Registry) interface{} {
	__antithesis_instrumentation__.Notify(578542)
	yn := &yaml.Node{
		Kind:  yaml.MappingNode,
		Style: yaml.FlowStyle,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "name"},
			{Kind: yaml.ScalarNode, Value: s.Name},
		},
	}
	if s.S1 != nil {
		__antithesis_instrumentation__.Notify(578546)
		yn.Content = append(yn.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "s1"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: r.MustGetName(t, s.S1)},
		)
	} else {
		__antithesis_instrumentation__.Notify(578547)
	}
	__antithesis_instrumentation__.Notify(578543)
	if s.S2 != nil {
		__antithesis_instrumentation__.Notify(578548)
		yn.Content = append(yn.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "s2"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: r.MustGetName(t, s.S2)},
		)
	} else {
		__antithesis_instrumentation__.Notify(578549)
	}
	__antithesis_instrumentation__.Notify(578544)
	if s.C != nil {
		__antithesis_instrumentation__.Notify(578550)
		yn.Content = append(yn.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "c"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: r.MustGetName(t, s.C)},
		)
	} else {
		__antithesis_instrumentation__.Notify(578551)
	}
	__antithesis_instrumentation__.Notify(578545)
	return yn
}

func (s *struct2) EncodeToYAML(t *testing.T, r *reltest.Registry) interface{} {
	__antithesis_instrumentation__.Notify(578552)
	return (*struct1)(s).EncodeToYAML(t, r)
}

func (c *container) EncodeToYAML(t *testing.T, r *reltest.Registry) interface{} {
	__antithesis_instrumentation__.Notify(578553)
	yn := &yaml.Node{
		Kind:  yaml.MappingNode,
		Style: yaml.FlowStyle,
	}
	if c.S1 != nil {
		__antithesis_instrumentation__.Notify(578556)
		yn.Content = append(yn.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "s1"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: r.MustGetName(t, c.S1)},
		)
	} else {
		__antithesis_instrumentation__.Notify(578557)
	}
	__antithesis_instrumentation__.Notify(578554)
	if c.S2 != nil {
		__antithesis_instrumentation__.Notify(578558)
		yn.Content = append(yn.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "s2"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: r.MustGetName(t, c.S2)},
		)
	} else {
		__antithesis_instrumentation__.Notify(578559)
	}
	__antithesis_instrumentation__.Notify(578555)
	return yn
}
