package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strconv"
	"strings"
)

func (ru ErrorDetail) GetInner() error {
	__antithesis_instrumentation__.Notify(158218)
	switch t := ru.GetValue().(type) {
	case *ErrorDetail_NotLeaseHolder:
		__antithesis_instrumentation__.Notify(158219)
		return t.NotLeaseHolder
	case *ErrorDetail_RangeNotFound:
		__antithesis_instrumentation__.Notify(158220)
		return t.RangeNotFound
	case *ErrorDetail_RangeKeyMismatch:
		__antithesis_instrumentation__.Notify(158221)
		return t.RangeKeyMismatch
	case *ErrorDetail_ReadWithinUncertaintyInterval:
		__antithesis_instrumentation__.Notify(158222)
		return t.ReadWithinUncertaintyInterval
	case *ErrorDetail_TransactionAborted:
		__antithesis_instrumentation__.Notify(158223)
		return t.TransactionAborted
	case *ErrorDetail_TransactionPush:
		__antithesis_instrumentation__.Notify(158224)
		return t.TransactionPush
	case *ErrorDetail_TransactionRetry:
		__antithesis_instrumentation__.Notify(158225)
		return t.TransactionRetry
	case *ErrorDetail_TransactionStatus:
		__antithesis_instrumentation__.Notify(158226)
		return t.TransactionStatus
	case *ErrorDetail_WriteIntent:
		__antithesis_instrumentation__.Notify(158227)
		return t.WriteIntent
	case *ErrorDetail_WriteTooOld:
		__antithesis_instrumentation__.Notify(158228)
		return t.WriteTooOld
	case *ErrorDetail_OpRequiresTxn:
		__antithesis_instrumentation__.Notify(158229)
		return t.OpRequiresTxn
	case *ErrorDetail_ConditionFailed:
		__antithesis_instrumentation__.Notify(158230)
		return t.ConditionFailed
	case *ErrorDetail_LeaseRejected:
		__antithesis_instrumentation__.Notify(158231)
		return t.LeaseRejected
	case *ErrorDetail_NodeUnavailable:
		__antithesis_instrumentation__.Notify(158232)
		return t.NodeUnavailable
	case *ErrorDetail_RaftGroupDeleted:
		__antithesis_instrumentation__.Notify(158233)
		return t.RaftGroupDeleted
	case *ErrorDetail_ReplicaCorruption:
		__antithesis_instrumentation__.Notify(158234)
		return t.ReplicaCorruption
	case *ErrorDetail_ReplicaTooOld:
		__antithesis_instrumentation__.Notify(158235)
		return t.ReplicaTooOld
	case *ErrorDetail_AmbiguousResult:
		__antithesis_instrumentation__.Notify(158236)
		return t.AmbiguousResult
	case *ErrorDetail_StoreNotFound:
		__antithesis_instrumentation__.Notify(158237)
		return t.StoreNotFound
	case *ErrorDetail_TransactionRetryWithProtoRefresh:
		__antithesis_instrumentation__.Notify(158238)
		return t.TransactionRetryWithProtoRefresh
	case *ErrorDetail_IntegerOverflow:
		__antithesis_instrumentation__.Notify(158239)
		return t.IntegerOverflow
	case *ErrorDetail_UnsupportedRequest:
		__antithesis_instrumentation__.Notify(158240)
		return t.UnsupportedRequest
	case *ErrorDetail_TimestampBefore:
		__antithesis_instrumentation__.Notify(158241)
		return t.TimestampBefore
	case *ErrorDetail_TxnAlreadyEncounteredError:
		__antithesis_instrumentation__.Notify(158242)
		return t.TxnAlreadyEncounteredError
	case *ErrorDetail_IntentMissing:
		__antithesis_instrumentation__.Notify(158243)
		return t.IntentMissing
	case *ErrorDetail_MergeInProgress:
		__antithesis_instrumentation__.Notify(158244)
		return t.MergeInProgress
	case *ErrorDetail_RangefeedRetry:
		__antithesis_instrumentation__.Notify(158245)
		return t.RangefeedRetry
	case *ErrorDetail_IndeterminateCommit:
		__antithesis_instrumentation__.Notify(158246)
		return t.IndeterminateCommit
	case *ErrorDetail_InvalidLeaseError:
		__antithesis_instrumentation__.Notify(158247)
		return t.InvalidLeaseError
	case *ErrorDetail_OptimisticEvalConflicts:
		__antithesis_instrumentation__.Notify(158248)
		return t.OptimisticEvalConflicts
	case *ErrorDetail_MinTimestampBoundUnsatisfiable:
		__antithesis_instrumentation__.Notify(158249)
		return t.MinTimestampBoundUnsatisfiable
	case *ErrorDetail_RefreshFailedError:
		__antithesis_instrumentation__.Notify(158250)
		return t.RefreshFailedError
	case *ErrorDetail_MVCCHistoryMutation:
		__antithesis_instrumentation__.Notify(158251)
		return t.MVCCHistoryMutation
	default:
		__antithesis_instrumentation__.Notify(158252)
		return nil
	}
}

func (ru RequestUnion) GetInner() Request {
	__antithesis_instrumentation__.Notify(158253)
	switch t := ru.GetValue().(type) {
	case *RequestUnion_Get:
		__antithesis_instrumentation__.Notify(158254)
		return t.Get
	case *RequestUnion_Put:
		__antithesis_instrumentation__.Notify(158255)
		return t.Put
	case *RequestUnion_ConditionalPut:
		__antithesis_instrumentation__.Notify(158256)
		return t.ConditionalPut
	case *RequestUnion_Increment:
		__antithesis_instrumentation__.Notify(158257)
		return t.Increment
	case *RequestUnion_Delete:
		__antithesis_instrumentation__.Notify(158258)
		return t.Delete
	case *RequestUnion_DeleteRange:
		__antithesis_instrumentation__.Notify(158259)
		return t.DeleteRange
	case *RequestUnion_ClearRange:
		__antithesis_instrumentation__.Notify(158260)
		return t.ClearRange
	case *RequestUnion_RevertRange:
		__antithesis_instrumentation__.Notify(158261)
		return t.RevertRange
	case *RequestUnion_Scan:
		__antithesis_instrumentation__.Notify(158262)
		return t.Scan
	case *RequestUnion_EndTxn:
		__antithesis_instrumentation__.Notify(158263)
		return t.EndTxn
	case *RequestUnion_AdminSplit:
		__antithesis_instrumentation__.Notify(158264)
		return t.AdminSplit
	case *RequestUnion_AdminUnsplit:
		__antithesis_instrumentation__.Notify(158265)
		return t.AdminUnsplit
	case *RequestUnion_AdminMerge:
		__antithesis_instrumentation__.Notify(158266)
		return t.AdminMerge
	case *RequestUnion_AdminTransferLease:
		__antithesis_instrumentation__.Notify(158267)
		return t.AdminTransferLease
	case *RequestUnion_AdminChangeReplicas:
		__antithesis_instrumentation__.Notify(158268)
		return t.AdminChangeReplicas
	case *RequestUnion_AdminRelocateRange:
		__antithesis_instrumentation__.Notify(158269)
		return t.AdminRelocateRange
	case *RequestUnion_HeartbeatTxn:
		__antithesis_instrumentation__.Notify(158270)
		return t.HeartbeatTxn
	case *RequestUnion_Gc:
		__antithesis_instrumentation__.Notify(158271)
		return t.Gc
	case *RequestUnion_PushTxn:
		__antithesis_instrumentation__.Notify(158272)
		return t.PushTxn
	case *RequestUnion_RecoverTxn:
		__antithesis_instrumentation__.Notify(158273)
		return t.RecoverTxn
	case *RequestUnion_ResolveIntent:
		__antithesis_instrumentation__.Notify(158274)
		return t.ResolveIntent
	case *RequestUnion_ResolveIntentRange:
		__antithesis_instrumentation__.Notify(158275)
		return t.ResolveIntentRange
	case *RequestUnion_Merge:
		__antithesis_instrumentation__.Notify(158276)
		return t.Merge
	case *RequestUnion_TruncateLog:
		__antithesis_instrumentation__.Notify(158277)
		return t.TruncateLog
	case *RequestUnion_RequestLease:
		__antithesis_instrumentation__.Notify(158278)
		return t.RequestLease
	case *RequestUnion_ReverseScan:
		__antithesis_instrumentation__.Notify(158279)
		return t.ReverseScan
	case *RequestUnion_ComputeChecksum:
		__antithesis_instrumentation__.Notify(158280)
		return t.ComputeChecksum
	case *RequestUnion_CheckConsistency:
		__antithesis_instrumentation__.Notify(158281)
		return t.CheckConsistency
	case *RequestUnion_InitPut:
		__antithesis_instrumentation__.Notify(158282)
		return t.InitPut
	case *RequestUnion_TransferLease:
		__antithesis_instrumentation__.Notify(158283)
		return t.TransferLease
	case *RequestUnion_LeaseInfo:
		__antithesis_instrumentation__.Notify(158284)
		return t.LeaseInfo
	case *RequestUnion_Export:
		__antithesis_instrumentation__.Notify(158285)
		return t.Export
	case *RequestUnion_QueryTxn:
		__antithesis_instrumentation__.Notify(158286)
		return t.QueryTxn
	case *RequestUnion_QueryIntent:
		__antithesis_instrumentation__.Notify(158287)
		return t.QueryIntent
	case *RequestUnion_QueryLocks:
		__antithesis_instrumentation__.Notify(158288)
		return t.QueryLocks
	case *RequestUnion_AdminScatter:
		__antithesis_instrumentation__.Notify(158289)
		return t.AdminScatter
	case *RequestUnion_AddSstable:
		__antithesis_instrumentation__.Notify(158290)
		return t.AddSstable
	case *RequestUnion_RecomputeStats:
		__antithesis_instrumentation__.Notify(158291)
		return t.RecomputeStats
	case *RequestUnion_Refresh:
		__antithesis_instrumentation__.Notify(158292)
		return t.Refresh
	case *RequestUnion_RefreshRange:
		__antithesis_instrumentation__.Notify(158293)
		return t.RefreshRange
	case *RequestUnion_Subsume:
		__antithesis_instrumentation__.Notify(158294)
		return t.Subsume
	case *RequestUnion_RangeStats:
		__antithesis_instrumentation__.Notify(158295)
		return t.RangeStats
	case *RequestUnion_AdminVerifyProtectedTimestamp:
		__antithesis_instrumentation__.Notify(158296)
		return t.AdminVerifyProtectedTimestamp
	case *RequestUnion_Migrate:
		__antithesis_instrumentation__.Notify(158297)
		return t.Migrate
	case *RequestUnion_QueryResolvedTimestamp:
		__antithesis_instrumentation__.Notify(158298)
		return t.QueryResolvedTimestamp
	case *RequestUnion_ScanInterleavedIntents:
		__antithesis_instrumentation__.Notify(158299)
		return t.ScanInterleavedIntents
	case *RequestUnion_Barrier:
		__antithesis_instrumentation__.Notify(158300)
		return t.Barrier
	case *RequestUnion_Probe:
		__antithesis_instrumentation__.Notify(158301)
		return t.Probe
	default:
		__antithesis_instrumentation__.Notify(158302)
		return nil
	}
}

