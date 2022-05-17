package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type Query struct {
	schema *Schema

	clauses []Clause

	variables []Var

	variableSlots map[Var]slotIdx

	entities []slotIdx

	slots []slot

	facts []fact

	filters []filter

	mu struct {
		syncutil.Mutex
		cached *evalContext
	}
}

type Result interface {
	Var(name Var) interface{}
}

type ResultIterator func(r Result) error

func NewQuery(sc *Schema, clauses ...Clause) (_ *Query, err error) {
	__antithesis_instrumentation__.Notify(578597)
	defer func() {
		__antithesis_instrumentation__.Notify(578599)
		switch r := recover().(type) {
		case nil:
			__antithesis_instrumentation__.Notify(578600)
			return
		case error:
			__antithesis_instrumentation__.Notify(578601)
			err = errors.Wrap(r, "failed to construct query")
		default:
			__antithesis_instrumentation__.Notify(578602)
			err = errors.AssertionFailedf("failed to construct query: %v", r)
		}
	}()
	__antithesis_instrumentation__.Notify(578598)
	q := newQuery(sc, clauses)
	return q, nil
}

func (q *Query) Iterate(db *Database, ri ResultIterator) error {
	__antithesis_instrumentation__.Notify(578603)
	ec := q.getEvalContext()
	defer q.putEvalContext(ec)
	return ec.Iterate(db, ri)
}

func (q *Query) getEvalContext() *evalContext {
	__antithesis_instrumentation__.Notify(578604)
	getCachedEvalContext := func() (ec *evalContext) {
		__antithesis_instrumentation__.Notify(578607)
		q.mu.Lock()
		defer q.mu.Unlock()
		ec, q.mu.cached = q.mu.cached, ec
		return ec
	}
	__antithesis_instrumentation__.Notify(578605)
	if ec := getCachedEvalContext(); ec != nil {
		__antithesis_instrumentation__.Notify(578608)
		return ec
	} else {
		__antithesis_instrumentation__.Notify(578609)
	}
	__antithesis_instrumentation__.Notify(578606)
	return newEvalContext(q)
}

func (q *Query) putEvalContext(ec *evalContext) {
	__antithesis_instrumentation__.Notify(578610)
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.mu.cached == nil {
		__antithesis_instrumentation__.Notify(578611)
		q.mu.cached = ec
	} else {
		__antithesis_instrumentation__.Notify(578612)
	}
}

func (q *Query) Entities() []Var {
	__antithesis_instrumentation__.Notify(578613)
	var entitySlots util.FastIntSet
	for _, slotIdx := range q.entities {
		__antithesis_instrumentation__.Notify(578617)
		entitySlots.Add(int(slotIdx))
	}
	__antithesis_instrumentation__.Notify(578614)
	vars := make([]Var, 0, len(q.entities))
	for v, slotIdx := range q.variableSlots {
		__antithesis_instrumentation__.Notify(578618)
		if !entitySlots.Contains(int(slotIdx)) {
			__antithesis_instrumentation__.Notify(578620)
			continue
		} else {
			__antithesis_instrumentation__.Notify(578621)
		}
		__antithesis_instrumentation__.Notify(578619)
		vars = append(vars, v)
	}
	__antithesis_instrumentation__.Notify(578615)
	sort.Slice(vars, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(578622)
		return q.variableSlots[vars[i]] < q.variableSlots[vars[j]]
	})
	__antithesis_instrumentation__.Notify(578616)
	return vars
}

func (q *Query) Clauses() Clauses {
	__antithesis_instrumentation__.Notify(578623)
	return q.clauses
}
