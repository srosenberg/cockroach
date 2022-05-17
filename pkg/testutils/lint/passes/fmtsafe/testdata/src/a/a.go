package a

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3"
)

var unsafeStr = "abc %d"

const constOk = "safe %d"

func init() {
	_ = recover()

	_ = errors.New(unsafeStr)

	_ = errors.New(unsafeStr)

	_ = errors.New("safestr")
	_ = errors.New(constOk)
	_ = errors.New("abo" + constOk)
	_ = errors.New("abo" + unsafeStr)

	_ = errors.Newf("safe %d", 123)
	_ = errors.Newf(constOk, 123)
	_ = errors.Newf(unsafeStr, 123)
	_ = errors.Newf("abo"+constOk, 123)
	_ = errors.Newf("abo"+unsafeStr, 123)

	ctx := context.Background()

	log.Errorf(ctx, "safe %d", 123)
	log.Errorf(ctx, constOk, 123)
	log.Errorf(ctx, unsafeStr, 123)
	log.Errorf(ctx, "abo"+constOk, 123)
	log.Errorf(ctx, "abo"+unsafeStr, 123)

	var m myLogger
	var l raft.Logger = m

	l.Infof("safe %d", 123)
	l.Infof(constOk, 123)
	l.Infof(unsafeStr, 123)
	l.Infof("abo"+constOk, 123)
	l.Infof("abo"+unsafeStr, 123)
}

type myLogger struct{}

func (m myLogger) Info(args ...interface{}) {
	__antithesis_instrumentation__.Notify(644689)
	log.Errorf(context.Background(), "", args...)
}

func (m myLogger) Infof(_ string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(644690)
	log.Errorf(context.Background(), "ignoredfmt", args...)
}
