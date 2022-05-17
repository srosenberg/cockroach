package reltest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type DatabaseTest struct {
	Data []string

	Indexes [][][]rel.Attr

	QueryCases []QueryTest
}

type QueryTest struct {
	Name string

	Query rel.Clauses

	ResVars []rel.Var

	Results [][]interface{}

	Entities []rel.Var

	ErrorRE string
}

func (tc DatabaseTest) run(t *testing.T, s Suite) {
	__antithesis_instrumentation__.Notify(578998)
	for _, databaseIndexes := range tc.databaseIndexes() {
		__antithesis_instrumentation__.Notify(578999)
		t.Run(fmt.Sprintf("%s", databaseIndexes), func(t *testing.T) {
			__antithesis_instrumentation__.Notify(579000)
			db, err := rel.NewDatabase(s.Schema, databaseIndexes)
			require.NoError(t, err)
			for _, k := range tc.Data {
				__antithesis_instrumentation__.Notify(579002)
				v := s.Registry.MustGetByName(t, k)
				require.NoError(t, db.Insert(v))
			}
			__antithesis_instrumentation__.Notify(579001)
			for _, qc := range tc.QueryCases {
				__antithesis_instrumentation__.Notify(579003)
				t.Run(qc.Name, func(t *testing.T) {
					__antithesis_instrumentation__.Notify(579004)
					qc.run(t, db)
				})
			}
		})
	}
}

func (qc QueryTest) run(t *testing.T, db *rel.Database) {
	__antithesis_instrumentation__.Notify(579005)
	var results [][]interface{}
	q, err := rel.NewQuery(db.Schema(), qc.Query...)
	if qc.ErrorRE != "" {
		__antithesis_instrumentation__.Notify(579010)
		require.Regexp(t, qc.ErrorRE, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(579011)
	}
	__antithesis_instrumentation__.Notify(579006)

	require.NoError(t, err)
	require.Equal(t, qc.Entities, q.Entities())
	require.NoError(t, q.Iterate(db, func(r rel.Result) error {
		__antithesis_instrumentation__.Notify(579012)
		var cur []interface{}
		for _, v := range qc.ResVars {
			__antithesis_instrumentation__.Notify(579014)
			cur = append(cur, r.Var(v))
		}
		__antithesis_instrumentation__.Notify(579013)
		results = append(results, cur)
		return nil
	}))
	__antithesis_instrumentation__.Notify(579007)
	expResults := append(qc.Results[:0:0], qc.Results...)
	findResultInExp := func(res []interface{}) (found bool) {
		__antithesis_instrumentation__.Notify(579015)
		for i, exp := range expResults {
			__antithesis_instrumentation__.Notify(579017)
			if reflect.DeepEqual(exp, res) {
				__antithesis_instrumentation__.Notify(579018)
				expResults = append(expResults[:i], expResults[i+1:]...)
				return true
			} else {
				__antithesis_instrumentation__.Notify(579019)
			}
		}
		__antithesis_instrumentation__.Notify(579016)
		return false
	}
	__antithesis_instrumentation__.Notify(579008)

	for _, res := range results {
		__antithesis_instrumentation__.Notify(579020)
		if !findResultInExp(res) {
			__antithesis_instrumentation__.Notify(579021)
			t.Fatalf("failed to find %v in %v", res, expResults)
		} else {
			__antithesis_instrumentation__.Notify(579022)
		}
	}
	__antithesis_instrumentation__.Notify(579009)
	require.Empty(t, expResults, "got", results)
}

func (tc DatabaseTest) encode(t *testing.T, r *Registry) *yaml.Node {
	__antithesis_instrumentation__.Notify(579023)
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			scalarYAML("indexes"),
			tc.encodeIndexes(),
			scalarYAML("data"),
			tc.encodeData(),
			scalarYAML("queries"),
			tc.encodeQueries(t, r),
		},
	}
}

func (tc DatabaseTest) encodeIndexes() *yaml.Node {
	__antithesis_instrumentation__.Notify(579024)
	databaseIndexesNode := yaml.Node{Kind: yaml.SequenceNode}
	for _, indexes := range tc.databaseIndexes() {
		__antithesis_instrumentation__.Notify(579026)
		indexesNode := yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
		for _, idx := range indexes {
			__antithesis_instrumentation__.Notify(579028)
			indexesNode.Content = append(indexesNode.Content, encodeAttrs(idx))
		}
		__antithesis_instrumentation__.Notify(579027)
		databaseIndexesNode.Content = append(databaseIndexesNode.Content, &indexesNode)
	}
	__antithesis_instrumentation__.Notify(579025)
	return &databaseIndexesNode
}

