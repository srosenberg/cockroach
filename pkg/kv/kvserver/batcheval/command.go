package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type DeclareKeysFunc func(
	rs ImmutableRangeState,
	header *roachpb.Header,
	request roachpb.Request,
	latchSpans, lockSpans *spanset.SpanSet,
	maxOffset time.Duration,
)

type ImmutableRangeState interface {
	GetRangeID() roachpb.RangeID
	GetStartKey() roachpb.RKey
}

type Command struct {
	DeclareKeys DeclareKeysFunc

	EvalRW func(context.Context, storage.ReadWriter, CommandArgs, roachpb.Response) (result.Result, error)
	EvalRO func(context.Context, storage.Reader, CommandArgs, roachpb.Response) (result.Result, error)
}

var cmds = make(map[roachpb.Method]Command)

func RegisterReadWriteCommand(
	method roachpb.Method,
	declare DeclareKeysFunc,
	impl func(context.Context, storage.ReadWriter, CommandArgs, roachpb.Response) (result.Result, error),
) {
	__antithesis_instrumentation__.Notify(97559)
	register(method, Command{
		DeclareKeys: declare,
		EvalRW:      impl,
	})
}

func RegisterReadOnlyCommand(
	method roachpb.Method,
	declare DeclareKeysFunc,
	impl func(context.Context, storage.Reader, CommandArgs, roachpb.Response) (result.Result, error),
) {
	__antithesis_instrumentation__.Notify(97560)
	register(method, Command{
		DeclareKeys: declare,
		EvalRO:      impl,
	})
}

func register(method roachpb.Method, command Command) {
	__antithesis_instrumentation__.Notify(97561)
	if _, ok := cmds[method]; ok {
		__antithesis_instrumentation__.Notify(97563)
		log.Fatalf(context.TODO(), "cannot overwrite previously registered method %v", method)
	} else {
		__antithesis_instrumentation__.Notify(97564)
	}
	__antithesis_instrumentation__.Notify(97562)
	cmds[method] = command
}

func UnregisterCommand(method roachpb.Method) {
	__antithesis_instrumentation__.Notify(97565)
	delete(cmds, method)
}

func LookupCommand(method roachpb.Method) (Command, bool) {
	__antithesis_instrumentation__.Notify(97566)
	cmd, ok := cmds[method]
	return cmd, ok
}