func (ru ResponseUnion) GetInner() Response {
	__antithesis_instrumentation__.Notify(158303)
	switch t := ru.GetValue().(type) {
	case *ResponseUnion_Get:
		__antithesis_instrumentation__.Notify(158304)
		return t.Get
	case *ResponseUnion_Put:
		__antithesis_instrumentation__.Notify(158305)
		return t.Put
	case *ResponseUnion_ConditionalPut:
		__antithesis_instrumentation__.Notify(158306)
		return t.ConditionalPut
	case *ResponseUnion_Increment:
		__antithesis_instrumentation__.Notify(158307)
		return t.Increment
	case *ResponseUnion_Delete:
		__antithesis_instrumentation__.Notify(158308)
		return t.Delete
	case *ResponseUnion_DeleteRange:
		__antithesis_instrumentation__.Notify(158309)
		return t.DeleteRange
	case *ResponseUnion_ClearRange:
		__antithesis_instrumentation__.Notify(158310)
		return t.ClearRange
	case *ResponseUnion_RevertRange:
		__antithesis_instrumentation__.Notify(158311)
		return t.RevertRange
	case *ResponseUnion_Scan:
		__antithesis_instrumentation__.Notify(158312)
		return t.Scan
	case *ResponseUnion_EndTxn:
		__antithesis_instrumentation__.Notify(158313)
		return t.EndTxn
	case *ResponseUnion_AdminSplit:
		__antithesis_instrumentation__.Notify(158314)
		return t.AdminSplit
	case *ResponseUnion_AdminUnsplit:
		__antithesis_instrumentation__.Notify(158315)
		return t.AdminUnsplit
	case *ResponseUnion_AdminMerge:
		__antithesis_instrumentation__.Notify(158316)
		return t.AdminMerge
	case *ResponseUnion_AdminTransferLease:
		__antithesis_instrumentation__.Notify(158317)
		return t.AdminTransferLease
	case *ResponseUnion_AdminChangeReplicas:
		__antithesis_instrumentation__.Notify(158318)
		return t.AdminChangeReplicas
	case *ResponseUnion_AdminRelocateRange:
		__antithesis_instrumentation__.Notify(158319)
		return t.AdminRelocateRange
	case *ResponseUnion_HeartbeatTxn:
		__antithesis_instrumentation__.Notify(158320)
		return t.HeartbeatTxn
	case *ResponseUnion_Gc:
		__antithesis_instrumentation__.Notify(158321)
		return t.Gc
	case *ResponseUnion_PushTxn:
		__antithesis_instrumentation__.Notify(158322)
		return t.PushTxn
	case *ResponseUnion_RecoverTxn:
		__antithesis_instrumentation__.Notify(158323)
		return t.RecoverTxn
	case *ResponseUnion_ResolveIntent:
		__antithesis_instrumentation__.Notify(158324)
		return t.ResolveIntent
	case *ResponseUnion_ResolveIntentRange:
		__antithesis_instrumentation__.Notify(158325)
		return t.ResolveIntentRange
	case *ResponseUnion_Merge:
		__antithesis_instrumentation__.Notify(158326)
		return t.Merge
	case *ResponseUnion_TruncateLog:
		__antithesis_instrumentation__.Notify(158327)
		return t.TruncateLog
	case *ResponseUnion_RequestLease:
		__antithesis_instrumentation__.Notify(158328)
		return t.RequestLease
	case *ResponseUnion_ReverseScan:
		__antithesis_instrumentation__.Notify(158329)
		return t.ReverseScan
	case *ResponseUnion_ComputeChecksum:
		__antithesis_instrumentation__.Notify(158330)
		return t.ComputeChecksum
	case *ResponseUnion_CheckConsistency:
		__antithesis_instrumentation__.Notify(158331)
		return t.CheckConsistency
	case *ResponseUnion_InitPut:
		__antithesis_instrumentation__.Notify(158332)
		return t.InitPut
	case *ResponseUnion_LeaseInfo:
		__antithesis_instrumentation__.Notify(158333)
		return t.LeaseInfo
	case *ResponseUnion_Export:
		__antithesis_instrumentation__.Notify(158334)
		return t.Export
	case *ResponseUnion_QueryTxn:
		__antithesis_instrumentation__.Notify(158335)
		return t.QueryTxn
	case *ResponseUnion_QueryIntent:
		__antithesis_instrumentation__.Notify(158336)
		return t.QueryIntent
	case *ResponseUnion_QueryLocks:
		__antithesis_instrumentation__.Notify(158337)
		return t.QueryLocks
	case *ResponseUnion_AdminScatter:
		__antithesis_instrumentation__.Notify(158338)
		return t.AdminScatter
	case *ResponseUnion_AddSstable:
		__antithesis_instrumentation__.Notify(158339)
		return t.AddSstable
	case *ResponseUnion_RecomputeStats:
		__antithesis_instrumentation__.Notify(158340)
		return t.RecomputeStats
	case *ResponseUnion_Refresh:
		__antithesis_instrumentation__.Notify(158341)
		return t.Refresh
	case *ResponseUnion_RefreshRange:
		__antithesis_instrumentation__.Notify(158342)
		return t.RefreshRange
	case *ResponseUnion_Subsume:
		__antithesis_instrumentation__.Notify(158343)
		return t.Subsume
	case *ResponseUnion_RangeStats:
		__antithesis_instrumentation__.Notify(158344)
		return t.RangeStats
	case *ResponseUnion_AdminVerifyProtectedTimestamp:
		__antithesis_instrumentation__.Notify(158345)
		return t.AdminVerifyProtectedTimestamp
	case *ResponseUnion_Migrate:
		__antithesis_instrumentation__.Notify(158346)
		return t.Migrate
	case *ResponseUnion_QueryResolvedTimestamp:
		__antithesis_instrumentation__.Notify(158347)
		return t.QueryResolvedTimestamp
	case *ResponseUnion_ScanInterleavedIntents:
		__antithesis_instrumentation__.Notify(158348)
		return t.ScanInterleavedIntents
	case *ResponseUnion_Barrier:
		__antithesis_instrumentation__.Notify(158349)
		return t.Barrier
	case *ResponseUnion_Probe:
		__antithesis_instrumentation__.Notify(158350)
		return t.Probe
	default:
		__antithesis_instrumentation__.Notify(158351)
		return nil
	}
}

func (ru *ErrorDetail) MustSetInner(r error) {
	__antithesis_instrumentation__.Notify(158352)
	ru.Reset()
	var union isErrorDetail_Value
	switch t := r.(type) {
	case *NotLeaseHolderError:
		__antithesis_instrumentation__.Notify(158354)
		union = &ErrorDetail_NotLeaseHolder{t}
	case *RangeNotFoundError:
		__antithesis_instrumentation__.Notify(158355)
		union = &ErrorDetail_RangeNotFound{t}
	case *RangeKeyMismatchError:
		__antithesis_instrumentation__.Notify(158356)
		union = &ErrorDetail_RangeKeyMismatch{t}
	case *ReadWithinUncertaintyIntervalError:
		__antithesis_instrumentation__.Notify(158357)
		union = &ErrorDetail_ReadWithinUncertaintyInterval{t}
	case *TransactionAbortedError:
		__antithesis_instrumentation__.Notify(158358)
		union = &ErrorDetail_TransactionAborted{t}
	case *TransactionPushError:
		__antithesis_instrumentation__.Notify(158359)
		union = &ErrorDetail_TransactionPush{t}
	case *TransactionRetryError:
		__antithesis_instrumentation__.Notify(158360)
		union = &ErrorDetail_TransactionRetry{t}
	case *TransactionStatusError:
		__antithesis_instrumentation__.Notify(158361)
		union = &ErrorDetail_TransactionStatus{t}
	case *WriteIntentError:
		__antithesis_instrumentation__.Notify(158362)
		union = &ErrorDetail_WriteIntent{t}
	case *WriteTooOldError:
		__antithesis_instrumentation__.Notify(158363)
		union = &ErrorDetail_WriteTooOld{t}
	case *OpRequiresTxnError:
		__antithesis_instrumentation__.Notify(158364)
		union = &ErrorDetail_OpRequiresTxn{t}
	case *ConditionFailedError:
		__antithesis_instrumentation__.Notify(158365)
		union = &ErrorDetail_ConditionFailed{t}
	case *LeaseRejectedError:
		__antithesis_instrumentation__.Notify(158366)
		union = &ErrorDetail_LeaseRejected{t}
	case *NodeUnavailableError:
		__antithesis_instrumentation__.Notify(158367)
		union = &ErrorDetail_NodeUnavailable{t}
	case *RaftGroupDeletedError:
		__antithesis_instrumentation__.Notify(158368)
		union = &ErrorDetail_RaftGroupDeleted{t}
	case *ReplicaCorruptionError:
		__antithesis_instrumentation__.Notify(158369)
		union = &ErrorDetail_ReplicaCorruption{t}
	case *ReplicaTooOldError:
		__antithesis_instrumentation__.Notify(158370)
		union = &ErrorDetail_ReplicaTooOld{t}
	case *AmbiguousResultError:
		__antithesis_instrumentation__.Notify(158371)
		union = &ErrorDetail_AmbiguousResult{t}
	case *StoreNotFoundError:
		__antithesis_instrumentation__.Notify(158372)
		union = &ErrorDetail_StoreNotFound{t}
	case *TransactionRetryWithProtoRefreshError:
		__antithesis_instrumentation__.Notify(158373)
		union = &ErrorDetail_TransactionRetryWithProtoRefresh{t}
	case *IntegerOverflowError:
		__antithesis_instrumentation__.Notify(158374)
		union = &ErrorDetail_IntegerOverflow{t}
	case *UnsupportedRequestError:
		__antithesis_instrumentation__.Notify(158375)
		union = &ErrorDetail_UnsupportedRequest{t}
	case *BatchTimestampBeforeGCError:
		__antithesis_instrumentation__.Notify(158376)
		union = &ErrorDetail_TimestampBefore{t}
	case *TxnAlreadyEncounteredErrorError:
		__antithesis_instrumentation__.Notify(158377)
		union = &ErrorDetail_TxnAlreadyEncounteredError{t}
	case *IntentMissingError:
		__antithesis_instrumentation__.Notify(158378)
		union = &ErrorDetail_IntentMissing{t}
	case *MergeInProgressError:
		__antithesis_instrumentation__.Notify(158379)
		union = &ErrorDetail_MergeInProgress{t}
	case *RangeFeedRetryError:
		__antithesis_instrumentation__.Notify(158380)
		union = &ErrorDetail_RangefeedRetry{t}
	case *IndeterminateCommitError:
		__antithesis_instrumentation__.Notify(158381)
		union = &ErrorDetail_IndeterminateCommit{t}
	case *InvalidLeaseError:
		__antithesis_instrumentation__.Notify(158382)
		union = &ErrorDetail_InvalidLeaseError{t}
	case *OptimisticEvalConflictsError:
		__antithesis_instrumentation__.Notify(158383)
		union = &ErrorDetail_OptimisticEvalConflicts{t}
	case *MinTimestampBoundUnsatisfiableError:
		__antithesis_instrumentation__.Notify(158384)
		union = &ErrorDetail_MinTimestampBoundUnsatisfiable{t}
	case *RefreshFailedError:
		__antithesis_instrumentation__.Notify(158385)
		union = &ErrorDetail_RefreshFailedError{t}
	case *MVCCHistoryMutationError:
		__antithesis_instrumentation__.Notify(158386)
		union = &ErrorDetail_MVCCHistoryMutation{t}
	default:
		__antithesis_instrumentation__.Notify(158387)
		panic(fmt.Sprintf("unsupported type %T for %T", r, ru))
	}
	__antithesis_instrumentation__.Notify(158353)
	ru.Value = union
}

