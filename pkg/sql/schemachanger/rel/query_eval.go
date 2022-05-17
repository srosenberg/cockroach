package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

type evalContext struct {
	q  *Query
	db *Database
	ri ResultIterator

	facts             []fact
	depth, cur        int
	slots             []slot
	filterSliceCaches map[int][]reflect.Value
}

func newEvalContext(q *Query) *evalContext {
	__antithesis_instrumentation__.Notify(578757)
	return &evalContext{
		q:     q,
		depth: len(q.entities),
		slots: append(make([]slot, 0, len(q.slots)), q.slots...),
		facts: q.facts,
	}
}

type evalResult evalContext

func (ec *evalResult) Var(name Var) interface{} {
	__antithesis_instrumentation__.Notify(578758)
	n, ok := ec.q.variableSlots[name]
	if !ok {
		__antithesis_instrumentation__.Notify(578760)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(578761)
	}
	__antithesis_instrumentation__.Notify(578759)
	return ec.slots[n].toInterface()
}

func (ec *evalContext) Iterate(db *Database, ri ResultIterator) error {
	__antithesis_instrumentation__.Notify(578762)
	if db.schema != ec.q.schema {
		__antithesis_instrumentation__.Notify(578766)
		return errors.Errorf(
			"query and database are not from the same schema: %s != %s",
			db.schema.name, ec.q.schema.name,
		)
	} else {
		__antithesis_instrumentation__.Notify(578767)
	}
	__antithesis_instrumentation__.Notify(578763)
	defer func() { __antithesis_instrumentation__.Notify(578768); ec.db, ec.ri = nil, nil }()
	__antithesis_instrumentation__.Notify(578764)
	ec.db, ec.ri = db, ri

	if ec.depth == 0 {
		__antithesis_instrumentation__.Notify(578769)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(578770)
	}
	__antithesis_instrumentation__.Notify(578765)
	return ec.iterateNext()
}

func (ec *evalContext) iterateNext() error {
	__antithesis_instrumentation__.Notify(578771)

	if ec.cur == ec.depth {
		__antithesis_instrumentation__.Notify(578775)
		if ec.haveUnboundSlots() || func() bool {
			__antithesis_instrumentation__.Notify(578777)
			return ec.checkFilters() == true
		}() == true {
			__antithesis_instrumentation__.Notify(578778)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(578779)
		}
		__antithesis_instrumentation__.Notify(578776)
		return ec.ri((*evalResult)(ec))
	} else {
		__antithesis_instrumentation__.Notify(578780)
	}
	__antithesis_instrumentation__.Notify(578772)

	if done, err := ec.maybeVisitAlreadyBoundEntity(); done || func() bool {
		__antithesis_instrumentation__.Notify(578781)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(578782)
		return err
	} else {
		__antithesis_instrumentation__.Notify(578783)
	}
	__antithesis_instrumentation__.Notify(578773)

	where, anyAttr, anyValues := ec.buildWhere()
	defer putValues(where)
	if len(anyValues) > 0 {
		__antithesis_instrumentation__.Notify(578784)
		for _, v := range anyValues {
			__antithesis_instrumentation__.Notify(578786)
			where.add(anyAttr, v.value)
			if err := ec.db.iterate(where, ec); err != nil {
				__antithesis_instrumentation__.Notify(578787)
				return err
			} else {
				__antithesis_instrumentation__.Notify(578788)
			}
		}
		__antithesis_instrumentation__.Notify(578785)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(578789)
	}
	__antithesis_instrumentation__.Notify(578774)

	return ec.db.iterate(where, ec)
}

