package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/abortspan"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/limit"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"golang.org/x/time/rate"
)

type Limiters struct {
	BulkIOWriteRate                      *rate.Limiter
	ConcurrentExportRequests             limit.ConcurrentRequestLimiter
	ConcurrentAddSSTableRequests         limit.ConcurrentRequestLimiter
	ConcurrentAddSSTableAsWritesRequests limit.ConcurrentRequestLimiter

	ConcurrentRangefeedIters         limit.ConcurrentRequestLimiter
	ConcurrentScanInterleavedIntents limit.ConcurrentRequestLimiter
}

type EvalContext interface {
	fmt.Stringer
	ClusterSettings() *cluster.Settings
	EvalKnobs() kvserverbase.BatchEvalTestingKnobs

	Clock() *hlc.Clock
	AbortSpan() *abortspan.AbortSpan
	GetConcurrencyManager() concurrency.Manager

	NodeID() roachpb.NodeID
	StoreID() roachpb.StoreID
	GetRangeID() roachpb.RangeID
	GetNodeLocality() roachpb.Locality

	IsFirstRange() bool
	GetFirstIndex() (uint64, error)
	GetTerm(uint64) (uint64, error)
	GetLeaseAppliedIndex() uint64

	Desc() *roachpb.RangeDescriptor
	ContainsKey(key roachpb.Key) bool

	CanCreateTxnRecord(
		ctx context.Context, txnID uuid.UUID, txnKey []byte, txnMinTS hlc.Timestamp,
	) (ok bool, minCommitTS hlc.Timestamp, reason roachpb.TransactionAbortedReason)

	GetMVCCStats() enginepb.MVCCStats

	GetMaxSplitQPS() (float64, bool)

	GetLastSplitQPS() float64

	GetGCThreshold() hlc.Timestamp
	ExcludeDataFromBackup() bool
	GetLastReplicaGCTimestamp(context.Context) (hlc.Timestamp, error)
	GetLease() (roachpb.Lease, roachpb.Lease)
	GetRangeInfo(context.Context) roachpb.RangeInfo

	GetCurrentReadSummary(ctx context.Context) rspb.ReadSummary

	GetClosedTimestamp(ctx context.Context) hlc.Timestamp

	GetExternalStorage(ctx context.Context, dest roachpb.ExternalStorage) (cloud.ExternalStorage, error)
	GetExternalStorageFromURI(ctx context.Context, uri string, user security.SQLUsername) (cloud.ExternalStorage,
		error)

	RevokeLease(context.Context, roachpb.LeaseSequence)

	WatchForMerge(ctx context.Context) error

	GetResponseMemoryAccount() *mon.BoundAccount

	GetMaxBytes() int64

	GetEngineCapacity() (roachpb.StoreCapacity, error)
}

type MockEvalCtx struct {
	ClusterSettings    *cluster.Settings
	Desc               *roachpb.RangeDescriptor
	StoreID            roachpb.StoreID
	Clock              *hlc.Clock
	Stats              enginepb.MVCCStats
	QPS                float64
	AbortSpan          *abortspan.AbortSpan
	GCThreshold        hlc.Timestamp
	Term, FirstIndex   uint64
	CanCreateTxn       func() (bool, hlc.Timestamp, roachpb.TransactionAbortedReason)
	Lease              roachpb.Lease
	CurrentReadSummary rspb.ReadSummary
	ClosedTimestamp    hlc.Timestamp
	RevokedLeaseSeq    roachpb.LeaseSequence
	MaxBytes           int64
}

func (m *MockEvalCtx) EvalContext() EvalContext {
	__antithesis_instrumentation__.Notify(97581)
	return &mockEvalCtxImpl{MockEvalCtx: m}
}

type mockEvalCtxImpl struct {
	*MockEvalCtx
}

