package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type MVCCLogicalOpType int

const (
	MVCCWriteValueOpType MVCCLogicalOpType = iota

	MVCCWriteIntentOpType

	MVCCUpdateIntentOpType

	MVCCCommitIntentOpType

	MVCCAbortIntentOpType
)

type MVCCLogicalOpDetails struct {
	Txn       enginepb.TxnMeta
	Key       roachpb.Key
	Timestamp hlc.Timestamp

	Safe bool
}

type OpLoggerBatch struct {
	Batch

	ops      []enginepb.MVCCLogicalOp
	opsAlloc bufalloc.ByteAllocator
}

func NewOpLoggerBatch(b Batch) *OpLoggerBatch {
	__antithesis_instrumentation__.Notify(641609)
	ol := &OpLoggerBatch{Batch: b}
	return ol
}

var _ Batch = &OpLoggerBatch{}

func (ol *OpLoggerBatch) LogLogicalOp(op MVCCLogicalOpType, details MVCCLogicalOpDetails) {
	__antithesis_instrumentation__.Notify(641610)
	ol.logLogicalOp(op, details)
	ol.Batch.LogLogicalOp(op, details)
}

func (ol *OpLoggerBatch) logLogicalOp(op MVCCLogicalOpType, details MVCCLogicalOpDetails) {
	__antithesis_instrumentation__.Notify(641611)
	if keys.IsLocal(details.Key) {
		__antithesis_instrumentation__.Notify(641613)

		if bytes.HasPrefix(details.Key, keys.LocalRangeLockTablePrefix) {
			__antithesis_instrumentation__.Notify(641615)
			panic(fmt.Sprintf("seeing locktable key %s", details.Key.String()))
		} else {
			__antithesis_instrumentation__.Notify(641616)
		}
		__antithesis_instrumentation__.Notify(641614)
		return
	} else {
		__antithesis_instrumentation__.Notify(641617)
	}
	__antithesis_instrumentation__.Notify(641612)

	switch op {
	case MVCCWriteValueOpType:
		__antithesis_instrumentation__.Notify(641618)
		if !details.Safe {
			__antithesis_instrumentation__.Notify(641627)
			ol.opsAlloc, details.Key = ol.opsAlloc.Copy(details.Key, 0)
		} else {
			__antithesis_instrumentation__.Notify(641628)
		}
		__antithesis_instrumentation__.Notify(641619)

		ol.recordOp(&enginepb.MVCCWriteValueOp{
			Key:       details.Key,
			Timestamp: details.Timestamp,
		})
	case MVCCWriteIntentOpType:
		__antithesis_instrumentation__.Notify(641620)
		if !details.Safe {
			__antithesis_instrumentation__.Notify(641629)
			ol.opsAlloc, details.Txn.Key = ol.opsAlloc.Copy(details.Txn.Key, 0)
		} else {
			__antithesis_instrumentation__.Notify(641630)
		}
		__antithesis_instrumentation__.Notify(641621)

		ol.recordOp(&enginepb.MVCCWriteIntentOp{
			TxnID:           details.Txn.ID,
			TxnKey:          details.Txn.Key,
			TxnMinTimestamp: details.Txn.MinTimestamp,
			Timestamp:       details.Timestamp,
		})
	case MVCCUpdateIntentOpType:
		__antithesis_instrumentation__.Notify(641622)
		ol.recordOp(&enginepb.MVCCUpdateIntentOp{
			TxnID:     details.Txn.ID,
			Timestamp: details.Timestamp,
		})
	case MVCCCommitIntentOpType:
		__antithesis_instrumentation__.Notify(641623)
		if !details.Safe {
			__antithesis_instrumentation__.Notify(641631)
			ol.opsAlloc, details.Key = ol.opsAlloc.Copy(details.Key, 0)
		} else {
			__antithesis_instrumentation__.Notify(641632)
		}
		__antithesis_instrumentation__.Notify(641624)

		ol.recordOp(&enginepb.MVCCCommitIntentOp{
			TxnID:     details.Txn.ID,
			Key:       details.Key,
			Timestamp: details.Timestamp,
		})
	case MVCCAbortIntentOpType:
		__antithesis_instrumentation__.Notify(641625)
		ol.recordOp(&enginepb.MVCCAbortIntentOp{
			TxnID: details.Txn.ID,
		})
	default:
		__antithesis_instrumentation__.Notify(641626)
		panic(fmt.Sprintf("unexpected op type %v", op))
	}
}

func (ol *OpLoggerBatch) recordOp(op interface{}) {
	__antithesis_instrumentation__.Notify(641633)
	ol.ops = append(ol.ops, enginepb.MVCCLogicalOp{})
	ol.ops[len(ol.ops)-1].MustSetValue(op)
}

func (ol *OpLoggerBatch) LogicalOps() []enginepb.MVCCLogicalOp {
	__antithesis_instrumentation__.Notify(641634)
	if ol == nil {
		__antithesis_instrumentation__.Notify(641636)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(641637)
	}
	__antithesis_instrumentation__.Notify(641635)
	return ol.ops
}
