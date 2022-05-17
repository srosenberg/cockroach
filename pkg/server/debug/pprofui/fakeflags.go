package pprofui

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/google/pprof/driver"
	"github.com/spf13/pflag"
)

type pprofFlags struct {
	args []string
	*pflag.FlagSet
}

var _ driver.FlagSet = &pprofFlags{}

func (pprofFlags) ExtraUsage() string {
	__antithesis_instrumentation__.Notify(190323)
	return ""
}

func (pprofFlags) AddExtraUsage(eu string) {
	__antithesis_instrumentation__.Notify(190324)
}

func (f pprofFlags) StringList(o, d, c string) *[]*string {
	__antithesis_instrumentation__.Notify(190325)
	return &[]*string{f.String(o, d, c)}
}

func (f pprofFlags) Parse(usage func()) []string {
	__antithesis_instrumentation__.Notify(190326)
	f.FlagSet.Usage = usage
	if err := f.FlagSet.Parse(f.args); err != nil {
		__antithesis_instrumentation__.Notify(190328)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(190329)
	}
	__antithesis_instrumentation__.Notify(190327)
	return f.FlagSet.Args()
}
