package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "reflect"

func TestingHook(ptr, val interface{}) func() {
	__antithesis_instrumentation__.Notify(644323)
	global := reflect.ValueOf(ptr).Elem()
	orig := reflect.New(global.Type()).Elem()
	orig.Set(global)
	global.Set(reflect.ValueOf(val))
	return func() { __antithesis_instrumentation__.Notify(644324); global.Set(orig) }
}
