// Package protectedts exports the interface to the protected timestamp
// subsystem which allows clients to prevent GC of expired data.
package protectedts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

var ErrNotExists = errors.New("protected timestamp record does not exist")

var ErrExists = errors.New("protected timestamp record already exists")

type Provider interface {
	Storage
	Cache
	Reconciler

	Start(context.Context, *stop.Stopper) error
	Metrics() metric.Struct
}

type Storage interface {
	Protect(context.Context, *kv.Txn, *ptpb.Record) error

	GetRecord(context.Context, *kv.Txn, uuid.UUID) (*ptpb.Record, error)

	MarkVerified(context.Context, *kv.Txn, uuid.UUID) error

	Release(context.Context, *kv.Txn, uuid.UUID) error

	GetMetadata(context.Context, *kv.Txn) (ptpb.Metadata, error)

	GetState(context.Context, *kv.Txn) (ptpb.State, error)

	UpdateTimestamp(ctx context.Context, txn *kv.Txn, id uuid.UUID, timestamp hlc.Timestamp) error
}

type Iterator func(*ptpb.Record) (wantMore bool)

type Cache interface {
	spanconfig.ProtectedTSReader

	Iterate(_ context.Context, from, to roachpb.Key, it Iterator) (asOf hlc.Timestamp)

	QueryRecord(_ context.Context, id uuid.UUID) (exists bool, asOf hlc.Timestamp)

	Refresh(_ context.Context, asOf hlc.Timestamp) error
}

type Reconciler interface {
	StartReconciler(ctx context.Context, stopper *stop.Stopper) error
}

func EmptyCache(c *hlc.Clock) Cache {
	__antithesis_instrumentation__.Notify(110401)
	return (*emptyCache)(c)
}

type emptyCache hlc.Clock

func (c *emptyCache) Iterate(
	_ context.Context, from, to roachpb.Key, it Iterator,
) (asOf hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(110402)
	return (*hlc.Clock)(c).Now()
}

func (c *emptyCache) QueryRecord(
	_ context.Context, id uuid.UUID,
) (exists bool, asOf hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(110403)
	return false, (*hlc.Clock)(c).Now()
}

func (c *emptyCache) Refresh(_ context.Context, asOf hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(110404)
	return nil
}

func (c *emptyCache) GetProtectionTimestamps(
	context.Context, roachpb.Span,
) (protectionTimestamps []hlc.Timestamp, asOf hlc.Timestamp, err error) {
	__antithesis_instrumentation__.Notify(110405)
	return protectionTimestamps, (*hlc.Clock)(c).Now(), nil
}
