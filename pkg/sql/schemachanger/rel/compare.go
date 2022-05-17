package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/errors"
)

func compare(a, b interface{}) (less, eq bool) {
	__antithesis_instrumentation__.Notify(578294)

	switch a := a.(type) {
	case *int:
		__antithesis_instrumentation__.Notify(578295)
		b := b.(*int)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578322)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578323)
		}
		__antithesis_instrumentation__.Notify(578296)
		return false, *a == *b
	case *int64:
		__antithesis_instrumentation__.Notify(578297)
		b := b.(*int64)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578324)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578325)
		}
		__antithesis_instrumentation__.Notify(578298)
		return false, *a == *b
	case *int32:
		__antithesis_instrumentation__.Notify(578299)
		b := b.(*int32)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578326)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578327)
		}
		__antithesis_instrumentation__.Notify(578300)
		return false, *a == *b
	case *int16:
		__antithesis_instrumentation__.Notify(578301)
		b := b.(*int16)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578328)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578329)
		}
		__antithesis_instrumentation__.Notify(578302)
		return false, *a == *b
	case *int8:
		__antithesis_instrumentation__.Notify(578303)
		b := b.(*int8)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578330)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578331)
		}
		__antithesis_instrumentation__.Notify(578304)
		return false, *a == *b
	case *uint:
		__antithesis_instrumentation__.Notify(578305)
		b := b.(*uint)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578332)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578333)
		}
		__antithesis_instrumentation__.Notify(578306)
		return false, *a == *b
	case *uint64:
		__antithesis_instrumentation__.Notify(578307)
		b := b.(*uint64)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578334)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578335)
		}
		__antithesis_instrumentation__.Notify(578308)
		return false, *a == *b
	case *uint32:
		__antithesis_instrumentation__.Notify(578309)
		b := b.(*uint32)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578336)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578337)
		}
		__antithesis_instrumentation__.Notify(578310)
		return false, *a == *b
	case *uint16:
		__antithesis_instrumentation__.Notify(578311)
		b := b.(*uint16)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578338)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578339)
		}
		__antithesis_instrumentation__.Notify(578312)
		return false, *a == *b
	case *uint8:
		__antithesis_instrumentation__.Notify(578313)
		b := b.(*uint8)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578340)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578341)
		}
		__antithesis_instrumentation__.Notify(578314)
		return false, *a == *b
	case *string:
		__antithesis_instrumentation__.Notify(578315)
		b := b.(*string)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578342)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578343)
		}
		__antithesis_instrumentation__.Notify(578316)
		return false, *a == *b
	case *uintptr:
		__antithesis_instrumentation__.Notify(578317)
		b := b.(*uintptr)
		if *a < *b {
			__antithesis_instrumentation__.Notify(578344)
			return true, false
		} else {
			__antithesis_instrumentation__.Notify(578345)
		}
		__antithesis_instrumentation__.Notify(578318)
		return false, *a == *b
	case reflect.Type:
		__antithesis_instrumentation__.Notify(578319)
		b := b.(reflect.Type)
		switch {
		case a == b:
			__antithesis_instrumentation__.Notify(578346)
			return false, true
		case a.PkgPath() == b.PkgPath():
			__antithesis_instrumentation__.Notify(578347)
			return a.String() < b.String(), false
		default:
			__antithesis_instrumentation__.Notify(578348)
			return a.PkgPath() < b.PkgPath(), false
		}
	default:
		__antithesis_instrumentation__.Notify(578320)

		av := reflect.ValueOf(a)
		bv := reflect.ValueOf(b)
		if av.Type().Kind() != reflect.Ptr || func() bool {
			__antithesis_instrumentation__.Notify(578349)
			return av.Type().Elem().Kind() != reflect.Struct == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(578350)
			return bv.Type().Kind() != reflect.Ptr == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(578351)
			return bv.Type().Elem().Kind() != reflect.Struct == true
		}() == true {
			__antithesis_instrumentation__.Notify(578352)
			panic(errors.AssertionFailedf("incomparable types %T and %T", a, b))
		} else {
			__antithesis_instrumentation__.Notify(578353)
		}
		__antithesis_instrumentation__.Notify(578321)
		ap, bp := av.Pointer(), bv.Pointer()
		return ap < bp, ap == bp
	}
}

var kindTypeMap = map[reflect.Kind]reflect.Type{
	reflect.Int:     reflect.TypeOf((*int)(nil)).Elem(),
	reflect.Int64:   reflect.TypeOf((*int64)(nil)).Elem(),
	reflect.Int32:   reflect.TypeOf((*int32)(nil)).Elem(),
	reflect.Int16:   reflect.TypeOf((*int16)(nil)).Elem(),
	reflect.Int8:    reflect.TypeOf((*int8)(nil)).Elem(),
	reflect.Uint:    reflect.TypeOf((*uint)(nil)).Elem(),
	reflect.Uint64:  reflect.TypeOf((*uint64)(nil)).Elem(),
	reflect.Uint32:  reflect.TypeOf((*uint32)(nil)).Elem(),
	reflect.Uint16:  reflect.TypeOf((*uint16)(nil)).Elem(),
	reflect.Uint8:   reflect.TypeOf((*uint8)(nil)).Elem(),
	reflect.Uintptr: reflect.TypeOf((*uintptr)(nil)).Elem(),
	reflect.String:  reflect.TypeOf((*string)(nil)).Elem(),
}

func isSupportScalarKind(kind reflect.Kind) bool {
	__antithesis_instrumentation__.Notify(578354)
	_, ok := kindTypeMap[kind]
	return kind != reflect.Ptr && func() bool {
		__antithesis_instrumentation__.Notify(578355)
		return ok == true
	}() == true
}

func getComparableType(t reflect.Type) reflect.Type {
	__antithesis_instrumentation__.Notify(578356)
	ct, ok := kindTypeMap[t.Kind()]
	if !ok {
		__antithesis_instrumentation__.Notify(578358)
		panic(errors.AssertionFailedf(
			"unsupported type %T of kind %v",
			t, t.Kind(),
		))
	} else {
		__antithesis_instrumentation__.Notify(578359)
	}
	__antithesis_instrumentation__.Notify(578357)
	return ct
}

func compareOn(attr ordinal, a, b *valuesMap) (less, eq bool) {
	__antithesis_instrumentation__.Notify(578360)
	av := a.get(attr)
	bv := b.get(attr)
	switch {
	case av == nil && func() bool {
		__antithesis_instrumentation__.Notify(578365)
		return bv == nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(578361)
		return false, true
	case av == nil:
		__antithesis_instrumentation__.Notify(578362)
		return false, false
	case bv == nil:
		__antithesis_instrumentation__.Notify(578363)
		return true, false
	default:
		__antithesis_instrumentation__.Notify(578364)
		return compare(av, bv)
	}
}

func compareEntities(a, b *entity) (less, eq bool) {
	__antithesis_instrumentation__.Notify(578366)
	ordinalSet.union(
		a.attrs, b.attrs,
	).forEach(func(attr ordinal) (wantMore bool) {
		__antithesis_instrumentation__.Notify(578368)
		less, eq = compareOn(attr, a.asMap(), b.asMap())
		return eq
	})
	__antithesis_instrumentation__.Notify(578367)
	return less, eq
}
