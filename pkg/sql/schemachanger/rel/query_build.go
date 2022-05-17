package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"
	"sort"

	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v2"
)

type queryBuilder struct {
	sc            *Schema
	variables     []Var
	variableSlots map[Var]slotIdx
	facts         []fact
	slots         []slot
	filters       []filter

	slotIsEntity []bool
}

func newQuery(sc *Schema, clauses Clauses) *Query {
	__antithesis_instrumentation__.Notify(578624)
	p := &queryBuilder{
		sc:            sc,
		variableSlots: map[Var]slotIdx{},
	}

	clauses = flattened(clauses)
	for _, t := range clauses {
		__antithesis_instrumentation__.Notify(578628)
		p.processClause(t)
	}
	__antithesis_instrumentation__.Notify(578625)

	entities := p.findEntitySlots()
	sort.SliceStable(p.facts, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(578629)
		if p.facts[i].variable == p.facts[j].variable {
			__antithesis_instrumentation__.Notify(578631)
			return p.facts[i].attr < p.facts[j].attr
		} else {
			__antithesis_instrumentation__.Notify(578632)
		}
		__antithesis_instrumentation__.Notify(578630)
		return p.facts[i].variable < p.facts[j].variable
	})
	__antithesis_instrumentation__.Notify(578626)

	if contradictionFound, contradiction := unifyReturningContradiction(
		p.facts, p.slots, nil,
	); contradictionFound {
		__antithesis_instrumentation__.Notify(578633)
		panic(errors.Errorf(
			"query contains contradiction on %v", sc.attrs[contradiction.attr],
		))
	} else {
		__antithesis_instrumentation__.Notify(578634)
	}
	__antithesis_instrumentation__.Notify(578627)
	return &Query{
		schema:        sc,
		variables:     p.variables,
		variableSlots: p.variableSlots,
		clauses:       clauses,
		entities:      entities,
		facts:         p.facts,
		slots:         p.slots,
		filters:       p.filters,
	}
}

func (p *queryBuilder) processClause(t Clause) {
	__antithesis_instrumentation__.Notify(578635)
	defer func() {
		__antithesis_instrumentation__.Notify(578637)
		if r := recover(); r != nil {
			__antithesis_instrumentation__.Notify(578638)
			rErr, ok := r.(error)
			if !ok {
				__antithesis_instrumentation__.Notify(578641)
				rErr = errors.AssertionFailedf("processClause: panic: %v", r)
			} else {
				__antithesis_instrumentation__.Notify(578642)
			}
			__antithesis_instrumentation__.Notify(578639)
			encoded, err := yaml.Marshal(t)
			if err != nil {
				__antithesis_instrumentation__.Notify(578643)
				panic(errors.CombineErrors(rErr, errors.Wrap(
					err, "failed to encode clause",
				)))
			} else {
				__antithesis_instrumentation__.Notify(578644)
			}
			__antithesis_instrumentation__.Notify(578640)
			panic(errors.Wrapf(
				rErr, "failed to process invalid clause %s", encoded,
			))
		} else {
			__antithesis_instrumentation__.Notify(578645)
		}
	}()
	__antithesis_instrumentation__.Notify(578636)
	switch t := t.(type) {
	case *tripleDecl:
		__antithesis_instrumentation__.Notify(578646)
		p.processTripleDecl(t)
	case *eqDecl:
		__antithesis_instrumentation__.Notify(578647)
		p.processEqDecl(t)
	case *filterDecl:
		__antithesis_instrumentation__.Notify(578648)
		p.processFilterDecl(t)
	case and:
		__antithesis_instrumentation__.Notify(578649)
		panic(errors.AssertionFailedf("and clauses should be flattened away"))
	default:
		__antithesis_instrumentation__.Notify(578650)
		panic(errors.AssertionFailedf("unknown clause type %T", t))
	}
}

