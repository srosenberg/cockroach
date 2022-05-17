package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(164263)

	var x [1]struct{}
	_ = x[NotLeaseHolderErrType-1]
	_ = x[RangeNotFoundErrType-2]
	_ = x[RangeKeyMismatchErrType-3]
	_ = x[ReadWithinUncertaintyIntervalErrType-4]
	_ = x[TransactionAbortedErrType-5]
	_ = x[TransactionPushErrType-6]
	_ = x[TransactionRetryErrType-7]
	_ = x[TransactionStatusErrType-8]
	_ = x[WriteIntentErrType-9]
	_ = x[WriteTooOldErrType-10]
	_ = x[OpRequiresTxnErrType-11]
	_ = x[ConditionFailedErrType-12]
	_ = x[LeaseRejectedErrType-13]
	_ = x[NodeUnavailableErrType-14]
	_ = x[RaftGroupDeletedErrType-16]
	_ = x[ReplicaCorruptionErrType-17]
	_ = x[ReplicaTooOldErrType-18]
	_ = x[AmbiguousResultErrType-26]
	_ = x[StoreNotFoundErrType-27]
	_ = x[TransactionRetryWithProtoRefreshErrType-28]
	_ = x[IntegerOverflowErrType-31]
	_ = x[UnsupportedRequestErrType-32]
	_ = x[BatchTimestampBeforeGCErrType-34]
	_ = x[TxnAlreadyEncounteredErrType-35]
	_ = x[IntentMissingErrType-36]
	_ = x[MergeInProgressErrType-37]
	_ = x[RangeFeedRetryErrType-38]
	_ = x[IndeterminateCommitErrType-39]
	_ = x[InvalidLeaseErrType-40]
	_ = x[OptimisticEvalConflictsErrType-41]
	_ = x[MinTimestampBoundUnsatisfiableErrType-42]
	_ = x[RefreshFailedErrType-43]
	_ = x[MVCCHistoryMutationErrType-44]
	_ = x[CommunicationErrType-22]
	_ = x[InternalErrType-25]
}

const (
	_ErrorDetailType_name_0 = "NotLeaseHolderErrTypeRangeNotFoundErrTypeRangeKeyMismatchErrTypeReadWithinUncertaintyIntervalErrTypeTransactionAbortedErrTypeTransactionPushErrTypeTransactionRetryErrTypeTransactionStatusErrTypeWriteIntentErrTypeWriteTooOldErrTypeOpRequiresTxnErrTypeConditionFailedErrTypeLeaseRejectedErrTypeNodeUnavailableErrType"
	_ErrorDetailType_name_1 = "RaftGroupDeletedErrTypeReplicaCorruptionErrTypeReplicaTooOldErrType"
	_ErrorDetailType_name_2 = "CommunicationErrType"
	_ErrorDetailType_name_3 = "InternalErrTypeAmbiguousResultErrTypeStoreNotFoundErrTypeTransactionRetryWithProtoRefreshErrType"
	_ErrorDetailType_name_4 = "IntegerOverflowErrTypeUnsupportedRequestErrType"
	_ErrorDetailType_name_5 = "BatchTimestampBeforeGCErrTypeTxnAlreadyEncounteredErrTypeIntentMissingErrTypeMergeInProgressErrTypeRangeFeedRetryErrTypeIndeterminateCommitErrTypeInvalidLeaseErrTypeOptimisticEvalConflictsErrTypeMinTimestampBoundUnsatisfiableErrTypeRefreshFailedErrTypeMVCCHistoryMutationErrType"
)

var (
	_ErrorDetailType_index_0 = [...]uint16{0, 21, 41, 64, 100, 125, 147, 170, 194, 212, 230, 250, 272, 292, 314}
	_ErrorDetailType_index_1 = [...]uint8{0, 23, 47, 67}
	_ErrorDetailType_index_3 = [...]uint8{0, 15, 37, 57, 96}
	_ErrorDetailType_index_4 = [...]uint8{0, 22, 47}
	_ErrorDetailType_index_5 = [...]uint16{0, 29, 57, 77, 99, 120, 146, 165, 195, 232, 252, 278}
)

func (i ErrorDetailType) String() string {
	__antithesis_instrumentation__.Notify(164264)
	switch {
	case 1 <= i && func() bool {
		__antithesis_instrumentation__.Notify(164272)
		return i <= 14 == true
	}() == true:
		__antithesis_instrumentation__.Notify(164265)
		i -= 1
		return _ErrorDetailType_name_0[_ErrorDetailType_index_0[i]:_ErrorDetailType_index_0[i+1]]
	case 16 <= i && func() bool {
		__antithesis_instrumentation__.Notify(164273)
		return i <= 18 == true
	}() == true:
		__antithesis_instrumentation__.Notify(164266)
		i -= 16
		return _ErrorDetailType_name_1[_ErrorDetailType_index_1[i]:_ErrorDetailType_index_1[i+1]]
	case i == 22:
		__antithesis_instrumentation__.Notify(164267)
		return _ErrorDetailType_name_2
	case 25 <= i && func() bool {
		__antithesis_instrumentation__.Notify(164274)
		return i <= 28 == true
	}() == true:
		__antithesis_instrumentation__.Notify(164268)
		i -= 25
		return _ErrorDetailType_name_3[_ErrorDetailType_index_3[i]:_ErrorDetailType_index_3[i+1]]
	case 31 <= i && func() bool {
		__antithesis_instrumentation__.Notify(164275)
		return i <= 32 == true
	}() == true:
		__antithesis_instrumentation__.Notify(164269)
		i -= 31
		return _ErrorDetailType_name_4[_ErrorDetailType_index_4[i]:_ErrorDetailType_index_4[i+1]]
	case 34 <= i && func() bool {
		__antithesis_instrumentation__.Notify(164276)
		return i <= 44 == true
	}() == true:
		__antithesis_instrumentation__.Notify(164270)
		i -= 34
		return _ErrorDetailType_name_5[_ErrorDetailType_index_5[i]:_ErrorDetailType_index_5[i+1]]
	default:
		__antithesis_instrumentation__.Notify(164271)
		return "ErrorDetailType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
