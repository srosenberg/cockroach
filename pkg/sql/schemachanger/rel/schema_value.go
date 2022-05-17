package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/errors"
)

var (
	reflectTypeType    = reflect.TypeOf((*reflect.Type)(nil)).Elem()
	emptyInterfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
)

func makeComparableValue(val interface{}) (typedValue, error) {
	__antithesis_instrumentation__.Notify(579240)
	if typ, isType := val.(reflect.Type); isType {
		__antithesis_instrumentation__.Notify(579243)
		return typedValue{
			typ:   reflectTypeType,
			value: typ,
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(579244)
	}
	__antithesis_instrumentation__.Notify(579241)
	vv := reflect.ValueOf(val)
	if err := checkNotNil(vv); err != nil {
		__antithesis_instrumentation__.Notify(579245)
		return typedValue{}, err
	} else {
		__antithesis_instrumentation__.Notify(579246)
	}
	__antithesis_instrumentation__.Notify(579242)
	typ := vv.Type()
	switch {
	case isSupportScalarKind(typ.Kind()):
		__antithesis_instrumentation__.Notify(579247)

		compType := getComparableType(typ)
		vvNew := reflect.New(vv.Type())
		vvNew.Elem().Set(vv)
		return typedValue{
			typ:   vv.Type(),
			value: vvNew.Convert(reflect.PtrTo(compType)).Interface(),
		}, nil
	case typ.Kind() == reflect.Ptr:
		__antithesis_instrumentation__.Notify(579248)
		switch {
		case isSupportScalarKind(typ.Elem().Kind()):
			__antithesis_instrumentation__.Notify(579250)
			compType := getComparableType(typ.Elem())
			return typedValue{
				typ:   vv.Type().Elem(),
				value: vv.Convert(reflect.PtrTo(compType)).Interface(),
			}, nil
		case typ.Elem().Kind() == reflect.Struct:
			__antithesis_instrumentation__.Notify(579251)
			return typedValue{
				typ:   vv.Type(),
				value: val,
			}, nil
		default:
			__antithesis_instrumentation__.Notify(579252)
			return typedValue{}, errors.Errorf(
				"unsupported pointer kind %v for type %T", typ.Elem().Kind(), val,
			)
		}
	default:
		__antithesis_instrumentation__.Notify(579249)
		return typedValue{}, errors.Errorf(
			"unsupported kind %v for type %T", typ.Kind(), val,
		)
	}
}

func checkNotNil(v reflect.Value) error {
	__antithesis_instrumentation__.Notify(579253)
	if !v.IsValid() {
		__antithesis_instrumentation__.Notify(579256)

		return errors.Errorf("invalid nil")
	} else {
		__antithesis_instrumentation__.Notify(579257)
	}
	__antithesis_instrumentation__.Notify(579254)
	if v.Kind() == reflect.Ptr && func() bool {
		__antithesis_instrumentation__.Notify(579258)
		return v.IsNil() == true
	}() == true {
		__antithesis_instrumentation__.Notify(579259)
		return errors.Errorf("invalid nil %v", v.Type())
	} else {
		__antithesis_instrumentation__.Notify(579260)
	}
	__antithesis_instrumentation__.Notify(579255)
	return nil
}
