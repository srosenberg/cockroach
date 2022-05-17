package ssh

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"regexp"
	"strings"
)

const shellMetachars = "|&;()<> \t\n$\\`"

func Escape1(arg string) string {
	__antithesis_instrumentation__.Notify(182532)
	if strings.ContainsAny(arg, shellMetachars) {
		__antithesis_instrumentation__.Notify(182534)

		e := regexp.MustCompile("([$`\"\\\\])").ReplaceAllString(arg, `\$1`)
		return fmt.Sprintf(`"%s"`, e)
	} else {
		__antithesis_instrumentation__.Notify(182535)
	}
	__antithesis_instrumentation__.Notify(182533)
	return arg
}

func Escape(args []string) string {
	__antithesis_instrumentation__.Notify(182536)
	escaped := make([]string, len(args))
	for i := range args {
		__antithesis_instrumentation__.Notify(182538)
		escaped[i] = Escape1(args[i])
	}
	__antithesis_instrumentation__.Notify(182537)
	return strings.Join(escaped, " ")
}
