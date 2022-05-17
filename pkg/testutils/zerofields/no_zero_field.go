package zerofields

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"reflect"

	"github.com/cockroachdb/errors"
)

type zeroFieldErr struct {
	field string
}

func (err zeroFieldErr) Error() string {
	__antithesis_instrumentation__.Notify(647037)
	return fmt.Sprintf("expected %s field to be non-zero", err.field)
}

func NoZeroField(v interface{}) error {
	__antithesis_instrumentation__.Notify(647038)
	ele := reflect.Indirect(reflect.ValueOf(v))
	eleT := ele.Type()
	for i := 0; i < ele.NumField(); i++ {
		__antithesis_instrumentation__.Notify(647040)
		f := ele.Field(i)
		n := eleT.Field(i).Name
		switch f.Kind() {
		case reflect.Struct:
			__antithesis_instrumentation__.Notify(647041)
			if err := NoZeroField(f.Interface()); err != nil {
				__antithesis_instrumentation__.Notify(647043)
				var zfe zeroFieldErr
				_ = errors.As(err, &zfe)
				zfe.field = fmt.Sprintf("%s.%s", n, zfe.field)
				return zfe
			} else {
				__antithesis_instrumentation__.Notify(647044)
			}
		default:
			__antithesis_instrumentation__.Notify(647042)
			zero := reflect.Zero(f.Type())
			if reflect.DeepEqual(f.Interface(), zero.Interface()) {
				__antithesis_instrumentation__.Notify(647045)
				switch field := eleT.Field(i).Name; field {
				case "XXX_NoUnkeyedLiteral", "XXX_DiscardUnknown", "XXX_sizecache":
					__antithesis_instrumentation__.Notify(647046)

				default:
					__antithesis_instrumentation__.Notify(647047)
					return zeroFieldErr{field: n}
				}
			} else {
				__antithesis_instrumentation__.Notify(647048)
			}
		}
	}
	__antithesis_instrumentation__.Notify(647039)
	return nil
}
