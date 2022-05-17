package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/errors"
)

type entity valuesMap

func (sc *Schema) EqualOn(attrs []Attr, a, b interface{}) (eq bool) {
	__antithesis_instrumentation__.Notify(578454)
	_, eq = sc.CompareOn(attrs, a, b)
	return eq
}

func (sc *Schema) CompareOn(attrs []Attr, a, b interface{}) (less, eq bool) {
	__antithesis_instrumentation__.Notify(578455)
	ae, err := toEntity(sc, a)
	if err != nil {
		__antithesis_instrumentation__.Notify(578459)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(578460)
	}
	__antithesis_instrumentation__.Notify(578456)
	be, err := toEntity(sc, b)
	if err != nil {
		__antithesis_instrumentation__.Notify(578461)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(578462)
	}
	__antithesis_instrumentation__.Notify(578457)
	defer putValues((*valuesMap)(ae))
	defer putValues((*valuesMap)(be))
	for _, a := range attrs {
		__antithesis_instrumentation__.Notify(578463)
		ord, err := sc.getOrdinal(a)
		if err != nil {
			__antithesis_instrumentation__.Notify(578465)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(578466)
		}
		__antithesis_instrumentation__.Notify(578464)
		if less, eq = compareOn(ord, (*valuesMap)(ae), (*valuesMap)(be)); !eq {
			__antithesis_instrumentation__.Notify(578467)
			return less, eq
		} else {
			__antithesis_instrumentation__.Notify(578468)
		}
	}
	__antithesis_instrumentation__.Notify(578458)
	return false, true
}

func (sc *Schema) IterateAttributes(
	entityI interface{}, f func(attribute Attr, value interface{}) error,
) (err error) {
	__antithesis_instrumentation__.Notify(578469)

	v, err := toEntity(sc, entityI)
	if err != nil {
		__antithesis_instrumentation__.Notify(578473)
		return err
	} else {
		__antithesis_instrumentation__.Notify(578474)
	}
	__antithesis_instrumentation__.Notify(578470)
	v.attrs.forEach(func(ord ordinal) (wantMore bool) {
		__antithesis_instrumentation__.Notify(578475)
		a := sc.attrs[ord]
		if isSystemAttribute(a) {
			__antithesis_instrumentation__.Notify(578478)
			return true
		} else {
			__antithesis_instrumentation__.Notify(578479)
		}
		__antithesis_instrumentation__.Notify(578476)
		tv, ok := v.getTypedValue(sc, ord)
		if !ok {
			__antithesis_instrumentation__.Notify(578480)
			err = errors.AssertionFailedf(
				"failed to get typed value for populated scalar attribute %v for %T",
				a, entityI,
			)
		} else {
			__antithesis_instrumentation__.Notify(578481)
			err = f(a, tv.toInterface())
		}
		__antithesis_instrumentation__.Notify(578477)
		return err == nil
	})
	__antithesis_instrumentation__.Notify(578471)
	if iterutil.Done(err) {
		__antithesis_instrumentation__.Notify(578482)
		err = nil
	} else {
		__antithesis_instrumentation__.Notify(578483)
	}
	__antithesis_instrumentation__.Notify(578472)
	return err
}

func (e *entity) getComparableValue(sc *Schema, attribute Attr) interface{} {
	__antithesis_instrumentation__.Notify(578484)
	return (*valuesMap)(e).get(sc.mustGetOrdinal(attribute))
}

func (e *entity) getTypeInfo(sc *Schema) *entityTypeSchema {
	__antithesis_instrumentation__.Notify(578485)
	return sc.entityTypeSchemas[e.getComparableValue(sc, Type).(reflect.Type)]
}

