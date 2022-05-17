package issues

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"regexp"
	"strings"
)

func firstNlines(input string, n int) string {
	__antithesis_instrumentation__.Notify(41047)
	if input == "" {
		__antithesis_instrumentation__.Notify(41050)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(41051)
	}
	__antithesis_instrumentation__.Notify(41048)
	pos := 0
	for pos < len(input) && func() bool {
		__antithesis_instrumentation__.Notify(41052)
		return n > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(41053)
		n--
		pos += strings.Index(input[pos:], "\n") + 1
	}
	__antithesis_instrumentation__.Notify(41049)
	return input[:pos]
}

func lastNlines(input string, n int) string {
	__antithesis_instrumentation__.Notify(41054)
	if input == "" {
		__antithesis_instrumentation__.Notify(41057)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(41058)
	}
	__antithesis_instrumentation__.Notify(41055)
	pos := len(input) - 1
	for pos >= 0 && func() bool {
		__antithesis_instrumentation__.Notify(41059)
		return n > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(41060)
		n--
		pos = strings.LastIndex(input[:pos], "\n")
	}
	__antithesis_instrumentation__.Notify(41056)
	return input[pos+1:]
}

type FatalOrPanic struct {
	LastLines,
	Error,
	FirstStack string
}

type RSGCrash struct {
	Error,
	Query,
	Schema string
}

type CondensedMessage string

var panicRE = regexp.MustCompile(`(?ms)^(panic:.*?\n)(goroutine \d+.*?\n)\n`)
var fatalRE = regexp.MustCompile(`(?ms)(^F\d{6}.*?\n)(goroutine \d+.*?\n)\n`)

var crasherRE = regexp.MustCompile(`(?s)( *rsg_test.go:\d{3}: Crash detected:.*?\n)(.*?;\n)`)
var reproRE = regexp.MustCompile(`(?s)( *rsg_test.go:\d{3}: To reproduce, use schema:)`)

func (s CondensedMessage) FatalOrPanic(numPrecedingLines int) (fop FatalOrPanic, ok bool) {
	__antithesis_instrumentation__.Notify(41061)
	ss := string(s)
	add := func(matches []int) {
		__antithesis_instrumentation__.Notify(41065)
		fop.LastLines = lastNlines(ss[:matches[2]], numPrecedingLines)
		fop.Error += ss[matches[2]:matches[3]]
		fop.FirstStack += ss[matches[4]:matches[5]]
		ok = true
	}
	__antithesis_instrumentation__.Notify(41062)
	if sl := panicRE.FindStringSubmatchIndex(ss); sl != nil {
		__antithesis_instrumentation__.Notify(41066)
		add(sl)
	} else {
		__antithesis_instrumentation__.Notify(41067)
	}
	__antithesis_instrumentation__.Notify(41063)
	if sl := fatalRE.FindStringSubmatchIndex(ss); sl != nil {
		__antithesis_instrumentation__.Notify(41068)
		add(sl)
	} else {
		__antithesis_instrumentation__.Notify(41069)
	}
	__antithesis_instrumentation__.Notify(41064)
	return fop, ok
}

func (s CondensedMessage) RSGCrash(lineLimit int) (c RSGCrash, ok bool) {
	__antithesis_instrumentation__.Notify(41070)
	ss := string(s)
	if cm := crasherRE.FindStringSubmatchIndex(ss); cm != nil {
		__antithesis_instrumentation__.Notify(41072)
		c.Error = ss[cm[2]:cm[3]]
		c.Query = firstNlines(ss[cm[4]:cm[5]], lineLimit)
		if rm := reproRE.FindStringSubmatchIndex(ss); rm != nil {
			__antithesis_instrumentation__.Notify(41074)

			c.Schema = firstNlines(ss[rm[2]:], lineLimit)
		} else {
			__antithesis_instrumentation__.Notify(41075)
		}
		__antithesis_instrumentation__.Notify(41073)
		return c, true
	} else {
		__antithesis_instrumentation__.Notify(41076)
	}
	__antithesis_instrumentation__.Notify(41071)
	return RSGCrash{}, false
}

func (s CondensedMessage) String() string {
	__antithesis_instrumentation__.Notify(41077)
	return s.Digest(30)
}

func (s CondensedMessage) Digest(n int) string {
	__antithesis_instrumentation__.Notify(41078)
	if fop, ok := s.FatalOrPanic(n); ok {
		__antithesis_instrumentation__.Notify(41081)
		return fop.LastLines + fop.Error + fop.FirstStack
	} else {
		__antithesis_instrumentation__.Notify(41082)
	}
	__antithesis_instrumentation__.Notify(41079)
	if c, ok := s.RSGCrash(n); ok {
		__antithesis_instrumentation__.Notify(41083)
		return c.Error + c.Query + c.Schema
	} else {
		__antithesis_instrumentation__.Notify(41084)
	}
	__antithesis_instrumentation__.Notify(41080)

	return lastNlines(string(s), n)
}
