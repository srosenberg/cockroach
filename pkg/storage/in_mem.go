package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/pebble/vfs"
)

func InMemFromFS(ctx context.Context, fs vfs.FS, dir string, opts ...ConfigOption) Engine {
	__antithesis_instrumentation__.Notify(639188)

	eng, err := Open(ctx, Location{dir: dir, fs: fs}, opts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(639190)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(639191)
	}
	__antithesis_instrumentation__.Notify(639189)
	return eng
}

func NewDefaultInMemForTesting(opts ...ConfigOption) Engine {
	__antithesis_instrumentation__.Notify(639192)
	eng, err := Open(context.Background(), InMemory(), ForTesting, MaxSize(1<<20), CombineOptions(opts...))
	if err != nil {
		__antithesis_instrumentation__.Notify(639194)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(639195)
	}
	__antithesis_instrumentation__.Notify(639193)
	return eng
}
