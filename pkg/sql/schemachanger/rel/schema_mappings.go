package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "reflect"

type SchemaOption interface {
	apply(*schemaMappings)
}

type EntityMappingOption interface {
	apply(typeMappings *entityMapping)
}

func AttrType(a Attr, typ reflect.Type) SchemaOption {
	__antithesis_instrumentation__.Notify(579232)
	return attrType{a: a, typ: typ}
}

func EntityMapping(typ reflect.Type, opts ...EntityMappingOption) SchemaOption {
	__antithesis_instrumentation__.Notify(579233)
	tm := entityMapping{typ: typ}
	for _, o := range opts {
		__antithesis_instrumentation__.Notify(579235)
		o.apply(&tm)
	}
	__antithesis_instrumentation__.Notify(579234)
	return tm
}

func EntityAttr(a Attr, selectors ...string) EntityMappingOption {
	__antithesis_instrumentation__.Notify(579236)
	return attrMapping{a: a, selectors: selectors}
}

type schemaMappings struct {
	attrTypes []attrType

	entityMappings []entityMapping
}

type attrType struct {
	a   Attr
	typ reflect.Type
}

func (a attrType) apply(mappings *schemaMappings) {
	__antithesis_instrumentation__.Notify(579237)
	mappings.attrTypes = append(mappings.attrTypes, a)
}

type entityMapping struct {
	typ          reflect.Type
	attrMappings []attrMapping
}

func (t entityMapping) apply(mappings *schemaMappings) {
	__antithesis_instrumentation__.Notify(579238)
	mappings.entityMappings = append(mappings.entityMappings, t)
}

type attrMapping struct {
	a         Attr
	selectors []string
}

func (a attrMapping) apply(tm *entityMapping) {
	__antithesis_instrumentation__.Notify(579239)
	tm.attrMappings = append(tm.attrMappings, a)
}
