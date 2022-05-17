package kvserverbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type BulkAdderOptions struct {
	Name string

	MinBufferSize int64

	MaxBufferSize func() int64

	SkipDuplicates bool

	DisallowShadowingBelow hlc.Timestamp

	BatchTimestamp hlc.Timestamp

	WriteAtBatchTimestamp bool

	InitialSplitsIfUnordered int
}

type BulkAdderFactory func(
	ctx context.Context, db *kv.DB, timestamp hlc.Timestamp, opts BulkAdderOptions,
) (BulkAdder, error)

type BulkAdder interface {
	Add(ctx context.Context, key roachpb.Key, value []byte) error

	Flush(ctx context.Context) error

	IsEmpty() bool

	CurrentBufferFill() float32

	GetSummary() roachpb.BulkOpSummary

	Close(ctx context.Context)

	SetOnFlush(func(summary roachpb.BulkOpSummary))
}

type DuplicateKeyError struct {
	Key   roachpb.Key
	Value []byte
}

func (d *DuplicateKeyError) Error() string {
	__antithesis_instrumentation__.Notify(101842)
	return fmt.Sprintf("duplicate key: %s", d.Key)
}
