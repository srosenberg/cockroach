package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/abortspan"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

type SpanSetReplicaEvalContext struct {
	i  batcheval.EvalContext
	ss spanset.SpanSet
}

var _ batcheval.EvalContext = &SpanSetReplicaEvalContext{}

func (rec *SpanSetReplicaEvalContext) AbortSpan() *abortspan.AbortSpan {
	__antithesis_instrumentation__.Notify(117192)
	return rec.i.AbortSpan()
}

func (rec *SpanSetReplicaEvalContext) EvalKnobs() kvserverbase.BatchEvalTestingKnobs {
	__antithesis_instrumentation__.Notify(117193)
	return rec.i.EvalKnobs()
}

func (rec *SpanSetReplicaEvalContext) StoreID() roachpb.StoreID {
	__antithesis_instrumentation__.Notify(117194)
	return rec.i.StoreID()
}

func (rec *SpanSetReplicaEvalContext) GetRangeID() roachpb.RangeID {
	__antithesis_instrumentation__.Notify(117195)
	return rec.i.GetRangeID()
}

func (rec *SpanSetReplicaEvalContext) ClusterSettings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(117196)
	return rec.i.ClusterSettings()
}

func (rec *SpanSetReplicaEvalContext) Clock() *hlc.Clock {
	__antithesis_instrumentation__.Notify(117197)
	return rec.i.Clock()
}

func (rec *SpanSetReplicaEvalContext) GetConcurrencyManager() concurrency.Manager {
	__antithesis_instrumentation__.Notify(117198)
	return rec.i.GetConcurrencyManager()
}

func (rec *SpanSetReplicaEvalContext) NodeID() roachpb.NodeID {
	__antithesis_instrumentation__.Notify(117199)
	return rec.i.NodeID()
}

func (rec *SpanSetReplicaEvalContext) GetNodeLocality() roachpb.Locality {
	__antithesis_instrumentation__.Notify(117200)
	return rec.i.GetNodeLocality()
}

func (rec *SpanSetReplicaEvalContext) GetFirstIndex() (uint64, error) {
	__antithesis_instrumentation__.Notify(117201)
	return rec.i.GetFirstIndex()
}

func (rec *SpanSetReplicaEvalContext) GetTerm(i uint64) (uint64, error) {
	__antithesis_instrumentation__.Notify(117202)
	return rec.i.GetTerm(i)
}

func (rec *SpanSetReplicaEvalContext) GetLeaseAppliedIndex() uint64 {
	__antithesis_instrumentation__.Notify(117203)
	return rec.i.GetLeaseAppliedIndex()
}

func (rec *SpanSetReplicaEvalContext) IsFirstRange() bool {
	__antithesis_instrumentation__.Notify(117204)
	return rec.i.IsFirstRange()
}

func (rec SpanSetReplicaEvalContext) Desc() *roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(117205)
	desc := rec.i.Desc()
	rec.ss.AssertAllowed(spanset.SpanReadOnly,
		roachpb.Span{Key: keys.RangeDescriptorKey(desc.StartKey)},
	)
	return desc
}

func (rec SpanSetReplicaEvalContext) ContainsKey(key roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(117206)
	desc := rec.Desc()
	return kvserverbase.ContainsKey(desc, key)
}

func (rec SpanSetReplicaEvalContext) GetMVCCStats() enginepb.MVCCStats {
	__antithesis_instrumentation__.Notify(117207)

	return rec.i.GetMVCCStats()
}

func (rec SpanSetReplicaEvalContext) GetMaxSplitQPS() (float64, bool) {
	__antithesis_instrumentation__.Notify(117208)
	return rec.i.GetMaxSplitQPS()
}

func (rec SpanSetReplicaEvalContext) GetLastSplitQPS() float64 {
	__antithesis_instrumentation__.Notify(117209)
	return rec.i.GetLastSplitQPS()
}

func (rec SpanSetReplicaEvalContext) CanCreateTxnRecord(
	ctx context.Context, txnID uuid.UUID, txnKey []byte, txnMinTS hlc.Timestamp,
) (bool, hlc.Timestamp, roachpb.TransactionAbortedReason) {
	__antithesis_instrumentation__.Notify(117210)
	rec.ss.AssertAllowed(spanset.SpanReadOnly,
		roachpb.Span{Key: keys.TransactionKey(txnKey, txnID)},
	)
	return rec.i.CanCreateTxnRecord(ctx, txnID, txnKey, txnMinTS)
}