func (e *entity) getTypedValue(sc *Schema, attr ordinal) (typedValue, bool) {
	__antithesis_instrumentation__.Notify(578486)
	val := (*valuesMap)(e).get(attr)
	if val == nil {
		__antithesis_instrumentation__.Notify(578489)
		return typedValue{}, false
	} else {
		__antithesis_instrumentation__.Notify(578490)
	}
	__antithesis_instrumentation__.Notify(578487)
	var typ reflect.Type

	if sc.attrs[attr] == Type {
		__antithesis_instrumentation__.Notify(578491)
		typ = reflectTypeType
	} else {
		__antithesis_instrumentation__.Notify(578492)
		if fi, ok := e.getTypeInfo(sc).attrFields[attr]; ok && func() bool {
			__antithesis_instrumentation__.Notify(578493)
			return !fi[0].isEntity == true
		}() == true {
			__antithesis_instrumentation__.Notify(578494)

			typ = fi[0].typ
		} else {
			__antithesis_instrumentation__.Notify(578495)
			typ = reflect.TypeOf(val)
		}
	}
	__antithesis_instrumentation__.Notify(578488)
	return typedValue{
		typ:   typ,
		value: val,
	}, true
}

func (e *entity) asMap() *valuesMap {
	__antithesis_instrumentation__.Notify(578496)
	return (*valuesMap)(e)
}

func toEntity(s *Schema, v interface{}) (*entity, error) {
	__antithesis_instrumentation__.Notify(578497)
	ti, value, err := getEntityValueInfo(s, v)
	if err != nil {
		__antithesis_instrumentation__.Notify(578500)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(578501)
	}
	__antithesis_instrumentation__.Notify(578498)

	e := getValues()
	e.add(s.mustGetOrdinal(Type), value.Type())
	e.add(s.mustGetOrdinal(Self), v)
	for _, field := range ti.fields {
		__antithesis_instrumentation__.Notify(578502)

		var val interface{}
		if field.isEntity {
			__antithesis_instrumentation__.Notify(578506)
			val = field.value(unsafe.Pointer(value.Pointer()))
		} else {
			__antithesis_instrumentation__.Notify(578507)
			val = field.comparableValue(unsafe.Pointer(value.Pointer()))
		}
		__antithesis_instrumentation__.Notify(578503)
		if field.isPtr && func() bool {
			__antithesis_instrumentation__.Notify(578508)
			return val == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(578509)
			continue
		} else {
			__antithesis_instrumentation__.Notify(578510)
			if val == nil {
				__antithesis_instrumentation__.Notify(578511)
				return nil, errors.AssertionFailedf(
					"got nil value for non-pointer scalar attribute %s of type %s",
					s.attrs[field.attr], ti.typ,
				)
			} else {
				__antithesis_instrumentation__.Notify(578512)
			}
		}
		__antithesis_instrumentation__.Notify(578504)
		if e.attrs.contains(field.attr) {
			__antithesis_instrumentation__.Notify(578513)
			return nil, errors.Errorf(
				"%v contains second non-nil entry for %v at %s",
				ti.typ, s.attrs[field.attr], field.path,
			)
		} else {
			__antithesis_instrumentation__.Notify(578514)
		}
		__antithesis_instrumentation__.Notify(578505)
		e.add(field.attr, val)
	}
	__antithesis_instrumentation__.Notify(578499)

	return (*entity)(e), nil
}

func getEntityValueInfo(s *Schema, v interface{}) (*entityTypeSchema, reflect.Value, error) {
	__antithesis_instrumentation__.Notify(578515)
	vv := reflect.ValueOf(v)
	if !vv.IsValid() {
		__antithesis_instrumentation__.Notify(578519)
		return nil, reflect.Value{}, errors.Errorf("invalid nil value")
	} else {
		__antithesis_instrumentation__.Notify(578520)
	}
	__antithesis_instrumentation__.Notify(578516)
	t, ok := s.entityTypeSchemas[vv.Type()]
	if !ok {
		__antithesis_instrumentation__.Notify(578521)
		return nil, reflect.Value{}, errors.Errorf("unknown type handler for %T", v)
	} else {
		__antithesis_instrumentation__.Notify(578522)
	}
	__antithesis_instrumentation__.Notify(578517)

	if vv.IsNil() {
		__antithesis_instrumentation__.Notify(578523)
		return nil, reflect.Value{}, errors.Errorf("invalid nil %T value", v)
	} else {
		__antithesis_instrumentation__.Notify(578524)
	}
	__antithesis_instrumentation__.Notify(578518)
	return t, vv, nil
}
