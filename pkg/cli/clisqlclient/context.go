package clisqlclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/cli/clicfg"

type Context struct {
	CliCtx *clicfg.Context

	Echo bool

	DebugMode bool

	EnableServerExecutionTimings bool
}

func (sqlConnCtx *Context) IsInteractive() bool {
	__antithesis_instrumentation__.Notify(28775)
	return sqlConnCtx.CliCtx != nil && func() bool {
		__antithesis_instrumentation__.Notify(28776)
		return sqlConnCtx.CliCtx.IsInteractive == true
	}() == true
}

func (sqlConnCtx *Context) EmbeddedMode() bool {
	__antithesis_instrumentation__.Notify(28777)
	return sqlConnCtx.CliCtx != nil && func() bool {
		__antithesis_instrumentation__.Notify(28778)
		return sqlConnCtx.CliCtx.EmbeddedMode == true
	}() == true
}