func (p *queryBuilder) processTripleDecl(fd *tripleDecl) {
	__antithesis_instrumentation__.Notify(578651)
	f := fact{
		variable: p.maybeAddVar(fd.entity, true),
		attr:     p.sc.mustGetOrdinal(fd.attribute),
	}
	f.value = p.processValueExpr(fd.value)
	p.typeCheck(f)
	p.facts = append(p.facts, f)
}

func (p *queryBuilder) processEqDecl(t *eqDecl) {
	__antithesis_instrumentation__.Notify(578652)
	varIdx := p.maybeAddVar(t.v, false)
	valueIdx := p.processValueExpr(t.expr)

	p.facts = append(p.facts,
		fact{
			variable: varIdx,
			attr:     p.sc.mustGetOrdinal(Self),
			value:    valueIdx,
		},
		fact{
			variable: varIdx,
			attr:     p.sc.mustGetOrdinal(Self),
			value:    varIdx,
		})
}

func (p *queryBuilder) processFilterDecl(t *filterDecl) {
	__antithesis_instrumentation__.Notify(578653)
	fv := reflect.ValueOf(t.predicateFunc)

	if err := checkNotNil(fv); err != nil {
		__antithesis_instrumentation__.Notify(578659)
		panic(errors.Wrapf(err, "nil filter function for variables %s", t.vars))
	} else {
		__antithesis_instrumentation__.Notify(578660)
	}
	__antithesis_instrumentation__.Notify(578654)
	if fv.Kind() != reflect.Func {
		__antithesis_instrumentation__.Notify(578661)
		panic(errors.Errorf(
			"non-function %T filter function for variables %s",
			t.predicateFunc, t.vars,
		))
	} else {
		__antithesis_instrumentation__.Notify(578662)
	}
	__antithesis_instrumentation__.Notify(578655)
	ft := fv.Type()
	if ft.NumOut() != 1 || func() bool {
		__antithesis_instrumentation__.Notify(578663)
		return ft.Out(0) != boolType == true
	}() == true {
		__antithesis_instrumentation__.Notify(578664)
		panic(errors.Errorf(
			"invalid non-bool return from %T filter function for variables %s",
			t.predicateFunc, t.vars,
		))
	} else {
		__antithesis_instrumentation__.Notify(578665)
	}
	__antithesis_instrumentation__.Notify(578656)
	if ft.NumIn() != len(t.vars) {
		__antithesis_instrumentation__.Notify(578666)
		panic(errors.Errorf(
			"invalid %T filter function for variables %s accepts %d inputs",
			t.predicateFunc, t.vars, ft.NumIn(),
		))
	} else {
		__antithesis_instrumentation__.Notify(578667)
	}
	__antithesis_instrumentation__.Notify(578657)

	slots := make([]slotIdx, len(t.vars))
	for i, v := range t.vars {
		__antithesis_instrumentation__.Notify(578668)
		slots[i] = p.maybeAddVar(v, false)

		checkSlotType(&p.slots[slots[i]], ft.In(i))
	}
	__antithesis_instrumentation__.Notify(578658)
	p.filters = append(p.filters, filter{
		input:     slots,
		predicate: fv,
	})
}

func (p *queryBuilder) processValueExpr(rawValue expr) slotIdx {
	__antithesis_instrumentation__.Notify(578669)
	switch v := rawValue.(type) {
	case Var:
		__antithesis_instrumentation__.Notify(578670)
		return p.maybeAddVar(v, false)
	case anyExpr:
		__antithesis_instrumentation__.Notify(578671)
		sd := slot{
			any: make([]typedValue, len(v)),
		}
		for i, vv := range v {
			__antithesis_instrumentation__.Notify(578676)
			tv, err := makeComparableValue(vv)
			if err != nil {
				__antithesis_instrumentation__.Notify(578678)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(578679)
			}
			__antithesis_instrumentation__.Notify(578677)
			sd.any[i] = tv
		}
		__antithesis_instrumentation__.Notify(578672)
		return p.fillSlot(sd, false)
	case valueExpr:
		__antithesis_instrumentation__.Notify(578673)
		tv, err := makeComparableValue(v.value)
		if err != nil {
			__antithesis_instrumentation__.Notify(578680)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(578681)
		}
		__antithesis_instrumentation__.Notify(578674)
		return p.fillSlot(slot{typedValue: tv}, false)
	default:
		__antithesis_instrumentation__.Notify(578675)
		panic(errors.AssertionFailedf("unknown expr type %T", rawValue))
	}
}