func (ru *RequestUnion) MustSetInner(r Request) {
	__antithesis_instrumentation__.Notify(158388)
	ru.Reset()
	var union isRequestUnion_Value
	switch t := r.(type) {
	case *GetRequest:
		__antithesis_instrumentation__.Notify(158390)
		union = &RequestUnion_Get{t}
	case *PutRequest:
		__antithesis_instrumentation__.Notify(158391)
		union = &RequestUnion_Put{t}
	case *ConditionalPutRequest:
		__antithesis_instrumentation__.Notify(158392)
		union = &RequestUnion_ConditionalPut{t}
	case *IncrementRequest:
		__antithesis_instrumentation__.Notify(158393)
		union = &RequestUnion_Increment{t}
	case *DeleteRequest:
		__antithesis_instrumentation__.Notify(158394)
		union = &RequestUnion_Delete{t}
	case *DeleteRangeRequest:
		__antithesis_instrumentation__.Notify(158395)
		union = &RequestUnion_DeleteRange{t}
	case *ClearRangeRequest:
		__antithesis_instrumentation__.Notify(158396)
		union = &RequestUnion_ClearRange{t}
	case *RevertRangeRequest:
		__antithesis_instrumentation__.Notify(158397)
		union = &RequestUnion_RevertRange{t}
	case *ScanRequest:
		__antithesis_instrumentation__.Notify(158398)
		union = &RequestUnion_Scan{t}
	case *EndTxnRequest:
		__antithesis_instrumentation__.Notify(158399)
		union = &RequestUnion_EndTxn{t}
	case *AdminSplitRequest:
		__antithesis_instrumentation__.Notify(158400)
		union = &RequestUnion_AdminSplit{t}
	case *AdminUnsplitRequest:
		__antithesis_instrumentation__.Notify(158401)
		union = &RequestUnion_AdminUnsplit{t}
	case *AdminMergeRequest:
		__antithesis_instrumentation__.Notify(158402)
		union = &RequestUnion_AdminMerge{t}
	case *AdminTransferLeaseRequest:
		__antithesis_instrumentation__.Notify(158403)
		union = &RequestUnion_AdminTransferLease{t}
	case *AdminChangeReplicasRequest:
		__antithesis_instrumentation__.Notify(158404)
		union = &RequestUnion_AdminChangeReplicas{t}
	case *AdminRelocateRangeRequest:
		__antithesis_instrumentation__.Notify(158405)
		union = &RequestUnion_AdminRelocateRange{t}
	case *HeartbeatTxnRequest:
		__antithesis_instrumentation__.Notify(158406)
		union = &RequestUnion_HeartbeatTxn{t}
	case *GCRequest:
		__antithesis_instrumentation__.Notify(158407)
		union = &RequestUnion_Gc{t}
	case *PushTxnRequest:
		__antithesis_instrumentation__.Notify(158408)
		union = &RequestUnion_PushTxn{t}
	case *RecoverTxnRequest:
		__antithesis_instrumentation__.Notify(158409)
		union = &RequestUnion_RecoverTxn{t}
	case *ResolveIntentRequest:
		__antithesis_instrumentation__.Notify(158410)
		union = &RequestUnion_ResolveIntent{t}
	case *ResolveIntentRangeRequest:
		__antithesis_instrumentation__.Notify(158411)
		union = &RequestUnion_ResolveIntentRange{t}
	case *MergeRequest:
		__antithesis_instrumentation__.Notify(158412)
		union = &RequestUnion_Merge{t}
	case *TruncateLogRequest:
		__antithesis_instrumentation__.Notify(158413)
		union = &RequestUnion_TruncateLog{t}
	case *RequestLeaseRequest:
		__antithesis_instrumentation__.Notify(158414)
		union = &RequestUnion_RequestLease{t}
	case *ReverseScanRequest:
		__antithesis_instrumentation__.Notify(158415)
		union = &RequestUnion_ReverseScan{t}
	case *ComputeChecksumRequest:
		__antithesis_instrumentation__.Notify(158416)
		union = &RequestUnion_ComputeChecksum{t}
	case *CheckConsistencyRequest:
		__antithesis_instrumentation__.Notify(158417)
		union = &RequestUnion_CheckConsistency{t}
	case *InitPutRequest:
		__antithesis_instrumentation__.Notify(158418)
		union = &RequestUnion_InitPut{t}
	case *TransferLeaseRequest:
		__antithesis_instrumentation__.Notify(158419)
		union = &RequestUnion_TransferLease{t}
	case *LeaseInfoRequest:
		__antithesis_instrumentation__.Notify(158420)
		union = &RequestUnion_LeaseInfo{t}
	case *ExportRequest:
		__antithesis_instrumentation__.Notify(158421)
		union = &RequestUnion_Export{t}
	case *QueryTxnRequest:
		__antithesis_instrumentation__.Notify(158422)
		union = &RequestUnion_QueryTxn{t}
	case *QueryIntentRequest:
		__antithesis_instrumentation__.Notify(158423)
		union = &RequestUnion_QueryIntent{t}
	case *QueryLocksRequest:
		__antithesis_instrumentation__.Notify(158424)
		union = &RequestUnion_QueryLocks{t}
	case *AdminScatterRequest:
		__antithesis_instrumentation__.Notify(158425)
		union = &RequestUnion_AdminScatter{t}
	case *AddSSTableRequest:
		__antithesis_instrumentation__.Notify(158426)
		union = &RequestUnion_AddSstable{t}
	case *RecomputeStatsRequest:
		__antithesis_instrumentation__.Notify(158427)
		union = &RequestUnion_RecomputeStats{t}
	case *RefreshRequest:
		__antithesis_instrumentation__.Notify(158428)
		union = &RequestUnion_Refresh{t}
	case *RefreshRangeRequest:
		__antithesis_instrumentation__.Notify(158429)
		union = &RequestUnion_RefreshRange{t}
	case *SubsumeRequest:
		__antithesis_instrumentation__.Notify(158430)
		union = &RequestUnion_Subsume{t}
	case *RangeStatsRequest:
		__antithesis_instrumentation__.Notify(158431)
		union = &RequestUnion_RangeStats{t}
	case *AdminVerifyProtectedTimestampRequest:
		__antithesis_instrumentation__.Notify(158432)
		union = &RequestUnion_AdminVerifyProtectedTimestamp{t}
	case *MigrateRequest:
		__antithesis_instrumentation__.Notify(158433)
		union = &RequestUnion_Migrate{t}
	case *QueryResolvedTimestampRequest:
		__antithesis_instrumentation__.Notify(158434)
		union = &RequestUnion_QueryResolvedTimestamp{t}
	case *ScanInterleavedIntentsRequest:
		__antithesis_instrumentation__.Notify(158435)
		union = &RequestUnion_ScanInterleavedIntents{t}
	case *BarrierRequest:
		__antithesis_instrumentation__.Notify(158436)
		union = &RequestUnion_Barrier{t}
	case *ProbeRequest:
		__antithesis_instrumentation__.Notify(158437)
		union = &RequestUnion_Probe{t}
	default:
		__antithesis_instrumentation__.Notify(158438)
		panic(fmt.Sprintf("unsupported type %T for %T", r, ru))
	}
	__antithesis_instrumentation__.Notify(158389)
	ru.Value = union
}

