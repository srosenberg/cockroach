package screl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	types "github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/errors"
)

func WalkDescIDs(e scpb.Element, f func(id *catid.DescID) error) error {
	__antithesis_instrumentation__.Notify(595050)
	return walk(reflect.TypeOf((*catid.DescID)(nil)), e, func(i interface{}) error {
		__antithesis_instrumentation__.Notify(595051)
		return f(i.(*catid.DescID))
	})
}

func WalkTypes(e scpb.Element, f func(t *types.T) error) error {
	__antithesis_instrumentation__.Notify(595052)
	return walk(reflect.TypeOf((*types.T)(nil)), e, func(i interface{}) error {
		__antithesis_instrumentation__.Notify(595053)
		return f(i.(*types.T))
	})
}

func WalkExpressions(e scpb.Element, f func(t *catpb.Expression) error) error {
	__antithesis_instrumentation__.Notify(595054)
	return walk(reflect.TypeOf((*catpb.Expression)(nil)), e, func(i interface{}) error {
		__antithesis_instrumentation__.Notify(595055)
		return f(i.(*catpb.Expression))
	})
}

func walk(wantType reflect.Type, toWalk interface{}, f func(interface{}) error) (err error) {
	__antithesis_instrumentation__.Notify(595056)
	defer func() {
		__antithesis_instrumentation__.Notify(595063)
		switch r := recover().(type) {
		case nil:
			__antithesis_instrumentation__.Notify(595065)
		case error:
			__antithesis_instrumentation__.Notify(595066)
			err = r
		default:
			__antithesis_instrumentation__.Notify(595067)
			err = errors.AssertionFailedf("failed to do walk: %v", r)
		}
		__antithesis_instrumentation__.Notify(595064)
		if iterutil.Done(err) {
			__antithesis_instrumentation__.Notify(595068)
			err = nil
		} else {
			__antithesis_instrumentation__.Notify(595069)
		}
	}()
	__antithesis_instrumentation__.Notify(595057)

	visit := func(v reflect.Value) {
		__antithesis_instrumentation__.Notify(595070)
		if err := f(v.Interface()); err != nil {
			__antithesis_instrumentation__.Notify(595071)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(595072)
		}
	}
	__antithesis_instrumentation__.Notify(595058)

	wantTypeIsPtr := wantType.Kind() == reflect.Ptr
	var wantTypeElem reflect.Type
	if wantTypeIsPtr {
		__antithesis_instrumentation__.Notify(595073)
		wantTypeElem = wantType.Elem()
	} else {
		__antithesis_instrumentation__.Notify(595074)
	}
	__antithesis_instrumentation__.Notify(595059)
	maybeVisit := func(v reflect.Value) bool {
		__antithesis_instrumentation__.Notify(595075)
		if vt := v.Type(); vt == wantType && func() bool {
			__antithesis_instrumentation__.Notify(595077)
			return (!wantTypeIsPtr || func() bool {
				__antithesis_instrumentation__.Notify(595078)
				return !v.IsNil() == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(595079)
			visit(v)
		} else {
			__antithesis_instrumentation__.Notify(595080)
			if wantTypeIsPtr && func() bool {
				__antithesis_instrumentation__.Notify(595081)
				return wantTypeElem == vt == true
			}() == true {
				__antithesis_instrumentation__.Notify(595082)
				visit(v.Addr())
			} else {
				__antithesis_instrumentation__.Notify(595083)
				return false
			}
		}
		__antithesis_instrumentation__.Notify(595076)
		return true
	}
	__antithesis_instrumentation__.Notify(595060)
	var walk func(value reflect.Value)
	walk = func(value reflect.Value) {
		__antithesis_instrumentation__.Notify(595084)
		if maybeVisit(value) {
			__antithesis_instrumentation__.Notify(595086)
			return
		} else {
			__antithesis_instrumentation__.Notify(595087)
		}
		__antithesis_instrumentation__.Notify(595085)
		switch value.Kind() {
		case reflect.Array, reflect.Slice:
			__antithesis_instrumentation__.Notify(595088)
			for i := 0; i < value.Len(); i++ {
				__antithesis_instrumentation__.Notify(595093)
				walk(value.Index(i).Addr().Elem())
			}
		case reflect.Ptr:
			__antithesis_instrumentation__.Notify(595089)
			if !value.IsNil() {
				__antithesis_instrumentation__.Notify(595094)
				walk(value.Elem())
			} else {
				__antithesis_instrumentation__.Notify(595095)
			}
		case reflect.Struct:
			__antithesis_instrumentation__.Notify(595090)
			for i := 0; i < value.NumField(); i++ {
				__antithesis_instrumentation__.Notify(595096)
				if f := value.Field(i); f.CanAddr() && func() bool {
					__antithesis_instrumentation__.Notify(595097)
					return value.Type().Field(i).IsExported() == true
				}() == true {
					__antithesis_instrumentation__.Notify(595098)
					walk(f.Addr().Elem())
				} else {
					__antithesis_instrumentation__.Notify(595099)
				}
			}
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String,
			reflect.Interface:
			__antithesis_instrumentation__.Notify(595091)
		default:
			__antithesis_instrumentation__.Notify(595092)
			panic(errors.AssertionFailedf(
				"cannot walk values of kind %v, type %v", value.Kind(), value.Type(),
			))
		}
	}
	__antithesis_instrumentation__.Notify(595061)
	v := reflect.ValueOf(toWalk)
	if !v.IsValid() || func() bool {
		__antithesis_instrumentation__.Notify(595100)
		return v.Kind() != reflect.Ptr == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(595101)
		return v.IsNil() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(595102)
		return !v.Elem().CanAddr() == true
	}() == true {
		__antithesis_instrumentation__.Notify(595103)
		return errors.Errorf("invalid value for walking of type %T", toWalk)
	} else {
		__antithesis_instrumentation__.Notify(595104)
	}
	__antithesis_instrumentation__.Notify(595062)
	walk(v.Elem())
	return nil
}
