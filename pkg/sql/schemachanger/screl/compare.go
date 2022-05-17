package screl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/rel"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
)

var equalityAttrs = func() []rel.Attr {
	__antithesis_instrumentation__.Notify(594996)
	s := make([]rel.Attr, 0, AttrMax)
	s = append(s, rel.Type)
	for a := Attr(1); a <= AttrMax; a++ {
		__antithesis_instrumentation__.Notify(594998)
		s = append(s, a)
	}
	__antithesis_instrumentation__.Notify(594997)
	return s
}()

func EqualElements(a, b scpb.Element) bool {
	__antithesis_instrumentation__.Notify(594999)
	return Schema.EqualOn(equalityAttrs, a, b)
}

func CompareElements(a, b scpb.Element) (less, eq bool) {
	__antithesis_instrumentation__.Notify(595000)
	return Schema.CompareOn(equalityAttrs, a, b)
}
