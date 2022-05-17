package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var escapeRE = regexp.MustCompile(`_u[0-9a-fA-F]{2,8}_`)
var kafkaDisallowedRE = regexp.MustCompile(`[^a-zA-Z0-9\._\-]`)
var avroDisallowedRE = regexp.MustCompile(`[^A-Za-z0-9_]`)

func escapeRune(r rune) string {
	__antithesis_instrumentation__.Notify(17551)
	if r <= 1<<16 {
		__antithesis_instrumentation__.Notify(17553)
		return fmt.Sprintf(`_u%04x_`, r)
	} else {
		__antithesis_instrumentation__.Notify(17554)
	}
	__antithesis_instrumentation__.Notify(17552)
	return fmt.Sprintf(`_u%08x_`, r)
}

func SQLNameToKafkaName(s string) string {
	__antithesis_instrumentation__.Notify(17555)
	if s == `.` {
		__antithesis_instrumentation__.Notify(17558)
		return escapeRune('.')
	} else {
		__antithesis_instrumentation__.Notify(17559)
		if s == `..` {
			__antithesis_instrumentation__.Notify(17560)
			return escapeRune('.') + escapeRune('.')
		} else {
			__antithesis_instrumentation__.Notify(17561)
		}
	}
	__antithesis_instrumentation__.Notify(17556)
	s = escapeSQLName(s, kafkaDisallowedRE)
	if len(s) > 249 {
		__antithesis_instrumentation__.Notify(17562)

		return s[:249]
	} else {
		__antithesis_instrumentation__.Notify(17563)
	}
	__antithesis_instrumentation__.Notify(17557)
	return s
}

func KafkaNameToSQLName(s string) string {
	__antithesis_instrumentation__.Notify(17564)
	return unescapeSQLName(s)
}

func SQLNameToAvroName(s string) string {
	__antithesis_instrumentation__.Notify(17565)
	r, firstSize := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		__antithesis_instrumentation__.Notify(17568)

		return s
	} else {
		__antithesis_instrumentation__.Notify(17569)
	}
	__antithesis_instrumentation__.Notify(17566)

	if r >= '0' && func() bool {
		__antithesis_instrumentation__.Notify(17570)
		return r <= '9' == true
	}() == true {
		__antithesis_instrumentation__.Notify(17571)
		return escapeRune(r) + escapeSQLName(s[firstSize:], avroDisallowedRE)
	} else {
		__antithesis_instrumentation__.Notify(17572)
	}
	__antithesis_instrumentation__.Notify(17567)
	return escapeSQLName(s, avroDisallowedRE)
}

func AvroNameToSQLName(s string) string {
	__antithesis_instrumentation__.Notify(17573)
	return unescapeSQLName(s)
}

func escapeSQLName(s string, disallowedRE *regexp.Regexp) string {
	__antithesis_instrumentation__.Notify(17574)

	s = escapeRE.ReplaceAllStringFunc(s, func(match string) string {
		__antithesis_instrumentation__.Notify(17577)
		var ret strings.Builder
		for _, r := range match {
			__antithesis_instrumentation__.Notify(17579)
			ret.WriteString(escapeRune(r))
		}
		__antithesis_instrumentation__.Notify(17578)
		return ret.String()
	})
	__antithesis_instrumentation__.Notify(17575)

	s = disallowedRE.ReplaceAllStringFunc(s, func(match string) string {
		__antithesis_instrumentation__.Notify(17580)
		var ret strings.Builder
		for _, r := range match {
			__antithesis_instrumentation__.Notify(17582)
			ret.WriteString(escapeRune(r))
		}
		__antithesis_instrumentation__.Notify(17581)
		return ret.String()
	})
	__antithesis_instrumentation__.Notify(17576)
	return s
}

func unescapeSQLName(s string) string {
	__antithesis_instrumentation__.Notify(17583)
	var buf [utf8.UTFMax]byte
	s = escapeRE.ReplaceAllStringFunc(s, func(match string) string {
		__antithesis_instrumentation__.Notify(17585)

		hex := match[2 : len(match)-1]
		r, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			__antithesis_instrumentation__.Notify(17587)

			return match
		} else {
			__antithesis_instrumentation__.Notify(17588)
		}
		__antithesis_instrumentation__.Notify(17586)
		n := utf8.EncodeRune(buf[:utf8.UTFMax], rune(r))
		return string(buf[:n])
	})
	__antithesis_instrumentation__.Notify(17584)
	return s
}
