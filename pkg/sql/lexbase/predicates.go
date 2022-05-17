package lexbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"unicode"
	"unicode/utf8"
)

func isASCII(s string) bool {
	__antithesis_instrumentation__.Notify(500176)
	for _, c := range s {
		__antithesis_instrumentation__.Notify(500178)
		if c > unicode.MaxASCII {
			__antithesis_instrumentation__.Notify(500179)
			return false
		} else {
			__antithesis_instrumentation__.Notify(500180)
		}
	}
	__antithesis_instrumentation__.Notify(500177)
	return true
}

func IsDigit(ch int) bool {
	__antithesis_instrumentation__.Notify(500181)
	return ch >= '0' && func() bool {
		__antithesis_instrumentation__.Notify(500182)
		return ch <= '9' == true
	}() == true
}

func IsHexDigit(ch int) bool {
	__antithesis_instrumentation__.Notify(500183)
	return (ch >= '0' && func() bool {
		__antithesis_instrumentation__.Notify(500184)
		return ch <= '9' == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(500185)
		return (ch >= 'a' && func() bool {
			__antithesis_instrumentation__.Notify(500186)
			return ch <= 'f' == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(500187)
		return (ch >= 'A' && func() bool {
			__antithesis_instrumentation__.Notify(500188)
			return ch <= 'F' == true
		}() == true) == true
	}() == true
}

var reservedOrLookaheadKeywords = make(map[string]struct{})

func init() {
	for s := range reservedKeywords {
		reservedOrLookaheadKeywords[s] = struct{}{}
	}
	for _, s := range []string{
		"between",
		"ilike",
		"in",
		"like",
		"of",
		"ordinality",
		"similar",
		"time",
		"generated",
		"reset",
		"role",
		"user",
		"on",
	} {
		reservedOrLookaheadKeywords[s] = struct{}{}
	}
}

func isReservedKeyword(s string) bool {
	__antithesis_instrumentation__.Notify(500189)
	_, ok := reservedOrLookaheadKeywords[s]
	return ok
}

func IsBareIdentifier(s string) bool {
	__antithesis_instrumentation__.Notify(500190)
	if len(s) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(500193)
		return !IsIdentStart(int(s[0])) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(500194)
		return (s[0] >= 'A' && func() bool {
			__antithesis_instrumentation__.Notify(500195)
			return s[0] <= 'Z' == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(500196)
		return false
	} else {
		__antithesis_instrumentation__.Notify(500197)
	}
	__antithesis_instrumentation__.Notify(500191)

	isASCII := s[0] < utf8.RuneSelf
	for i := 1; i < len(s); i++ {
		__antithesis_instrumentation__.Notify(500198)
		if !IsIdentMiddle(int(s[i])) {
			__antithesis_instrumentation__.Notify(500201)
			return false
		} else {
			__antithesis_instrumentation__.Notify(500202)
		}
		__antithesis_instrumentation__.Notify(500199)
		if s[i] >= 'A' && func() bool {
			__antithesis_instrumentation__.Notify(500203)
			return s[i] <= 'Z' == true
		}() == true {
			__antithesis_instrumentation__.Notify(500204)

			return false
		} else {
			__antithesis_instrumentation__.Notify(500205)
		}
		__antithesis_instrumentation__.Notify(500200)
		if s[i] >= utf8.RuneSelf {
			__antithesis_instrumentation__.Notify(500206)
			isASCII = false
		} else {
			__antithesis_instrumentation__.Notify(500207)
		}
	}
	__antithesis_instrumentation__.Notify(500192)
	return isASCII || func() bool {
		__antithesis_instrumentation__.Notify(500208)
		return NormalizeName(s) == s == true
	}() == true
}

func IsIdentStart(ch int) bool {
	__antithesis_instrumentation__.Notify(500209)
	return (ch >= 'A' && func() bool {
		__antithesis_instrumentation__.Notify(500210)
		return ch <= 'Z' == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(500211)
		return (ch >= 'a' && func() bool {
			__antithesis_instrumentation__.Notify(500212)
			return ch <= 'z' == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(500213)
		return (ch >= 128 && func() bool {
			__antithesis_instrumentation__.Notify(500214)
			return ch <= 255 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(500215)
		return (ch == '_') == true
	}() == true
}

func IsIdentMiddle(ch int) bool {
	__antithesis_instrumentation__.Notify(500216)
	return IsIdentStart(ch) || func() bool {
		__antithesis_instrumentation__.Notify(500217)
		return IsDigit(ch) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(500218)
		return ch == '$' == true
	}() == true
}