func (ru *ResponseUnion) MustSetInner(r Response) {
	__antithesis_instrumentation__.Notify(158439)
	ru.Reset()
	var union isResponseUnion_Value
	switch t := r.(type) {
	case *GetResponse:
		__antithesis_instrumentation__.Notify(158441)
		union = &ResponseUnion_Get{t}
	case *PutResponse:
		__antithesis_instrumentation__.Notify(158442)
		union = &ResponseUnion_Put{t}
	case *ConditionalPutResponse:
		__antithesis_instrumentation__.Notify(158443)
		union = &ResponseUnion_ConditionalPut{t}
	case *IncrementResponse:
		__antithesis_instrumentation__.Notify(158444)
		union = &ResponseUnion_Increment{t}
	case *DeleteResponse:
		__antithesis_instrumentation__.Notify(158445)
		union = &ResponseUnion_Delete{t}
	case *DeleteRangeResponse:
		__antithesis_instrumentation__.Notify(158446)
		union = &ResponseUnion_DeleteRange{t}
	case *ClearRangeResponse:
		__antithesis_instrumentation__.Notify(158447)
		union = &ResponseUnion_ClearRange{t}
	case *RevertRangeResponse:
		__antithesis_instrumentation__.Notify(158448)
		union = &ResponseUnion_RevertRange{t}
	case *ScanResponse:
		__antithesis_instrumentation__.Notify(158449)
		union = &ResponseUnion_Scan{t}
	case *EndTxnResponse:
		__antithesis_instrumentation__.Notify(158450)
		union = &ResponseUnion_EndTxn{t}
	case *AdminSplitResponse:
		__antithesis_instrumentation__.Notify(158451)
		union = &ResponseUnion_AdminSplit{t}
	case *AdminUnsplitResponse:
		__antithesis_instrumentation__.Notify(158452)
		union = &ResponseUnion_AdminUnsplit{t}
	case *AdminMergeResponse:
		__antithesis_instrumentation__.Notify(158453)
		union = &ResponseUnion_AdminMerge{t}
	case *AdminTransferLeaseResponse:
		__antithesis_instrumentation__.Notify(158454)
		union = &ResponseUnion_AdminTransferLease{t}
	case *AdminChangeReplicasResponse:
		__antithesis_instrumentation__.Notify(158455)
		union = &ResponseUnion_AdminChangeReplicas{t}
	case *AdminRelocateRangeResponse:
		__antithesis_instrumentation__.Notify(158456)
		union = &ResponseUnion_AdminRelocateRange{t}
	case *HeartbeatTxnResponse:
		__antithesis_instrumentation__.Notify(158457)
		union = &ResponseUnion_HeartbeatTxn{t}
	case *GCResponse:
		__antithesis_instrumentation__.Notify(158458)
		union = &ResponseUnion_Gc{t}
	case *PushTxnResponse:
		__antithesis_instrumentation__.Notify(158459)
		union = &ResponseUnion_PushTxn{t}
	case *RecoverTxnResponse:
		__antithesis_instrumentation__.Notify(158460)
		union = &ResponseUnion_RecoverTxn{t}
	case *ResolveIntentResponse:
		__antithesis_instrumentation__.Notify(158461)
		union = &ResponseUnion_ResolveIntent{t}
	case *ResolveIntentRangeResponse:
		__antithesis_instrumentation__.Notify(158462)
		union = &ResponseUnion_ResolveIntentRange{t}
	case *MergeResponse:
		__antithesis_instrumentation__.Notify(158463)
		union = &ResponseUnion_Merge{t}
	case *TruncateLogResponse:
		__antithesis_instrumentation__.Notify(158464)
		union = &ResponseUnion_TruncateLog{t}
	case *RequestLeaseResponse:
		__antithesis_instrumentation__.Notify(158465)
		union = &ResponseUnion_RequestLease{t}
	case *ReverseScanResponse:
		__antithesis_instrumentation__.Notify(158466)
		union = &ResponseUnion_ReverseScan{t}
	case *ComputeChecksumResponse:
		__antithesis_instrumentation__.Notify(158467)
		union = &ResponseUnion_ComputeChecksum{t}
	case *CheckConsistencyResponse:
		__antithesis_instrumentation__.Notify(158468)
		union = &ResponseUnion_CheckConsistency{t}
	case *InitPutResponse:
		__antithesis_instrumentation__.Notify(158469)
		union = &ResponseUnion_InitPut{t}
	case *LeaseInfoResponse:
		__antithesis_instrumentation__.Notify(158470)
		union = &ResponseUnion_LeaseInfo{t}
	case *ExportResponse:
		__antithesis_instrumentation__.Notify(158471)
		union = &ResponseUnion_Export{t}
	case *QueryTxnResponse:
		__antithesis_instrumentation__.Notify(158472)
		union = &ResponseUnion_QueryTxn{t}
	case *QueryIntentResponse:
		__antithesis_instrumentation__.Notify(158473)
		union = &ResponseUnion_QueryIntent{t}
	case *QueryLocksResponse:
		__antithesis_instrumentation__.Notify(158474)
		union = &ResponseUnion_QueryLocks{t}
	case *AdminScatterResponse:
		__antithesis_instrumentation__.Notify(158475)
		union = &ResponseUnion_AdminScatter{t}
	case *AddSSTableResponse:
		__antithesis_instrumentation__.Notify(158476)
		union = &ResponseUnion_AddSstable{t}
	case *RecomputeStatsResponse:
		__antithesis_instrumentation__.Notify(158477)
		union = &ResponseUnion_RecomputeStats{t}
	case *RefreshResponse:
		__antithesis_instrumentation__.Notify(158478)
		union = &ResponseUnion_Refresh{t}
	case *RefreshRangeResponse:
		__antithesis_instrumentation__.Notify(158479)
		union = &ResponseUnion_RefreshRange{t}
	case *SubsumeResponse:
		__antithesis_instrumentation__.Notify(158480)
		union = &ResponseUnion_Subsume{t}
	case *RangeStatsResponse:
		__antithesis_instrumentation__.Notify(158481)
		union = &ResponseUnion_RangeStats{t}
	case *AdminVerifyProtectedTimestampResponse:
		__antithesis_instrumentation__.Notify(158482)
		union = &ResponseUnion_AdminVerifyProtectedTimestamp{t}
	case *MigrateResponse:
		__antithesis_instrumentation__.Notify(158483)
		union = &ResponseUnion_Migrate{t}
	case *QueryResolvedTimestampResponse:
		__antithesis_instrumentation__.Notify(158484)
		union = &ResponseUnion_QueryResolvedTimestamp{t}
	case *ScanInterleavedIntentsResponse:
		__antithesis_instrumentation__.Notify(158485)
		union = &ResponseUnion_ScanInterleavedIntents{t}
	case *BarrierResponse:
		__antithesis_instrumentation__.Notify(158486)
		union = &ResponseUnion_Barrier{t}
	case *ProbeResponse:
		__antithesis_instrumentation__.Notify(158487)
		union = &ResponseUnion_Probe{t}
	default:
		__antithesis_instrumentation__.Notify(158488)
		panic(fmt.Sprintf("unsupported type %T for %T", r, ru))
	}
	__antithesis_instrumentation__.Notify(158440)
	ru.Value = union
}

type reqCounts [48]int32

func (ba *BatchRequest) getReqCounts() reqCounts {
	__antithesis_instrumentation__.Notify(158489)
	var counts reqCounts
	for _, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(158491)
		switch ru.GetValue().(type) {
		case *RequestUnion_Get:
			__antithesis_instrumentation__.Notify(158492)
			counts[0]++
		case *RequestUnion_Put:
			__antithesis_instrumentation__.Notify(158493)
			counts[1]++
		case *RequestUnion_ConditionalPut:
			__antithesis_instrumentation__.Notify(158494)
			counts[2]++
		case *RequestUnion_Increment:
			__antithesis_instrumentation__.Notify(158495)
			counts[3]++
		case *RequestUnion_Delete:
			__antithesis_instrumentation__.Notify(158496)
			counts[4]++
		case *RequestUnion_DeleteRange:
			__antithesis_instrumentation__.Notify(158497)
			counts[5]++
		case *RequestUnion_ClearRange:
			__antithesis_instrumentation__.Notify(158498)
			counts[6]++
		case *RequestUnion_RevertRange:
			__antithesis_instrumentation__.Notify(158499)
			counts[7]++
		case *RequestUnion_Scan:
			__antithesis_instrumentation__.Notify(158500)
			counts[8]++
		case *RequestUnion_EndTxn:
			__antithesis_instrumentation__.Notify(158501)
			counts[9]++
		case *RequestUnion_AdminSplit:
			__antithesis_instrumentation__.Notify(158502)
			counts[10]++
		case *RequestUnion_AdminUnsplit:
			__antithesis_instrumentation__.Notify(158503)
			counts[11]++
		case *RequestUnion_AdminMerge:
			__antithesis_instrumentation__.Notify(158504)
			counts[12]++
		case *RequestUnion_AdminTransferLease:
			__antithesis_instrumentation__.Notify(158505)
			counts[13]++
		case *RequestUnion_AdminChangeReplicas:
			__antithesis_instrumentation__.Notify(158506)
			counts[14]++
		case *RequestUnion_AdminRelocateRange:
			__antithesis_instrumentation__.Notify(158507)
			counts[15]++
		case *RequestUnion_HeartbeatTxn:
			__antithesis_instrumentation__.Notify(158508)
			counts[16]++
		case *RequestUnion_Gc:
			__antithesis_instrumentation__.Notify(158509)
			counts[17]++
		case *RequestUnion_PushTxn:
			__antithesis_instrumentation__.Notify(158510)
			counts[18]++
		case *RequestUnion_RecoverTxn:
			__antithesis_instrumentation__.Notify(158511)
			counts[19]++
		case *RequestUnion_ResolveIntent:
			__antithesis_instrumentation__.Notify(158512)
			counts[20]++
		case *RequestUnion_ResolveIntentRange:
			__antithesis_instrumentation__.Notify(158513)
			counts[21]++
		case *RequestUnion_Merge:
			__antithesis_instrumentation__.Notify(158514)
			counts[22]++
		case *RequestUnion_TruncateLog:
			__antithesis_instrumentation__.Notify(158515)
			counts[23]++
		case *RequestUnion_RequestLease:
			__antithesis_instrumentation__.Notify(158516)
			counts[24]++
		case *RequestUnion_ReverseScan:
			__antithesis_instrumentation__.Notify(158517)
			counts[25]++
		case *RequestUnion_ComputeChecksum:
			__antithesis_instrumentation__.Notify(158518)
			counts[26]++
		case *RequestUnion_CheckConsistency:
			__antithesis_instrumentation__.Notify(158519)
			counts[27]++
		case *RequestUnion_InitPut:
			__antithesis_instrumentation__.Notify(158520)
			counts[28]++
		case *RequestUnion_TransferLease:
			__antithesis_instrumentation__.Notify(158521)
			counts[29]++
		case *RequestUnion_LeaseInfo:
			__antithesis_instrumentation__.Notify(158522)
			counts[30]++
		case *RequestUnion_Export:
			__antithesis_instrumentation__.Notify(158523)
			counts[31]++
		case *RequestUnion_QueryTxn:
			__antithesis_instrumentation__.Notify(158524)
			counts[32]++
		case *RequestUnion_QueryIntent:
			__antithesis_instrumentation__.Notify(158525)
			counts[33]++
		case *RequestUnion_QueryLocks:
			__antithesis_instrumentation__.Notify(158526)
			counts[34]++
		case *RequestUnion_AdminScatter:
			__antithesis_instrumentation__.Notify(158527)
			counts[35]++
		case *RequestUnion_AddSstable:
			__antithesis_instrumentation__.Notify(158528)
			counts[36]++
		case *RequestUnion_RecomputeStats:
			__antithesis_instrumentation__.Notify(158529)
			counts[37]++
		case *RequestUnion_Refresh:
			__antithesis_instrumentation__.Notify(158530)
			counts[38]++
		case *RequestUnion_RefreshRange:
			__antithesis_instrumentation__.Notify(158531)
			counts[39]++
		case *RequestUnion_Subsume:
			__antithesis_instrumentation__.Notify(158532)
			counts[40]++
		case *RequestUnion_RangeStats:
			__antithesis_instrumentation__.Notify(158533)
			counts[41]++
		case *RequestUnion_AdminVerifyProtectedTimestamp:
			__antithesis_instrumentation__.Notify(158534)
			counts[42]++
		case *RequestUnion_Migrate:
			__antithesis_instrumentation__.Notify(158535)
			counts[43]++
		case *RequestUnion_QueryResolvedTimestamp:
			__antithesis_instrumentation__.Notify(158536)
			counts[44]++
		case *RequestUnion_ScanInterleavedIntents:
			__antithesis_instrumentation__.Notify(158537)
			counts[45]++
		case *RequestUnion_Barrier:
			__antithesis_instrumentation__.Notify(158538)
			counts[46]++
		case *RequestUnion_Probe:
			__antithesis_instrumentation__.Notify(158539)
			counts[47]++
		default:
			__antithesis_instrumentation__.Notify(158540)
			panic(fmt.Sprintf("unsupported request: %+v", ru))
		}
	}
	__antithesis_instrumentation__.Notify(158490)
	return counts
}