func encodeAttrs(idx []rel.Attr) *yaml.Node {
	__antithesis_instrumentation__.Notify(579029)
	indexNode := yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
	for _, attr := range idx {
		__antithesis_instrumentation__.Notify(579031)
		indexNode.Content = append(indexNode.Content, scalarYAML(attr.String()))
	}
	__antithesis_instrumentation__.Notify(579030)
	return &indexNode
}

func (tc DatabaseTest) encodeData() *yaml.Node {
	__antithesis_instrumentation__.Notify(579032)
	dataNode := yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
	for _, k := range tc.Data {
		__antithesis_instrumentation__.Notify(579034)
		dataNode.Content = append(dataNode.Content, scalarYAML(k))
	}
	__antithesis_instrumentation__.Notify(579033)
	return &dataNode
}

func (tc DatabaseTest) encodeQueries(t *testing.T, r *Registry) *yaml.Node {
	__antithesis_instrumentation__.Notify(579035)
	queriesNode := yaml.Node{Kind: yaml.MappingNode}

	encodeValues := func(t *testing.T, v []interface{}) *yaml.Node {
		__antithesis_instrumentation__.Notify(579041)
		var seq yaml.Node
		seq.Kind = yaml.SequenceNode
		seq.Style = yaml.FlowStyle
		for _, v := range v {
			__antithesis_instrumentation__.Notify(579043)
			name, ok := r.GetName(v)
			if ok {
				__antithesis_instrumentation__.Notify(579044)
				seq.Content = append(seq.Content, scalarYAML(name))
			} else {
				__antithesis_instrumentation__.Notify(579045)
				if typ, isType := v.(reflect.Type); isType {
					__antithesis_instrumentation__.Notify(579046)
					seq.Content = append(seq.Content, scalarYAML(typ.String()))
				} else {
					__antithesis_instrumentation__.Notify(579047)
					var content yaml.Node
					require.NoError(t, content.Encode(v))
					seq.Content = append(seq.Content, &content)
				}
			}
		}
		__antithesis_instrumentation__.Notify(579042)
		return &seq
	}
	__antithesis_instrumentation__.Notify(579036)
	encodeResults := func(t *testing.T, results [][]interface{}) *yaml.Node {
		__antithesis_instrumentation__.Notify(579048)
		var res yaml.Node
		res.Kind = yaml.SequenceNode
		for _, r := range results {
			__antithesis_instrumentation__.Notify(579050)
			res.Content = append(res.Content, encodeValues(t, r))
		}
		__antithesis_instrumentation__.Notify(579049)
		return &res
	}
	__antithesis_instrumentation__.Notify(579037)
	encodeVars := func(vars []rel.Var) *yaml.Node {
		__antithesis_instrumentation__.Notify(579051)
		n := yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
		for _, v := range vars {
			__antithesis_instrumentation__.Notify(579053)
			n.Content = append(n.Content, scalarYAML("$"+string(v)))
		}
		__antithesis_instrumentation__.Notify(579052)
		return &n
	}
	__antithesis_instrumentation__.Notify(579038)
	addQuery := func(t *testing.T, qt QueryTest) {
		__antithesis_instrumentation__.Notify(579054)
		var query yaml.Node
		require.NoError(t, query.Encode(qt.Query))

		var qtNode *yaml.Node
		if qt.ErrorRE != "" {
			__antithesis_instrumentation__.Notify(579056)
			qtNode = &yaml.Node{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					scalarYAML("query"),
					&query,
					scalarYAML("error"),
					scalarYAML(qt.ErrorRE),
				},
			}
		} else {
			__antithesis_instrumentation__.Notify(579057)
			qtNode = &yaml.Node{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					scalarYAML("query"),
					&query,
					scalarYAML("entities"),
					encodeVars(qt.Entities),
					scalarYAML("result-vars"),
					encodeVars(qt.ResVars),
					scalarYAML("results"),
					encodeResults(t, qt.Results),
				},
			}
		}
		__antithesis_instrumentation__.Notify(579055)
		queriesNode.Content = append(queriesNode.Content,
			scalarYAML(qt.Name), qtNode,
		)
	}
	__antithesis_instrumentation__.Notify(579039)
	for _, qc := range tc.QueryCases {
		__antithesis_instrumentation__.Notify(579058)
		addQuery(t, qc)
	}
	__antithesis_instrumentation__.Notify(579040)
	return &queriesNode
}

func (tc DatabaseTest) databaseIndexes() [][][]rel.Attr {
	__antithesis_instrumentation__.Notify(579059)
	if len(tc.Indexes) == 0 {
		__antithesis_instrumentation__.Notify(579061)
		return [][][]rel.Attr{{}}
	} else {
		__antithesis_instrumentation__.Notify(579062)
	}
	__antithesis_instrumentation__.Notify(579060)
	return tc.Indexes
}
