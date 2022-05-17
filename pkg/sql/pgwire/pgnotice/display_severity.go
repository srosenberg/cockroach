package pgnotice

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"
)

type DisplaySeverity uint32

const (
	DisplaySeverityError = iota

	DisplaySeverityWarning

	DisplaySeverityNotice

	DisplaySeverityLog

	DisplaySeverityDebug1

	DisplaySeverityDebug2

	DisplaySeverityDebug3

	DisplaySeverityDebug4

	DisplaySeverityDebug5
)

func ParseDisplaySeverity(k string) (DisplaySeverity, bool) {
	__antithesis_instrumentation__.Notify(560893)
	s, ok := namesToDisplaySeverity[strings.ToLower(k)]
	return s, ok
}

func (ns DisplaySeverity) String() string {
	__antithesis_instrumentation__.Notify(560894)
	if ns >= DisplaySeverity(len(noticeDisplaySeverityNames)) {
		__antithesis_instrumentation__.Notify(560896)
		return fmt.Sprintf("DisplaySeverity(%d)", ns)
	} else {
		__antithesis_instrumentation__.Notify(560897)
	}
	__antithesis_instrumentation__.Notify(560895)
	return noticeDisplaySeverityNames[ns]
}

var noticeDisplaySeverityNames = [...]string{
	DisplaySeverityDebug5:  "debug5",
	DisplaySeverityDebug4:  "debug4",
	DisplaySeverityDebug3:  "debug3",
	DisplaySeverityDebug2:  "debug2",
	DisplaySeverityDebug1:  "debug1",
	DisplaySeverityLog:     "log",
	DisplaySeverityNotice:  "notice",
	DisplaySeverityWarning: "warning",
	DisplaySeverityError:   "error",
}

var namesToDisplaySeverity = map[string]DisplaySeverity{}

func ValidDisplaySeverities() []string {
	__antithesis_instrumentation__.Notify(560898)
	ret := make([]string, 0, len(namesToDisplaySeverity))
	for _, s := range noticeDisplaySeverityNames {
		__antithesis_instrumentation__.Notify(560900)
		ret = append(ret, s)
	}
	__antithesis_instrumentation__.Notify(560899)
	return ret
}

func init() {
	for k, v := range noticeDisplaySeverityNames {
		namesToDisplaySeverity[v] = DisplaySeverity(k)
	}
}