var requestNames = []string{
	"Get",
	"Put",
	"CPut",
	"Inc",
	"Del",
	"DelRng",
	"ClearRng",
	"RevertRng",
	"Scan",
	"EndTxn",
	"AdmSplit",
	"AdmUnsplit",
	"AdmMerge",
	"AdmTransferLease",
	"AdmChangeReplicas",
	"AdmRelocateRng",
	"HeartbeatTxn",
	"Gc",
	"PushTxn",
	"RecoverTxn",
	"ResolveIntent",
	"ResolveIntentRng",
	"Merge",
	"TruncLog",
	"RequestLease",
	"RevScan",
	"ComputeChksum",
	"ChkConsistency",
	"InitPut",
	"TransferLease",
	"LeaseInfo",
	"Export",
	"QueryTxn",
	"QueryIntent",
	"QueryLocks",
	"AdmScatter",
	"AddSstable",
	"RecomputeStats",
	"Refresh",
	"RefreshRng",
	"Subsume",
	"RngStats",
	"AdmVerifyProtectedTimestamp",
	"Migrate",
	"QueryResolvedTimestamp",
	"ScanInterleavedIntents",
	"Barrier",
	"Probe",
}

func (ba *BatchRequest) Summary() string {
	__antithesis_instrumentation__.Notify(158541)
	var b strings.Builder
	ba.WriteSummary(&b)
	return b.String()
}

func (ba *BatchRequest) WriteSummary(b *strings.Builder) {
	__antithesis_instrumentation__.Notify(158542)
	if len(ba.Requests) == 0 {
		__antithesis_instrumentation__.Notify(158544)
		b.WriteString("empty batch")
		return
	} else {
		__antithesis_instrumentation__.Notify(158545)
	}
	__antithesis_instrumentation__.Notify(158543)
	counts := ba.getReqCounts()
	var tmp [10]byte
	var comma bool
	for i, v := range counts {
		__antithesis_instrumentation__.Notify(158546)
		if v != 0 {
			__antithesis_instrumentation__.Notify(158547)
			if comma {
				__antithesis_instrumentation__.Notify(158549)
				b.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(158550)
			}
			__antithesis_instrumentation__.Notify(158548)
			comma = true

			b.Write(strconv.AppendInt(tmp[:0], int64(v), 10))
			b.WriteString(" ")
			b.WriteString(requestNames[i])
		} else {
			__antithesis_instrumentation__.Notify(158551)
		}
	}
}

type getResponseAlloc struct {
	union ResponseUnion_Get
	resp  GetResponse
}
type putResponseAlloc struct {
	union ResponseUnion_Put
	resp  PutResponse
}
type conditionalPutResponseAlloc struct {
	union ResponseUnion_ConditionalPut
	resp  ConditionalPutResponse
}
type incrementResponseAlloc struct {
	union ResponseUnion_Increment
	resp  IncrementResponse
}
type deleteResponseAlloc struct {
	union ResponseUnion_Delete
	resp  DeleteResponse
}
type deleteRangeResponseAlloc struct {
	union ResponseUnion_DeleteRange
	resp  DeleteRangeResponse
}
type clearRangeResponseAlloc struct {
	union ResponseUnion_ClearRange
	resp  ClearRangeResponse
}
type revertRangeResponseAlloc struct {
	union ResponseUnion_RevertRange
	resp  RevertRangeResponse
}
type scanResponseAlloc struct {
	union ResponseUnion_Scan
	resp  ScanResponse
}
type endTxnResponseAlloc struct {
	union ResponseUnion_EndTxn
	resp  EndTxnResponse
}
type adminSplitResponseAlloc struct {
	union ResponseUnion_AdminSplit
	resp  AdminSplitResponse
}
type adminUnsplitResponseAlloc struct {
	union ResponseUnion_AdminUnsplit
	resp  AdminUnsplitResponse
}
type adminMergeResponseAlloc struct {
	union ResponseUnion_AdminMerge
	resp  AdminMergeResponse
}
type adminTransferLeaseResponseAlloc struct {
	union ResponseUnion_AdminTransferLease
	resp  AdminTransferLeaseResponse
}
type adminChangeReplicasResponseAlloc struct {
	union ResponseUnion_AdminChangeReplicas
	resp  AdminChangeReplicasResponse
}
type adminRelocateRangeResponseAlloc struct {
	union ResponseUnion_AdminRelocateRange
	resp  AdminRelocateRangeResponse
}
type heartbeatTxnResponseAlloc struct {
	union ResponseUnion_HeartbeatTxn
	resp  HeartbeatTxnResponse
}
type gCResponseAlloc struct {
	union ResponseUnion_Gc
	resp  GCResponse
}
type pushTxnResponseAlloc struct {
	union ResponseUnion_PushTxn
	resp  PushTxnResponse
}
type recoverTxnResponseAlloc struct {
	union ResponseUnion_RecoverTxn
	resp  RecoverTxnResponse
}
type resolveIntentResponseAlloc struct {
	union ResponseUnion_ResolveIntent
	resp  ResolveIntentResponse
}
type resolveIntentRangeResponseAlloc struct {
	union ResponseUnion_ResolveIntentRange
	resp  ResolveIntentRangeResponse
}
type mergeResponseAlloc struct {
	union ResponseUnion_Merge
	resp  MergeResponse
}
type truncateLogResponseAlloc struct {
	union ResponseUnion_TruncateLog
	resp  TruncateLogResponse
}
type requestLeaseResponseAlloc struct {
	union ResponseUnion_RequestLease
	resp  RequestLeaseResponse
}
type reverseScanResponseAlloc struct {
	union ResponseUnion_ReverseScan
	resp  ReverseScanResponse
}
type computeChecksumResponseAlloc struct {
	union ResponseUnion_ComputeChecksum
	resp  ComputeChecksumResponse
}
type checkConsistencyResponseAlloc struct {
	union ResponseUnion_CheckConsistency
	resp  CheckConsistencyResponse
}
type initPutResponseAlloc struct {
	union ResponseUnion_InitPut
	resp  InitPutResponse
}
type leaseInfoResponseAlloc struct {
	union ResponseUnion_LeaseInfo
	resp  LeaseInfoResponse
}
type exportResponseAlloc struct {
	union ResponseUnion_Export
	resp  ExportResponse
}
type queryTxnResponseAlloc struct {
	union ResponseUnion_QueryTxn
	resp  QueryTxnResponse
}
type queryIntentResponseAlloc struct {
	union ResponseUnion_QueryIntent
	resp  QueryIntentResponse
}
type queryLocksResponseAlloc struct {
	union ResponseUnion_QueryLocks
	resp  QueryLocksResponse
}
type adminScatterResponseAlloc struct {
	union ResponseUnion_AdminScatter
	resp  AdminScatterResponse
}
type addSSTableResponseAlloc struct {
	union ResponseUnion_AddSstable
	resp  AddSSTableResponse
}
type recomputeStatsResponseAlloc struct {
	union ResponseUnion_RecomputeStats
	resp  RecomputeStatsResponse
}
type refreshResponseAlloc struct {
	union ResponseUnion_Refresh
	resp  RefreshResponse
}
type refreshRangeResponseAlloc struct {
	union ResponseUnion_RefreshRange
	resp  RefreshRangeResponse
}
type subsumeResponseAlloc struct {
	union ResponseUnion_Subsume
	resp  SubsumeResponse
}
type rangeStatsResponseAlloc struct {
	union ResponseUnion_RangeStats
	resp  RangeStatsResponse
}
type adminVerifyProtectedTimestampResponseAlloc struct {
	union ResponseUnion_AdminVerifyProtectedTimestamp
	resp  AdminVerifyProtectedTimestampResponse
}
type migrateResponseAlloc struct {
	union ResponseUnion_Migrate
	resp  MigrateResponse
}
type queryResolvedTimestampResponseAlloc struct {
	union ResponseUnion_QueryResolvedTimestamp
	resp  QueryResolvedTimestampResponse
}
type scanInterleavedIntentsResponseAlloc struct {
	union ResponseUnion_ScanInterleavedIntents
	resp  ScanInterleavedIntentsResponse
}
type barrierResponseAlloc struct {
	union ResponseUnion_Barrier
	resp  BarrierResponse
}
type probeResponseAlloc struct {
	union ResponseUnion_Probe
	resp  ProbeResponse
}

