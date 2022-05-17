package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func (s *Store) insertRangeLogEvent(
	ctx context.Context, txn *kv.Txn, event kvserverpb.RangeLogEvent,
) error {
	__antithesis_instrumentation__.Notify(108683)

	var info string
	if event.Info != nil {
		__antithesis_instrumentation__.Notify(108691)
		info = event.Info.String()
	} else {
		__antithesis_instrumentation__.Notify(108692)
	}
	__antithesis_instrumentation__.Notify(108684)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(108693)
		log.Infof(ctx, "Range Event: %q, range: %d, info: %s",
			event.EventType, event.RangeID, info)
	} else {
		__antithesis_instrumentation__.Notify(108694)
	}
	__antithesis_instrumentation__.Notify(108685)

	const insertEventTableStmt = `
	INSERT INTO system.rangelog (
		timestamp, "rangeID", "storeID", "eventType", "otherRangeID", info
	)
	VALUES(
		$1, $2, $3, $4, $5, $6
	)
	`
	args := []interface{}{
		event.Timestamp,
		event.RangeID,
		event.StoreID,
		event.EventType.String(),
		nil,
		nil,
	}
	if event.OtherRangeID != 0 {
		__antithesis_instrumentation__.Notify(108695)
		args[4] = event.OtherRangeID
	} else {
		__antithesis_instrumentation__.Notify(108696)
	}
	__antithesis_instrumentation__.Notify(108686)
	if event.Info != nil {
		__antithesis_instrumentation__.Notify(108697)
		infoBytes, err := json.Marshal(*event.Info)
		if err != nil {
			__antithesis_instrumentation__.Notify(108699)
			return err
		} else {
			__antithesis_instrumentation__.Notify(108700)
		}
		__antithesis_instrumentation__.Notify(108698)
		args[5] = string(infoBytes)
	} else {
		__antithesis_instrumentation__.Notify(108701)
	}
	__antithesis_instrumentation__.Notify(108687)

	switch event.EventType {
	case kvserverpb.RangeLogEventType_split:
		__antithesis_instrumentation__.Notify(108702)
		s.metrics.RangeSplits.Inc(1)
	case kvserverpb.RangeLogEventType_merge:
		__antithesis_instrumentation__.Notify(108703)
		s.metrics.RangeMerges.Inc(1)
	case kvserverpb.RangeLogEventType_add_voter:
		__antithesis_instrumentation__.Notify(108704)
		s.metrics.RangeAdds.Inc(1)
	case kvserverpb.RangeLogEventType_remove_voter:
		__antithesis_instrumentation__.Notify(108705)
		s.metrics.RangeRemoves.Inc(1)
	default:
		__antithesis_instrumentation__.Notify(108706)
	}
	__antithesis_instrumentation__.Notify(108688)

	rows, err := s.cfg.SQLExecutor.ExecEx(ctx, "log-range-event", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		insertEventTableStmt, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(108707)
		return err
	} else {
		__antithesis_instrumentation__.Notify(108708)
	}
	__antithesis_instrumentation__.Notify(108689)
	if rows != 1 {
		__antithesis_instrumentation__.Notify(108709)
		return errors.Errorf("%d rows affected by log insertion; expected exactly one row affected.", rows)
	} else {
		__antithesis_instrumentation__.Notify(108710)
	}
	__antithesis_instrumentation__.Notify(108690)
	return nil
}

func (s *Store) logSplit(
	ctx context.Context, txn *kv.Txn, updatedDesc, newDesc roachpb.RangeDescriptor,
) error {
	__antithesis_instrumentation__.Notify(108711)
	if !s.cfg.LogRangeEvents {
		__antithesis_instrumentation__.Notify(108713)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(108714)
	}
	__antithesis_instrumentation__.Notify(108712)
	return s.insertRangeLogEvent(ctx, txn, kvserverpb.RangeLogEvent{
		Timestamp:    selectEventTimestamp(s, txn.ReadTimestamp()),
		RangeID:      updatedDesc.RangeID,
		EventType:    kvserverpb.RangeLogEventType_split,
		StoreID:      s.StoreID(),
		OtherRangeID: newDesc.RangeID,
		Info: &kvserverpb.RangeLogEvent_Info{
			UpdatedDesc: &updatedDesc,
			NewDesc:     &newDesc,
		},
	})
}