func (ec *evalContext) visit(e *entity) error {
	__antithesis_instrumentation__.Notify(578790)

	var slotsFilled util.FastIntSet
	defer func() {
		__antithesis_instrumentation__.Notify(578796)
		slotsFilled.ForEach(func(i int) {
			__antithesis_instrumentation__.Notify(578797)
			ec.slots[i].typedValue = typedValue{}
		})
	}()
	__antithesis_instrumentation__.Notify(578791)

	if foundContradiction := maybeSet(
		ec.slots, ec.q.entities[ec.cur],
		typedValue{
			typ:   e.getTypeInfo(ec.db.Schema()).typ,
			value: e.getComparableValue(ec.db.schema, Self),
		},
		&slotsFilled,
	); foundContradiction {
		__antithesis_instrumentation__.Notify(578798)
		return errors.AssertionFailedf(
			"unexpectedly found contradiction setting entity",
		)
	} else {
		__antithesis_instrumentation__.Notify(578799)
	}
	__antithesis_instrumentation__.Notify(578792)

	setEntitySlots := func() (foundContradiction bool) {
		__antithesis_instrumentation__.Notify(578800)

		for _, f := range ec.facts {
			__antithesis_instrumentation__.Notify(578802)
			if f.variable != ec.q.entities[ec.cur] {
				__antithesis_instrumentation__.Notify(578805)
				continue
			} else {
				__antithesis_instrumentation__.Notify(578806)
			}
			__antithesis_instrumentation__.Notify(578803)

			tv, ok := e.getTypedValue(ec.db.schema, f.attr)
			if !ok {
				__antithesis_instrumentation__.Notify(578807)
				return true
			} else {
				__antithesis_instrumentation__.Notify(578808)
			}
			__antithesis_instrumentation__.Notify(578804)
			if contradiction := maybeSet(
				ec.slots, f.value, tv, &slotsFilled,
			); contradiction {
				__antithesis_instrumentation__.Notify(578809)
				return true
			} else {
				__antithesis_instrumentation__.Notify(578810)
			}
		}
		__antithesis_instrumentation__.Notify(578801)

		return false
	}
	__antithesis_instrumentation__.Notify(578793)

	if contradiction := setEntitySlots() || func() bool {
		__antithesis_instrumentation__.Notify(578811)
		return unify(
			ec.facts, ec.slots, &slotsFilled,
		) == true
	}() == true; contradiction {
		__antithesis_instrumentation__.Notify(578812)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(578813)
	}
	__antithesis_instrumentation__.Notify(578794)

	ec.cur++
	defer func() { __antithesis_instrumentation__.Notify(578814); ec.cur-- }()
	__antithesis_instrumentation__.Notify(578795)
	return ec.iterateNext()
}

func (ec *evalContext) haveUnboundSlots() bool {
	__antithesis_instrumentation__.Notify(578815)
	for _, v := range ec.q.variableSlots {
		__antithesis_instrumentation__.Notify(578817)
		if ec.slots[v].value == nil {
			__antithesis_instrumentation__.Notify(578818)
			return true
		} else {
			__antithesis_instrumentation__.Notify(578819)
		}
	}
	__antithesis_instrumentation__.Notify(578816)
	return false
}

func (ec *evalContext) checkFilters() (done bool) {
	__antithesis_instrumentation__.Notify(578820)
	for i := range ec.q.filters {
		__antithesis_instrumentation__.Notify(578822)
		if done = ec.checkFilter(i); done {
			__antithesis_instrumentation__.Notify(578823)
			return true
		} else {
			__antithesis_instrumentation__.Notify(578824)
		}
	}
	__antithesis_instrumentation__.Notify(578821)
	return false
}

func (ec *evalContext) checkFilter(i int) bool {
	__antithesis_instrumentation__.Notify(578825)
	f := ec.q.filters[i]
	ins := ec.getFilterInput(i)
	defer func() {
		__antithesis_instrumentation__.Notify(578828)
		for i := range f.input {
			__antithesis_instrumentation__.Notify(578829)
			ins[i] = reflect.Value{}
		}
	}()
	__antithesis_instrumentation__.Notify(578826)
	for i, idx := range f.input {
		__antithesis_instrumentation__.Notify(578830)
		inI := ec.slots[idx].typedValue.toInterface()
		in := reflect.ValueOf(inI)

		inType := f.predicate.Type().In(i)
		if in.Type() != inType {
			__antithesis_instrumentation__.Notify(578832)
			if in.Type().ConvertibleTo(inType) {
				__antithesis_instrumentation__.Notify(578833)
				in = in.Convert(inType)
			} else {
				__antithesis_instrumentation__.Notify(578834)
				return true
			}
		} else {
			__antithesis_instrumentation__.Notify(578835)
		}
		__antithesis_instrumentation__.Notify(578831)
		ins[i] = in
	}
	__antithesis_instrumentation__.Notify(578827)

	outs := f.predicate.Call(ins)
	return !outs[0].Bool()
}