func (p *queryBuilder) maybeAddVar(v Var, entity bool) slotIdx {
	__antithesis_instrumentation__.Notify(578682)
	id, exists := p.variableSlots[v]
	if exists {
		__antithesis_instrumentation__.Notify(578684)
		if entity && func() bool {
			__antithesis_instrumentation__.Notify(578686)
			return !p.slotIsEntity[id] == true
		}() == true {
			__antithesis_instrumentation__.Notify(578687)
			p.slotIsEntity[id] = entity
		} else {
			__antithesis_instrumentation__.Notify(578688)
		}
		__antithesis_instrumentation__.Notify(578685)
		return id
	} else {
		__antithesis_instrumentation__.Notify(578689)
	}
	__antithesis_instrumentation__.Notify(578683)
	id = p.fillSlot(slot{}, entity)
	p.variables = append(p.variables, v)
	p.variableSlots[v] = id
	return id
}

func (p *queryBuilder) fillSlot(sd slot, isEntity bool) slotIdx {
	__antithesis_instrumentation__.Notify(578690)
	s := slotIdx(len(p.slots))
	p.slots = append(p.slots, sd)
	p.slotIsEntity = append(p.slotIsEntity, isEntity)
	return s
}

func (p *queryBuilder) findEntitySlots() (entitySlots []slotIdx) {
	__antithesis_instrumentation__.Notify(578691)
	for i := range p.slots {
		__antithesis_instrumentation__.Notify(578693)
		if p.slotIsEntity[i] {
			__antithesis_instrumentation__.Notify(578694)
			entitySlots = append(entitySlots, slotIdx(i))
		} else {
			__antithesis_instrumentation__.Notify(578695)
		}
	}
	__antithesis_instrumentation__.Notify(578692)
	return entitySlots
}

func (p *queryBuilder) typeCheck(f fact) {
	__antithesis_instrumentation__.Notify(578696)
	s := &p.slots[f.value]
	if s.empty() && func() bool {
		__antithesis_instrumentation__.Notify(578698)
		return s.any == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(578699)
		return
	} else {
		__antithesis_instrumentation__.Notify(578700)
	}
	__antithesis_instrumentation__.Notify(578697)
	switch f.attr {
	case p.sc.mustGetOrdinal(Type):
		__antithesis_instrumentation__.Notify(578701)
		checkSlotType(s, reflectTypeType)
	default:
		__antithesis_instrumentation__.Notify(578702)
		checkSlotType(s, p.sc.attrTypes[f.attr])
	}
}

var boolType = reflect.TypeOf((*bool)(nil)).Elem()

func checkSlotType(s *slot, exp reflect.Type) {
	__antithesis_instrumentation__.Notify(578703)
	if !s.empty() {
		__antithesis_instrumentation__.Notify(578705)
		if err := checkType(s.typ, exp); err != nil {
			__antithesis_instrumentation__.Notify(578706)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(578707)
		}
	} else {
		__antithesis_instrumentation__.Notify(578708)
	}
	__antithesis_instrumentation__.Notify(578704)
	for i := range s.any {
		__antithesis_instrumentation__.Notify(578709)
		if err := checkType(s.any[i].typ, exp); err != nil {
			__antithesis_instrumentation__.Notify(578710)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(578711)
		}
	}
}
