package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"unsafe"

	"github.com/cockroachdb/errors"
)

func (sc *Schema) GetAttribute(attribute Attr, v interface{}) (interface{}, error) {
	__antithesis_instrumentation__.Notify(579218)
	ord, err := sc.getOrdinal(attribute)
	if err != nil {
		__antithesis_instrumentation__.Notify(579223)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(579224)
	}
	__antithesis_instrumentation__.Notify(579219)
	ti, value, err := getEntityValueInfo(sc, v)
	if err != nil {
		__antithesis_instrumentation__.Notify(579225)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(579226)
	}
	__antithesis_instrumentation__.Notify(579220)

	fi, ok := ti.attrFields[ord]
	if !ok {
		__antithesis_instrumentation__.Notify(579227)
		return nil, errors.Errorf(
			"no field defined on %v for %v", ti.typ, attribute,
		)
	} else {
		__antithesis_instrumentation__.Notify(579228)
	}
	__antithesis_instrumentation__.Notify(579221)

	for i := range fi {
		__antithesis_instrumentation__.Notify(579229)
		got := fi[i].value(unsafe.Pointer(value.Pointer()))
		if got != nil {
			__antithesis_instrumentation__.Notify(579230)
			return got, nil
		} else {
			__antithesis_instrumentation__.Notify(579231)
		}
	}
	__antithesis_instrumentation__.Notify(579222)

	return nil, nil
}