func (ec *evalContext) buildWhere() (where *valuesMap, anyAttr ordinal, anyValues []typedValue) {
	__antithesis_instrumentation__.Notify(578836)
	where = getValues()

	for _, f := range ec.facts {
		__antithesis_instrumentation__.Notify(578838)
		if f.variable != ec.q.entities[ec.cur] {
			__antithesis_instrumentation__.Notify(578840)
			continue
		} else {
			__antithesis_instrumentation__.Notify(578841)
		}
		__antithesis_instrumentation__.Notify(578839)
		s := ec.slots[f.value]
		if !s.empty() {
			__antithesis_instrumentation__.Notify(578842)
			where.add(f.attr, s.value)
		} else {
			__antithesis_instrumentation__.Notify(578843)
			if anyValues == nil && func() bool {
				__antithesis_instrumentation__.Notify(578844)
				return s.any != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(578845)
				anyAttr, anyValues = f.attr, s.any
			} else {
				__antithesis_instrumentation__.Notify(578846)
			}
		}
	}
	__antithesis_instrumentation__.Notify(578837)
	return where, anyAttr, anyValues
}

func unify(facts []fact, s []slot, slotsFilled *util.FastIntSet) (contradictionFound bool) {
	__antithesis_instrumentation__.Notify(578847)
	contradictionFound, _ = unifyReturningContradiction(facts, s, slotsFilled)
	return contradictionFound
}

func unifyReturningContradiction(
	facts []fact, s []slot, slotsFilled *util.FastIntSet,
) (contradictionFound bool, contradicted fact) {
	__antithesis_instrumentation__.Notify(578848)

	for {
		__antithesis_instrumentation__.Notify(578849)
		var somethingChanged bool
		var prev, cur *fact
		for i := 1; i < len(facts); i++ {
			__antithesis_instrumentation__.Notify(578851)
			prev, cur = &facts[i-1], &facts[i]
			if prev.variable != cur.variable || func() bool {
				__antithesis_instrumentation__.Notify(578854)
				return prev.attr != cur.attr == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(578855)
				return prev.value == cur.value == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(578856)
				return s[prev.value].eq(s[cur.value]) == true
			}() == true {
				__antithesis_instrumentation__.Notify(578857)
				continue
			} else {
				__antithesis_instrumentation__.Notify(578858)
			}
			__antithesis_instrumentation__.Notify(578852)

			var unified bool
			for _, v := range []struct {
				dst, src slotIdx
			}{
				{prev.value, cur.value},
				{cur.value, prev.value},
			} {
				__antithesis_instrumentation__.Notify(578859)
				if s[v.dst].empty() {
					__antithesis_instrumentation__.Notify(578860)
					if contradictionFound = maybeSet(
						s, v.dst, s[v.src].typedValue, slotsFilled,
					); contradictionFound {
						__antithesis_instrumentation__.Notify(578862)
						return true, *cur
					} else {
						__antithesis_instrumentation__.Notify(578863)
					}
					__antithesis_instrumentation__.Notify(578861)
					unified = true
					break
				} else {
					__antithesis_instrumentation__.Notify(578864)
				}
			}
			__antithesis_instrumentation__.Notify(578853)
			if unified {
				__antithesis_instrumentation__.Notify(578865)
				somethingChanged = true
			} else {
				__antithesis_instrumentation__.Notify(578866)
				return true, *cur
			}
		}
		__antithesis_instrumentation__.Notify(578850)
		if !somethingChanged {
			__antithesis_instrumentation__.Notify(578867)
			return false, fact{}
		} else {
			__antithesis_instrumentation__.Notify(578868)
		}
	}
}

func (ec *evalContext) maybeVisitAlreadyBoundEntity() (done bool, _ error) {
	__antithesis_instrumentation__.Notify(578869)
	s := &ec.slots[ec.q.entities[ec.cur]]
	if s.empty() {
		__antithesis_instrumentation__.Notify(578872)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(578873)
	}
	__antithesis_instrumentation__.Notify(578870)
	e, ok := ec.db.entities[s.value]

	if !ok {
		__antithesis_instrumentation__.Notify(578874)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(578875)
	}
	__antithesis_instrumentation__.Notify(578871)
	return true, ec.visit(e)
}

func (ec *evalContext) getFilterInput(i int) (ins []reflect.Value) {
	__antithesis_instrumentation__.Notify(578876)
	if ec.filterSliceCaches == nil {
		__antithesis_instrumentation__.Notify(578879)
		ec.filterSliceCaches = make(map[int][]reflect.Value)
	} else {
		__antithesis_instrumentation__.Notify(578880)
	}
	__antithesis_instrumentation__.Notify(578877)
	c, ok := ec.filterSliceCaches[i]
	if !ok {
		__antithesis_instrumentation__.Notify(578881)
		c = make([]reflect.Value, len(ec.q.filters[i].input))
		ec.filterSliceCaches[i] = c
	} else {
		__antithesis_instrumentation__.Notify(578882)
	}
	__antithesis_instrumentation__.Notify(578878)
	return c
}
