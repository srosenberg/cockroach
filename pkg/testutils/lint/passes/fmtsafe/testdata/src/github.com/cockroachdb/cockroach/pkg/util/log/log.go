package log

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
)

func Errorf(ctx context.Context, format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(644691)
	fmt.Println(fmt.Sprintf(format, args...))
}

func Error(ctx context.Context, msg string) {
	__antithesis_instrumentation__.Notify(644692)
	fmt.Println(msg)
}
