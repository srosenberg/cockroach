package clicfg

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "fmt"

func (cliCtx *Context) PrintlnUnlessEmbedded(args ...interface{}) {
	__antithesis_instrumentation__.Notify(28149)
	if !cliCtx.EmbeddedMode {
		__antithesis_instrumentation__.Notify(28150)
		fmt.Println(args...)
	} else {
		__antithesis_instrumentation__.Notify(28151)
	}
}

func (cliCtx *Context) PrintfUnlessEmbedded(f string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(28152)
	if !cliCtx.EmbeddedMode {
		__antithesis_instrumentation__.Notify(28153)
		fmt.Printf(f, args...)
	} else {
		__antithesis_instrumentation__.Notify(28154)
	}
}
