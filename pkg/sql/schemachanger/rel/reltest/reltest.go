// Package reltest provides tools for testing the rel package.
package reltest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"flag"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var rewrite bool

func init() {
	flag.BoolVar(&rewrite, "rewrite", false, "set to rewrite the test output")
}

type Suite struct {
	Name string

	Schema *rel.Schema

	Registry *Registry

	AttributeTests []AttributeTestCase

	DatabaseTests []DatabaseTest

	ComparisonTests []ComparisonTests
}

func (s Suite) Run(t *testing.T) {
	__antithesis_instrumentation__.Notify(579083)
	for _, tc := range s.DatabaseTests {
		__antithesis_instrumentation__.Notify(579087)
		t.Run("database", func(t *testing.T) {
			__antithesis_instrumentation__.Notify(579088)
			tc.run(t, s)
		})
	}
	__antithesis_instrumentation__.Notify(579084)
	t.Run("attributes", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(579089)
		for _, tc := range s.AttributeTests {
			__antithesis_instrumentation__.Notify(579090)
			t.Run(tc.Entity, func(t *testing.T) {
				__antithesis_instrumentation__.Notify(579091)
				tc.run(t, s)
			})
		}
	})
	__antithesis_instrumentation__.Notify(579085)
	t.Run("comparison", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(579092)
		for _, tc := range s.ComparisonTests {
			__antithesis_instrumentation__.Notify(579093)
			t.Run(strings.Join(tc.Entities, ","), func(t *testing.T) {
				__antithesis_instrumentation__.Notify(579094)
				tc.run(t, s)
			})
		}
	})
	__antithesis_instrumentation__.Notify(579086)
	t.Run("yaml", func(t *testing.T) {
		__antithesis_instrumentation__.Notify(579095)
		s.writeYAML(t)
	})
}

func (s Suite) writeYAML(t *testing.T) {
	__antithesis_instrumentation__.Notify(579096)
	out, err := yaml.Marshal(s.toYAML(t))
	require.NoError(t, err)
	tdp := testutils.TestDataPath(t, s.Name)
	if rewrite {
		__antithesis_instrumentation__.Notify(579097)
		require.NoError(t, ioutil.WriteFile(tdp, out, 0777))
	} else {
		__antithesis_instrumentation__.Notify(579098)
		exp, err := ioutil.ReadFile(tdp)
		require.NoError(t, err)
		require.Equal(t, exp, out)
	}
}

func (s Suite) toYAML(t *testing.T) *yaml.Node {
	__antithesis_instrumentation__.Notify(579099)
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			scalarYAML("name"),
			scalarYAML(s.Name),
			scalarYAML("data"),
			s.encodeData(t),
			scalarYAML("attributes"),
			s.encodeAttributes(t),
			scalarYAML("queries"),
			s.encodeQueries(t),
			scalarYAML("comparisons"),
			s.encodeComparisons(t),
		},
	}
}

func (s Suite) encodeData(t *testing.T) *yaml.Node {
	__antithesis_instrumentation__.Notify(579100)
	encodeValue := func(name string) *yaml.Node {
		__antithesis_instrumentation__.Notify(579103)
		return s.Registry.valueToYAML(t, name)
	}
	__antithesis_instrumentation__.Notify(579101)
	n := yaml.Node{Kind: yaml.MappingNode}
	for _, name := range s.Registry.names {
		__antithesis_instrumentation__.Notify(579104)
		n.Content = append(n.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: name},
			encodeValue(name),
		)
	}
	__antithesis_instrumentation__.Notify(579102)
	return &n
}

func (s Suite) encodeQueries(t *testing.T) *yaml.Node {
	__antithesis_instrumentation__.Notify(579105)
	queries := &yaml.Node{Kind: yaml.SequenceNode}
	for _, q := range s.DatabaseTests {
		__antithesis_instrumentation__.Notify(579107)
		queries.Content = append(queries.Content,
			q.encode(t, s.Registry),
		)
	}
	__antithesis_instrumentation__.Notify(579106)
	return queries
}

func (s Suite) encodeAttributes(t *testing.T) *yaml.Node {
	__antithesis_instrumentation__.Notify(579108)
	n := yaml.Node{Kind: yaml.MappingNode}
	for _, tc := range s.AttributeTests {
		__antithesis_instrumentation__.Notify(579110)
		n.Content = append(n.Content,
			scalarYAML(tc.Entity),
			tc.encode(t, s),
		)
	}
	__antithesis_instrumentation__.Notify(579109)
	return &n
}

func (s Suite) encodeComparisons(t *testing.T) *yaml.Node {
	__antithesis_instrumentation__.Notify(579111)
	n := yaml.Node{Kind: yaml.SequenceNode}
	for _, ct := range s.ComparisonTests {
		__antithesis_instrumentation__.Notify(579113)
		n.Content = append(n.Content, ct.encode(t))
	}
	__antithesis_instrumentation__.Notify(579112)
	return &n
}

func scalarYAML(value string) *yaml.Node {
	__antithesis_instrumentation__.Notify(579114)
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
	}
}