func (ba *BatchRequest) CreateReply() *BatchResponse {
	__antithesis_instrumentation__.Notify(158552)
	br := &BatchResponse{}
	br.Responses = make([]ResponseUnion, len(ba.Requests))

	counts := ba.getReqCounts()

	var buf0 []getResponseAlloc
	var buf1 []putResponseAlloc
	var buf2 []conditionalPutResponseAlloc
	var buf3 []incrementResponseAlloc
	var buf4 []deleteResponseAlloc
	var buf5 []deleteRangeResponseAlloc
	var buf6 []clearRangeResponseAlloc
	var buf7 []revertRangeResponseAlloc
	var buf8 []scanResponseAlloc
	var buf9 []endTxnResponseAlloc
	var buf10 []adminSplitResponseAlloc
	var buf11 []adminUnsplitResponseAlloc
	var buf12 []adminMergeResponseAlloc
	var buf13 []adminTransferLeaseResponseAlloc
	var buf14 []adminChangeReplicasResponseAlloc
	var buf15 []adminRelocateRangeResponseAlloc
	var buf16 []heartbeatTxnResponseAlloc
	var buf17 []gCResponseAlloc
	var buf18 []pushTxnResponseAlloc
	var buf19 []recoverTxnResponseAlloc
	var buf20 []resolveIntentResponseAlloc
	var buf21 []resolveIntentRangeResponseAlloc
	var buf22 []mergeResponseAlloc
	var buf23 []truncateLogResponseAlloc
	var buf24 []requestLeaseResponseAlloc
	var buf25 []reverseScanResponseAlloc
	var buf26 []computeChecksumResponseAlloc
	var buf27 []checkConsistencyResponseAlloc
	var buf28 []initPutResponseAlloc
	var buf29 []requestLeaseResponseAlloc
	var buf30 []leaseInfoResponseAlloc
	var buf31 []exportResponseAlloc
	var buf32 []queryTxnResponseAlloc
	var buf33 []queryIntentResponseAlloc
	var buf34 []queryLocksResponseAlloc
	var buf35 []adminScatterResponseAlloc
	var buf36 []addSSTableResponseAlloc
	var buf37 []recomputeStatsResponseAlloc
	var buf38 []refreshResponseAlloc
	var buf39 []refreshRangeResponseAlloc
	var buf40 []subsumeResponseAlloc
	var buf41 []rangeStatsResponseAlloc
	var buf42 []adminVerifyProtectedTimestampResponseAlloc
	var buf43 []migrateResponseAlloc
	var buf44 []queryResolvedTimestampResponseAlloc
	var buf45 []scanInterleavedIntentsResponseAlloc
	var buf46 []barrierResponseAlloc
	var buf47 []probeResponseAlloc

	for i, r := range ba.Requests {
		__antithesis_instrumentation__.Notify(158554)
		switch r.GetValue().(type) {
		case *RequestUnion_Get:
			__antithesis_instrumentation__.Notify(158555)
			if buf0 == nil {
				__antithesis_instrumentation__.Notify(158652)
				buf0 = make([]getResponseAlloc, counts[0])
			} else {
				__antithesis_instrumentation__.Notify(158653)
			}
			__antithesis_instrumentation__.Notify(158556)
			buf0[0].union.Get = &buf0[0].resp
			br.Responses[i].Value = &buf0[0].union
			buf0 = buf0[1:]
		case *RequestUnion_Put:
			__antithesis_instrumentation__.Notify(158557)
			if buf1 == nil {
				__antithesis_instrumentation__.Notify(158654)
				buf1 = make([]putResponseAlloc, counts[1])
			} else {
				__antithesis_instrumentation__.Notify(158655)
			}
			__antithesis_instrumentation__.Notify(158558)
			buf1[0].union.Put = &buf1[0].resp
			br.Responses[i].Value = &buf1[0].union
			buf1 = buf1[1:]
		case *RequestUnion_ConditionalPut:
			__antithesis_instrumentation__.Notify(158559)
			if buf2 == nil {
				__antithesis_instrumentation__.Notify(158656)
				buf2 = make([]conditionalPutResponseAlloc, counts[2])
			} else {
				__antithesis_instrumentation__.Notify(158657)
			}
			__antithesis_instrumentation__.Notify(158560)
			buf2[0].union.ConditionalPut = &buf2[0].resp
			br.Responses[i].Value = &buf2[0].union
			buf2 = buf2[1:]
		case *RequestUnion_Increment:
			__antithesis_instrumentation__.Notify(158561)
			if buf3 == nil {
				__antithesis_instrumentation__.Notify(158658)
				buf3 = make([]incrementResponseAlloc, counts[3])
			} else {
				__antithesis_instrumentation__.Notify(158659)
			}
			__antithesis_instrumentation__.Notify(158562)
			buf3[0].union.Increment = &buf3[0].resp
			br.Responses[i].Value = &buf3[0].union
			buf3 = buf3[1:]
		case *RequestUnion_Delete:
			__antithesis_instrumentation__.Notify(158563)
			if buf4 == nil {
				__antithesis_instrumentation__.Notify(158660)
				buf4 = make([]deleteResponseAlloc, counts[4])
			} else {
				__antithesis_instrumentation__.Notify(158661)
			}
			__antithesis_instrumentation__.Notify(158564)
			buf4[0].union.Delete = &buf4[0].resp
			br.Responses[i].Value = &buf4[0].union
			buf4 = buf4[1:]
		case *RequestUnion_DeleteRange:
			__antithesis_instrumentation__.Notify(158565)
			if buf5 == nil {
				__antithesis_instrumentation__.Notify(158662)
				buf5 = make([]deleteRangeResponseAlloc, counts[5])
			} else {
				__antithesis_instrumentation__.Notify(158663)
			}
			__antithesis_instrumentation__.Notify(158566)
			buf5[0].union.DeleteRange = &buf5[0].resp
			br.Responses[i].Value = &buf5[0].union
			buf5 = buf5[1:]
		case *RequestUnion_ClearRange:
			__antithesis_instrumentation__.Notify(158567)
			if buf6 == nil {
				__antithesis_instrumentation__.Notify(158664)
				buf6 = make([]clearRangeResponseAlloc, counts[6])
			} else {
				__antithesis_instrumentation__.Notify(158665)
			}
			__antithesis_instrumentation__.Notify(158568)
			buf6[0].union.ClearRange = &buf6[0].resp
			br.Responses[i].Value = &buf6[0].union
			buf6 = buf6[1:]
		case *RequestUnion_RevertRange:
			__antithesis_instrumentation__.Notify(158569)
			if buf7 == nil {
				__antithesis_instrumentation__.Notify(158666)
				buf7 = make([]revertRangeResponseAlloc, counts[7])
			} else {
				__antithesis_instrumentation__.Notify(158667)
			}
			__antithesis_instrumentation__.Notify(158570)
			buf7[0].union.RevertRange = &buf7[0].resp
			br.Responses[i].Value = &buf7[0].union
			buf7 = buf7[1:]
		case *RequestUnion_Scan:
			__antithesis_instrumentation__.Notify(158571)
			if buf8 == nil {
				__antithesis_instrumentation__.Notify(158668)
				buf8 = make([]scanResponseAlloc, counts[8])
			} else {
				__antithesis_instrumentation__.Notify(158669)
			}
			__antithesis_instrumentation__.Notify(158572)
			buf8[0].union.Scan = &buf8[0].resp
			br.Responses[i].Value = &buf8[0].union
			buf8 = buf8[1:]
		case *RequestUnion_EndTxn:
			__antithesis_instrumentation__.Notify(158573)
			if buf9 == nil {
				__antithesis_instrumentation__.Notify(158670)
				buf9 = make([]endTxnResponseAlloc, counts[9])
			} else {
				__antithesis_instrumentation__.Notify(158671)
			}
			__antithesis_instrumentation__.Notify(158574)
			buf9[0].union.EndTxn = &buf9[0].resp
			br.Responses[i].Value = &buf9[0].union
			buf9 = buf9[1:]
		case *RequestUnion_AdminSplit:
			__antithesis_instrumentation__.Notify(158575)
			if buf10 == nil {
				__antithesis_instrumentation__.Notify(158672)
				buf10 = make([]adminSplitResponseAlloc, counts[10])
			} else {
				__antithesis_instrumentation__.Notify(158673)
			}
			__antithesis_instrumentation__.Notify(158576)
			buf10[0].union.AdminSplit = &buf10[0].resp
			br.Responses[i].Value = &buf10[0].union
			buf10 = buf10[1:]
		case *RequestUnion_AdminUnsplit:
			__antithesis_instrumentation__.Notify(158577)
			if buf11 == nil {
				__antithesis_instrumentation__.Notify(158674)
				buf11 = make([]adminUnsplitResponseAlloc, counts[11])
			} else {
				__antithesis_instrumentation__.Notify(158675)
			}
			__antithesis_instrumentation__.Notify(158578)
			buf11[0].union.AdminUnsplit = &buf11[0].resp
			br.Responses[i].Value = &buf11[0].union
			buf11 = buf11[1:]
		case *RequestUnion_AdminMerge:
			__antithesis_instrumentation__.Notify(158579)
			if buf12 == nil {
				__antithesis_instrumentation__.Notify(158676)
				buf12 = make([]adminMergeResponseAlloc, counts[12])
			} else {
				__antithesis_instrumentation__.Notify(158677)
			}
			__antithesis_instrumentation__.Notify(158580)
			buf12[0].union.AdminMerge = &buf12[0].resp
			br.Responses[i].Value = &buf12[0].union
			buf12 = buf12[1:]
		case *RequestUnion_AdminTransferLease:
			__antithesis_instrumentation__.Notify(158581)
			if buf13 == nil {
				__antithesis_instrumentation__.Notify(158678)
				buf13 = make([]adminTransferLeaseResponseAlloc, counts[13])
			} else {
				__antithesis_instrumentation__.Notify(158679)
			}
			__antithesis_instrumentation__.Notify(158582)
			buf13[0].union.AdminTransferLease = &buf13[0].resp
			br.Responses[i].Value = &buf13[0].union
			buf13 = buf13[1:]
		case *RequestUnion_AdminChangeReplicas:
			__antithesis_instrumentation__.Notify(158583)
			if buf14 == nil {
				__antithesis_instrumentation__.Notify(158680)
				buf14 = make([]adminChangeReplicasResponseAlloc, counts[14])
			} else {
				__antithesis_instrumentation__.Notify(158681)
			}
			__antithesis_instrumentation__.Notify(158584)
			buf14[0].union.AdminChangeReplicas = &buf14[0].resp
			br.Responses[i].Value = &buf14[0].union
			buf14 = buf14[1:]
		case *RequestUnion_AdminRelocateRange:
			__antithesis_instrumentation__.Notify(158585)
			if buf15 == nil {
				__antithesis_instrumentation__.Notify(158682)
				buf15 = make([]adminRelocateRangeResponseAlloc, counts[15])
			} else {
				__antithesis_instrumentation__.Notify(158683)
			}
			__antithesis_instrumentation__.Notify(158586)
			buf15[0].union.AdminRelocateRange = &buf15[0].resp
			br.Responses[i].Value = &buf15[0].union
			buf15 = buf15[1:]
		case *RequestUnion_HeartbeatTxn:
			__antithesis_instrumentation__.Notify(158587)
			if buf16 == nil {
				__antithesis_instrumentation__.Notify(158684)
				buf16 = make([]heartbeatTxnResponseAlloc, counts[16])
			} else {
				__antithesis_instrumentation__.Notify(158685)
			}
			__antithesis_instrumentation__.Notify(158588)
			buf16[0].union.HeartbeatTxn = &buf16[0].resp
			br.Responses[i].Value = &buf16[0].union
			buf16 = buf16[1:]
		case *RequestUnion_Gc:
			__antithesis_instrumentation__.Notify(158589)
			if buf17 == nil {
				__antithesis_instrumentation__.Notify(158686)
				buf17 = make([]gCResponseAlloc, counts[17])
			} else {
				__antithesis_instrumentation__.Notify(158687)
			}
			__antithesis_instrumentation__.Notify(158590)
			buf17[0].union.Gc = &buf17[0].resp
			br.Responses[i].Value = &buf17[0].union
			buf17 = buf17[1:]
		case *RequestUnion_PushTxn:
			__antithesis_instrumentation__.Notify(158591)
			if buf18 == nil {
				__antithesis_instrumentation__.Notify(158688)
				buf18 = make([]pushTxnResponseAlloc, counts[18])
			} else {
				__antithesis_instrumentation__.Notify(158689)
			}
			__antithesis_instrumentation__.Notify(158592)
			buf18[0].union.PushTxn = &buf18[0].resp
			br.Responses[i].Value = &buf18[0].union
			buf18 = buf18[1:]
		case *RequestUnion_RecoverTxn:
			__antithesis_instrumentation__.Notify(158593)
			if buf19 == nil {
				__antithesis_instrumentation__.Notify(158690)
				buf19 = make([]recoverTxnResponseAlloc, counts[19])
			} else {
				__antithesis_instrumentation__.Notify(158691)
			}
			__antithesis_instrumentation__.Notify(158594)
			buf19[0].union.RecoverTxn = &buf19[0].resp
			br.Responses[i].Value = &buf19[0].union
			buf19 = buf19[1:]
		case *RequestUnion_ResolveIntent:
			__antithesis_instrumentation__.Notify(158595)
			if buf20 == nil {
				__antithesis_instrumentation__.Notify(158692)
				buf20 = make([]resolveIntentResponseAlloc, counts[20])
			} else {
				__antithesis_instrumentation__.Notify(158693)
			}
			__antithesis_instrumentation__.Notify(158596)
			buf20[0].union.ResolveIntent = &buf20[0].resp
			br.Responses[i].Value = &buf20[0].union
			buf20 = buf20[1:]
		case *RequestUnion_ResolveIntentRange:
			__antithesis_instrumentation__.Notify(158597)
			if buf21 == nil {
				__antithesis_instrumentation__.Notify(158694)
				buf21 = make([]resolveIntentRangeResponseAlloc, counts[21])
			} else {
				__antithesis_instrumentation__.Notify(158695)
			}
			__antithesis_instrumentation__.Notify(158598)
			buf21[0].union.ResolveIntentRange = &buf21[0].resp
			br.Responses[i].Value = &buf21[0].union
			buf21 = buf21[1:]
		case *RequestUnion_Merge:
			__antithesis_instrumentation__.Notify(158599)
			if buf22 == nil {
				__antithesis_instrumentation__.Notify(158696)
				buf22 = make([]mergeResponseAlloc, counts[22])
			} else {
				__antithesis_instrumentation__.Notify(158697)
			}
			__antithesis_instrumentation__.Notify(158600)
			buf22[0].union.Merge = &buf22[0].resp
			br.Responses[i].Value = &buf22[0].union
			buf22 = buf22[1:]
		case *RequestUnion_TruncateLog:
			__antithesis_instrumentation__.Notify(158601)
			if buf23 == nil {
				__antithesis_instrumentation__.Notify(158698)
				buf23 = make([]truncateLogResponseAlloc, counts[23])
			} else {
				__antithesis_instrumentation__.Notify(158699)
			}
			__antithesis_instrumentation__.Notify(158602)
			buf23[0].union.TruncateLog = &buf23[0].resp
			br.Responses[i].Value = &buf23[0].union
			buf23 = buf23[1:]
		case *RequestUnion_RequestLease:
			__antithesis_instrumentation__.Notify(158603)
			if buf24 == nil {
				__antithesis_instrumentation__.Notify(158700)
				buf24 = make([]requestLeaseResponseAlloc, counts[24])
			} else {
				__antithesis_instrumentation__.Notify(158701)
			}
			__antithesis_instrumentation__.Notify(158604)
			buf24[0].union.RequestLease = &buf24[0].resp
			br.Responses[i].Value = &buf24[0].union
			buf24 = buf24[1:]
		case *RequestUnion_ReverseScan:
			__antithesis_instrumentation__.Notify(158605)
			if buf25 == nil {
				__antithesis_instrumentation__.Notify(158702)
				buf25 = make([]reverseScanResponseAlloc, counts[25])
			} else {
				__antithesis_instrumentation__.Notify(158703)
			}
			__antithesis_instrumentation__.Notify(158606)
			buf25[0].union.ReverseScan = &buf25[0].resp
			br.Responses[i].Value = &buf25[0].union
			buf25 = buf25[1:]
		case *RequestUnion_ComputeChecksum:
			__antithesis_instrumentation__.Notify(158607)
			if buf26 == nil {
				__antithesis_instrumentation__.Notify(158704)
				buf26 = make([]computeChecksumResponseAlloc, counts[26])
			} else {
				__antithesis_instrumentation__.Notify(158705)
			}
			__antithesis_instrumentation__.Notify(158608)
			buf26[0].union.ComputeChecksum = &buf26[0].resp
			br.Responses[i].Value = &buf26[0].union
			buf26 = buf26[1:]
		case *RequestUnion_CheckConsistency:
			__antithesis_instrumentation__.Notify(158609)
			if buf27 == nil {
				__antithesis_instrumentation__.Notify(158706)
				buf27 = make([]checkConsistencyResponseAlloc, counts[27])
			} else {
				__antithesis_instrumentation__.Notify(158707)
			}
			__antithesis_instrumentation__.Notify(158610)
			buf27[0].union.CheckConsistency = &buf27[0].resp
			br.Responses[i].Value = &buf27[0].union
			buf27 = buf27[1:]
		case *RequestUnion_InitPut:
			__antithesis_instrumentation__.Notify(158611)
			if buf28 == nil {
				__antithesis_instrumentation__.Notify(158708)
				buf28 = make([]initPutResponseAlloc, counts[28])
			} else {
				__antithesis_instrumentation__.Notify(158709)
			}
			__antithesis_instrumentation__.Notify(158612)
			buf28[0].union.InitPut = &buf28[0].resp
			br.Responses[i].Value = &buf28[0].union
			buf28 = buf28[1:]
		case *RequestUnion_TransferLease:
			__antithesis_instrumentation__.Notify(158613)
			if buf29 == nil {
				__antithesis_instrumentation__.Notify(158710)
				buf29 = make([]requestLeaseResponseAlloc, counts[29])
			} else {
				__antithesis_instrumentation__.Notify(158711)
			}
			__antithesis_instrumentation__.Notify(158614)
			buf29[0].union.RequestLease = &buf29[0].resp
			br.Responses[i].Value = &buf29[0].union
			buf29 = buf29[1:]
		case *RequestUnion_LeaseInfo:
			__antithesis_instrumentation__.Notify(158615)
			if buf30 == nil {
				__antithesis_instrumentation__.Notify(158712)
				buf30 = make([]leaseInfoResponseAlloc, counts[30])
			} else {
				__antithesis_instrumentation__.Notify(158713)
			}
			__antithesis_instrumentation__.Notify(158616)
			buf30[0].union.LeaseInfo = &buf30[0].resp
			br.Responses[i].Value = &buf30[0].union
			buf30 = buf30[1:]
		case *RequestUnion_Export:
			__antithesis_instrumentation__.Notify(158617)
			if buf31 == nil {
				__antithesis_instrumentation__.Notify(158714)
				buf31 = make([]exportResponseAlloc, counts[31])
			} else {
				__antithesis_instrumentation__.Notify(158715)
			}
			__antithesis_instrumentation__.Notify(158618)
			buf31[0].union.Export = &buf31[0].resp
			br.Responses[i].Value = &buf31[0].union
			buf31 = buf31[1:]
		case *RequestUnion_QueryTxn:
			__antithesis_instrumentation__.Notify(158619)
			if buf32 == nil {
				__antithesis_instrumentation__.Notify(158716)
				buf32 = make([]queryTxnResponseAlloc, counts[32])
			} else {
				__antithesis_instrumentation__.Notify(158717)
			}
			__antithesis_instrumentation__.Notify(158620)
			buf32[0].union.QueryTxn = &buf32[0].resp
			br.Responses[i].Value = &buf32[0].union
			buf32 = buf32[1:]
		case *RequestUnion_QueryIntent:
			__antithesis_instrumentation__.Notify(158621)
			if buf33 == nil {
				__antithesis_instrumentation__.Notify(158718)
				buf33 = make([]queryIntentResponseAlloc, counts[33])
			} else {
				__antithesis_instrumentation__.Notify(158719)
			}
			__antithesis_instrumentation__.Notify(158622)
			buf33[0].union.QueryIntent = &buf33[0].resp
			br.Responses[i].Value = &buf33[0].union
			buf33 = buf33[1:]
		case *RequestUnion_QueryLocks:
			__antithesis_instrumentation__.Notify(158623)
			if buf34 == nil {
				__antithesis_instrumentation__.Notify(158720)
				buf34 = make([]queryLocksResponseAlloc, counts[34])
			} else {
				__antithesis_instrumentation__.Notify(158721)
			}
			__antithesis_instrumentation__.Notify(158624)
			buf34[0].union.QueryLocks = &buf34[0].resp
			br.Responses[i].Value = &buf34[0].union
			buf34 = buf34[1:]
		case *RequestUnion_AdminScatter:
			__antithesis_instrumentation__.Notify(158625)
			if buf35 == nil {
				__antithesis_instrumentation__.Notify(158722)
				buf35 = make([]adminScatterResponseAlloc, counts[35])
			} else {
				__antithesis_instrumentation__.Notify(158723)
			}
			__antithesis_instrumentation__.Notify(158626)
			buf35[0].union.AdminScatter = &buf35[0].resp
			br.Responses[i].Value = &buf35[0].union
			buf35 = buf35[1:]
		case *RequestUnion_AddSstable:
			__antithesis_instrumentation__.Notify(158627)
			if buf36 == nil {
				__antithesis_instrumentation__.Notify(158724)
				buf36 = make([]addSSTableResponseAlloc, counts[36])
			} else {
				__antithesis_instrumentation__.Notify(158725)
			}
			__antithesis_instrumentation__.Notify(158628)
			buf36[0].union.AddSstable = &buf36[0].resp
			br.Responses[i].Value = &buf36[0].union
			buf36 = buf36[1:]
		case *RequestUnion_RecomputeStats:
			__antithesis_instrumentation__.Notify(158629)
			if buf37 == nil {
				__antithesis_instrumentation__.Notify(158726)
				buf37 = make([]recomputeStatsResponseAlloc, counts[37])
			} else {
				__antithesis_instrumentation__.Notify(158727)
			}
			__antithesis_instrumentation__.Notify(158630)
			buf37[0].union.RecomputeStats = &buf37[0].resp
			br.Responses[i].Value = &buf37[0].union
			buf37 = buf37[1:]
		case *RequestUnion_Refresh:
			__antithesis_instrumentation__.Notify(158631)
			if buf38 == nil {
				__antithesis_instrumentation__.Notify(158728)
				buf38 = make([]refreshResponseAlloc, counts[38])
			} else {
				__antithesis_instrumentation__.Notify(158729)
			}
			__antithesis_instrumentation__.Notify(158632)
			buf38[0].union.Refresh = &buf38[0].resp
			br.Responses[i].Value = &buf38[0].union
			buf38 = buf38[1:]
		case *RequestUnion_RefreshRange:
			__antithesis_instrumentation__.Notify(158633)
			if buf39 == nil {
				__antithesis_instrumentation__.Notify(158730)
				buf39 = make([]refreshRangeResponseAlloc, counts[39])
			} else {
				__antithesis_instrumentation__.Notify(158731)
			}
			__antithesis_instrumentation__.Notify(158634)
			buf39[0].union.RefreshRange = &buf39[0].resp
			br.Responses[i].Value = &buf39[0].union
			buf39 = buf39[1:]
		case *RequestUnion_Subsume:
			__antithesis_instrumentation__.Notify(158635)
			if buf40 == nil {
				__antithesis_instrumentation__.Notify(158732)
				buf40 = make([]subsumeResponseAlloc, counts[40])
			} else {
				__antithesis_instrumentation__.Notify(158733)
			}
			__antithesis_instrumentation__.Notify(158636)
			buf40[0].union.Subsume = &buf40[0].resp
			br.Responses[i].Value = &buf40[0].union
			buf40 = buf40[1:]
		case *RequestUnion_RangeStats:
			__antithesis_instrumentation__.Notify(158637)
			if buf41 == nil {
				__antithesis_instrumentation__.Notify(158734)
				buf41 = make([]rangeStatsResponseAlloc, counts[41])
			} else {
				__antithesis_instrumentation__.Notify(158735)
			}
			__antithesis_instrumentation__.Notify(158638)
			buf41[0].union.RangeStats = &buf41[0].resp
			br.Responses[i].Value = &buf41[0].union
			buf41 = buf41[1:]
		case *RequestUnion_AdminVerifyProtectedTimestamp:
			__antithesis_instrumentation__.Notify(158639)
			if buf42 == nil {
				__antithesis_instrumentation__.Notify(158736)
				buf42 = make([]adminVerifyProtectedTimestampResponseAlloc, counts[42])
			} else {
				__antithesis_instrumentation__.Notify(158737)
			}
			__antithesis_instrumentation__.Notify(158640)
			buf42[0].union.AdminVerifyProtectedTimestamp = &buf42[0].resp
			br.Responses[i].Value = &buf42[0].union
			buf42 = buf42[1:]
		case *RequestUnion_Migrate:
			__antithesis_instrumentation__.Notify(158641)
			if buf43 == nil {
				__antithesis_instrumentation__.Notify(158738)
				buf43 = make([]migrateResponseAlloc, counts[43])
			} else {
				__antithesis_instrumentation__.Notify(158739)
			}
			__antithesis_instrumentation__.Notify(158642)
			buf43[0].union.Migrate = &buf43[0].resp
			br.Responses[i].Value = &buf43[0].union
			buf43 = buf43[1:]
		case *RequestUnion_QueryResolvedTimestamp:
			__antithesis_instrumentation__.Notify(158643)
			if buf44 == nil {
				__antithesis_instrumentation__.Notify(158740)
				buf44 = make([]queryResolvedTimestampResponseAlloc, counts[44])
			} else {
				__antithesis_instrumentation__.Notify(158741)
			}
			__antithesis_instrumentation__.Notify(158644)
			buf44[0].union.QueryResolvedTimestamp = &buf44[0].resp
			br.Responses[i].Value = &buf44[0].union
			buf44 = buf44[1:]
		case *RequestUnion_ScanInterleavedIntents:
			__antithesis_instrumentation__.Notify(158645)
			if buf45 == nil {
				__antithesis_instrumentation__.Notify(158742)
				buf45 = make([]scanInterleavedIntentsResponseAlloc, counts[45])
			} else {
				__antithesis_instrumentation__.Notify(158743)
			}
			__antithesis_instrumentation__.Notify(158646)
			buf45[0].union.ScanInterleavedIntents = &buf45[0].resp
			br.Responses[i].Value = &buf45[0].union
			buf45 = buf45[1:]
		case *RequestUnion_Barrier:
			__antithesis_instrumentation__.Notify(158647)
			if buf46 == nil {
				__antithesis_instrumentation__.Notify(158744)
				buf46 = make([]barrierResponseAlloc, counts[46])
			} else {
				__antithesis_instrumentation__.Notify(158745)
			}
			__antithesis_instrumentation__.Notify(158648)
			buf46[0].union.Barrier = &buf46[0].resp
			br.Responses[i].Value = &buf46[0].union
			buf46 = buf46[1:]
		case *RequestUnion_Probe:
			__antithesis_instrumentation__.Notify(158649)
			if buf47 == nil {
				__antithesis_instrumentation__.Notify(158746)
				buf47 = make([]probeResponseAlloc, counts[47])
			} else {
				__antithesis_instrumentation__.Notify(158747)
			}
			__antithesis_instrumentation__.Notify(158650)
			buf47[0].union.Probe = &buf47[0].resp
			br.Responses[i].Value = &buf47[0].union
			buf47 = buf47[1:]
		default:
			__antithesis_instrumentation__.Notify(158651)
			panic(fmt.Sprintf("unsupported request: %+v", r))
		}
	}
	__antithesis_instrumentation__.Notify(158553)
	return br
}

