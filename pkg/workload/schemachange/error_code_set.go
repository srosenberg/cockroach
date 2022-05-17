package schemachange

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
)

type errorCodeSet map[pgcode.Code]bool

func makeExpectedErrorSet() errorCodeSet {
	__antithesis_instrumentation__.Notify(695736)
	return errorCodeSet(map[pgcode.Code]bool{})
}

func (set errorCodeSet) merge(otherSet errorCodeSet) {
	__antithesis_instrumentation__.Notify(695737)
	for code := range otherSet {
		__antithesis_instrumentation__.Notify(695738)
		set[code] = true
	}
}

func (set errorCodeSet) add(code pgcode.Code) {
	__antithesis_instrumentation__.Notify(695739)
	set[code] = true
}

func (set errorCodeSet) reset() {
	__antithesis_instrumentation__.Notify(695740)
	for k := range set {
		__antithesis_instrumentation__.Notify(695741)
		delete(set, k)
	}
}

func (set errorCodeSet) contains(code pgcode.Code) bool {
	__antithesis_instrumentation__.Notify(695742)
	_, ok := set[code]
	return ok
}

func (set errorCodeSet) String() string {
	__antithesis_instrumentation__.Notify(695743)
	var codes []string
	for code := range set {
		__antithesis_instrumentation__.Notify(695745)
		codes = append(codes, code.String())
	}
	__antithesis_instrumentation__.Notify(695744)
	sort.Strings(codes)
	return strings.Join(codes, ",")
}

func (set errorCodeSet) empty() bool {
	__antithesis_instrumentation__.Notify(695746)
	return len(set) == 0
}

type codesWithConditions []struct {
	code      pgcode.Code
	condition bool
}

func (c codesWithConditions) add(s errorCodeSet) {
	__antithesis_instrumentation__.Notify(695747)
	for _, cc := range c {
		__antithesis_instrumentation__.Notify(695748)
		if cc.condition {
			__antithesis_instrumentation__.Notify(695749)
			s.add(cc.code)
		} else {
			__antithesis_instrumentation__.Notify(695750)
		}
	}
}
