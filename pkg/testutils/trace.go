package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"regexp"

	"github.com/cockroachdb/errors"
)

func MatchInOrder(s string, res ...string) error {
	__antithesis_instrumentation__.Notify(647028)
	sPos := 0
	for i := range res {
		__antithesis_instrumentation__.Notify(647030)
		reStr := "(?ms)" + res[i]
		re, err := regexp.Compile(reStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(647033)
			return errors.Wrapf(err, "regexp %d (%q) does not compile", i, reStr)
		} else {
			__antithesis_instrumentation__.Notify(647034)
		}
		__antithesis_instrumentation__.Notify(647031)
		loc := re.FindStringIndex(s[sPos:])
		if loc == nil {
			__antithesis_instrumentation__.Notify(647035)

			return errors.Errorf(
				"unable to find regexp %d (%q) in remaining string:\n\n%s\n\nafter having matched:\n\n%s",
				i, reStr, s[sPos:], s[:sPos],
			)
		} else {
			__antithesis_instrumentation__.Notify(647036)
		}
		__antithesis_instrumentation__.Notify(647032)
		sPos += loc[1]
	}
	__antithesis_instrumentation__.Notify(647029)
	return nil
}