func (rec SpanSetReplicaEvalContext) GetGCThreshold() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(117211)
	rec.ss.AssertAllowed(spanset.SpanReadOnly,
		roachpb.Span{Key: keys.RangeGCThresholdKey(rec.GetRangeID())},
	)
	return rec.i.GetGCThreshold()
}

func (rec SpanSetReplicaEvalContext) ExcludeDataFromBackup() bool {
	__antithesis_instrumentation__.Notify(117212)
	return rec.i.ExcludeDataFromBackup()
}

func (rec SpanSetReplicaEvalContext) String() string {
	__antithesis_instrumentation__.Notify(117213)
	return rec.i.String()
}

func (rec SpanSetReplicaEvalContext) GetLastReplicaGCTimestamp(
	ctx context.Context,
) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(117214)
	if err := rec.ss.CheckAllowed(spanset.SpanReadOnly,
		roachpb.Span{Key: keys.RangeLastReplicaGCTimestampKey(rec.GetRangeID())},
	); err != nil {
		__antithesis_instrumentation__.Notify(117216)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(117217)
	}
	__antithesis_instrumentation__.Notify(117215)
	return rec.i.GetLastReplicaGCTimestamp(ctx)
}

func (rec SpanSetReplicaEvalContext) GetLease() (roachpb.Lease, roachpb.Lease) {
	__antithesis_instrumentation__.Notify(117218)
	rec.ss.AssertAllowed(spanset.SpanReadOnly,
		roachpb.Span{Key: keys.RangeLeaseKey(rec.GetRangeID())},
	)
	return rec.i.GetLease()
}

func (rec SpanSetReplicaEvalContext) GetRangeInfo(ctx context.Context) roachpb.RangeInfo {
	__antithesis_instrumentation__.Notify(117219)

	rec.Desc()
	rec.GetLease()

	return rec.i.GetRangeInfo(ctx)
}

func (rec *SpanSetReplicaEvalContext) GetCurrentReadSummary(ctx context.Context) rspb.ReadSummary {
	__antithesis_instrumentation__.Notify(117220)

	desc := rec.i.Desc()
	rec.ss.AssertAllowed(spanset.SpanReadWrite, roachpb.Span{
		Key:    keys.MakeRangeKeyPrefix(desc.StartKey),
		EndKey: keys.MakeRangeKeyPrefix(desc.EndKey),
	})
	rec.ss.AssertAllowed(spanset.SpanReadWrite, roachpb.Span{
		Key:    desc.StartKey.AsRawKey(),
		EndKey: desc.EndKey.AsRawKey(),
	})
	return rec.i.GetCurrentReadSummary(ctx)
}

func (rec *SpanSetReplicaEvalContext) GetClosedTimestamp(ctx context.Context) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(117221)
	return rec.i.GetClosedTimestamp(ctx)
}

func (rec *SpanSetReplicaEvalContext) GetExternalStorage(
	ctx context.Context, dest roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(117222)
	return rec.i.GetExternalStorage(ctx, dest)
}

func (rec *SpanSetReplicaEvalContext) GetExternalStorageFromURI(
	ctx context.Context, uri string, user security.SQLUsername,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(117223)
	return rec.i.GetExternalStorageFromURI(ctx, uri, user)
}

func (rec *SpanSetReplicaEvalContext) RevokeLease(ctx context.Context, seq roachpb.LeaseSequence) {
	__antithesis_instrumentation__.Notify(117224)
	rec.i.RevokeLease(ctx, seq)
}

func (rec *SpanSetReplicaEvalContext) WatchForMerge(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(117225)
	return rec.i.WatchForMerge(ctx)
}

func (rec *SpanSetReplicaEvalContext) GetResponseMemoryAccount() *mon.BoundAccount {
	__antithesis_instrumentation__.Notify(117226)
	return rec.i.GetResponseMemoryAccount()
}

func (rec *SpanSetReplicaEvalContext) GetMaxBytes() int64 {
	__antithesis_instrumentation__.Notify(117227)
	return rec.i.GetMaxBytes()
}

func (rec *SpanSetReplicaEvalContext) GetEngineCapacity() (roachpb.StoreCapacity, error) {
	__antithesis_instrumentation__.Notify(117228)
	return rec.i.GetEngineCapacity()
}
