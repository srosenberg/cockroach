package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "reflect"

type Clause interface {
	clause()
}

func newTriple(entity Var, a Attr, value expr) Clause {
	__antithesis_instrumentation__.Notify(578883)
	return &tripleDecl{
		entity:    entity,
		attribute: a,
		value:     value,
	}
}

func (v Var) AttrEq(a Attr, value interface{}) Clause {
	__antithesis_instrumentation__.Notify(578884)
	return newTriple(v, a, valueExpr{value: value})
}

func (v Var) AttrIn(a Attr, values ...interface{}) Clause {
	__antithesis_instrumentation__.Notify(578885)
	return newTriple(v, a, (anyExpr)(values))
}

func (v Var) AttrEqVar(a Attr, value Var) Clause {
	__antithesis_instrumentation__.Notify(578886)
	return newTriple(v, a, value)
}

func (v Var) Eq(value interface{}) Clause {
	__antithesis_instrumentation__.Notify(578887)
	return &eqDecl{v, valueExpr{value: value}}
}

func (v Var) In(disjuncts ...interface{}) Clause {
	__antithesis_instrumentation__.Notify(578888)
	return &eqDecl{v, (anyExpr)(disjuncts)}
}

func (v Var) Type(valueForTypeOf interface{}, moreValuesForTypeOf ...interface{}) Clause {
	__antithesis_instrumentation__.Notify(578889)
	typ := reflect.TypeOf(valueForTypeOf)
	if len(moreValuesForTypeOf) == 0 {
		__antithesis_instrumentation__.Notify(578892)
		return v.AttrEq(Type, typ)
	} else {
		__antithesis_instrumentation__.Notify(578893)
	}
	__antithesis_instrumentation__.Notify(578890)

	types := make([]interface{}, 0, len(moreValuesForTypeOf)+1)
	types = append(types, typ)
	for _, v := range moreValuesForTypeOf {
		__antithesis_instrumentation__.Notify(578894)
		types = append(types, reflect.TypeOf(v))
	}
	__antithesis_instrumentation__.Notify(578891)
	return v.AttrIn(Type, types...)
}

func (v Var) Entities(attr Attr, entities ...Var) Clause {
	__antithesis_instrumentation__.Notify(578895)
	terms := make([]Clause, len(entities))
	for i, e := range entities {
		__antithesis_instrumentation__.Notify(578897)
		terms[i] = e.AttrEqVar(attr, v)
	}
	__antithesis_instrumentation__.Notify(578896)
	return And(terms...)
}

func And(terms ...Clause) Clause {
	__antithesis_instrumentation__.Notify(578898)
	return (and)(terms)
}

func Filter(name string, vars ...Var) func(predicateFunc interface{}) Clause {
	__antithesis_instrumentation__.Notify(578899)
	return func(predicateFunc interface{}) Clause {
		__antithesis_instrumentation__.Notify(578900)
		return &filterDecl{
			name:          name,
			vars:          vars,
			predicateFunc: predicateFunc,
		}
	}
}