func CreateRequest(method Method) Request {
	__antithesis_instrumentation__.Notify(158748)
	switch method {
	case Get:
		__antithesis_instrumentation__.Notify(158749)
		return &GetRequest{}
	case Put:
		__antithesis_instrumentation__.Notify(158750)
		return &PutRequest{}
	case ConditionalPut:
		__antithesis_instrumentation__.Notify(158751)
		return &ConditionalPutRequest{}
	case Increment:
		__antithesis_instrumentation__.Notify(158752)
		return &IncrementRequest{}
	case Delete:
		__antithesis_instrumentation__.Notify(158753)
		return &DeleteRequest{}
	case DeleteRange:
		__antithesis_instrumentation__.Notify(158754)
		return &DeleteRangeRequest{}
	case ClearRange:
		__antithesis_instrumentation__.Notify(158755)
		return &ClearRangeRequest{}
	case RevertRange:
		__antithesis_instrumentation__.Notify(158756)
		return &RevertRangeRequest{}
	case Scan:
		__antithesis_instrumentation__.Notify(158757)
		return &ScanRequest{}
	case EndTxn:
		__antithesis_instrumentation__.Notify(158758)
		return &EndTxnRequest{}
	case AdminSplit:
		__antithesis_instrumentation__.Notify(158759)
		return &AdminSplitRequest{}
	case AdminUnsplit:
		__antithesis_instrumentation__.Notify(158760)
		return &AdminUnsplitRequest{}
	case AdminMerge:
		__antithesis_instrumentation__.Notify(158761)
		return &AdminMergeRequest{}
	case AdminTransferLease:
		__antithesis_instrumentation__.Notify(158762)
		return &AdminTransferLeaseRequest{}
	case AdminChangeReplicas:
		__antithesis_instrumentation__.Notify(158763)
		return &AdminChangeReplicasRequest{}
	case AdminRelocateRange:
		__antithesis_instrumentation__.Notify(158764)
		return &AdminRelocateRangeRequest{}
	case HeartbeatTxn:
		__antithesis_instrumentation__.Notify(158765)
		return &HeartbeatTxnRequest{}
	case GC:
		__antithesis_instrumentation__.Notify(158766)
		return &GCRequest{}
	case PushTxn:
		__antithesis_instrumentation__.Notify(158767)
		return &PushTxnRequest{}
	case RecoverTxn:
		__antithesis_instrumentation__.Notify(158768)
		return &RecoverTxnRequest{}
	case ResolveIntent:
		__antithesis_instrumentation__.Notify(158769)
		return &ResolveIntentRequest{}
	case ResolveIntentRange:
		__antithesis_instrumentation__.Notify(158770)
		return &ResolveIntentRangeRequest{}
	case Merge:
		__antithesis_instrumentation__.Notify(158771)
		return &MergeRequest{}
	case TruncateLog:
		__antithesis_instrumentation__.Notify(158772)
		return &TruncateLogRequest{}
	case RequestLease:
		__antithesis_instrumentation__.Notify(158773)
		return &RequestLeaseRequest{}
	case ReverseScan:
		__antithesis_instrumentation__.Notify(158774)
		return &ReverseScanRequest{}
	case ComputeChecksum:
		__antithesis_instrumentation__.Notify(158775)
		return &ComputeChecksumRequest{}
	case CheckConsistency:
		__antithesis_instrumentation__.Notify(158776)
		return &CheckConsistencyRequest{}
	case InitPut:
		__antithesis_instrumentation__.Notify(158777)
		return &InitPutRequest{}
	case TransferLease:
		__antithesis_instrumentation__.Notify(158778)
		return &TransferLeaseRequest{}
	case LeaseInfo:
		__antithesis_instrumentation__.Notify(158779)
		return &LeaseInfoRequest{}
	case Export:
		__antithesis_instrumentation__.Notify(158780)
		return &ExportRequest{}
	case QueryTxn:
		__antithesis_instrumentation__.Notify(158781)
		return &QueryTxnRequest{}
	case QueryIntent:
		__antithesis_instrumentation__.Notify(158782)
		return &QueryIntentRequest{}
	case QueryLocks:
		__antithesis_instrumentation__.Notify(158783)
		return &QueryLocksRequest{}
	case AdminScatter:
		__antithesis_instrumentation__.Notify(158784)
		return &AdminScatterRequest{}
	case AddSSTable:
		__antithesis_instrumentation__.Notify(158785)
		return &AddSSTableRequest{}
	case RecomputeStats:
		__antithesis_instrumentation__.Notify(158786)
		return &RecomputeStatsRequest{}
	case Refresh:
		__antithesis_instrumentation__.Notify(158787)
		return &RefreshRequest{}
	case RefreshRange:
		__antithesis_instrumentation__.Notify(158788)
		return &RefreshRangeRequest{}
	case Subsume:
		__antithesis_instrumentation__.Notify(158789)
		return &SubsumeRequest{}
	case RangeStats:
		__antithesis_instrumentation__.Notify(158790)
		return &RangeStatsRequest{}
	case AdminVerifyProtectedTimestamp:
		__antithesis_instrumentation__.Notify(158791)
		return &AdminVerifyProtectedTimestampRequest{}
	case Migrate:
		__antithesis_instrumentation__.Notify(158792)
		return &MigrateRequest{}
	case QueryResolvedTimestamp:
		__antithesis_instrumentation__.Notify(158793)
		return &QueryResolvedTimestampRequest{}
	case ScanInterleavedIntents:
		__antithesis_instrumentation__.Notify(158794)
		return &ScanInterleavedIntentsRequest{}
	case Barrier:
		__antithesis_instrumentation__.Notify(158795)
		return &BarrierRequest{}
	case Probe:
		__antithesis_instrumentation__.Notify(158796)
		return &ProbeRequest{}
	default:
		__antithesis_instrumentation__.Notify(158797)
		panic(fmt.Sprintf("unsupported method: %+v", method))
	}
}
