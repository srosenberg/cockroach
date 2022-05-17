package clisqlexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/cli/clicfg"

type Context struct {
	CliCtx *clicfg.Context

	TerminalOutput bool

	TableDisplayFormat TableDisplayFormat

	TableBorderMode int

	ShowTimes bool

	VerboseTimings bool
}

func (sqlExecCtx *Context) IsInteractive() bool {
	__antithesis_instrumentation__.Notify(28981)
	return sqlExecCtx.CliCtx != nil && func() bool {
		__antithesis_instrumentation__.Notify(28982)
		return sqlExecCtx.CliCtx.IsInteractive == true
	}() == true
}
