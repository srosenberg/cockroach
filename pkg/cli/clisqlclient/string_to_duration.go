package clisqlclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"regexp"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
)

func stringToDuration(s string) (time.Duration, error) {
	__antithesis_instrumentation__.Notify(28971)
	m := intervalRe.FindStringSubmatch(s)
	if m == nil {
		__antithesis_instrumentation__.Notify(28973)
		return 0, errors.Newf("invalid format: %q", s)
	} else {
		__antithesis_instrumentation__.Notify(28974)
	}
	__antithesis_instrumentation__.Notify(28972)
	th, e1 := strconv.Atoi(m[1])
	tm, e2 := strconv.Atoi(m[2])
	ts, e3 := strconv.Atoi(m[3])
	us := m[4] + "000000"[:6-len(m[4])]
	tus, e4 := strconv.Atoi(us)
	return (time.Duration(th)*time.Hour +
			time.Duration(tm)*time.Minute +
			time.Duration(ts)*time.Second +
			time.Duration(tus)*time.Microsecond),
		errors.CombineErrors(e1,
			errors.CombineErrors(e2,
				errors.CombineErrors(e3, e4)))
}

var intervalRe = regexp.MustCompile(`^(\d{2,}):(\d{2}):(\d{2})(?:\.(\d{1,6}))?$`)