func (s *Store) logMerge(
	ctx context.Context, txn *kv.Txn, updatedLHSDesc, rhsDesc roachpb.RangeDescriptor,
) error {
	__antithesis_instrumentation__.Notify(108715)
	if !s.cfg.LogRangeEvents {
		__antithesis_instrumentation__.Notify(108717)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(108718)
	}
	__antithesis_instrumentation__.Notify(108716)
	return s.insertRangeLogEvent(ctx, txn, kvserverpb.RangeLogEvent{
		Timestamp:    selectEventTimestamp(s, txn.ReadTimestamp()),
		RangeID:      updatedLHSDesc.RangeID,
		EventType:    kvserverpb.RangeLogEventType_merge,
		StoreID:      s.StoreID(),
		OtherRangeID: rhsDesc.RangeID,
		Info: &kvserverpb.RangeLogEvent_Info{
			UpdatedDesc: &updatedLHSDesc,
			RemovedDesc: &rhsDesc,
		},
	})
}

func (s *Store) logChange(
	ctx context.Context,
	txn *kv.Txn,
	changeType roachpb.ReplicaChangeType,
	replica roachpb.ReplicaDescriptor,
	desc roachpb.RangeDescriptor,
	reason kvserverpb.RangeLogEventReason,
	details string,
) error {
	__antithesis_instrumentation__.Notify(108719)
	if !s.cfg.LogRangeEvents {
		__antithesis_instrumentation__.Notify(108722)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(108723)
	}
	__antithesis_instrumentation__.Notify(108720)

	var logType kvserverpb.RangeLogEventType
	var info kvserverpb.RangeLogEvent_Info
	switch changeType {
	case roachpb.ADD_VOTER:
		__antithesis_instrumentation__.Notify(108724)
		logType = kvserverpb.RangeLogEventType_add_voter
		info = kvserverpb.RangeLogEvent_Info{
			AddedReplica: &replica,
			UpdatedDesc:  &desc,
			Reason:       reason,
			Details:      details,
		}
	case roachpb.REMOVE_VOTER:
		__antithesis_instrumentation__.Notify(108725)
		logType = kvserverpb.RangeLogEventType_remove_voter
		info = kvserverpb.RangeLogEvent_Info{
			RemovedReplica: &replica,
			UpdatedDesc:    &desc,
			Reason:         reason,
			Details:        details,
		}
	case roachpb.ADD_NON_VOTER:
		__antithesis_instrumentation__.Notify(108726)
		logType = kvserverpb.RangeLogEventType_add_non_voter
		info = kvserverpb.RangeLogEvent_Info{
			AddedReplica: &replica,
			UpdatedDesc:  &desc,
			Reason:       reason,
			Details:      details,
		}
	case roachpb.REMOVE_NON_VOTER:
		__antithesis_instrumentation__.Notify(108727)
		logType = kvserverpb.RangeLogEventType_remove_non_voter
		info = kvserverpb.RangeLogEvent_Info{
			RemovedReplica: &replica,
			UpdatedDesc:    &desc,
			Reason:         reason,
			Details:        details,
		}
	default:
		__antithesis_instrumentation__.Notify(108728)
		return errors.Errorf("unknown replica change type %s", changeType)
	}
	__antithesis_instrumentation__.Notify(108721)

	return s.insertRangeLogEvent(ctx, txn, kvserverpb.RangeLogEvent{
		Timestamp: selectEventTimestamp(s, txn.ReadTimestamp()),
		RangeID:   desc.RangeID,
		EventType: logType,
		StoreID:   s.StoreID(),
		Info:      &info,
	})
}

func selectEventTimestamp(s *Store, input hlc.Timestamp) time.Time {
	__antithesis_instrumentation__.Notify(108729)
	if input.IsEmpty() {
		__antithesis_instrumentation__.Notify(108731)
		return s.Clock().PhysicalTime()
	} else {
		__antithesis_instrumentation__.Notify(108732)
	}
	__antithesis_instrumentation__.Notify(108730)
	return input.GoTime()
}
