package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Method int

func (Method) SafeValue() { __antithesis_instrumentation__.Notify(176870) }

const (
	Get Method = iota

	Put

	ConditionalPut

	Increment

	Delete

	DeleteRange

	ClearRange

	RevertRange

	Scan

	ReverseScan

	EndTxn

	AdminSplit

	AdminUnsplit

	AdminMerge

	AdminTransferLease

	AdminChangeReplicas

	AdminRelocateRange

	HeartbeatTxn

	GC

	PushTxn

	RecoverTxn

	QueryLocks

	QueryTxn

	QueryIntent

	ResolveIntent

	ResolveIntentRange

	Merge

	TruncateLog

	RequestLease

	TransferLease

	LeaseInfo

	ComputeChecksum

	CheckConsistency

	InitPut

	WriteBatch

	Export

	AdminScatter

	AddSSTable

	Migrate

	RecomputeStats

	Refresh

	RefreshRange

	Subsume

	RangeStats

	AdminVerifyProtectedTimestamp

	QueryResolvedTimestamp

	ScanInterleavedIntents

	Barrier

	Probe

	NumMethods
)