func (m *mockEvalCtxImpl) String() string {
	__antithesis_instrumentation__.Notify(97582)
	return "mock"
}
func (m *mockEvalCtxImpl) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(97583)
	return m.MockEvalCtx.ClusterSettings
}
func (m *mockEvalCtxImpl) EvalKnobs() kvserverbase.BatchEvalTestingKnobs {
	__antithesis_instrumentation__.Notify(97584)
	return kvserverbase.BatchEvalTestingKnobs{}
}
func (m *mockEvalCtxImpl) Clock() *hlc.Clock {
	__antithesis_instrumentation__.Notify(97585)
	return m.MockEvalCtx.Clock
}
func (m *mockEvalCtxImpl) AbortSpan() *abortspan.AbortSpan {
	__antithesis_instrumentation__.Notify(97586)
	return m.MockEvalCtx.AbortSpan
}
func (m *mockEvalCtxImpl) GetConcurrencyManager() concurrency.Manager {
	__antithesis_instrumentation__.Notify(97587)
	panic("unimplemented")
}
func (m *mockEvalCtxImpl) NodeID() roachpb.NodeID {
	__antithesis_instrumentation__.Notify(97588)
	panic("unimplemented")
}
func (m *mockEvalCtxImpl) GetNodeLocality() roachpb.Locality {
	__antithesis_instrumentation__.Notify(97589)
	panic("unimplemented")
}
func (m *mockEvalCtxImpl) StoreID() roachpb.StoreID {
	__antithesis_instrumentation__.Notify(97590)
	return m.MockEvalCtx.StoreID
}
func (m *mockEvalCtxImpl) GetRangeID() roachpb.RangeID {
	__antithesis_instrumentation__.Notify(97591)
	return m.MockEvalCtx.Desc.RangeID
}
func (m *mockEvalCtxImpl) IsFirstRange() bool {
	__antithesis_instrumentation__.Notify(97592)
	panic("unimplemented")
}
func (m *mockEvalCtxImpl) GetFirstIndex() (uint64, error) {
	__antithesis_instrumentation__.Notify(97593)
	return m.FirstIndex, nil
}
func (m *mockEvalCtxImpl) GetTerm(uint64) (uint64, error) {
	__antithesis_instrumentation__.Notify(97594)
	return m.Term, nil
}
func (m *mockEvalCtxImpl) GetLeaseAppliedIndex() uint64 {
	__antithesis_instrumentation__.Notify(97595)
	panic("unimplemented")
}
func (m *mockEvalCtxImpl) Desc() *roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(97596)
	return m.MockEvalCtx.Desc
}
func (m *mockEvalCtxImpl) ContainsKey(key roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(97597)
	return false
}
func (m *mockEvalCtxImpl) GetMVCCStats() enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(97598)
	return m.Stats
}
func (m *mockEvalCtxImpl) GetMaxSplitQPS() (float64, bool) {
	__antithesis_instrumentation__.Notify(97599)
	return m.QPS, true
}
func (m *mockEvalCtxImpl) GetLastSplitQPS() float64 {
	__antithesis_instrumentation__.Notify(97600)
	return m.QPS
}
func (m *mockEvalCtxImpl) CanCreateTxnRecord(
	context.Context, uuid.UUID, []byte, hlc.Timestamp,
) (bool, hlc.Timestamp, roachpb.TransactionAbortedReason) {
	__antithesis_instrumentation__.Notify(97601)
	return m.CanCreateTxn()
}
func (m *mockEvalCtxImpl) GetGCThreshold() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(97602)
	return m.GCThreshold
}
func (m *mockEvalCtxImpl) ExcludeDataFromBackup() bool {
	__antithesis_instrumentation__.Notify(97603)
	return false
}
func (m *mockEvalCtxImpl) GetLastReplicaGCTimestamp(context.Context) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(97604)
	panic("unimplemented")
}
func (m *mockEvalCtxImpl) GetLease() (roachpb.Lease, roachpb.Lease) {
	__antithesis_instrumentation__.Notify(97605)
	return m.Lease, roachpb.Lease{}
}
func (m *mockEvalCtxImpl) GetRangeInfo(ctx context.Context) roachpb.RangeInfo {
	__antithesis_instrumentation__.Notify(97606)
	return roachpb.RangeInfo{Desc: *m.Desc(), Lease: m.Lease}
}
func (m *mockEvalCtxImpl) GetCurrentReadSummary(ctx context.Context) rspb.ReadSummary {
	__antithesis_instrumentation__.Notify(97607)
	return m.CurrentReadSummary
}
func (m *mockEvalCtxImpl) GetClosedTimestamp(ctx context.Context) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(97608)
	return m.ClosedTimestamp
}
func (m *mockEvalCtxImpl) GetExternalStorage(
	ctx context.Context, dest roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(97609)
	panic("unimplemented")
}
func (m *mockEvalCtxImpl) GetExternalStorageFromURI(
	ctx context.Context, uri string, user security.SQLUsername,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(97610)
	panic("unimplemented")
}
func (m *mockEvalCtxImpl) RevokeLease(_ context.Context, seq roachpb.LeaseSequence) {
	__antithesis_instrumentation__.Notify(97611)
	m.RevokedLeaseSeq = seq
}
func (m *mockEvalCtxImpl) WatchForMerge(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(97612)
	panic("unimplemented")
}
func (m *mockEvalCtxImpl) GetResponseMemoryAccount() *mon.BoundAccount {
	__antithesis_instrumentation__.Notify(97613)

	return nil
}
func (m *mockEvalCtxImpl) GetMaxBytes() int64 {
	__antithesis_instrumentation__.Notify(97614)
	if m.MaxBytes != 0 {
		__antithesis_instrumentation__.Notify(97616)
		return m.MaxBytes
	} else {
		__antithesis_instrumentation__.Notify(97617)
	}
	__antithesis_instrumentation__.Notify(97615)
	return math.MaxInt64
}
func (m *mockEvalCtxImpl) GetEngineCapacity() (roachpb.StoreCapacity, error) {
	__antithesis_instrumentation__.Notify(97618)
	return roachpb.StoreCapacity{Available: 1, Capacity: 1}, nil
}
