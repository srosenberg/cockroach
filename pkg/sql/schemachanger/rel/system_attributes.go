package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type systemAttribute int8

const (
	_ systemAttribute = 64 - iota

	Type

	Self

	maxUserAttribute ordinal = 64 - iota
)

func isSystemAttribute(a Attr) bool {
	__antithesis_instrumentation__.Notify(579261)
	_, isSystemAttr := a.(systemAttribute)
	return isSystemAttr
}

var _ Attr = systemAttribute(0)
