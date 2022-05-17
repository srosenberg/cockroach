package lexbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var normalize = unicode.SpecialCase{
	unicode.CaseRange{
		Lo: 0x0130,
		Hi: 0x0130,
		Delta: [unicode.MaxCase]rune{
			0x49 - 0x130,
			0x69 - 0x130,
			0x49 - 0x130,
		},
	},
	unicode.CaseRange{
		Lo: 0x0131,
		Hi: 0x0131,
		Delta: [unicode.MaxCase]rune{
			0x49 - 0x131,
			0x69 - 0x131,
			0x49 - 0x131,
		},
	},
}

func NormalizeName(n string) string {
	__antithesis_instrumentation__.Notify(500168)
	lower := strings.Map(normalize.ToLower, n)
	if isASCII(lower) {
		__antithesis_instrumentation__.Notify(500170)
		return lower
	} else {
		__antithesis_instrumentation__.Notify(500171)
	}
	__antithesis_instrumentation__.Notify(500169)
	return norm.NFC.String(lower)
}

func NormalizeString(s string) string {
	__antithesis_instrumentation__.Notify(500172)
	if isASCII(s) {
		__antithesis_instrumentation__.Notify(500174)
		return s
	} else {
		__antithesis_instrumentation__.Notify(500175)
	}
	__antithesis_instrumentation__.Notify(500173)
	return norm.NFC.String(s)
}
