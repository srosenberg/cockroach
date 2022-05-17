// Package rules contains rules to:
//  - generate dependency edges for a graph which contains op edges,
//  - mark certain op-edges as no-op.
package rules

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan/internal/scgraph"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func ApplyDepRules(g *scgraph.Graph) error {
	__antithesis_instrumentation__.Notify(594184)
	for _, dr := range registry.depRules {
		__antithesis_instrumentation__.Notify(594186)
		start := timeutil.Now()
		var added int
		if err := dr.q.Iterate(g.Database(), func(r rel.Result) error {
			__antithesis_instrumentation__.Notify(594188)
			from := r.Var(dr.from).(*screl.Node)
			to := r.Var(dr.to).(*screl.Node)
			added++
			return g.AddDepEdge(
				dr.name, dr.kind, from.Target, from.CurrentStatus, to.Target, to.CurrentStatus,
			)
		}); err != nil {
			__antithesis_instrumentation__.Notify(594189)
			return err
		} else {
			__antithesis_instrumentation__.Notify(594190)
		}
		__antithesis_instrumentation__.Notify(594187)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(594191)
			log.Infof(
				context.TODO(), "applying dep rule %s %d took %v",
				dr.name, added, timeutil.Since(start),
			)
		} else {
			__antithesis_instrumentation__.Notify(594192)
		}
	}
	__antithesis_instrumentation__.Notify(594185)
	return nil
}

func ApplyOpRules(g *scgraph.Graph) (*scgraph.Graph, error) {
	__antithesis_instrumentation__.Notify(594193)
	db := g.Database()
	m := make(map[*screl.Node][]string)
	for _, rule := range registry.opRules {
		__antithesis_instrumentation__.Notify(594196)
		var added int
		start := timeutil.Now()
		err := rule.q.Iterate(db, func(r rel.Result) error {
			__antithesis_instrumentation__.Notify(594199)
			added++
			n := r.Var(rule.from).(*screl.Node)
			m[n] = append(m[n], rule.name)
			return nil
		})
		__antithesis_instrumentation__.Notify(594197)
		if err != nil {
			__antithesis_instrumentation__.Notify(594200)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(594201)
		}
		__antithesis_instrumentation__.Notify(594198)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(594202)
			log.Infof(
				context.TODO(), "applying op rule %s %d took %v",
				rule.name, added, timeutil.Since(start),
			)
		} else {
			__antithesis_instrumentation__.Notify(594203)
		}
	}
	__antithesis_instrumentation__.Notify(594194)

	ret := g.ShallowClone()
	for from, rules := range m {
		__antithesis_instrumentation__.Notify(594204)
		if opEdge, ok := g.GetOpEdgeFrom(from); ok {
			__antithesis_instrumentation__.Notify(594205)
			ret.MarkAsNoOp(opEdge, rules...)
		} else {
			__antithesis_instrumentation__.Notify(594206)
		}
	}
	__antithesis_instrumentation__.Notify(594195)
	return ret, nil
}

var registry struct {
	depRules []registeredDepRule
	opRules  []registeredOpRule
}

type registeredDepRule struct {
	name     string
	from, to rel.Var
	q        *rel.Query
	kind     scgraph.DepEdgeKind
}

type registeredOpRule struct {
	name string
	from rel.Var
	q    *rel.Query
}

func registerDepRule(
	ruleName string, edgeKind scgraph.DepEdgeKind, from, to rel.Var, query *rel.Query,
) {
	__antithesis_instrumentation__.Notify(594207)
	registry.depRules = append(registry.depRules, registeredDepRule{
		name: ruleName,
		kind: edgeKind,
		from: from,
		to:   to,
		q:    query,
	})
}

func registerOpRule(ruleName string, from rel.Var, q *rel.Query) {
	__antithesis_instrumentation__.Notify(594208)
	registry.opRules = append(registry.opRules, registeredOpRule{
		name: ruleName,
		from: from,
		q:    q,
	})
}
