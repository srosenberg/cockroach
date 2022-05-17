package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"
	"sort"
	"strings"
	"unsafe"

	"github.com/cockroachdb/errors"
)

type Schema struct {
	name              string
	attrs             []Attr
	attrTypes         []reflect.Type
	attrToOrdinal     map[Attr]ordinal
	entityTypeSchemas map[reflect.Type]*entityTypeSchema
}

func NewSchema(name string, m ...SchemaOption) (_ *Schema, err error) {
	__antithesis_instrumentation__.Notify(579115)
	defer func() {
		__antithesis_instrumentation__.Notify(579117)
		switch r := recover().(type) {
		case nil:
			__antithesis_instrumentation__.Notify(579118)
			return
		case error:
			__antithesis_instrumentation__.Notify(579119)
			err = errors.Wrap(r, "failed to construct schema")
		default:
			__antithesis_instrumentation__.Notify(579120)
			err = errors.AssertionFailedf("failed to construct schema: %v", r)
		}
	}()
	__antithesis_instrumentation__.Notify(579116)
	sc := buildSchema(name, m...)
	return sc, nil
}

func MustSchema(name string, m ...SchemaOption) *Schema {
	__antithesis_instrumentation__.Notify(579121)
	return buildSchema(name, m...)
}

type entityTypeSchema struct {
	typ        reflect.Type
	fields     []fieldInfo
	attrFields map[ordinal][]fieldInfo
}

type fieldInfo struct {
	path            string
	typ             reflect.Type
	attr            ordinal
	comparableValue func(unsafe.Pointer) interface{}
	value           func(unsafe.Pointer) interface{}
	isPtr, isEntity bool
}

func buildSchema(name string, opts ...SchemaOption) *Schema {
	__antithesis_instrumentation__.Notify(579122)
	var m schemaMappings
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(579126)
		opt.apply(&m)
	}
	__antithesis_instrumentation__.Notify(579123)
	sb := &schemaBuilder{
		Schema: &Schema{
			name:              name,
			attrToOrdinal:     make(map[Attr]ordinal),
			entityTypeSchemas: make(map[reflect.Type]*entityTypeSchema),
		},
		m: m,
	}

	sb.maybeAddAttribute(Self, emptyInterfaceType)
	sb.maybeAddAttribute(Type, reflectTypeType)
	for _, t := range m.attrTypes {
		__antithesis_instrumentation__.Notify(579127)
		sb.maybeAddAttribute(t.a, t.typ)
	}
	__antithesis_instrumentation__.Notify(579124)

	for _, tm := range m.entityMappings {
		__antithesis_instrumentation__.Notify(579128)
		sb.maybeAddTypeMapping(tm.typ, tm.attrMappings)
	}
	__antithesis_instrumentation__.Notify(579125)
	return sb.Schema
}

type schemaBuilder struct {
	*Schema
	m schemaMappings
}

func (sb *schemaBuilder) maybeAddAttribute(a Attr, typ reflect.Type) ordinal {
	__antithesis_instrumentation__.Notify(579129)

	ord, exists := sb.attrToOrdinal[a]
	if !exists {
		__antithesis_instrumentation__.Notify(579132)
		ord = ordinal(len(sb.attrs))
		if ord >= maxUserAttribute {
			__antithesis_instrumentation__.Notify(579134)
			panic(errors.Errorf("too many attributes"))
		} else {
			__antithesis_instrumentation__.Notify(579135)
		}
		__antithesis_instrumentation__.Notify(579133)
		sb.attrs = append(sb.attrs, a)
		sb.attrTypes = append(sb.attrTypes, typ)
		sb.attrToOrdinal[a] = ord
		return ord
	} else {
		__antithesis_instrumentation__.Notify(579136)
	}
	__antithesis_instrumentation__.Notify(579130)
	prev := sb.attrTypes[ord]
	if err := checkType(typ, prev); err != nil {
		__antithesis_instrumentation__.Notify(579137)
		panic(errors.Wrapf(err, "type mismatch for %v", a))
	} else {
		__antithesis_instrumentation__.Notify(579138)
	}
	__antithesis_instrumentation__.Notify(579131)
	return ord
}

