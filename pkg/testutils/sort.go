package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/cockroachdb/errors"
)

var _ sort.Interface = structSorter{}

type structSorter struct {
	v          reflect.Value
	fieldNames []string
}

func (ss structSorter) Len() int {
	__antithesis_instrumentation__.Notify(646037)
	return ss.v.Len()
}

func (ss structSorter) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(646038)
	v1 := reflect.Indirect(ss.v.Index(i))
	v2 := reflect.Indirect(ss.v.Index(j))
	return ss.fieldIsLess(v1, v2, 0)
}

func (ss structSorter) fieldIsLess(v1, v2 reflect.Value, fieldNum int) bool {
	__antithesis_instrumentation__.Notify(646039)
	fieldName := ss.fieldNames[fieldNum]
	lastField := len(ss.fieldNames) == fieldNum+1

	f1 := v1.FieldByName(fieldName)
	if !f1.IsValid() {
		__antithesis_instrumentation__.Notify(646043)
		panic(fmt.Sprintf("couldn't get field %s", fieldName))
	} else {
		__antithesis_instrumentation__.Notify(646044)
	}
	__antithesis_instrumentation__.Notify(646040)
	f2 := v2.FieldByName(fieldName)
	if !f2.IsValid() {
		__antithesis_instrumentation__.Notify(646045)
		panic(fmt.Sprintf("couldn't get field %s", fieldName))
	} else {
		__antithesis_instrumentation__.Notify(646046)
	}
	__antithesis_instrumentation__.Notify(646041)

	switch f1.Kind() {
	case reflect.String:
		__antithesis_instrumentation__.Notify(646047)
		if !lastField && func() bool {
			__antithesis_instrumentation__.Notify(646058)
			return f1.String() == f2.String() == true
		}() == true {
			__antithesis_instrumentation__.Notify(646059)
			return ss.fieldIsLess(v1, v2, fieldNum+1)
		} else {
			__antithesis_instrumentation__.Notify(646060)
		}
		__antithesis_instrumentation__.Notify(646048)
		return f1.String() < f2.String()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		__antithesis_instrumentation__.Notify(646049)
		if !lastField && func() bool {
			__antithesis_instrumentation__.Notify(646061)
			return f1.Int() == f2.Int() == true
		}() == true {
			__antithesis_instrumentation__.Notify(646062)
			return ss.fieldIsLess(v1, v2, fieldNum+1)
		} else {
			__antithesis_instrumentation__.Notify(646063)
		}
		__antithesis_instrumentation__.Notify(646050)
		return f1.Int() < f2.Int()

	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		__antithesis_instrumentation__.Notify(646051)
		if !lastField && func() bool {
			__antithesis_instrumentation__.Notify(646064)
			return f1.Uint() == f2.Uint() == true
		}() == true {
			__antithesis_instrumentation__.Notify(646065)
			return ss.fieldIsLess(v1, v2, fieldNum+1)
		} else {
			__antithesis_instrumentation__.Notify(646066)
		}
		__antithesis_instrumentation__.Notify(646052)
		return f1.Uint() < f2.Uint()

	case reflect.Float32, reflect.Float64:
		__antithesis_instrumentation__.Notify(646053)
		if !lastField && func() bool {
			__antithesis_instrumentation__.Notify(646067)
			return f1.Float() == f2.Float() == true
		}() == true {
			__antithesis_instrumentation__.Notify(646068)
			return ss.fieldIsLess(v1, v2, fieldNum+1)
		} else {
			__antithesis_instrumentation__.Notify(646069)
		}
		__antithesis_instrumentation__.Notify(646054)
		return f1.Float() < f2.Float()

	case reflect.Bool:
		__antithesis_instrumentation__.Notify(646055)
		if !lastField && func() bool {
			__antithesis_instrumentation__.Notify(646070)
			return f1.Bool() == f2.Bool() == true
		}() == true {
			__antithesis_instrumentation__.Notify(646071)
			return ss.fieldIsLess(v1, v2, fieldNum+1)
		} else {
			__antithesis_instrumentation__.Notify(646072)
		}
		__antithesis_instrumentation__.Notify(646056)
		return !f1.Bool() && func() bool {
			__antithesis_instrumentation__.Notify(646073)
			return f2.Bool() == true
		}() == true
	default:
		__antithesis_instrumentation__.Notify(646057)
	}
	__antithesis_instrumentation__.Notify(646042)

	panic(fmt.Sprintf("can't handle sort key type %d", uint(f1.Kind())))
}

func (ss structSorter) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(646074)

	t := reflect.ValueOf(ss.v.Index(i).Interface())
	ss.v.Index(i).Set(ss.v.Index(j))
	ss.v.Index(j).Set(t)
}

func SortStructs(s interface{}, fieldNames ...string) {
	__antithesis_instrumentation__.Notify(646075)

	structs := reflect.ValueOf(s)
	if structs.Kind() != reflect.Slice {
		__antithesis_instrumentation__.Notify(646079)
		panic(errors.AssertionFailedf("expected slice, got %T", s))
	} else {
		__antithesis_instrumentation__.Notify(646080)
	}
	__antithesis_instrumentation__.Notify(646076)
	elemType := structs.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		__antithesis_instrumentation__.Notify(646081)
		elemType = elemType.Elem()
	} else {
		__antithesis_instrumentation__.Notify(646082)
	}
	__antithesis_instrumentation__.Notify(646077)
	if elemType.Kind() != reflect.Struct {
		__antithesis_instrumentation__.Notify(646083)
		panic(errors.AssertionFailedf("%s is not a struct or pointer to struct", structs.Elem()))
	} else {
		__antithesis_instrumentation__.Notify(646084)
	}
	__antithesis_instrumentation__.Notify(646078)

	sort.Sort(structSorter{structs, fieldNames})
}
