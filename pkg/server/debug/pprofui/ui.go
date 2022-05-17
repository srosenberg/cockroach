package pprofui

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/logtags"
)

func pprofCtx(ctx context.Context) context.Context {
	__antithesis_instrumentation__.Notify(190440)
	return logtags.AddTag(ctx, "pprof", nil)
}

type fakeUI struct{}

func (*fakeUI) ReadLine(prompt string) (string, error) {
	__antithesis_instrumentation__.Notify(190441)
	return "", io.EOF
}

func (*fakeUI) Print(args ...interface{}) {
	__antithesis_instrumentation__.Notify(190442)
	msg := fmt.Sprint(args...)
	log.InfofDepth(pprofCtx(context.Background()), 1, "%s", msg)
}

func (*fakeUI) PrintErr(args ...interface{}) {
	__antithesis_instrumentation__.Notify(190443)
	msg := fmt.Sprint(args...)
	log.WarningfDepth(pprofCtx(context.Background()), 1, "%s", msg)
}

func (*fakeUI) IsTerminal() bool {
	__antithesis_instrumentation__.Notify(190444)
	return false
}

func (*fakeUI) WantBrowser() bool {
	__antithesis_instrumentation__.Notify(190445)
	return false
}

func (*fakeUI) SetAutoComplete(complete func(string) string) {
	__antithesis_instrumentation__.Notify(190446)
}