func checkType(typ, exp reflect.Type) error {
	__antithesis_instrumentation__.Notify(579139)
	switch exp.Kind() {
	case reflect.Interface:
		__antithesis_instrumentation__.Notify(579141)
		if !typ.Implements(exp) {
			__antithesis_instrumentation__.Notify(579143)
			return errors.Errorf("%v does not implement %v", typ, exp)
		} else {
			__antithesis_instrumentation__.Notify(579144)
		}
	default:
		__antithesis_instrumentation__.Notify(579142)
		if typ != exp {
			__antithesis_instrumentation__.Notify(579145)
			return errors.Errorf("%v is not %v", typ, exp)
		} else {
			__antithesis_instrumentation__.Notify(579146)
		}
	}
	__antithesis_instrumentation__.Notify(579140)
	return nil
}

func (sb *schemaBuilder) maybeAddTypeMapping(t reflect.Type, attributeMappings []attrMapping) {
	__antithesis_instrumentation__.Notify(579147)
	isStructPointer := func(tt reflect.Type) bool {
		__antithesis_instrumentation__.Notify(579153)
		return tt.Kind() == reflect.Ptr && func() bool {
			__antithesis_instrumentation__.Notify(579154)
			return tt.Elem().Kind() == reflect.Struct == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(579148)

	if !isStructPointer(t) {
		__antithesis_instrumentation__.Notify(579155)
		panic(errors.Errorf("%v is not a pointer to a struct", t))
	} else {
		__antithesis_instrumentation__.Notify(579156)
	}
	__antithesis_instrumentation__.Notify(579149)
	var fieldInfos []fieldInfo
	for _, am := range attributeMappings {
		__antithesis_instrumentation__.Notify(579157)
		for _, sel := range am.selectors {
			__antithesis_instrumentation__.Notify(579158)
			fieldInfos = append(fieldInfos,
				sb.addTypeAttrMapping(am.a, t, sel))
		}
	}
	__antithesis_instrumentation__.Notify(579150)
	sort.Slice(fieldInfos, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(579159)
		return fieldInfos[i].attr < fieldInfos[j].attr
	})
	__antithesis_instrumentation__.Notify(579151)
	attributeFields := make(map[ordinal][]fieldInfo)

	for i := 0; i < len(fieldInfos); {
		__antithesis_instrumentation__.Notify(579160)
		cur := fieldInfos[i].attr
		j := i + 1
		for ; j < len(fieldInfos); j++ {
			__antithesis_instrumentation__.Notify(579162)
			if fieldInfos[j].attr != cur {
				__antithesis_instrumentation__.Notify(579163)
				break
			} else {
				__antithesis_instrumentation__.Notify(579164)
			}
		}
		__antithesis_instrumentation__.Notify(579161)
		attributeFields[cur] = fieldInfos[i:j]
		i = j
	}
	__antithesis_instrumentation__.Notify(579152)
	sb.entityTypeSchemas[t] = &entityTypeSchema{
		typ:        t,
		fields:     fieldInfos,
		attrFields: attributeFields,
	}
}

func (sb *schemaBuilder) addTypeAttrMapping(a Attr, t reflect.Type, sel string) fieldInfo {
	__antithesis_instrumentation__.Notify(579165)
	offset, cur := getOffsetAndTypeFromSelector(t, sel)

	isPtr := cur.Kind() == reflect.Ptr
	isStructPtr := isPtr && func() bool {
		__antithesis_instrumentation__.Notify(579170)
		return cur.Elem().Kind() == reflect.Struct == true
	}() == true
	isScalarPtr := isPtr && func() bool {
		__antithesis_instrumentation__.Notify(579171)
		return isSupportScalarKind(cur.Elem().Kind()) == true
	}() == true
	if !isScalarPtr && func() bool {
		__antithesis_instrumentation__.Notify(579172)
		return !isStructPtr == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(579173)
		return !isSupportScalarKind(cur.Kind()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(579174)
		panic(errors.Errorf(
			"selector %q of %v has unsupported type %v",
			sel, t, cur,
		))
	} else {
		__antithesis_instrumentation__.Notify(579175)
	}
	__antithesis_instrumentation__.Notify(579166)

	typ := cur
	if isScalarPtr {
		__antithesis_instrumentation__.Notify(579176)
		typ = cur.Elem()
	} else {
		__antithesis_instrumentation__.Notify(579177)
	}
	__antithesis_instrumentation__.Notify(579167)
	ord := sb.maybeAddAttribute(a, typ)

	f := fieldInfo{
		path:     sel,
		attr:     ord,
		isEntity: isStructPtr,
		isPtr:    isPtr,
		typ:      typ,
	}
	makeValueGetter := func(t reflect.Type, offset uintptr) func(u unsafe.Pointer) reflect.Value {
		__antithesis_instrumentation__.Notify(579178)
		return func(u unsafe.Pointer) reflect.Value {
			__antithesis_instrumentation__.Notify(579179)
			return reflect.NewAt(t, unsafe.Pointer(uintptr(u)+offset))
		}
	}
	__antithesis_instrumentation__.Notify(579168)
	getPtrValue := func(vg func(pointer unsafe.Pointer) reflect.Value) func(u unsafe.Pointer) interface{} {
		__antithesis_instrumentation__.Notify(579180)
		return func(u unsafe.Pointer) interface{} {
			__antithesis_instrumentation__.Notify(579181)
			got := vg(u)
			if got.Elem().IsNil() {
				__antithesis_instrumentation__.Notify(579183)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(579184)
			}
			__antithesis_instrumentation__.Notify(579182)
			return got.Elem().Interface()
		}
	}
	{
		__antithesis_instrumentation__.Notify(579185)
		vg := makeValueGetter(cur, offset)
		if isStructPtr {
			__antithesis_instrumentation__.Notify(579186)
			f.value = getPtrValue(vg)
		} else {
			__antithesis_instrumentation__.Notify(579187)
			if isScalarPtr {
				__antithesis_instrumentation__.Notify(579188)
				f.value = func(u unsafe.Pointer) interface{} {
					__antithesis_instrumentation__.Notify(579189)
					got := vg(u)
					if got.Elem().IsNil() {
						__antithesis_instrumentation__.Notify(579191)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(579192)
					}
					__antithesis_instrumentation__.Notify(579190)
					return got.Elem().Elem().Interface()
				}
			} else {
				__antithesis_instrumentation__.Notify(579193)
				f.value = func(u unsafe.Pointer) interface{} {
					__antithesis_instrumentation__.Notify(579194)
					return vg(u).Elem().Interface()
				}
			}
		}
	}
	{
		__antithesis_instrumentation__.Notify(579195)
		if isStructPtr {
			__antithesis_instrumentation__.Notify(579196)
			f.comparableValue = getPtrValue(makeValueGetter(cur, offset))
		} else {
			__antithesis_instrumentation__.Notify(579197)
			compType := getComparableType(typ)
			if isScalarPtr {
				__antithesis_instrumentation__.Notify(579199)
				compType = reflect.PtrTo(compType)
			} else {
				__antithesis_instrumentation__.Notify(579200)
			}
			__antithesis_instrumentation__.Notify(579198)
			vg := makeValueGetter(compType, offset)
			if isScalarPtr {
				__antithesis_instrumentation__.Notify(579201)
				f.comparableValue = getPtrValue(vg)
			} else {
				__antithesis_instrumentation__.Notify(579202)
				f.comparableValue = func(u unsafe.Pointer) interface{} {
					__antithesis_instrumentation__.Notify(579203)
					return vg(u).Interface()
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(579169)
	return f
}

func getOffsetAndTypeFromSelector(
	structPointer reflect.Type, selector string,
) (uintptr, reflect.Type) {
	__antithesis_instrumentation__.Notify(579204)
	names := strings.Split(selector, ".")
	var offset uintptr
	cur := structPointer.Elem()
	for _, n := range names {
		__antithesis_instrumentation__.Notify(579206)
		sf, ok := cur.FieldByName(n)
		if !ok {
			__antithesis_instrumentation__.Notify(579208)
			panic(errors.Errorf("%v.%s is not a field", structPointer, selector))
		} else {
			__antithesis_instrumentation__.Notify(579209)
		}
		__antithesis_instrumentation__.Notify(579207)
		offset += sf.Offset
		cur = sf.Type
	}
	__antithesis_instrumentation__.Notify(579205)
	return offset, cur
}

func (sc *Schema) mustGetOrdinal(attribute Attr) ordinal {
	__antithesis_instrumentation__.Notify(579210)
	ord, err := sc.getOrdinal(attribute)
	if err != nil {
		__antithesis_instrumentation__.Notify(579212)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579213)
	}
	__antithesis_instrumentation__.Notify(579211)
	return ord
}

func (sc *Schema) getOrdinal(attribute Attr) (ordinal, error) {
	__antithesis_instrumentation__.Notify(579214)
	ord, ok := sc.attrToOrdinal[attribute]
	if !ok {
		__antithesis_instrumentation__.Notify(579216)
		return 0, errors.Errorf("unknown attribute %s in schema %s", attribute, sc.name)
	} else {
		__antithesis_instrumentation__.Notify(579217)
	}
	__antithesis_instrumentation__.Notify(579215)
	return ord, nil
}
