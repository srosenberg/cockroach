package loqrecovery

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/loqrecovery/loqrecoverypb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

func writeReplicaRecoveryStoreRecord(
	uuid uuid.UUID,
	timestamp int64,
	update loqrecoverypb.ReplicaUpdate,
	report PrepareReplicaReport,
	readWriter storage.ReadWriter,
) error {
	__antithesis_instrumentation__.Notify(110036)
	record := loqrecoverypb.ReplicaRecoveryRecord{
		Timestamp:       timestamp,
		RangeID:         report.RangeID(),
		StartKey:        update.StartKey,
		EndKey:          update.StartKey,
		OldReplicaID:    report.OldReplica.ReplicaID,
		NewReplica:      update.NewReplica,
		RangeDescriptor: report.Descriptor,
	}

	data, err := protoutil.Marshal(&record)
	if err != nil {
		__antithesis_instrumentation__.Notify(110039)
		return errors.Wrap(err, "failed to marshal update record entry")
	} else {
		__antithesis_instrumentation__.Notify(110040)
	}
	__antithesis_instrumentation__.Notify(110037)
	if err := readWriter.PutUnversioned(
		keys.StoreUnsafeReplicaRecoveryKey(uuid), data); err != nil {
		__antithesis_instrumentation__.Notify(110041)
		return err
	} else {
		__antithesis_instrumentation__.Notify(110042)
	}
	__antithesis_instrumentation__.Notify(110038)
	return nil
}

func RegisterOfflineRecoveryEvents(
	ctx context.Context,
	readWriter storage.ReadWriter,
	registerEvent func(context.Context, loqrecoverypb.ReplicaRecoveryRecord) (bool, error),
) (int, error) {
	__antithesis_instrumentation__.Notify(110043)
	successCount := 0
	var processingErrors error

	iter := readWriter.NewMVCCIterator(
		storage.MVCCKeyIterKind, storage.IterOptions{
			LowerBound: keys.LocalStoreUnsafeReplicaRecoveryKeyMin,
			UpperBound: keys.LocalStoreUnsafeReplicaRecoveryKeyMax,
		})
	defer iter.Close()

	iter.SeekGE(storage.MVCCKey{Key: keys.LocalStoreUnsafeReplicaRecoveryKeyMin})
	for ; ; iter.Next() {
		__antithesis_instrumentation__.Notify(110046)
		valid, err := iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(110052)
			processingErrors = errors.CombineErrors(processingErrors,
				errors.Wrapf(err, "failed to iterate replica recovery record keys"))
			break
		} else {
			__antithesis_instrumentation__.Notify(110053)
		}
		__antithesis_instrumentation__.Notify(110047)
		if !valid {
			__antithesis_instrumentation__.Notify(110054)
			break
		} else {
			__antithesis_instrumentation__.Notify(110055)
		}
		__antithesis_instrumentation__.Notify(110048)

		record := loqrecoverypb.ReplicaRecoveryRecord{}
		if err := iter.ValueProto(&record); err != nil {
			__antithesis_instrumentation__.Notify(110056)
			processingErrors = errors.CombineErrors(processingErrors, errors.Wrapf(err,
				"failed to deserialize replica recovery event at key %s", iter.Key()))
			continue
		} else {
			__antithesis_instrumentation__.Notify(110057)
		}
		__antithesis_instrumentation__.Notify(110049)
		removeEvent, err := registerEvent(ctx, record)
		if err != nil {
			__antithesis_instrumentation__.Notify(110058)
			processingErrors = errors.CombineErrors(processingErrors,
				errors.Wrapf(err, "replica recovery record processing failed"))
			continue
		} else {
			__antithesis_instrumentation__.Notify(110059)
		}
		__antithesis_instrumentation__.Notify(110050)
		if removeEvent {
			__antithesis_instrumentation__.Notify(110060)
			if err := readWriter.ClearUnversioned(iter.UnsafeKey().Key); err != nil {
				__antithesis_instrumentation__.Notify(110061)
				processingErrors = errors.CombineErrors(processingErrors, errors.Wrapf(
					err, "failed to delete replica recovery record at key %s", iter.Key()))
				continue
			} else {
				__antithesis_instrumentation__.Notify(110062)
			}
		} else {
			__antithesis_instrumentation__.Notify(110063)
		}
		__antithesis_instrumentation__.Notify(110051)
		successCount++
	}
	__antithesis_instrumentation__.Notify(110044)
	if processingErrors != nil {
		__antithesis_instrumentation__.Notify(110064)
		return 0, errors.Wrapf(processingErrors,
			"failed to fully process replica recovery records, successfully processed %d", successCount)
	} else {
		__antithesis_instrumentation__.Notify(110065)
	}
	__antithesis_instrumentation__.Notify(110045)
	return successCount, nil
}

func UpdateRangeLogWithRecovery(
	ctx context.Context,
	sqlExec func(ctx context.Context, stmt string, args ...interface{}) (int, error),
	event loqrecoverypb.ReplicaRecoveryRecord,
) error {
	__antithesis_instrumentation__.Notify(110066)
	const insertEventTableStmt = `
	INSERT INTO system.rangelog (
		timestamp, "rangeID", "storeID", "eventType", "otherRangeID", info
	)
	VALUES(
		$1, $2, $3, $4, $5, $6
	)
	`
	updateInfo := kvserverpb.RangeLogEvent_Info{
		UpdatedDesc:  &event.RangeDescriptor,
		AddedReplica: &event.NewReplica,
		Reason:       kvserverpb.ReasonUnsafeRecovery,
		Details:      "Performed unsafe range loss of quorum recovery",
	}
	infoBytes, err := json.Marshal(updateInfo)
	if err != nil {
		__antithesis_instrumentation__.Notify(110070)
		return errors.Wrap(err, "failed to serialize a RangeLog info entry")
	} else {
		__antithesis_instrumentation__.Notify(110071)
	}
	__antithesis_instrumentation__.Notify(110067)
	args := []interface{}{
		timeutil.Unix(0, event.Timestamp),
		event.RangeID,
		event.NewReplica.StoreID,
		kvserverpb.RangeLogEventType_unsafe_quorum_recovery.String(),
		nil,
		string(infoBytes),
	}

	rows, err := sqlExec(ctx, insertEventTableStmt, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(110072)
		return errors.Wrap(err, "failed to insert a RangeLog entry")
	} else {
		__antithesis_instrumentation__.Notify(110073)
	}
	__antithesis_instrumentation__.Notify(110068)
	if rows != 1 {
		__antithesis_instrumentation__.Notify(110074)
		return errors.Errorf("%d row(s) affected by RangeLog insert while expected 1",
			rows)
	} else {
		__antithesis_instrumentation__.Notify(110075)
	}
	__antithesis_instrumentation__.Notify(110069)
	return nil
}
