package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/util"
)

type slotIdx int

type fact struct {
	variable slotIdx
	attr     ordinal
	value    slotIdx
}

type slot struct {
	typedValue

	any []typedValue
}

type typedValue struct {
	typ   reflect.Type
	value interface{}
}

func (tv typedValue) toInterface() interface{} {
	__antithesis_instrumentation__.Notify(578712)
	if tv.typ == reflectTypeType {
		__antithesis_instrumentation__.Notify(578715)
		return tv.value.(reflect.Type)
	} else {
		__antithesis_instrumentation__.Notify(578716)
	}
	__antithesis_instrumentation__.Notify(578713)
	if tv.typ.Kind() == reflect.Ptr {
		__antithesis_instrumentation__.Notify(578717)
		if tv.typ.Elem().Kind() == reflect.Struct {
			__antithesis_instrumentation__.Notify(578719)
			return tv.value
		} else {
			__antithesis_instrumentation__.Notify(578720)
		}
		__antithesis_instrumentation__.Notify(578718)
		return reflect.ValueOf(tv.value).Convert(tv.typ).Interface()
	} else {
		__antithesis_instrumentation__.Notify(578721)
	}
	__antithesis_instrumentation__.Notify(578714)
	return reflect.ValueOf(tv.value).Convert(reflect.PtrTo(tv.typ)).Elem().Interface()
}

func (s *slot) eq(other slot) bool {
	__antithesis_instrumentation__.Notify(578722)

	switch {
	case s.value == nil && func() bool {
		__antithesis_instrumentation__.Notify(578727)
		return other.value == nil == true
	}() == true:
		__antithesis_instrumentation__.Notify(578723)
		return true
	case s.value == nil:
		__antithesis_instrumentation__.Notify(578724)
		return false
	case other.value == nil:
		__antithesis_instrumentation__.Notify(578725)
		return false
	default:
		__antithesis_instrumentation__.Notify(578726)
		_, eq := compare(s.value, other.value)
		return eq
	}
}

func (s *slot) empty() bool {
	__antithesis_instrumentation__.Notify(578728)
	return s.value == nil
}

func maybeSet(
	slots []slot, idx slotIdx, tv typedValue, set *util.FastIntSet,
) (foundContradiction bool) {
	__antithesis_instrumentation__.Notify(578729)
	s := &slots[idx]
	check := func() (shouldSet, foundContradiction bool) {
		__antithesis_instrumentation__.Notify(578733)
		if !s.empty() {
			__antithesis_instrumentation__.Notify(578736)
			if _, eq := compare(s.value, tv.value); !eq {
				__antithesis_instrumentation__.Notify(578738)
				return false, true
			} else {
				__antithesis_instrumentation__.Notify(578739)
			}
			__antithesis_instrumentation__.Notify(578737)
			return false, false
		} else {
			__antithesis_instrumentation__.Notify(578740)
		}
		__antithesis_instrumentation__.Notify(578734)

		if s.any != nil {
			__antithesis_instrumentation__.Notify(578741)
			var foundMatch bool
			for _, v := range s.any {
				__antithesis_instrumentation__.Notify(578743)
				if tv.typ != v.typ {
					__antithesis_instrumentation__.Notify(578745)
					continue
				} else {
					__antithesis_instrumentation__.Notify(578746)
				}
				__antithesis_instrumentation__.Notify(578744)
				if _, foundMatch = compare(v.value, tv.value); foundMatch {
					__antithesis_instrumentation__.Notify(578747)
					break
				} else {
					__antithesis_instrumentation__.Notify(578748)
				}
			}
			__antithesis_instrumentation__.Notify(578742)
			if !foundMatch {
				__antithesis_instrumentation__.Notify(578749)
				return false, true
			} else {
				__antithesis_instrumentation__.Notify(578750)
			}
		} else {
			__antithesis_instrumentation__.Notify(578751)
		}
		__antithesis_instrumentation__.Notify(578735)
		return true, false
	}
	__antithesis_instrumentation__.Notify(578730)
	shouldSet, contradiction := check()
	if !shouldSet || func() bool {
		__antithesis_instrumentation__.Notify(578752)
		return contradiction == true
	}() == true {
		__antithesis_instrumentation__.Notify(578753)
		return contradiction
	} else {
		__antithesis_instrumentation__.Notify(578754)
	}
	__antithesis_instrumentation__.Notify(578731)
	s.typedValue = tv
	if set != nil {
		__antithesis_instrumentation__.Notify(578755)
		set.Add(int(idx))
	} else {
		__antithesis_instrumentation__.Notify(578756)
	}
	__antithesis_instrumentation__.Notify(578732)
	return false
}

type filter struct {
	input     []slotIdx
	predicate reflect.Value
}
